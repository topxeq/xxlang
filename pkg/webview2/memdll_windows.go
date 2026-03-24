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
	rtlMoveMemory    = kernel32DLL.NewProc("RtlMoveMemory")
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
	handle     uintptr
	baseAddr   uintptr
	size       uint32
	exports    map[string]uintptr
	ordinalMap map[uint16]uintptr
}

var (
	moduleCache     = make(map[string]*MemoryModule)
	moduleCacheLock sync.RWMutex
)

// loadFromMemory loads a DLL from memory data.
// It returns a MemoryModule that can be used to get procedure addresses.
func loadFromMemory(data []byte) (*MemoryModule, error) {
	// Parse DOS header
	if len(data) < 64 {
		return nil, fmt.Errorf("data too small for DOS header")
	}

	dosHeader := (*ImageDOSHeader)(unsafe.Pointer(&data[0]))
	if binary.LittleEndian.Uint16(data[0:2]) != MZSignature {
		return nil, fmt.Errorf("invalid DOS signature")
	}

	// Parse PE header
	peOffset := dosHeader.NewHeaderAddr
	if int(peOffset)+4 > len(data) {
		return nil, fmt.Errorf("PE header offset out of bounds")
	}

	// Check PE signature
	peSig := binary.LittleEndian.Uint32(data[peOffset : peOffset+4])
	if peSig != PESignature {
		return nil, fmt.Errorf("invalid PE signature")
	}

	// Parse file header
	fileHeaderOffset := peOffset + 4
	fileHeader := (*ImageFileHeader)(unsafe.Pointer(&data[fileHeaderOffset]))

	// Determine if 32-bit or 64-bit
	optHeaderOffset := fileHeaderOffset + 20
	magic := binary.LittleEndian.Uint16(data[optHeaderOffset : optHeaderOffset+2])

	var imageBase uint64
	var sizeOfImage uint32
	var entryPoint uint32
	var numSections uint16
	var sectionsOffset uintptr
	var is64Bit bool

	switch magic {
	case ImageNTOptionalHeader64Magic:
		is64Bit = true
		optHeader := (*ImageOptionalHeader64)(unsafe.Pointer(&data[optHeaderOffset]))
		imageBase = optHeader.ImageBase
		sizeOfImage = optHeader.SizeOfImage
		entryPoint = optHeader.AddressOfEntryPoint
		numSections = fileHeader.NumberOfSections
		sectionsOffset = uintptr(optHeaderOffset + uint32(fileHeader.SizeOfOptionalHeader))
	case ImageNTOptionalHeader32Magic:
		is64Bit = false
		optHeader := (*ImageOptionalHeader32)(unsafe.Pointer(&data[optHeaderOffset]))
		imageBase = uint64(optHeader.ImageBase)
		sizeOfImage = optHeader.SizeOfImage
		entryPoint = optHeader.AddressOfEntryPoint
		numSections = fileHeader.NumberOfSections
		sectionsOffset = uintptr(optHeaderOffset + uint32(fileHeader.SizeOfOptionalHeader))
	default:
		return nil, fmt.Errorf("unknown PE magic: 0x%x", magic)
	}

	// Allocate memory for the image
	// Try to allocate at preferred base first
	baseAddr, _, err := virtualAlloc.Call(
		uintptr(imageBase),
		uintptr(sizeOfImage),
		uintptr(MEM_COMMIT|MEM_RESERVE),
		uintptr(PAGE_EXECUTE_READWRITE),
	)
	if baseAddr == 0 {
		// Allocate at any address
		baseAddr, _, err = virtualAlloc.Call(
			0,
			uintptr(sizeOfImage),
			uintptr(MEM_COMMIT|MEM_RESERVE),
			uintptr(PAGE_EXECUTE_READWRITE),
		)
		if baseAddr == 0 {
			return nil, fmt.Errorf("failed to allocate memory: %v", err)
		}
	}

	delta := uint64(baseAddr) - imageBase

	// Copy headers
	rtlMoveMemory.Call(
		baseAddr,
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(binary.LittleEndian.Uint32(data[optHeaderOffset+56:optHeaderOffset+60])), // SizeOfHeaders
	)

	// Map sections
	for i := uint16(0); i < numSections; i++ {
		section := (*ImageSectionHeader)(unsafe.Pointer(&data[sectionsOffset + uintptr(i)*40]))
		if section.SizeOfRawData == 0 {
			continue
		}

		destAddr := baseAddr + uintptr(section.VirtualAddress)
		srcAddr := uintptr(unsafe.Pointer(&data[section.PointerToRawData]))

		rtlMoveMemory.Call(
			destAddr,
			srcAddr,
			uintptr(section.SizeOfRawData),
		)
	}

	// Process relocations if needed
	if delta != 0 {
		var relocDir ImageDataDirectory
		if is64Bit {
			optHeader := (*ImageOptionalHeader64)(unsafe.Pointer(&data[optHeaderOffset]))
			relocDir = optHeader.DataDirectory[ImageDirectoryEntryBaseReloc]
		} else {
			optHeader := (*ImageOptionalHeader32)(unsafe.Pointer(&data[optHeaderOffset]))
			relocDir = optHeader.DataDirectory[ImageDirectoryEntryBaseReloc]
		}

		if relocDir.Size > 0 {
			if err := processRelocations(baseAddr, delta, data, relocDir.VirtualAddress, relocDir.Size, is64Bit); err != nil {
				virtualFree.Call(baseAddr, 0, uintptr(MEM_RELEASE))
				return nil, fmt.Errorf("relocation failed: %w", err)
			}
		}
	}

	// Process imports
	var importDir ImageDataDirectory
	if is64Bit {
		optHeader := (*ImageOptionalHeader64)(unsafe.Pointer(&data[optHeaderOffset]))
		importDir = optHeader.DataDirectory[ImageDirectoryEntryImport]
	} else {
		optHeader := (*ImageOptionalHeader32)(unsafe.Pointer(&data[optHeaderOffset]))
		importDir = optHeader.DataDirectory[ImageDirectoryEntryImport]
	}

	if importDir.Size > 0 {
		// Need to read from allocated memory, not from original data
		importData := (*[0]byte)(unsafe.Pointer(baseAddr + uintptr(importDir.VirtualAddress)))
		if err := processImports(baseAddr, importData, is64Bit); err != nil {
			virtualFree.Call(baseAddr, 0, uintptr(MEM_RELEASE))
			return nil, fmt.Errorf("import processing failed: %w", err)
		}
	}

	// Process exports
	exports := make(map[string]uintptr)
	ordinalMap := make(map[uint16]uintptr)

	var exportDir ImageDataDirectory
	if is64Bit {
		optHeader := (*ImageOptionalHeader64)(unsafe.Pointer(&data[optHeaderOffset]))
		exportDir = optHeader.DataDirectory[ImageDirectoryEntryExport]
	} else {
		optHeader := (*ImageOptionalHeader32)(unsafe.Pointer(&data[optHeaderOffset]))
		exportDir = optHeader.DataDirectory[ImageDirectoryEntryExport]
	}

	if exportDir.Size > 0 {
		exportData := (*ImageExportDirectory)(unsafe.Pointer(baseAddr + uintptr(exportDir.VirtualAddress)))
		loadExports(baseAddr, exportData, exports, ordinalMap)
	}

	module := &MemoryModule{
		handle:     baseAddr,
		baseAddr:   baseAddr,
		size:       sizeOfImage,
		exports:    exports,
		ordinalMap: ordinalMap,
	}

	// Call DllMain with DLL_PROCESS_ATTACH
	if entryPoint != 0 {
		dllMain := baseAddr + uintptr(entryPoint)
		callDllMain(dllMain, baseAddr, 1) // DLL_PROCESS_ATTACH = 1
	}

	return module, nil
}

