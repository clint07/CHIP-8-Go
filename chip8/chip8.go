package CHIP8

type Chip8 struct {
	vm *VM
	ppu *PPU
	apu *APU
}

func (self *Chip8) Init() {
	// Initialize VM
	self.vm = &VM{}
	self.vm.Init()

	// Create PPU
	self.ppu = &PPU{}
	self.ppu.Init()

	// Create APU
	self.apu = &APU{}
	self.apu.Init()
}

func (self *Chip8) Load(filename *string) error {
	if err := self.vm.Load(filename); err != nil {
		return err
	}

	return nil
}

func (self *Chip8) Run() {
	// Print ROM for sanity sake
	self.vm.printRAM()

	// Run ROM
	for i := 0; i < 2000; i++ {
		// Emulate a cycle
		self.vm.Cycle()
		// Check draw flag
		// Draw

		// Record key press
	}
}

func (self *Chip8) Shutdown() {
	self.ppu.destroy()
}