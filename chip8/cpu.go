package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
)

type Cpu struct {
	RAM   [4096]byte    // CHIP-8 is capable of accessing 4KB (4,096 bytes) of RAM.
	GFX   [64 * 32]byte // CHIP-8 screen is 64x32 pixels.
	Stack [16]uint16    // 16 16-bit stack used for saving addresses before subroutines.

	V [16]byte // 16 8-bit Registers: V0 - VE are general registers and VF is a flag register.

	PC uint16 // 16-bit Program counter. All programs start at 0x200.
	SP uint16 // 16-bit Stack pointer
	I  uint // Address register

	DT byte // Delay timer
	ST byte // Sound timer

	Key [16]uint

	RS int  // ROM Size: length of CHIP-8 program byte array
	DF bool // Draw Flag
}

func (cpu *Cpu) LoadROM(filename *string) error {
	// Read file into byte array
	rom, err := ioutil.ReadFile(*filename)
	if err != nil {
		return err
	}

	// Save ROM size
	cpu.RS = len(rom)

	// Move the PC to 0x200 (512 byte)
	cpu.PC = 0x200

	// Copy program byte array into RAM
	for i, b := range rom {
		cpu.RAM[cpu.PC+uint16(i)] = b
	}

	return nil
}

// Helpful for debugging
func (cpu *Cpu) printRAM() {
	for i := 0; i < cpu.RS+512; i++ {
		fmt.Printf("%d: %X\n", i, cpu.RAM[i])
	}
}

// Helpful for debugging
func (cpu *Cpu) printRegisters() {
	fmt.Printf("\nPC: %d     SP: %d     I: %d\n", cpu.PC, cpu.SP, cpu.I)
	fmt.Printf("Stack: %v\n", cpu.Stack)

	for i := range cpu.V {
		fmt.Printf("V%X: %x\t", i, cpu.V[i])
	}

	fmt.Println()
}

// Each opcode is 2 bytes, but RAM is a byte array, so it must be accessed twice to create the opcode.
//
// RAM[PC] = 0x01 (1 byte)
// RAM[PC + 1] = 0xFE (1 byte)
// opcode = RAM[PC] + RAM[PC + 1] = 0x01FE
func (cpu *Cpu) getOpCode(PC uint16) uint16 {
	opCode1 := uint16(cpu.RAM[PC])
	opCode2 := uint16(cpu.RAM[PC+1])
	opCode := opCode1<<8 | opCode2

	//fmt.Printf("1st OpCode: %X\n2nd OpCode: %X\n", opCode1, opCode2)
	fmt.Printf("OpCode: %X\n", opCode)

	return opCode
}

func (cpu *Cpu) Cycle() {
	// Debug
	cpu.printRegisters()

	// Get opcode
	opCode := cpu.getOpCode(cpu.PC)

	// Execute code
	cpu.execute(opCode)

	// Increment PC
	cpu.PC += 2
}