// processRelocations handles base relocations.
func processRelocations(baseAddr uintptr, delta uint64, data []byte, relocVA, relocSize uint32, is64Bit bool) error {
	offset := uintptr(0)
	for offset < uintptr(relocSize) {
		// Read relocation block header from the allocated memory
		block := (*ImageBaseRelocation)(unsafe.Pointer(baseAddr + uintptr(relocVA) + offset))
		if block.SizeOfBlock == 0 {
			break
		}

		numEntries := (block.SizeOfBlock - 8) / 2
		entries := (*[0]uint16)(unsafe.Pointer(baseAddr + uintptr(relocVA) + offset + 8))

		for i := uint32(0); i < uint32(numEntries); i++ {
			entry := entries[i]
			relType := entry >> 12
			relOffset := uintptr(entry & 0xFFF)

			targetAddr := baseAddr + uintptr(block.VirtualAddress) + relOffset

			switch relType {
			case ImageRelBasedAbsolute:
				// No relocation needed
			case ImageRelBasedHighLow:
				// 32-bit relocation
				val := *(*uint32)(unsafe.Pointer(targetAddr))
				val += uint32(delta)
				*(*uint32)(unsafe.Pointer(targetAddr)) = val
			case ImageRelBasedDir64:
				// 64-bit relocation
				val := *(*uint64)(unsafe.Pointer(targetAddr))
				val += delta
				*(*uint64)(unsafe.Pointer(targetAddr)) = val
			}
		}

		offset += uintptr(block.SizeOfBlock)
	}
	return nil
}

