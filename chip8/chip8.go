package main

import (
	"fmt"
	"io/ioutil"
)

type CHIP8 struct {
	// CHIP-8 is capable of accessing 4KB (4,096 bytes) of RAM.
	// Because the first 512 bytes are reserved for the interpreter, CHIP-8 programs start at address 0x200 (512).
	RAM [4096]byte

	// CHIP-8 screen is 64x32 pixels.
	GFX [64 * 32]byte

	// 16 16-bit stack used for saving addresses before subroutines.
	Stack [16]uint

	// 16 8-bit Registers: V0 - VE are general registers and VF is a flag register.
	V [16]byte

	// 16-bit Program counter. All programs start at 0x200.
	PC uint

	// 8-bit Stack pointer
	SP uint

	// Address register
	I uint

	// Delay timer
	DT int64

	// Sound timer
	ST int64

	// Keys
	Key [16]uint

	// ROM Size: length of CHIP-8 program byte array
	RS int
}

func (chip8 *CHIP8) LoadROM(filename *string) error {
	// Read file into byte array
	rom, err := ioutil.ReadFile(*filename)
	if err != nil {
		return err
	}

	// Save ROM size
	chip8.RS = len(rom)

	// Copy program byte array into RAM
	for i := uint(0); i < uint(len(rom)); i++ {
		chip8.RAM[i+chip8.PC] = rom[i]
	}

	return nil
}

// Helpful for debugging
func (chip8 *CHIP8) printRAM() {
	for i := 0; i < chip8.RS+512; i++ {
		fmt.Printf("%d: %x\n", i, chip8.RAM[i])
	}
}
