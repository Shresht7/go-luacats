package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Shresht7/go-luacats"
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
	err := luacats.Generate(source, target)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Generated LuaCATS")
}