// processImports resolves import addresses.
func processImports(baseAddr uintptr, importData *[0]byte, is64Bit bool) error {
	offset := uintptr(0)
	ptrSize := 4
	if is64Bit {
		ptrSize = 8
	}

	for {
		desc := (*ImageImportDescriptor)(unsafe.Pointer(uintptr(unsafe.Pointer(importData)) + offset))
		if desc.Name == 0 {
			break
		}

		// Get DLL name - desc.Name is RVA, add baseAddr to get actual pointer
		dllName := ptrToString(baseAddr + uintptr(desc.Name))
		if dllName == "" {
			offset += 20
			continue
		}

		// Load the DLL
		dllHandle, _, _ := loadLibraryW.Call(uintptr(unsafe.Pointer(stringToUTF16Ptr(dllName))))
		if dllHandle == 0 {
			// Try without .dll extension
			dllHandle, _, _ = loadLibraryW.Call(uintptr(unsafe.Pointer(stringToUTF16Ptr(dllName + ".dll"))))
			if dllHandle == 0 {
				offset += 20
				continue
			}
		}

		// Process thunks
		thunkOffset := uintptr(desc.FirstThunk)
		origThunkOffset := uintptr(desc.OriginalFirstThunk)
		if origThunkOffset == 0 {
			origThunkOffset = thunkOffset
		}

		for {
			var thunkVal uint64
			if is64Bit {
				thunkVal = uint64(*(*uintptr)(unsafe.Pointer(baseAddr + origThunkOffset)))
			} else {
				thunkVal = uint64(*(*uint32)(unsafe.Pointer(baseAddr + origThunkOffset)))
			}

			if thunkVal == 0 {
				break
			}

			var procAddr uintptr
			if thunkVal&0x8000000000000000 != 0 || thunkVal&0x80000000 != 0 {
				// Import by ordinal
				ordinal := uint16(thunkVal & 0xFFFF)
				procAddr, _, _ = getProcAddress.Call(dllHandle, uintptr(ordinal))
			} else {
				// Import by name
				nameOffset := uintptr(thunkVal & 0x7FFFFFFFFFFFFFFF)
				if nameOffset != 0 {
					namePtr := baseAddr + nameOffset + 2 // Skip hint
					nameStr := ptrToString(namePtr)
					if nameStr != "" {
						nameBytes := []byte(nameStr)
						procAddr, _, _ = getProcAddress.Call(dllHandle, uintptr(unsafe.Pointer(&nameBytes[0])))
					}
				}
			}

			if procAddr != 0 {
				if is64Bit {
					*(*uintptr)(unsafe.Pointer(baseAddr + thunkOffset)) = procAddr
				} else {
					*(*uint32)(unsafe.Pointer(baseAddr + thunkOffset)) = uint32(procAddr)
				}
			}

			thunkOffset += uintptr(ptrSize)
			origThunkOffset += uintptr(ptrSize)
		}

		offset += 20
	}
	return nil
}

// loadExports loads export table.
func loadExports(baseAddr uintptr, exportDir *ImageExportDirectory, exports map[string]uintptr, ordinalMap map[uint16]uintptr) {
	names := (*[0]uint32)(unsafe.Pointer(baseAddr + uintptr(exportDir.AddressOfNames)))
	functions := (*[0]uint32)(unsafe.Pointer(baseAddr + uintptr(exportDir.AddressOfFunctions)))
	ordinals := (*[0]uint16)(unsafe.Pointer(baseAddr + uintptr(exportDir.AddressOfNameOrdinals)))

	for i := uint32(0); i < exportDir.NumberOfNames; i++ {
		nameOffset := names[i]
		name := ptrToString(baseAddr + uintptr(nameOffset))
		ordinal := ordinals[i]
		funcOffset := functions[ordinal]
		exports[name] = baseAddr + uintptr(funcOffset)
		ordinalMap[ordinal] = baseAddr + uintptr(funcOffset)
	}
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
		// Call DllMain with DLL_PROCESS_DETACH
		// We would need to store entry point for this

		ret, _, _ := virtualFree.Call(m.baseAddr, 0, uintptr(MEM_RELEASE))
		if ret == 0 {
			return fmt.Errorf("failed to free memory")
		}
		m.baseAddr = 0
	}
	return nil
}

// Helper functions

func ptrToString(ptr uintptr) string {
	if ptr == 0 {
		return ""
	}
	var length int
	for {
		b := *(*byte)(unsafe.Pointer(ptr + uintptr(length)))
		if b == 0 {
			break
		}
		length++
	}
	return string((*[0]byte)(unsafe.Pointer(ptr))[:length:length])
}

func stringToUTF16Ptr(s string) *uint16 {
	ptr, _ := syscall.UTF16PtrFromString(s)
	return ptr
}

func callDllMain(entry, hinst uintptr, reason uint32) {
	// DllMain signature: BOOL WINAPI DllMain(HINSTANCE hinstDLL, DWORD fdwReason, LPVOID lpvReserved)
	syscall.Syscall(entry, 3, hinst, uintptr(reason), 0)
}
