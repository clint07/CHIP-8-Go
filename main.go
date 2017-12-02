package main

import (
	"flag"
	"github.com/clint07/Go CHIP-8/chip8"
	"time"
)

func main() {
	// Parse command line arguments
	filename := flag.String("file", "", "ROM filename")
	flag.Parse()

	// Initialize CHIP-8
	chip8 := CHIP8.Chip8{}
	chip8.Init()

	// Load ROM
	if err := chip8.Load(filename); err != nil {
		panic(err)
	}

	// Run ROM
	chip8.Run()
	time.Sleep(5 * time.Second)

	// Shutdown CHIP-8
	chip8.Shutdown()
}
