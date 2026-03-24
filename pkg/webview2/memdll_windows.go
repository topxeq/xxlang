//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file implements memory DLL loading for Windows.
package webview2

import (
	"encoding/binary"
	"fmt"
	"sync"
	"syscall"
	"unsafe"
)

var (
	kernel32DLL      = syscall.NewLazyDLL("kernel32.dll")
	virtualAlloc     = kernel32DLL.NewProc("VirtualAlloc")
	virtualProtect   = kernel32DLL.NewProc("VirtualProtect")
	virtualFree      = kernel32DLL.NewProc("VirtualFree")
	getProcAddress   = kernel32DLL.NewProc("GetProcAddress")
	loadLibraryW     = kernel32DLL.NewProc("LoadLibraryW")
	getModuleHandleW = kernel32DLL.NewProc("GetModuleHandleW")
	flushInstructionCache = kernel32DLL.NewProc("FlushInstructionCache")
	getCurrentProcess = kernel32DLL.NewProc("GetCurrentProcess")
)

// Memory allocation constants
const (
	MEM_COMMIT  = 0x00001000
	MEM_RESERVE = 0x00002000
	MEM_RELEASE = 0x8000

	// Memory protection constants
	PAGE_NOACCESS          = 0x01
	PAGE_READONLY          = 0x02
	PAGE_READWRITE         = 0x04
	PAGE_WRITECOPY         = 0x08
	PAGE_EXECUTE           = 0x10
	PAGE_EXECUTE_READ      = 0x20
	PAGE_EXECUTE_READWRITE = 0x40
	PAGE_EXECUTE_WRITECOPY = 0x80
)

// MemoryModule represents a module loaded in memory.
type MemoryModule struct {
	baseAddr uintptr
	size     uint32
	exports  map[string]uintptr
}

var (
	moduleCache     = make(map[string]*MemoryModule)
	moduleCacheLock sync.RWMutex
)

// memory represents a block of allocated memory with ReadWriteSeeker interface.
type memory struct {
	base uintptr
	pos  int64
	size uint32
}

// newMemory creates a memory wrapper around allocated memory.
func newMemory(base uintptr, size uint32) *memory {
	return &memory{base: base, size: size}
}

// Read implements io.Reader.
func (m *memory) Read(b []byte) (n int, err error) {
	if m.pos >= int64(m.size) {
		return 0, nil
	}
	available := int64(m.size) - m.pos
	if available < int64(len(b)) {
		b = b[:available]
	}
	for i := range b {
		b[i] = *(*byte)(unsafe.Pointer(m.base + uintptr(m.pos + int64(i))))
	}
	m.pos += int64(len(b))
	return len(b), nil
}

// Write implements io.Writer.
func (m *memory) Write(b []byte) (n int, err error) {
	if m.pos >= int64(m.size) {
		return 0, nil
	}
	available := int64(m.size) - m.pos
	if available < int64(len(b)) {
		b = b[:available]
	}
	for i, v := range b {
		*(*byte)(unsafe.Pointer(m.base + uintptr(m.pos + int64(i)))) = v
	}
	// Flush instruction cache for code sections
	proc, _, _ := getCurrentProcess.Call()
	flushInstructionCache.Call(
		proc,
		m.base+uintptr(m.pos),
		uintptr(len(b)),
	)
	m.pos += int64(len(b))
	return len(b), nil
}

// Seek implements io.Seeker.
func (m *memory) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0: // io.SeekStart
		m.pos = offset
	case 1: // io.SeekCurrent
		m.pos += offset
	case 2: // io.SeekEnd
		m.pos = int64(m.size) + offset
	}
	if m.pos < 0 {
		m.pos = 0
	}
	return m.pos, nil
}

// Addr returns the base address.
func (m *memory) Addr() uintptr {
	return m.base
}