func (cpu *Cpu) execute(opCode uint16) error {
	vx := byte((opCode & 0x0F00) >> 8)
	vy := byte((opCode & 0x00F0) >> 4)

	nnn := uint16(opCode & 0x0FFF)
	kk := byte(opCode & 0x00FF)
	n := byte(opCode & 0x000F)

	if opCode == 0x00E0 {
		// Instruction 00E0: Clear the display.
		cpu.clear()

	} else if opCode == 0x00EE {
		// Instruction 00EE: Return from a subroutine.
		return cpu.ret()

	} else if (opCode & 0xF000) == 0x1000 {
		// Instruction 1nnn: Jump to location nnn.
		cpu.jump(nnn)

	} else if (opCode & 0xF000) == 0x2000 {
		// Instruction 2nnn: Call subroutine at nnn.
		cpu.call(nnn)

	} else if (opCode & 0xF000) == 0x3000 {
		// Instruction 3xkk: Skip next instructionif Vx = kk.
		cpu.skipIf(vx, kk)

	} else if (opCode & 0xF000) == 0x4000 {
		// Instruction 4xkk: Skip next instruction if Vx != kk.
		cpu.skipIfNot(vx, kk)

	} else if (opCode & 0xF000) == 0x5000 {
		// Instruction 5xy0: Skip next isntruction if Vx = Vy.
		cpu.skipIfXY(vx, vy)

	} else if (opCode & 0xF000) == 0x6000 {
		// Instruction 6xkk: Set Vx = kk.
		cpu.load(vx, kk)

	} else if (opCode & 0xF000) == 0x7000 {
		// Instruction 7xkk: Set Vx = Vx + kk.
		cpu.add(vx, kk)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xy0: Set Vx = Vy.
		cpu.loadXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8001 {
		// Instruction 8xy1: Set Vx = Vx | Vy.
		cpu.orXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8002 {
		// Instruction 8xy2: Set Vx = Vx & Vy.
		cpu.andXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8003 {
		// Instruction 8xy3: Set Vx = Vx ^ Vy.
		cpu.xorXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8004 {
		// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
		cpu.addXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8005 {
		// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
		cpu.subXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8006 {
		// Instruction 8xy6: Set Vx = Vx SHR 1.
		cpu.shiftRight(vx)

	} else if (opCode & 0xF00F) == 0x8007 {
		// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
		cpu.subYX(vx, vy)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xyE: Set Vx = Vx SHL 1.
		cpu.shiftLeft(vx)

	} else if (opCode & 0xF00F) == 0x9000 {
		// Instruction 9xy0: Skip next instruction if Vx != Vy.
		cpu.skipIfNotXY(vx, vy)

	} else if (opCode & 0xF000) == 0xA000 {
		// Instruction Annn: Set I = nnnn.
		cpu.loadI(nnn)

	} else if (opCode & 0xF000) == 0xB000 {
		// Instruction Bnnn: Jump to location nnn + V0.
		cpu.jumpV0(nnn)

	} else if (opCode & 0xF000) == 0xC000 {
		// Instruction Cxkk: Set Vx = random byte AND kk.
		cpu.rand(vx, kk)

	} else if (opCode & 0xF000) == 0xD000 {
		// Instruction Dxyn: Display nbyte sprite starting at memory
		// location I at (Vx, Vy), set Vf = collusion.
		cpu.draw(vx, vy, n)

	} else if (opCode & 0xF0FF) == 0xE09E {
		// Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.
		cpu.skipIfKey(vx)

	} else if (opCode & 0xF0FF) == 0xE0A1 {
		// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
		cpu.skipIfKeyNot(vx)

	} else if (opCode & 0xF0FF) == 0xF007 {
		// Instruction Fx07: Set Vx = delay timer value.
		cpu.loadXDT(vx)

	} else if (opCode & 0xF0FF) == 0xF00A {
		// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
		cpu.loadKey(vx)

	} else if (opCode & 0xF0FF) == 0xF015 {
		// Instruction Fx15: Set delay timer = Vx.
		cpu.loadDTX(vx)

	} else if (opCode & 0xF0FF) == 0xF018 {
		// Instruction Fx18: Set sounder timer = Vx.
		cpu.loadSTX(vx)

	} else if (opCode & 0xF0FF) == 0xF01E {
		// Instruction Fx1E : Set I = I + Vx.
		cpu.addIX(vx)

	} else if (opCode & 0xF0FF) == 0xF029 {
		// Instruction Fx29: Set I = location of sprite for digit Vx.
		cpu.loadIX(vx)

	} else if (opCode & 0xF0FF) == 0xF033 {
		// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, I+2.
		cpu.loadBCD(vx)

	} else if (opCode & 0xF0FF) == 0xF055 {
		// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
		cpu.saveV(vx)

	} else if (opCode & 0xF0FF) == 0xF065 {
		// Instruction Fx65: Read registers V0 through Vx in memory starting at location I.
		cpu.loadV(vx)

	} else {
		fmt.Errorf("Unknown instruction: %X\n\n", opCode)
	}

	return nil
}

// Instruction 00E0: Clear the display.
func (cpu *Cpu) clear() {
	fmt.Println("Instruction 00E0: Clear the display.")

	// Zero out gfx
	for i := range cpu.GFX {
		cpu.GFX[i] = 0
	}

	// Set draw flag
	cpu.DF = true
}

// Instruction 00EE: Return from a subroutine.
// The interpreter sets the program counter to the address at the top of the stack,
// then subtracts 1 from the stack pointer.
func (cpu *Cpu) ret() error {
	fmt.Println("Instruction 00EE: Return from a subroutine.")

	// Decrement stack pointer and error if it's below 0.
	if cpu.SP -= 1; cpu.SP < 0 {
		return fmt.Errorf("stack pointer out of bounds: %d", cpu.SP)
	}

	cpu.PC = cpu.Stack[cpu.SP]

	fmt.Printf("New PC: %d", cpu.PC)
	return nil
}

// Instruction 1nnn: Jump to location nnn.
// The interpreter sets the program counter to nnn.
func (cpu *Cpu) jump(nnn uint16) {
	fmt.Println("Instruction 1nnn: Jump to location nnn.")
	fmt.Printf("nnn: %d\n", nnn)

	cpu.PC = nnn

	fmt.Printf("New PC: %d\n", cpu.PC)
}

// Instruction 2nnn: Call subroutine at nnn.
// The interpreter increments the stack pointer, then puts the current PC on the top of the stack.
// The PC is then set to nnn.
func (cpu *Cpu) call(nnn uint16) error {
	fmt.Println("Instruction 2nnn: Call subroutine at nnn.")
	fmt.Printf("nnn: %d\n", nnn)

	cpu.Stack[cpu.SP] = cpu.PC
	cpu.PC = nnn

	// Increment stack pointer and error if it's above it's length
	if cpu.SP += 1; cpu.SP > uint16(len(cpu.Stack)) {
		fmt.Errorf("stack pointer out of points: %d", cpu.SP)
	}

	fmt.Printf("New Stack: %v\nnew SP: %d\tPC: %d\n", cpu.Stack, cpu.SP, cpu.PC)
	return nil
}

// Instruction 3xkk: Skip next instruction if Vx = kk.
// The interpreter compares register Vx to kk, and if they are equal,
// increments the program counter by 2.
func (cpu *Cpu) skipIf(vx byte, kk byte) {
	fmt.Println("Instruction 3xkk: Skip next instruction if Vx == kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	if cpu.V[vx] == kk {
		cpu.PC += 2
	}

	fmt.Printf("New PC: %d\n", cpu.PC)
}

// Instruction 4xkk: Skip next instruction if Vx != kk.
// The interpreter compares register Vx to kk, and if they are not equal,
// increments the program counter by 2.
func (cpu *Cpu) skipIfNot(vx byte, kk byte) {
	fmt.Println("Instruction 4xkk: Skip next instruction if Vx != kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	if cpu.V[vx] != kk {
		cpu.PC += 2
	}

	fmt.Printf("New PC: %d\n", cpu.PC)
}

// Instruction 5xy0: Skip next instruction if Vx = Vy.
// The interpreter compares register Vx to register Vy, and if they are equal,
// increments the program counter by 2.
func (cpu *Cpu) skipIfXY(vx byte, vy byte) {
	fmt.Println("Instruction 5xy0: Skip next isntruction if Vx = Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if vx == vy {
		cpu.PC += 2
	}

	fmt.Printf("New PC: %d\n", cpu.PC)
}

// Instruction 6xkk: Set Vx = kk.
// The interpreter puts the value kk into register Vx.
func (cpu *Cpu) load(vx byte, kk byte) {
	fmt.Println("Instruction 6xkk: Set Vx = kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	cpu.V[vx] = kk

	fmt.Printf("New V%X: %X\n", vx, cpu.V[vx])
}

// Instruction 7xkk: Set Vx = Vx + kk.
// Adds the value kk to the value of register Vx, then stores the result in Vx.
func (cpu *Cpu) add(vx byte, kk byte) {
	fmt.Println("Instruction 7xkk: Set Vx = Vx + kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	cpu.V[vx] += kk

	fmt.Printf("New V%X: %X\n", vx, cpu.V[vx])
}

// Instruction 8xy0: Set Vx = Vy.
// Stores the value of register Vy in register Vx.
func (cpu *Cpu) loadXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy0: Set Vx = Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] = cpu.V[vy]

	fmt.Printf("New V%X: %X\n", vx, cpu.V[vx])
}

// Instruction 8xy1: Set Vx = Vx OR Vy.
// Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
// A bitwise OR compares the corrseponding bits from two values, and if either bit is 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (cpu *Cpu) orXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy1: Set Vx = Vx | Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] = cpu.V[vx] | cpu.V[vy]

	fmt.Printf("New V%X: %X", vx, cpu.V[vx])
}

// Instruction 8xy2: Set Vx = Vx AND Vy.
// Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
// A bitwise AND compares the corrseponding bits from two values, and if both bits are 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (cpu *Cpu) andXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy2: Set Vx = Vx & Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] = cpu.V[vx] & cpu.V[vy]

	fmt.Printf("New V%X: %X", vx, cpu.V[vx])
}

// Instruction 8xy3: Set Vx = Vx XOR Vy.
// Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx.
// An exclusive OR compares the corrseponding bits from two values,
// and if the bits are not both the same, then the corresponding bit in the result is set to 1.
// Otherwise, it is 0.
func (cpu *Cpu) xorXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy3: Set Vx = Vx ^ Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] = cpu.V[vx] ^ cpu.V[vy]

	fmt.Printf("New V%X: %X", vx, cpu.V[vx])
}

// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
// VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
func (cpu *Cpu) addXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] = cpu.V[vx] + cpu.V[vy]

	fmt.Printf("New V%X: %X", vx, cpu.V[vx])
}

// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
// and the results stored in Vx.
func (cpu *Cpu) subXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] = cpu.V[vx] + cpu.V[vy]

	fmt.Printf("New V%X: %X", vx, cpu.V[vx])
}

// Instruction 8xy6: Set Vx = Vx SHR 1.
// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
// Then Vx is divided by 2.
func (cpu *Cpu) shiftRight(vx byte) {
	fmt.Println("Instruction 8xy6: Set Vx = Vx SHR 1.")
	fmt.Printf("Vx: %X\n", vx)

	cpu.V[0xF] = cpu.V[vx] & 0x1

	// Another way to divide by 2
	cpu.V[vx] = cpu.V[vx] >> 1

	fmt.Printf("New V%X: %X\tVF: %X", vx, cpu.V[vx], cpu.V[0xF])
}

// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
// If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy,
// and the results stored in Vx.
func (cpu *Cpu) subYX(vx byte, vy byte) {
	fmt.Println("Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if vy > vx {
		cpu.V[0xF] = 1
	} else {
		cpu.V[0xF] = 0
	}

	cpu.V[vx] = cpu.V[vy] - cpu.V[vx]

	fmt.Printf("New V%X: %d\tVF: %d\n", vx, cpu.V[vx], cpu.V[0xF])
}

// Instruction 8xyE: Set Vx = Vx SHL 1.
// If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
// Then Vx is multiplied by 2.
func (cpu *Cpu) shiftLeft(vx byte) {
	fmt.Println("Instruction 8xyE: Set Vx = Vx SHL 1.")
	fmt.Printf("Vx: %X\n", vx)

	// Get the most significant bit in a byte
	cpu.V[0xF] = cpu.V[vx] >> 7

	// Multiple by 2
	cpu.V[vx] = cpu.V[vx] << 1

	fmt.Printf("New V%X: %d\tVF: %d\n", vx, cpu.V[vx], cpu.V[0xF])
}

