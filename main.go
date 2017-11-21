package main

import (
	"flag"
	"github.com/clint07/CHIP-8/chip8"
)

func main() {
	// Parse command line arguments
	filename := flag.String("file", "", "ROM filename")
	flag.Parse()

	// Initialize VM
	chip8 := CHIP8.Chip8{}
	chip8.Init()

	// Load ROM
	if err := chip8.Load(filename); err != nil {
		panic(err)
	}

	// Run ROM
	chip8.Run()
}
