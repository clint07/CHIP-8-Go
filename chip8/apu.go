package CHIP8

import "fmt"

type APU struct {
}

func (apu *APU) beep() {
	fmt.Print("\x07")
}

