package CHIP8

import (
	"testing"
)

// Instruction 00E0: Clear the display.
func TestClear(t *testing.T) {

}

// TODO test PC, SP, sound and delay timer

// Instruction 00EE: Return from a subroutine.
// The CPU sets the program counter to the address at the top of the stack,
// then subtracts 1 from the stack pointer.
func TestRet(t *testing.T) {
	cpu := &CPU{}
	cpu.PC = 0xFF
	cpu.Stack[cpu.SP] = 512
	cpu.SP += 1

	cpu.ret()

	if cpu.PC != 514 {
		t.Errorf("TestRet: failed to pop the stack and set PC to it. Expected: %d Received: %d", 514, cpu.PC)
	}

	if cpu.SP != 0 {
		t.Errorf("TestRet: failed to decrement SP after popping the stack. Expected: %d Received: %d", 0, cpu.SP)
	}

}

// Instruction 1nnn: Jump to location nnn.
// The CPU sets the program counter to nnn.
func TestJump(t *testing.T) {
	cpu := &CPU{}
	cpu.jump(512)

	if cpu.PC != 512 {
		t.Errorf("TestJump: failed to jump to instruction. Expected: %d Received: %d", 512, cpu.PC)
	}
}

// Instruction 2nnn: Call subroutine at nnn.
// The CPU increments the stack pointer, then puts the current PC on the top of the stack.
// The PC is then set to nnn.
func TestCall(t *testing.T) {
	cpu := &CPU{}
	cpu.PC = 512

	cpu.call(777)

	if cpu.SP != 1 {
		t.Errorf("TestCall: failed to increment SP. Expected: %d Received: %d", 1, cpu.SP)
	}

	if cpu.Stack[0] != 512 {
		t.Errorf("TestCall: failed to placed PC on the stack. Expected %d Received: %d", 512, cpu.Stack[0])
	}

	if cpu.PC != 777 {
		t.Errorf("TestCall: failed to jump to instruction. Expected: %d Received: %d", 777, cpu.PC)
	}
}

// Instruction 3xkk: Skip next instruction if Vx = kk.
// The CPU compares register Vx to kk, and if they are equal,
// increments the program counter by 2.
func TestSkipIf(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 7

	if cpu.skipIf(0x0, 7); cpu.PC != 4 {
		t.Errorf("TestSkipIf: failed to skip.")
	}

	if cpu.skipIf(0x0, 9); cpu.PC != 6 {
		t.Errorf("TestSkipIf: skipped by error")
	}
}

// Instruction 4xkk: Skip next instruction if Vx != kk.
// The CPU compares register Vx to kk, and if they are not equal,
// increments the program counter by 2.
func TestSkipIfNot(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 7

	if cpu.skipIf(0x0, 9); cpu.PC == 4 {
		t.Errorf("TestSkipIf: failed to skip")
	}

	if cpu.skipIf(0x0, 7); cpu.PC != 6 {
		t.Errorf("TestSkipIf: skipped by error.")
	}
}

// Instruction 5xy0: Skip next instruction if Vx = Vy.
// The CPU compares register Vx to register Vy, and if they are equal,
// increments the program counter by 2.
func TestSkipIfXY(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 7
	cpu.V[0xE] = 7

	if cpu.skipIfXY(0x0, 0xE); cpu.PC != 4 {
		t.Errorf("TestSkipIf: failed to skip")
	}

	if cpu.skipIfXY(0x0, 0xF); cpu.PC != 6 {
		t.Errorf("TestSkipIf: skipped by error.")
	}
}

// Instruction 6xkk: Set Vx = kk.
// The CPU puts the value kk into register Vx.
func TestLoad(t *testing.T) {
	cpu := &CPU{}

	if cpu.load(0x0, 7); cpu.V[0x0] != 7 {
		t.Errorf("TestLoad: failed to load %d into V%X", 7, 0x0)
	}


}

// Instruction 7xkk: Set Vx = Vx + kk.
// Adds the value kk to the value of register Vx, then stores the result in Vx.
func TestAdd(t *testing.T) {
	cpu := &CPU{}

	cpu.V[0x0] = 7

	if cpu.add(0x0, 7); cpu.V[0x0] != 14 {
		t.Errorf("TestAdd: failed to add %d to V%X. Expected: %d Result: %d", 7, 0x0, 14, cpu.V[0x0])
	}
}

