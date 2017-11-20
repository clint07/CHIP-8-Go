package main

import "flag"

func main() {
	// Parse command line arguments
	filename := flag.String("file", "", "ROM filename")
	flag.Parse()

	// Initialize CHIP8
	chip8 := &CHIP8{PC: 0x200}


	// Load ROM
	if err := chip8.LoadROM(filename); err != nil {
		panic(err)
	}

	// Print ROM for sanity sake
	chip8.printRAM()

	// Load game through arguments

	//for {
		// Emulate a cycle
		// Check draw flag
			// Draw

		// Record key press
	//}
}
