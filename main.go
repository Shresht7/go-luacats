package main

import (
	"fmt"
	"log"
	"os"
)

func main() {

	// Show help message if needed
	if len(os.Args) < 2 {
		log.Fatal("Usage: go-luacats <path-to-go-file-or-directory>")
	}

	// Extract filepath out of command-line arguments
	source := os.Args[1]
	target := ""
	if len(os.Args) > 2 {
		target = os.Args[2]
	}

	// Generate LuaCATS files
	err := Generate(source, target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Generated LuaCATS")
}