// Instruction 8xy0: Set Vx = Vy.
// Stores the value of register Vy in register Vx.
func TestLoadXY(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0xE] = 7

	if cpu.loadXY(0x0, 0xE); cpu.V[0x0] != 7 {
		t.Errorf("TestLoadXY: failed to store value in V%X to V%X. Expected: %d Result %d", 0xE, 0x0, 7, cpu.V[0x0])
	}
}

// Instruction 8xy1: Set Vx = Vx OR Vy.
// Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
// A bitwise OR compares the corresponding bits from two values, and if either bit is 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func TestOrXY(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 9
	cpu.V[0xE] = 7

	if cpu.orXY(0x0, 0xE); cpu.V[0x0] != 15 {
		t.Errorf("TestOrXY: failed to Or V%X and V%X. Expected: %d Result: %d", 0x0, 0xE, 15, cpu.V[0x0])
	}

	if cpu.V[0xE] != 7 {
		t.Errorf("TestOrXY: operated on the wrong register")
	}
}

// Instruction 8xy2: Set Vx = Vx AND Vy.
// Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
// A bitwise AND compares the corresponding bits from two values, and if both bits are 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func TestAndXY(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 9
	cpu.V[0xE] = 7

	if cpu.andXY(0x0, 0xE); cpu.V[0x0] != 1 {
		t.Errorf("TestAndXY: failed to And V%X and V%x, Expected: %d Result: %d", 0x0, 0xE, 1, cpu.V[0x0])
	}

	if cpu.V[0xE] != 7 {
		t.Errorf("TestAndXY: operated on the wrong register")
	}
}

// Instruction 8xy3: Set Vx = Vx XOR Vy.
// Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx.
// An exclusive OR compares the corrseponding bits from two values,
// and if the bits are not both the same, then the corresponding bit in the result is set to 1.
// Otherwise, it is 0.
func TestXorXY(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 9
	cpu.V[0xE] = 7

	if cpu.xorXY(0x0, 0xE); cpu.V[0x0] != 14 {
		t.Errorf("TestXorXY: failed to Xor V%X and V%x, Expected: %d Result: %d", 0x0, 0xE, 14, cpu.V[0x0])
	}

	if cpu.V[0xE] != 7 {
		t.Errorf("TestXOrXY: operated on the wrong register")
	}
}

// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
// VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
func TestAddXY(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 9
	cpu.V[0xE] = 7
	cpu.V[0xF] = 0

	if cpu.addXY(0x0, 0xE); cpu.V[0x0] != 16 {
		t.Errorf("TestAddXY: failed to add V%X and V%X. Expected: %d Result: %d", 0x0, 0xE, 16, cpu.V[0x0])
	} else if cpu.V[0xF] != 0 {
		t.Errorf("TestAddXY: failed to set the VF flag correctly. Expected: %d Result: %d", 0, cpu.V[0xF])
	}

	if cpu.V[0xE] != 7 {
		t.Errorf("TestAddXY: operated on the wrong register")
	}

	cpu.V[0xE] = 255
	if cpu.addXY(0x0, 0xE); cpu.V[0xF] != 1 {
		t.Errorf("TestAddXY: failed to set the VF flag correctly. Expected: %d Result: %d", 1, cpu.V[0xF])
	}

}

// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
// and the results stored in Vx.
func TestSubXY(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 9
	cpu.V[0xE] = 7
	cpu.V[0xF] = 0

	if cpu.subXY(0x0, 0xE); cpu.V[0x0] != 2 {
		t.Errorf("TestSubXY: failed to subtract V%X and V%X. Expected: %d Result: %d", 0x0, 0xE, 2, cpu.V[0x0])
	} else if cpu.V[0xF] != 1 {
		t.Errorf("TestAddXY: failed to set the VF flag correctly. Expected: %d Result: %d", 1, cpu.V[0xF])
	}
}