// Instruction 9xy0: Skip next instruction if Vx != Vy.
// The values of Vx and Vy are compared, and if they are not equal,
// the program counter is increased by 2.
func (cpu *Cpu) skipIfNotXY(vx byte, vy byte) {
	fmt.Println("Instruction 9xy0: Skip next instruction if Vx != Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if cpu.V[vx] != cpu.V[vy] {
		cpu.PC += 2
	}

	fmt.Printf("New PC: %d\n", cpu.PC)
}

// Instruction Annn: Set I = nnn.
// The value of register I is set to nnn.
func (cpu *Cpu) loadI(nnn uint16) {
	fmt.Println("Instruction Annn: Set I = nnnn.")
	fmt.Printf("nnn: %X\n", nnn)

	cpu.I = uint(nnn)

	fmt.Printf("New I: %X", cpu.I)
}

// Instruction Bnnn: Jump to location nnn + V0.
// The program counter is set to nnn plus the value of V0.
func (cpu *Cpu) jumpV0(nnn uint16) {
	fmt.Println("Instruction Bnnn: Jump to location nnn + V0.")
	fmt.Printf("nnn: %X\n", nnn)

	cpu.PC = uint16(cpu.V[0x0]) + nnn

	fmt.Printf("New PC: %d\n", cpu.PC)
}

