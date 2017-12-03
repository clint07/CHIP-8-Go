package main

import (
	"flag"
	"github.com/clint07/CHIP-8/chip8"
	"strconv"
)

func main() {
	// Parse command line arguments
	flagFilename := flag.String("file", "", "ROM filename")
	flagFps := flag.String("fps", "120", "120 FPS recommended unless using ROMs such as a clock ROM")
	flag.Parse()

	// Initialize CHIP-8
	chip8 := CHIP8.Chip8{}
	chip8.Init()

	// Load ROM
	if err := chip8.Load(flagFilename); err != nil {
		panic(err)
	}

	// Run ROM
	fps, err := strconv.Atoi(*flagFps)
	if err != nil {
		panic(err)
	}

	chip8.Run(fps)

	// Shutdown CHIP-8
	chip8.Shutdown()
}