// Instruction 8xy6: Set Vx = Vx SHR 1.
// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
// Then Vx is divided by 2.
func TestShiftRight(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 0x04

	if cpu.shiftRight(0x0); cpu.V[0x0] != 2 {
		t.Errorf("TestShiftRight: failed to shift right on V%X. Expected: %d Result: %d", 0x0, 2, cpu.V[0x0])
	} else if cpu.V[0xF] != 0 {
		t.Errorf("TestShiftRight: failed to set the VF flag correctly. Expected: %d Result: %d", 0, cpu.V[0xF])
	}


	cpu.V[0x0] = 0x5
	if cpu.shiftRight(0x0); cpu.V[0x0] != 2 {
		t.Errorf("TestShiftRight: failed to shift right on V%X. Expected: %d Result: %d", 0x0, 2, cpu.V[0x0])
	} else if cpu.V[0xF] != 1 {
		t.Errorf("TestShiftRight: failed to set the VF flag correctly. Expected: %d Result: %d", 1, cpu.V[0xF])
	}

}

// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
// If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy,
// and the results stored in Vx.
func TestSubYX(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 7
	cpu.V[0xE] = 9

	if cpu.subYX(0x0, 0xE); cpu.V[0x0] != 2 {
		t.Errorf("TestSubYX: failed to subtract V%X and V%X. Expected: %d Result: %d", 0x0, 0xE, 2, cpu.V[0x0])
	} else if cpu.V[0xF] != 1 {
		t.Errorf("TestsubYX: failed to set the VF flag correctly. Expected: %d Result %d", 1, cpu.V[0xF])
	}


	cpu.V[0x0] = 9
	cpu.V[0xE] = 7

	if cpu.subYX(0x0, 0xE); cpu.V[0x0] != 254 {
		t.Errorf("TestSubYX: failed to subtract V%X and V%X. Expected: %d Result: %d", 0x0, 0xE, 254, cpu.V[0x0])
	} else if cpu.V[0xF] != 0 {
		t.Errorf("TestsubYX: failed to set the VF flag correctly. Expected: %d Result %d", 0, cpu.V[0xF])
	}
}

// Instruction 8xyE: Set Vx = Vx SHL 1.
// If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
// Then Vx is multiplied by 2.
func TestShiftLeft(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 128

	if cpu.shiftLeft(0x0); cpu.V[0x0] != 0 {
		t.Errorf("TestShiftLeft: failed to shift left on V%X. Expected: %d Result: %d", 0x0, 0, cpu.V[0x0])
	} else if cpu.V[0xF] != 1 {
		t.Errorf("TestShiftLeft: failed to set the VF flag correctly. Expected: %d Result %d", 1, cpu.V[0xf])
	}

}

// Instruction 9xy0: Skip next instruction if Vx != Vy.
// The values of Vx and Vy are compared, and if they are not equal,
// the program counter is increased by 2.
func TestSkipIfNotXY(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0X0] = 7
	cpu.V[0xE] = 9

	if cpu.skipIfNotXY(0x0, 0xE); cpu.PC != 4 {
		t.Errorf("TestSkipIfNotXY: failed to skip instruction. Expected: %d Result %d", 4, cpu.PC)
	}

	cpu.V[0xE] = 7

	if cpu.skipIfNotXY(0x0, 0xE); cpu.PC != 6 {
		t.Errorf("TestSkipIfNotXY: failed to not skip instruction. Expected: %d Result %d", 6, cpu.PC)
	}

}

// Instruction Annn: Set I = nnn.
// The value of register I is set to nnn.
func TestLoadI(t *testing.T) {
	cpu := &CPU{}

	if cpu.loadI(7); cpu.I != 7 {
		t.Errorf("TestLoadI: failed to load nnn into I. Expected: %d Result %d", 7, cpu.I)
	}
}

// Instruction Bnnn: Jump to location nnn + V0.
// The program counter is set to nnn plus the value of V0.
func TestJumpV0(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 6

	if cpu.jumpV0(8); cpu.PC != 14 {
		t.Errorf("TestJumpV0: failed to jump nnn times plus V0. Expected: %d Result %d", 14, cpu.PC)
	}
}

