package main

import (
	"fmt"
	"io/ioutil"
)

type CHIP8 struct {
	RAM   [4096]byte   // CHIP-8 is capable of accessing 4KB (4,096 bytes) of RAM.
	GFX   [64][32]byte // CHIP-8 screen is 64x32 pixels.
	Stack [16]uint     // 16 16-bit stack used for saving addresses before subroutines.

	V [16]byte // 16 8-bit Registers: V0 - VE are general registers and VF is a flag register.

	PC uint // 16-bit Program counter. All programs start at 0x200.
	SP uint // 8-bit Stack pointer
	I  uint // Address register

	DT int64 // Delay timer
	ST int64 // Sound timer

	Key [16]uint

	RS int  // ROM Size: length of CHIP-8 program byte array
	DF uint // Draw Flag
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
	for i, b := range rom {
		chip8.RAM[uint(i)+chip8.PC] = b
	}

	return nil
}

// Helpful for debugging
func (chip8 *CHIP8) printRAM() {
	for i := 0; i < chip8.RS+512; i++ {
		fmt.Printf("%d: %X\n", i, chip8.RAM[i])
	}
}

// Each opcode is 2 bytes, but RAM is a byte array, so it must be accessed twice to create the opcode.
//
// RAM[PC] = 0x01 (1 byte)
// RAM[PC + 1] = 0xFE (1 byte)
// opcode = RAM[PC] + RAM[PC + 1] = 0x01FE
func (chip8 *CHIP8) getOpCode(PC uint) uint16 {
	opCode1 := uint16(chip8.RAM[PC])
	opCode2 := uint16(chip8.RAM[PC+1])
	opCode := opCode1<<8 | opCode2

	//fmt.Printf("1st OpCode: %X\n2nd OpCode: %X\n", opCode1, opCode2)
	fmt.Printf("OpCode: %X\n", opCode)

	return opCode
}

func (chip8 *CHIP8) Cycle() {
	fmt.Printf("\nPC: %d\n", chip8.PC)
	// Get opcode and increment PC twice
	opCode := chip8.getOpCode(chip8.PC)
	chip8.PC += 2

	// Execute code
	chip8.execute(opCode)
}