// loadFromMemory loads a DLL from memory data.
func loadFromMemory(data []byte) (*MemoryModule, error) {
	// Parse DOS header
	if len(data) < 64 {
		return nil, fmt.Errorf("data too small for DOS header")
	}

	// Check MZ signature
	if data[0] != 'M' || data[1] != 'Z' {
		return nil, fmt.Errorf("invalid DOS signature")
	}

	dosHeader := (*ImageDOSHeader)(unsafe.Pointer(&data[0]))

	// Parse PE header
	peOffset := dosHeader.NewHeaderAddr
	if int(peOffset)+264 > len(data) {
		return nil, fmt.Errorf("PE header offset out of bounds")
	}

	// Check PE signature
	if data[peOffset] != 'P' || data[peOffset+1] != 'E' || data[peOffset+2] != 0 || data[peOffset+3] != 0 {
		return nil, fmt.Errorf("invalid PE signature")
	}

	// Parse file header
	fileHeader := (*ImageFileHeader)(unsafe.Pointer(&data[peOffset+4]))

	// Determine if 32-bit or 64-bit
	optHeaderOffset := peOffset + 24
	magic := binary.LittleEndian.Uint16(data[optHeaderOffset : optHeaderOffset+2])

	var imageBase uint64
	var sizeOfImage uint32
	var entryPoint uint32
	var is64Bit bool
	var optHeader interface{}

	switch magic {
	case ImageNTOptionalHeader64Magic:
		is64Bit = true
		opt := (*ImageOptionalHeader64)(unsafe.Pointer(&data[optHeaderOffset]))
		optHeader = opt
		imageBase = opt.ImageBase
		sizeOfImage = opt.SizeOfImage
		entryPoint = opt.AddressOfEntryPoint
	case ImageNTOptionalHeader32Magic:
		is64Bit = false
		opt := (*ImageOptionalHeader32)(unsafe.Pointer(&data[optHeaderOffset]))
		optHeader = opt
		imageBase = uint64(opt.ImageBase)
		sizeOfImage = opt.SizeOfImage
		entryPoint = opt.AddressOfEntryPoint
	default:
		return nil, fmt.Errorf("unknown PE magic: 0x%x", magic)
	}

	// Allocate memory for the image
	baseAddr, _, _ := virtualAlloc.Call(
		uintptr(imageBase),
		uintptr(sizeOfImage),
		uintptr(MEM_COMMIT|MEM_RESERVE),
		uintptr(PAGE_EXECUTE_READWRITE),
	)
	if baseAddr == 0 {
		baseAddr, _, _ = virtualAlloc.Call(
			0,
			uintptr(sizeOfImage),
			uintptr(MEM_COMMIT|MEM_RESERVE),
			uintptr(PAGE_EXECUTE_READWRITE),
		)
		if baseAddr == 0 {
			return nil, fmt.Errorf("failed to allocate memory")
		}
	}

	delta := uint64(baseAddr) - imageBase
	mem := newMemory(baseAddr, sizeOfImage)

	// Get section info
	numSections := fileHeader.NumberOfSections
	sectionsOffset := int(optHeaderOffset) + int(fileHeader.SizeOfOptionalHeader)

	// Copy headers
	var sizeOfHeaders uint32
	switch h := optHeader.(type) {
	case *ImageOptionalHeader64:
		sizeOfHeaders = h.SizeOfHeaders
	case *ImageOptionalHeader32:
		sizeOfHeaders = h.SizeOfHeaders
	}
	mem.Write(data[:sizeOfHeaders])

	// Map sections
	for i := uint16(0); i < numSections; i++ {
		section := (*ImageSectionHeader)(unsafe.Pointer(&data[sectionsOffset + int(i)*40]))
		if section.SizeOfRawData == 0 {
			continue
		}
		destAddr := baseAddr + uintptr(section.VirtualAddress)
		srcData := data[section.PointerToRawData : section.PointerToRawData+section.SizeOfRawData]
		for j, b := range srcData {
			*(*byte)(unsafe.Pointer(destAddr + uintptr(j))) = b
		}
	}

	// Process relocations if needed
	if delta != 0 {
		var relocDir ImageDataDirectory
		switch h := optHeader.(type) {
		case *ImageOptionalHeader64:
			relocDir = h.DataDirectory[ImageDirectoryEntryBaseReloc]
		case *ImageOptionalHeader32:
			relocDir = h.DataDirectory[ImageDirectoryEntryBaseReloc]
		}

		if relocDir.Size > 0 {
			if err := processRelocations(mem, relocDir.VirtualAddress, relocDir.Size, delta, is64Bit); err != nil {
				virtualFree.Call(baseAddr, 0, uintptr(MEM_RELEASE))
				return nil, fmt.Errorf("relocation failed: %w", err)
			}
		}
	}

	// Process imports
	var importDir ImageDataDirectory
	switch h := optHeader.(type) {
	case *ImageOptionalHeader64:
		importDir = h.DataDirectory[ImageDirectoryEntryImport]
	case *ImageOptionalHeader32:
		importDir = h.DataDirectory[ImageDirectoryEntryImport]
	}

	if importDir.Size > 0 {
		if err := processImports(mem, importDir.VirtualAddress, importDir.Size, is64Bit); err != nil {
			virtualFree.Call(baseAddr, 0, uintptr(MEM_RELEASE))
			return nil, fmt.Errorf("import processing failed: %w", err)
		}
	}

	// Process exports
	exports := make(map[string]uintptr)
	var exportDir ImageDataDirectory
	switch h := optHeader.(type) {
	case *ImageOptionalHeader64:
		exportDir = h.DataDirectory[ImageDirectoryEntryExport]
	case *ImageOptionalHeader32:
		exportDir = h.DataDirectory[ImageDirectoryEntryExport]
	}

	if exportDir.Size > 0 {
		if err := loadExports(mem, exportDir.VirtualAddress, exports); err != nil {
			virtualFree.Call(baseAddr, 0, uintptr(MEM_RELEASE))
			return nil, fmt.Errorf("export loading failed: %w", err)
		}
	}

	// Call DllMain with DLL_PROCESS_ATTACH
	if entryPoint != 0 {
		dllMain := baseAddr + uintptr(entryPoint)
		syscall.Syscall(dllMain, 3, baseAddr, 1, 0) // DLL_PROCESS_ATTACH = 1
	}

	return &MemoryModule{
		baseAddr: baseAddr,
		size:     sizeOfImage,
		exports:  exports,
	}, nil
}

