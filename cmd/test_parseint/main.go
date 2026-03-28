package main

import (
	"fmt"
	"strconv"
)

func main() {
	s := "1234567890123456789"
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Value:", v)
	}
	
	// Try base 10
	v2, err2 := strconv.ParseInt(s, 10, 64)
	if err2 != nil {
		fmt.Println("Error:", err2)
	} else {
		fmt.Println("Value (base 10):", v2)
	}
}
