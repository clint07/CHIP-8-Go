package CHIP8

//import (
//	"github.com/veandco/go-sdl2/sdl"
//	"fmt"
//)

type Chip8 struct {
	cpu *CPU
	ppu *PPU
	apu *APU
}

func (self *Chip8) Init() {
	// Initialize CPU
	self.cpu = &CPU{}
	self.cpu.Init()

	// Create PPU
	self.ppu = &PPU{}
	self.ppu.Init()

	// Create APU
	self.apu = &APU{}
	self.apu.Init()
}

func (self *Chip8) Load(filename *string) error {
	if err := self.cpu.LoadROM(filename); err != nil {
		return err
	}

	return nil
}

func (self *Chip8) Run() {
	// Print ROM for sanity sake
	self.cpu.printRAM()

	// Run ROM
	for {
		// Emulate a cycle
		self.cpu.Cycle()

		// Check draw flag
		if self.cpu.DF {
			// Draw
			self.ppu.Draw(&self.cpu.GFX)

			// Don't forget to set the draw flag back
			self.cpu.DF = false
		}

		// Check keyboard input
		if exit := self.ppu.Poll(&self.cpu.Key); exit {
			break
		}
	}
}


func (self *Chip8) Shutdown() {
	self.ppu.destroy()
}