// processRelocations handles base relocations.
func processRelocations(mem *memory, relocVA, relocSize uint32, delta uint64, is64Bit bool) error {
	offset := uint32(0)
	for offset < relocSize {
		// Read relocation block header
		mem.Seek(int64(relocVA+offset), 0)
		var blockVA, blockSize uint32
		binary.Read(mem, binary.LittleEndian, &blockVA)
		binary.Read(mem, binary.LittleEndian, &blockSize)

		if blockSize == 0 {
			break
		}

		numEntries := (blockSize - 8) / 2
		for i := uint32(0); i < numEntries; i++ {
			var entry uint16
			binary.Read(mem, binary.LittleEndian, &entry)

			relType := entry >> 12
			relOffset := uint32(entry & 0xFFF)

			targetAddr := mem.Addr() + uintptr(blockVA) + uintptr(relOffset)

			switch relType {
			case ImageRelBasedHighLow:
				val := *(*uint32)(unsafe.Pointer(targetAddr))
				val += uint32(delta)
				*(*uint32)(unsafe.Pointer(targetAddr)) = val
			case ImageRelBasedDir64:
				val := *(*uint64)(unsafe.Pointer(targetAddr))
				val += delta
				*(*uint64)(unsafe.Pointer(targetAddr)) = val
			}
		}

		offset += blockSize
	}
	return nil
}

