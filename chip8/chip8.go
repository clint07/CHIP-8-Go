package CHIP8

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
	for i := 0; i < 20; i++ {
		// Emulate a cycle
		self.cpu.Cycle()
		// Check draw flag
		// Draw

		// Record key press
	}
}

func (self *Chip8) Shutdown() {
	self.ppu.destroy()
}
