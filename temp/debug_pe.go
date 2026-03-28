package main

import (
    "encoding/binary"
    "fmt"
    "os"
    "unsafe"
)

type ImageDOSHeader struct {
    _            [16]uint16
    NewHeaderAddr uint32
}

func main() {
    data, err := os.ReadFile("pkg/webview2/WebView2Loader_x64.dll")
    if err != nil {
        fmt.Println("Error reading file:", err)
        return
    }
    
    fmt.Printf("File size: %d bytes\n", len(data))
    fmt.Printf("First 2 bytes (MZ): %x %x\n", data[0], data[1])
    
    dosHeader := (*ImageDOSHeader)(unsafe.Pointer(&data[0]))
    fmt.Printf("PE header offset: 0x%x (%d)\n", dosHeader.NewHeaderAddr, dosHeader.NewHeaderAddr)
    
    peOffset := dosHeader.NewHeaderAddr
    if int(peOffset)+4 > len(data) {
        fmt.Println("PE offset out of bounds!")
        return
    }
    
    peSig := binary.LittleEndian.Uint32(data[peOffset : peOffset+4])
    fmt.Printf("PE signature bytes: %x %x %x %x\n", data[peOffset], data[peOffset+1], data[peOffset+2], data[peOffset+3])
    fmt.Printf("PE signature value: 0x%x (expected 0x4550)\n", peSig)
}
