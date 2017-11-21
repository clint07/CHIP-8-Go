package main

type Chip8 struct {
	vm *VM
	ppu *PPU
	apu *APU
}

func (chip8 *Chip8) Init() {
	// Initialize VM
	chip8.vm = &VM{}

	// Load VM fonts

	// Create PPU

	// Create APU
}

func (chip8 *Chip8) Load(filename *string) error {
	if err := chip8.vm.Load(filename); err != nil {
		return err
	}

	return nil
}

func (chip8 *Chip8) Run() {
	// Print ROM for sanity sake
	chip8.vm.printRAM()

	// Run ROM
	for i := 0; i < 2000; i++ {
		// Emulate a cycle
		chip8.vm.Cycle()
		// Check draw flag
		// Draw

		// Record key press
	}
}