// Instruction Dxyn: Display n-byte sprite starting at memory location I at (Vx, Vy),
// set VF = collision.
//
// The CPU reads n bytes from memory, starting at the address stored in I.
// These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
// Sprites are XORed onto the existing screen. If this causes any pixels to be erased,
// VF is set to 1, otherwise it is set to 0. If the sprite is positioned so part of it
// is outside the coordinates of the display, it wraps around to the opposite side of the screen.
// See instruction 8xy3 for more information on XOR, and section 2.4, Display,
// for more information on the Chip-8 screen and sprites.
func TestDraw(t *testing.T) {

}

// Instruction Ex9E: Skip next instruction if key with the value of Vx is pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the down position, PC is increased by 2.
func TestSkipIfKey(t *testing.T) {
	cpu := &CPU{}

	cpu.Key[0x0] = true
	if cpu.skipIfKey(0x0); cpu.PC != 4 {
		t.Errorf("TestSkipIfKey: failed to properly increment PC. Expected: %d Result: %d", 4, cpu.PC)
	}

	cpu.Key[0x0] = false
	if cpu.skipIfKey(0x0); cpu.PC != 6 {
		t.Errorf("TestSkipIfSky: failed to properly increment PC. Expected: %d Result: %d", 6, cpu.PC)
	}
}

// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the up position, PC is increased by 2.
func TestSkipIfKeyNot(t *testing.T) {
	cpu := &CPU{}

	cpu.Key[0x0] = false
	if cpu.skipIfKeyNot(0x0); cpu.PC != 4 {
		t.Errorf("TestSkipIfKeyNot: failed to properly increment PC. Expected: %d Result: %d", 4, cpu.PC)
	}

	cpu.Key[0x0] = true
	if cpu.skipIfKeyNot(0x0); cpu.PC != 6 {
		t.Errorf("TestSkipIfKeyNot: failed to properly increment PC. Expected: %d Result: %d", 6, cpu.PC)
	}
}

// Instruction Fx07: Set Vx = delay timer value.
// The value of DT is placed into Vx.
func TestLoadXDT(t *testing.T) {
	cpu := &CPU{}
	cpu.DT = 7

	if cpu.loadXDT(0xE); cpu.V[0xE] != 7 {
		t.Errorf("TestLoadXDT: failed to load the delay timer into V%X. Expected: %d Result %d", 0xE, 7, cpu.V[0xE])
	}
}

// Instruction Fx15: Set delay timer = Vx.
// DT is set equal to the value of Vx.
func TestLoadDTX(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0xE] = 7

	if cpu.loadDTX(0xE); cpu.DT != 7 {
		t.Errorf("TestLoadDTX: failed to load V%X into delay timer. Expected: %d  Result %d", 0xE, 7, cpu.DT)
	}
}

// Instruction Fx18: Set sound timer = Vx.
// ST is set equal to the value of Vx.
func TestLoadSTX(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0xE] = 7

	if cpu.loadSTX(0xE); cpu.ST != 7 {
		t.Errorf("TestLoadSTC: failed to load V%X into sound timer. Expected: %d Result: %d", 0xE, 7, cpu.ST)
	}
}

// Instruction Fx1E: Set I = I + Vx.
// The values of I and Vx are added, and the results are stored in I.
func TestAddIX(t *testing.T) {
	cpu := &CPU{}
	cpu.V[0x0] = 7

	if cpu.addIX(0x0); cpu.I != 7 {
		t.Errorf("TestAddIX: failed to add V%X and I. Expected: %d Result: %d", 0x0, 7, cpu.I)
	}
}

// Instruction Fx29: Set I = location of sprite for digit Vx.
// The value of I is set to the location for the hexadecimal sprite corresponding
// to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
func TestLoadIX(t *testing.T) {

}

// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, and I+2.
// The CPU takes the decimal value of Vx, and places the hundreds digit in memory
// at location in I, the tens digit at location I+1, and the ones digit at location I+2.
func TestLoadBCD(t *testing.T) {

}

// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
// The CPU copies the values of registers V0 through Vx into memory,
// starting at the address in I.
func TestSaveV(t *testing.T) {

}

// Instruction Fx65: Read registers V0 through Vx from memory starting at location I.
// The CPU reads values from memory starting at location I into registers V0 through Vx.
func TestLoadV(t *testing.T) {

}