// processImports resolves import addresses.
func processImports(mem *memory, importVA, importSize uint32, is64Bit bool) error {
	ptrSize := 4
	if is64Bit {
		ptrSize = 8
	}

	offset := uint32(0)
	for offset < importSize {
		// Read import descriptor
		mem.Seek(int64(importVA+offset), 0)
		var desc ImageImportDescriptor
		binary.Read(mem, binary.LittleEndian, &desc)

		if desc.Name == 0 {
			break
		}

		// Read DLL name
		mem.Seek(int64(desc.Name), 0)
		dllName := readString(mem)
		if dllName == "" {
			offset += 20
			continue
		}

		// Load the DLL
		dllNamePtr, _ := syscall.UTF16PtrFromString(dllName)
		dllHandle, _, _ := loadLibraryW.Call(uintptr(unsafe.Pointer(dllNamePtr)))
		if dllHandle == 0 {
			dllNamePtr, _ = syscall.UTF16PtrFromString(dllName + ".dll")
			dllHandle, _, _ = loadLibraryW.Call(uintptr(unsafe.Pointer(dllNamePtr)))
			if dllHandle == 0 {
				offset += 20
				continue
			}
		}

		// Process thunks
		thunkOffset := desc.FirstThunk
		origThunkOffset := desc.OriginalFirstThunk
		if origThunkOffset == 0 {
			origThunkOffset = thunkOffset
		}

		for {
			mem.Seek(int64(origThunkOffset), 0)
			var thunkVal uint64
			if is64Bit {
				var v uint64
				binary.Read(mem, binary.LittleEndian, &v)
				thunkVal = v
			} else {
				var v uint32
				binary.Read(mem, binary.LittleEndian, &v)
				thunkVal = uint64(v)
			}

			if thunkVal == 0 {
				break
			}

			var procAddr uintptr
			if (is64Bit && thunkVal&0x8000000000000000 != 0) || (!is64Bit && thunkVal&0x80000000 != 0) {
				// Import by ordinal
				ordinal := uint16(thunkVal & 0xFFFF)
				procAddr, _, _ = getProcAddress.Call(dllHandle, uintptr(ordinal))
			} else {
				// Import by name
				mem.Seek(int64(thunkVal+2), 0)
				funcName := readString(mem)
				if funcName != "" {
					namePtr, _ := syscall.BytePtrFromString(funcName)
					procAddr, _, _ = getProcAddress.Call(dllHandle, uintptr(unsafe.Pointer(namePtr)))
				}
			}

			if procAddr != 0 {
				target := mem.Addr() + uintptr(thunkOffset)
				if is64Bit {
					*(*uintptr)(unsafe.Pointer(target)) = procAddr
				} else {
					*(*uint32)(unsafe.Pointer(target)) = uint32(procAddr)
				}
			}

			thunkOffset += uint32(ptrSize)
			origThunkOffset += uint32(ptrSize)
		}

		offset += 20
	}
	return nil
}

// loadExports loads export table.
func loadExports(mem *memory, exportVA uint32, exports map[string]uintptr) error {
	// Read export directory
	mem.Seek(int64(exportVA), 0)
	var header ImageExportDirectory
	if err := binary.Read(mem, binary.LittleEndian, &header); err != nil {
		return err
	}

	if header.NumberOfNames == 0 {
		return nil
	}

	// Read function addresses
	functions := make([]uint32, header.NumberOfFunctions)
	mem.Seek(int64(header.AddressOfFunctions), 0)
	for i := range functions {
		binary.Read(mem, binary.LittleEndian, &functions[i])
	}

	// Read name ordinals
	nameOrdinals := make([]uint16, header.NumberOfNames)
	mem.Seek(int64(header.AddressOfNameOrdinals), 0)
	for i := range nameOrdinals {
		binary.Read(mem, binary.LittleEndian, &nameOrdinals[i])
	}

	// Read name addresses
	nameAddresses := make([]uint32, header.NumberOfNames)
	mem.Seek(int64(header.AddressOfNames), 0)
	for i := range nameAddresses {
		binary.Read(mem, binary.LittleEndian, &nameAddresses[i])
	}

	// Build export map
	for i := uint32(0); i < header.NumberOfNames; i++ {
		mem.Seek(int64(nameAddresses[i]), 0)
		name := readString(mem)
		if name != "" {
			ordinal := nameOrdinals[i]
			exports[name] = mem.Addr() + uintptr(functions[ordinal])
		}
	}

	return nil
}

// readString reads a null-terminated string from memory.
func readString(mem *memory) string {
	var b []byte
	for {
		var v byte
		if err := binary.Read(mem, binary.LittleEndian, &v); err != nil || v == 0 {
			break
		}
		b = append(b, v)
	}
	return string(b)
}

// GetProc returns the address of a procedure by name.
func (m *MemoryModule) GetProc(name string) (uintptr, error) {
	if addr, ok := m.exports[name]; ok {
		return addr, nil
	}
	return 0, fmt.Errorf("procedure %s not found", name)
}

// Free releases the module memory.
func (m *MemoryModule) Free() error {
	if m.baseAddr != 0 {
		ret, _, _ := virtualFree.Call(m.baseAddr, 0, uintptr(MEM_RELEASE))
		if ret == 0 {
			return fmt.Errorf("failed to free memory")
		}
		m.baseAddr = 0
	}
	return nil
}
