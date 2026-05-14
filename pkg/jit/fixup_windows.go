//go:build windows && amd64
// +build windows,amd64

package jit

type fixup struct {
	offset int
	label  string
	size   int
}
