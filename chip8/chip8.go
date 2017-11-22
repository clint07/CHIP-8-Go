package CHIP8

type Chip8 struct {
	vm *VM
	ppu *PPU
	apu *APU
}

func (this *Chip8) Init() {
	// Initialize VM
	this.vm = &VM{}
	this.vm.Init()

	// Create PPU
	this.ppu = &PPU{}
	this.ppu.Init()

	// Create APU
	this.apu = &APU{}
	this.apu.Init()
}

func (this *Chip8) Load(filename *string) error {
	if err := this.vm.Load(filename); err != nil {
		return err
	}

	return nil
}

func (this *Chip8) Run() {
	// Print ROM for sanity sake
	this.vm.printRAM()

	// Run ROM
	for i := 0; i < 2000; i++ {
		// Emulate a cycle
		this.vm.Cycle()
		// Check draw flag
		// Draw

		// Record key press
	}
}

func (this *Chip8) Shutdown() {
	this.ppu.destroy()
}