func (chip8 *CHIP8) execute(opCode uint16) error {
	vx := byte((opCode & 0x0F00) >> 8)
	vy := byte((opCode & 0x00F0) >> 4)

	nnn := uint16(opCode & 0x0FFF)
	kk := byte(opCode & 0x00FF)
	n := byte(opCode & 0x000F)

	if opCode == 0x00E0 {
		// Instruction 00E0: Clear the display.
		chip8.clear()

	} else if opCode == 0x00EE {
		// Instruction 00EE: Return from a subroutine.
		chip8.ret()

	} else if (opCode & 0xF000) == 0x1000 {
		// Instruction 1nnn: Jump to location nnn.
		chip8.jump(nnn)

	} else if (opCode & 0xF000) == 0x2000 {
		// Instruction 2nnn: Call subroutine at nnn.
		chip8.call(nnn)

	} else if (opCode & 0xF000) == 0x3000 {
		// Instruction 3xkk: Skip next instructionif Vx = kk.
		chip8.skipIf(vx, kk)

	} else if (opCode & 0xF000) == 0x4000 {
		// Instruction 4xkk: Skip next instruction if Vx != kk.
		chip8.skipIfNot(vx, kk)

	} else if (opCode & 0xF000) == 0x5000 {
		// Instruction 5xy0: Skip next isntruction if Vx = Vy.
		chip8.skipIfXY(vx, vy)

	} else if (opCode & 0xF000) == 0x6000 {
		// Instruction 6xkk: Set Vx = kk.
		chip8.load(vx, kk)

	} else if (opCode & 0xF000) == 0x7000 {
		// Instruction 7xkk: Set Vx = Vx + kk.
		chip8.add(vx, kk)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xy0: Set Vx = Vy.
		chip8.loadXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8001 {
		// Instruction 8xy1: Set Vx = Vx | Vy.
		chip8.orXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8002 {
		// Instruction 8xy2: Set Vx = Vx & Vy.
		chip8.andXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8003 {
		// Instruction 8xy3: Set Vx = Vx ^ Vy.
		chip8.xorXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8004 {
		// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
		chip8.addXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8005 {
		// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
		chip8.subXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8006 {
		// Instruction 8xy6: Set Vx = Vx SHR 1.
		chip8.shiftRight(vx)

	} else if (opCode & 0xF00F) == 0x8007 {
		// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
		chip8.subnXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xyE: Set Vx = Vx SHL 1.
		chip8.shiftLeft(vx)

	} else if (opCode & 0xF00F) == 0x9000 {
		// Instruction 9xy0: Skip next instruction if Vx != Vy.
		chip8.skipIfNotXY(vx, vy)

	} else if (opCode & 0xF000) == 0xA000 {
		// Instruction Annn: Set I = nnnn.
		chip8.loadI(nnn)

	} else if (opCode & 0xF000) == 0xB000 {
		// Instruction Bnnn: Jump to location nnn + V0.
		chip8.jumpV0(nnn)

	} else if (opCode & 0xF000) == 0xC000 {
		// Instruction Cxkk: Set Vx = random byte AND kk.
		chip8.rand(vx)

	} else if (opCode & 0xF000) == 0xD000 {
		// Instruction Dxyn: Display nbyte sprite starting at memory
		// location I at (Vx, Vy), set Vf = collusion.
		chip8.draw(vx, vy, n)

	} else if (opCode & 0xF0FF) == 0xE09E {
		// Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.
		chip8.skipIfKey(vx)

	} else if (opCode & 0xF0FF) == 0xE0A1 {
		// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
		chip8.skipIfKeyNot(vx)

	} else if (opCode & 0xF0FF) == 0xF007 {
		// Instruction Fx07: Set Vx = delay timer value.
		chip8.loadXDT(vx)

	} else if (opCode & 0xF0FF) == 0xF00A {
		// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
		chip8.loadKey(vx)

	} else if (opCode & 0xF0FF) == 0xF015 {
		// Instruction Fx15: Set delay timer = Vx.
		chip8.loadDTX(vx)

	} else if (opCode & 0xF0FF) == 0xF018 {
		// Instruction Fx18: Set sounder timer = Vx.
		chip8.loadSTX(vx)

	} else if (opCode & 0xF0FF) == 0xF01E {
		// Instruction Fx1E : Set I = I + Vx.
		chip8.addIX(vx)

	} else if (opCode & 0xF0FF) == 0xF029 {
		// Instruction Fx29: Set I = location of sprite for digit Vx.
		chip8.loadIX(vx)

	} else if (opCode & 0xF0FF) == 0xF033 {
		// Instruction Fx33: Store BCD represention of Vx in memory locations I, I+1, I+2.
		chip8.loadBCD(vx)

	} else if (opCode & 0xF0FF) == 0xF055 {
		// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
		chip8.saveV(vx)

	} else if (opCode & 0xF0FF) == 0xF065 {
		// Instruction Fx65: Read registers V0 through Vx in memory starting at location I.
		chip8.loadV(vx)

	} else {
		fmt.Errorf("Unknown instruction: %X\n\n", opCode)
	}

	return nil
}

// Instruction 00E0: Clear the display.
func (chip8 *CHIP8) clear() {
	fmt.Println("Instruction 00E0: Clear the display.")

}

// Instruction 00EE: Return from a subroutine.
// The interpreter sets the program counter to the address at the top of the stack,
// then subtracts 1 from the stack pointer.
func (chip8 *CHIP8) ret() {
	fmt.Println("Instruction 00EE: Return from a subroutine.")
}

// Instruction 1nnn: Jump to location nnn.
// The interpreter sets the program counter to nnn.
func (chip8 *CHIP8) jump(nnn uint16) {
	fmt.Println("Instruction 1nnn: Jump to location nnn.")
	fmt.Printf("nnn: %X\n", nnn)
}

// Instruction 2nnn: Call subroutine at nnn.
// The interpreter increments the stack pointer, then puts the current PC on the top of the stack.
// The PC is then set to nnn.
func (chip8 *CHIP8) call(nnn uint16) {
	fmt.Println("Instruction 2nnn: Call subroutine at nnn.")
	fmt.Printf("nnn: %X\n", nnn)
}

// Instruction 3xkk: Skip next instruction if Vx = kk.
// The interpreter compares register Vx to kk, and if they are equal,
// increments the program counter by 2.
func (chip8 *CHIP8) skipIf(vx byte, kk byte) {
	fmt.Println("Instruction 3xkk: Skip next instructionif Vx = kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)
}

// Instruction 4xkk: Skip next instruction if Vx != kk.
// The interpreter compares register Vx to kk, and if they are not equal,
// increments the program counter by 2.
func (chip8 *CHIP8) skipIfNot(vx byte, kk byte) {
	fmt.Println("Instruction 4xkk: Skip next instruction if Vx != kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)
}

// Instruction 5xy0: Skip next instruction if Vx = Vy.
// The interpreter compares register Vx to register Vy, and if they are equal,
// increments the program counter by 2.
func (chip8 *CHIP8) skipIfXY(vx byte, vy byte) {
	fmt.Println("Instruction 5xy0: Skip next isntruction if Vx = Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction 6xkk: Set Vx = kk.
// The interpreter puts the value kk into register Vx.
func (chip8 *CHIP8) load(vx byte, kk byte) {
	fmt.Println("Instruction 6xkk: Set Vx = kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)
}

// Instruction 7xkk: Set Vx = Vx + kk.
// Adds the value kk to the value of register Vx, then stores the result in Vx.
func (chip8 *CHIP8) add(vx byte, kk byte) {
	fmt.Println("Instruction 7xkk: Set Vx = Vx + kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)
}

// Instruction 8xy0: Set Vx = Vy.
// Stores the value of register Vy in register Vx.
func (chip8 *CHIP8) loadXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy0: Set Vx = Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction 8xy1: Set Vx = Vx OR Vy.
// Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
// A bitwise OR compares the corrseponding bits from two values, and if either bit is 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (chip8 *CHIP8) orXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy1: Set Vx = Vx | Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction 8xy2: Set Vx = Vx AND Vy.
// Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
// A bitwise AND compares the corrseponding bits from two values, and if both bits are 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (chip8 *CHIP8) andXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy2: Set Vx = Vx & Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction 8xy3: Set Vx = Vx XOR Vy.
// Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx.
// An exclusive OR compares the corrseponding bits from two values,
// and if the bits are not both the same, then the corresponding bit in the result is set to 1.
// Otherwise, it is 0.
func (chip8 *CHIP8) xorXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy3: Set Vx = Vx ^ Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
// VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
func (chip8 *CHIP8) addXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
// and the results stored in Vx.
func (chip8 *CHIP8) subXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction 8xy6: Set Vx = Vx SHR 1.
// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
// Then Vx is divided by 2.
func (chip8 *CHIP8) shiftRight(vx byte) {
	fmt.Println("Instruction 8xy6: Set Vx = Vx SHR 1.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
// If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy,
// and the results stored in Vx.
func (chip8 *CHIP8) subnXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction 8xyE: Set Vx = Vx SHL 1.
// If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
// Then Vx is multiplied by 2.
func (chip8 *CHIP8) shiftLeft(vx byte) {
	fmt.Println("Instruction 8xyE: Set Vx = Vx SHL 1.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction 9xy0: Skip next instruction if Vx != Vy.
// The values of Vx and Vy are compared, and if they are not equal,
// the program counter is increased by 2.
func (chip8 *CHIP8) skipIfNotXY(vx byte, vy byte) {
	fmt.Println("Instruction 9xy0: Skip next instruction if Vx != Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)
}

// Instruction Annn: Set I = nnn.
// The value of register I is set to nnn.
func (chip8 *CHIP8) loadI(nnn uint16) {
	fmt.Println("Instruction Annn: Set I = nnnn.")
	fmt.Printf("nnn: %X\n", nnn)

}

// Instruction Bnnn: Jump to location nnn + V0.
// The program counter is set to nnn plus the value of V0.
func (chip8 *CHIP8) jumpV0(nnn uint16) {
	fmt.Println("Instruction Bnnn: Jump to location nnn + V0.")
	fmt.Printf("nnn: %X\n", nnn)
}

// Instruction Cxkk: Set Vx = random byte AND kk.
// The interpreter generates a random number from 0 to 255,
// which is then ANDed with the value kk. The results are stored in Vx.
// See instruction 8xy2 for more information on AND.
func (chip8 *CHIP8) rand(vx byte) {
	fmt.Println("Instruction Cxkk: Set Vx = random byte AND kk.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Dxyn: Display n-byte sprite starting at memory location I at (Vx, Vy),
// set VF = collision.
//
// The interpreter reads n bytes from memory, starting at the address stored in I.
// These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
// Sprites are XORed onto the existing screen. If this causes any pixels to be erased,
// VF is set to 1, otherwise it is set to 0. If the sprite is positioned so part of it
// is outside the coordinates of the display, it wraps around to the opposite side of the screen.
// See instruction 8xy3 for more information on XOR, and section 2.4, Display,
// for more information on the Chip-8 screen and sprites.
func (chip8 *CHIP8) draw(vx byte, vy byte, n byte) {
	fmt.Println("Instruction Dxyn: Display nbyte sprite starting at memory location I at (Vx, Vy), " +
		"set Vf = collusion.")
	fmt.Printf("Vx: %X\tVy: %X\tn: %X\n", vx, vy, n)
}

// Instruction Ex9E: Skip next instruction if key with the value of Vx is pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the down position, PC is increased by 2.
func (chip8 *CHIP8) skipIfKey(vx byte) {
	fmt.Println("Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the up position, PC is increased by 2.
func (chip8 *CHIP8) skipIfKeyNot(vx byte) {
	fmt.Println("Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx07: Set Vx = delay timer value.
// The value of DT is placed into Vx.
func (chip8 *CHIP8) loadXDT(vx byte) {
	fmt.Println("Instruction Fx07: Set Vx = delay timer value.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
// All execution stops until a key is pressed, then the value of that key is stored in Vx.
func (chip8 *CHIP8) loadKey(vx byte) {
	fmt.Println("Instruction Fx0A: Wait for a key press, store the value of the key in Vx.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx15: Set delay timer = Vx.
// DT is set equal to the value of Vx.
func (chip8 *CHIP8) loadDTX(vx byte) {
	fmt.Println("Instruction Fx15: Set delay timer = Vx.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx18: Set sound timer = Vx.
// ST is set equal to the value of Vx.
func (chip8 *CHIP8) loadSTX(vx byte) {
	fmt.Println("Instruction Fx18: Set sounder timer = Vx.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx1E: Set I = I + Vx.
// The values of I and Vx are added, and the results are stored in I.
func (chip8 *CHIP8) addIX(vx byte) {
	fmt.Println("Instruction Fx1E : Set I = I + Vx.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx29: Set I = location of sprite for digit Vx.
// The value of I is set to the location for the hexadecimal sprite corresponding
// to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
func (chip8 *CHIP8) loadIX(vx byte) {
	fmt.Println("Instruction Fx29: Set I = location of sprite for digit Vx.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, and I+2.
// The interpreter takes the decimal value of Vx, and places the hundreds digit in memory
// at location in I, the tens digit at location I+1, and the ones digit at location I+2.
func (chip8 *CHIP8) loadBCD(vx byte) {
	fmt.Println("Instruction Fx33: Store BCD represention of Vx in memory locations I, I+1, I+2.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
// The interpreter copies the values of registers V0 through Vx into memory,
// starting at the address in I.
func (chip8 *CHIP8) saveV(vx byte) {
	fmt.Println("Instruction Fx55: Store registers V0 through Vx in memory starting at location I.")
	fmt.Printf("Vx: %X\n", vx)
}

// Instruction Fx65: Read registers V0 through Vx from memory starting at location I.
// The interpreter reads values from memory starting at location I into registers V0 through Vx.
func (chip8 *CHIP8) loadV(vx byte) {
	fmt.Println("Instruction Fx65: Read registers V0 through Vx in memory starting at location I.")
	fmt.Printf("Vx: %X\n", vx)
}