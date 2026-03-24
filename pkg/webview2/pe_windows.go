//go:build windows

// Package webview2 provides WebView2 bindings for Xxlang.
// This file contains PE format definitions for memory DLL loading.
package webview2

// PE format constants
const (
	// Signatures
	MZSignature = 0x5A4D // "MZ"
	PESignature = 0x00004550 // "PE\0\0"

	// Optional header magic
	ImageNTOptionalHeader32Magic = 0x10b
	ImageNTOptionalHeader64Magic = 0x20b

	// Machine types
	ImageFileMachineI386   = 0x014c
	ImageFileMachineAMD64  = 0x8664
	ImageFileMachineARM64  = 0xAA64

	// DLL characteristics
	ImageDLLCharacteristicsDynamicBase = 0x0040

	// Directory entries
	ImageDirectoryEntryExport    = 0
	ImageDirectoryEntryImport    = 1
	ImageDirectoryEntryBaseReloc = 5
	ImageDirectoryEntryTLS       = 9

	// Section characteristics
	ImageSectionCharacteristicsMemoryExecute = 0x20000000
	ImageSectionCharacteristicsMemoryRead    = 0x40000000
	ImageSectionCharacteristicsMemoryWrite   = 0x80000000

	// Relocation types
	ImageRelBasedAbsolute = 0
	ImageRelBasedHighLow  = 3
	ImageRelBasedDir64    = 10
)

// ImageDOSHeader represents the DOS header.
// e_lfanew (NewHeaderAddr) is at offset 0x3C (60 bytes).
type ImageDOSHeader struct {
	_             [30]uint16 // 60 bytes of DOS header fields
	NewHeaderAddr uint32     // e_lfanew: offset to PE header
}

// ImageFileHeader represents the COFF file header.
type ImageFileHeader struct {
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}

// ImageDataDirectory represents a data directory entry.
type ImageDataDirectory struct {
	VirtualAddress uint32
	Size           uint32
}

// ImageOptionalHeader64 represents the 64-bit optional header.
type ImageOptionalHeader64 struct {
	Magic                       uint16
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32
	BaseOfCode                  uint32
	ImageBase                   uint64
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32
	SizeOfHeaders               uint32
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	SizeOfStackReserve          uint64
	SizeOfStackCommit           uint64
	SizeOfHeapReserve           uint64
	SizeOfHeapCommit            uint64
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
	DataDirectory               [16]ImageDataDirectory
}

// ImageOptionalHeader32 represents the 32-bit optional header.
type ImageOptionalHeader32 struct {
	Magic                       uint16
	MajorLinkerVersion          uint8
	MinorLinkerVersion          uint8
	SizeOfCode                  uint32
	SizeOfInitializedData       uint32
	SizeOfUninitializedData     uint32
	AddressOfEntryPoint         uint32
	BaseOfCode                  uint32
	BaseOfData                  uint32
	ImageBase                   uint32
	SectionAlignment            uint32
	FileAlignment               uint32
	MajorOperatingSystemVersion uint16
	MinorOperatingSystemVersion uint16
	MajorImageVersion           uint16
	MinorImageVersion           uint16
	MajorSubsystemVersion       uint16
	MinorSubsystemVersion       uint16
	Win32VersionValue           uint32
	SizeOfImage                 uint32
	SizeOfHeaders               uint32
	CheckSum                    uint32
	Subsystem                   uint16
	DllCharacteristics          uint16
	SizeOfStackReserve          uint32
	SizeOfStackCommit           uint32
	SizeOfHeapReserve           uint32
	SizeOfHeapCommit            uint32
	LoaderFlags                 uint32
	NumberOfRvaAndSizes         uint32
	DataDirectory               [16]ImageDataDirectory
}

// ImageSectionHeader represents a section header.
type ImageSectionHeader struct {
	Name                 [8]byte
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

// ImageBaseRelocation represents a base relocation block.
type ImageBaseRelocation struct {
	VirtualAddress uint32
	SizeOfBlock    uint32
}

// ImageImportDescriptor represents an import descriptor.
type ImageImportDescriptor struct {
	OriginalFirstThunk uint32
	TimeDateStamp      uint32
	ForwarderChain     uint32
	Name               uint32
	FirstThunk         uint32
}

// ImageExportDirectory represents an export directory.
type ImageExportDirectory struct {
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfNames        uint32
	AddressOfNameOrdinals uint32
}
