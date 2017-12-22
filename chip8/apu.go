package CHIP8

import "fmt"

type APU struct {
}

func (apu *APU) beep() {
	// Simple audio output that uses the system's alert sound to emulate a Chip-8 beep
	fmt.Print("\x07")
}