// Instruction Cxkk: Set Vx = random byte AND kk.
// The interpreter generates a random number from 0 to 255,
// which is then ANDed with the value kk. The results are stored in Vx.
// See instruction 8xy2 for more information on AND.
func (cpu *Cpu) rand(vx byte, kk byte) {
	fmt.Println("Instruction Cxkk: Set Vx = random byte AND kk.")
	fmt.Printf("Vx: %X\n", vx)

	r := byte(rand.Intn(256))
	cpu.V[vx] = kk & r

	fmt.Printf("New V%X: %X", vx, cpu.V[vx])
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
func (cpu *Cpu) draw(vx byte, vy byte, n byte) {
	fmt.Println("Instruction Dxyn: Display nbyte sprite starting at memory location I at (Vx, Vy), " +
		"set Vf = collusion.")
	fmt.Printf("Vx: %X\tVy: %X\tn: %X\n", vx, vy, n)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Ex9E: Skip next instruction if key with the value of Vx is pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the down position, PC is increased by 2.
func (cpu *Cpu) skipIfKey(vx byte) {
	fmt.Println("Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the up position, PC is increased by 2.
func (cpu *Cpu) skipIfKeyNot(vx byte) {
	fmt.Println("Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx07: Set Vx = delay timer value.
// The value of DT is placed into Vx.
func (cpu *Cpu) loadXDT(vx byte) {
	fmt.Println("Instruction Fx07: Set Vx = delay timer value.")
	fmt.Printf("Vx: %X\n", vx)

	cpu.V[vx] = cpu.DT
	fmt.Printf("New V%X: %X", vx, cpu.V[vx])
}

// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
// All execution stops until a key is pressed, then the value of that key is stored in Vx.
func (cpu *Cpu) loadKey(vx byte) {
	fmt.Println("Instruction Fx0A: Wait for a key press, store the value of the key in Vx.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx15: Set delay timer = Vx.
// DT is set equal to the value of Vx.
func (cpu *Cpu) loadDTX(vx byte) {
	fmt.Println("Instruction Fx15: Set delay timer = Vx.")
	fmt.Printf("Vx: %X\n", vx)

	cpu.DT = cpu.V[vx]

	fmt.Printf("New DT: %d", cpu.DT)
}

// Instruction Fx18: Set sound timer = Vx.
// ST is set equal to the value of Vx.
func (cpu *Cpu) loadSTX(vx byte) {
	fmt.Println("Instruction Fx18: Set sounder timer = Vx.")
	fmt.Printf("Vx: %X\n", vx)

	cpu.ST = cpu.V[vx]

	fmt.Printf("New ST: %d", cpu.ST)
}

// Instruction Fx1E: Set I = I + Vx.
// The values of I and Vx are added, and the results are stored in I.
func (cpu *Cpu) addIX(vx byte) {
	fmt.Println("Instruction Fx1E : Set I = I + Vx.")
	fmt.Printf("Vx: %X\n", vx)

	cpu.I = cpu.I + uint(cpu.V[vx])

	fmt.Printf("New I: %X", cpu.I)
}

// Instruction Fx29: Set I = location of sprite for digit Vx.
// The value of I is set to the location for the hexadecimal sprite corresponding
// to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
func (cpu *Cpu) loadIX(vx byte) {
	fmt.Println("Instruction Fx29: Set I = location of sprite for digit Vx.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, and I+2.
// The interpreter takes the decimal value of Vx, and places the hundreds digit in memory
// at location in I, the tens digit at location I+1, and the ones digit at location I+2.
func (cpu *Cpu) loadBCD(vx byte) {
	fmt.Println("Instruction Fx33: Store BCD represention of Vx in memory locations I, I+1, I+2.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
// The interpreter copies the values of registers V0 through Vx into memory,
// starting at the address in I.
func (cpu *Cpu) saveV(vx byte) {
	fmt.Println("Instruction Fx55: Store registers V0 through Vx in memory starting at location I.")
	fmt.Printf("Vx: %X\n", vx)

	for i:= uint(0); i <= uint(vx); i++ {
		cpu.RAM[cpu.I + i] = cpu.V[i]
	}

	fmt.Printf("New ")
	for i:= uint(0); i <= uint(vx); i++ {
		fmt.Printf("I+%d: %X", i, cpu.RAM[cpu.I + i])
	}
	fmt.Println()
}

// Instruction Fx65: Read registers V0 through Vx from memory starting at location I.
// The interpreter reads values from memory starting at location I into registers V0 through Vx.
func (cpu *Cpu) loadV(vx byte) {
	fmt.Println("Instruction Fx65: Read registers V0 through Vx in memory starting at location I.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")

	fmt.Printf("New ")
	for i := range cpu.V {
		fmt.Printf("V%X: %x\t", i, cpu.V[i])
	}
	fmt.Println()
}
