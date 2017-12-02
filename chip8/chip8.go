package CHIP8

import (
	"time"
	"fmt"
	"os"
)

type Chip8 struct {
	cpu *CPU
	ppu *PPU
	apu *APU
}

func (chip8 *Chip8) Init() {
	// Initialize CPU
	chip8.cpu = &CPU{}
	chip8.cpu.Init()

	// Create PPU
	chip8.ppu = &PPU{}
	chip8.ppu.Init()

	// Create APU
	chip8.apu = &APU{}
	chip8.apu.Init()
}

func (chip8 *Chip8) Load(filename *string) error {
	if err := chip8.cpu.LoadROM(filename); err != nil {
		return err
	}

	return nil
}

func (chip8 *Chip8) Run() {
	// Print ROM for sanity sake
	chip8.cpu.printRAM()
	tick := time.Tick(16 * time.Millisecond)

	// Run ROM
	for {
		select {
			case <- tick:

			// Emulate a cycle. Panic if error has occurred.
			if err := chip8.cpu.Cycle(); err != nil {
				panic(err)
			}

			// Check draw flag
			if chip8.cpu.DF {
				// Draw
				chip8.ppu.Draw(&chip8.cpu.GFX)

				// Don't forget to set the draw flag back
				chip8.cpu.DF = false
			}

			// Check keyboard input
			if exit := chip8.ppu.Poll(&chip8.cpu.Key); exit {
				break
			}
		}
	}
}


func (chip8 *Chip8) Shutdown() {
	chip8.ppu.destroy()
}
