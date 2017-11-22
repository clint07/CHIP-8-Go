package CHIP8

import (
	"fmt"
	"io/ioutil"
	"math/rand"
)

type VM struct {
	RAM   [4096]byte    // CHIP-8 is capable of accessing 4KB (4,096 bytes) of RAM.
	GFX   [64 * 32]byte // CHIP-8 screen is 64x32 pixels.
	Stack [16]uint16    // 16 16-bit stack used for saving addresses before subroutines.

	V [16]byte // 16 8-bit Registers: V0 - VE are general registers and VF is a flag register.

	PC uint16 // 16-bit Program counter. All programs start at 0x200.
	SP uint16 // 16-bit Stack pointer
	I  uint   // Address register

	DT byte // Delay timer
	ST byte // Sound timer

	Key [16]uint

	RS int  // ROM Size: length of CHIP-8 program byte array
	DF bool // Draw Flag
}

func (this *VM) Init() {
	fmt.Println("TODO: LOAD FONTS")
}

func (this *VM) Load(filename *string) error {
	// Read file into byte array
	rom, err := ioutil.ReadFile(*filename)
	if err != nil {
		return err
	}

	// Save ROM size
	this.RS = len(rom)

	// Move the PC to 0x200 (512 byte)
	this.PC = 0x200

	// Copy program byte array into RAM
	for i, b := range rom {
		this.RAM[this.PC+uint16(i)] = b
	}

	return nil
}

// Helpful for debugging
func (this *VM) printRAM() {
	for i := 0; i < this.RS+512; i++ {
		fmt.Printf("%d: %X\n", i, this.RAM[i])
	}
}

// Helpful for debugging
func (this *VM) printRegisters() {
	fmt.Printf("\nPC: %d     SP: %d     I: %d\n", this.PC, this.SP, this.I)
	fmt.Printf("Stack: %v\n", this.Stack)

	for i := range this.V {
		fmt.Printf("V%X: %x\t", i, this.V[i])
	}

	fmt.Println()
}

// Each opcode is 2 bytes, but RAM is a byte array, so it must be accessed twice to create the opcode.
//
// RAM[PC] = 0x01 (1 byte)
// RAM[PC + 1] = 0xFE (1 byte)
// opcode = RAM[PC] + RAM[PC + 1] = 0x01FE
func (this *VM) getOpCode(PC uint16) uint16 {
	opCode1 := uint16(this.RAM[PC])
	opCode2 := uint16(this.RAM[PC+1])
	opCode := opCode1<<8 | opCode2

	//fmt.Printf("1st OpCode: %X\n2nd OpCode: %X\n", opCode1, opCode2)
	fmt.Printf("OpCode: %X\n", opCode)

	return opCode
}

func (this *VM) Cycle() {
	// Debug
	this.printRegisters()

	// Get opcode
	opCode := this.getOpCode(this.PC)

	// Execute code
	this.execute(opCode)

	// Increment PC
	this.PC += 2
}

func (this *VM) execute(opCode uint16) error {
	vx := byte((opCode & 0x0F00) >> 8)
	vy := byte((opCode & 0x00F0) >> 4)

	nnn := uint16(opCode & 0x0FFF)
	kk := byte(opCode & 0x00FF)
	n := byte(opCode & 0x000F)

	if opCode == 0x00E0 {
		// Instruction 00E0: Clear the display.
		this.clear()

	} else if opCode == 0x00EE {
		// Instruction 00EE: Return from a subroutine.
		return this.ret()

	} else if (opCode & 0xF000) == 0x1000 {
		// Instruction 1nnn: Jump to location nnn.
		this.jump(nnn)

	} else if (opCode & 0xF000) == 0x2000 {
		// Instruction 2nnn: Call subroutine at nnn.
		this.call(nnn)

	} else if (opCode & 0xF000) == 0x3000 {
		// Instruction 3xkk: Skip next instructionif Vx = kk.
		this.skipIf(vx, kk)

	} else if (opCode & 0xF000) == 0x4000 {
		// Instruction 4xkk: Skip next instruction if Vx != kk.
		this.skipIfNot(vx, kk)

	} else if (opCode & 0xF000) == 0x5000 {
		// Instruction 5xy0: Skip next isntruction if Vx = Vy.
		this.skipIfXY(vx, vy)

	} else if (opCode & 0xF000) == 0x6000 {
		// Instruction 6xkk: Set Vx = kk.
		this.load(vx, kk)

	} else if (opCode & 0xF000) == 0x7000 {
		// Instruction 7xkk: Set Vx = Vx + kk.
		this.add(vx, kk)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xy0: Set Vx = Vy.
		this.loadXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8001 {
		// Instruction 8xy1: Set Vx = Vx | Vy.
		this.orXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8002 {
		// Instruction 8xy2: Set Vx = Vx & Vy.
		this.andXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8003 {
		// Instruction 8xy3: Set Vx = Vx ^ Vy.
		this.xorXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8004 {
		// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
		this.addXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8005 {
		// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
		this.subXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8006 {
		// Instruction 8xy6: Set Vx = Vx SHR 1.
		this.shiftRight(vx)

	} else if (opCode & 0xF00F) == 0x8007 {
		// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
		this.subYX(vx, vy)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xyE: Set Vx = Vx SHL 1.
		this.shiftLeft(vx)

	} else if (opCode & 0xF00F) == 0x9000 {
		// Instruction 9xy0: Skip next instruction if Vx != Vy.
		this.skipIfNotXY(vx, vy)

	} else if (opCode & 0xF000) == 0xA000 {
		// Instruction Annn: Set I = nnnn.
		this.loadI(nnn)

	} else if (opCode & 0xF000) == 0xB000 {
		// Instruction Bnnn: Jump to location nnn + V0.
		this.jumpV0(nnn)

	} else if (opCode & 0xF000) == 0xC000 {
		// Instruction Cxkk: Set Vx = random byte AND kk.
		this.rand(vx, kk)

	} else if (opCode & 0xF000) == 0xD000 {
		// Instruction Dxyn: Display nbyte sprite starting at memory
		// location I at (Vx, Vy), set Vf = collusion.
		this.draw(vx, vy, n)

	} else if (opCode & 0xF0FF) == 0xE09E {
		// Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.
		this.skipIfKey(vx)

	} else if (opCode & 0xF0FF) == 0xE0A1 {
		// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
		this.skipIfKeyNot(vx)

	} else if (opCode & 0xF0FF) == 0xF007 {
		// Instruction Fx07: Set Vx = delay timer value.
		this.loadXDT(vx)

	} else if (opCode & 0xF0FF) == 0xF00A {
		// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
		this.loadKey(vx)

	} else if (opCode & 0xF0FF) == 0xF015 {
		// Instruction Fx15: Set delay timer = Vx.
		this.loadDTX(vx)

	} else if (opCode & 0xF0FF) == 0xF018 {
		// Instruction Fx18: Set sounder timer = Vx.
		this.loadSTX(vx)

	} else if (opCode & 0xF0FF) == 0xF01E {
		// Instruction Fx1E : Set I = I + Vx.
		this.addIX(vx)

	} else if (opCode & 0xF0FF) == 0xF029 {
		// Instruction Fx29: Set I = location of sprite for digit Vx.
		this.loadIX(vx)

	} else if (opCode & 0xF0FF) == 0xF033 {
		// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, I+2.
		this.loadBCD(vx)

	} else if (opCode & 0xF0FF) == 0xF055 {
		// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
		this.saveV(vx)

	} else if (opCode & 0xF0FF) == 0xF065 {
		// Instruction Fx65: Read registers V0 through Vx in memory starting at location I.
		this.loadV(vx)

	} else {
		fmt.Errorf("Unknown instruction: %X\n\n", opCode)
	}

	return nil
}

// Instruction 00E0: Clear the display.
func (this *VM) clear() {
	fmt.Println("Instruction 00E0: Clear the display.")

	// Zero out gfx
	for i := range this.GFX {
		this.GFX[i] = 0
	}

	// Set draw flag
	this.DF = true
}

// Instruction 00EE: Return from a subroutine.
// The VM sets the program counter to the address at the top of the stack,
// then subtracts 1 from the stack pointer.
func (this *VM) ret() error {
	fmt.Println("Instruction 00EE: Return from a subroutine.")

	// Decrement stack pointer and error if it's below 0.
	if this.SP -= 1; this.SP < 0 {
		return fmt.Errorf("stack pointer out of bounds: %d", this.SP)
	}

	this.PC = this.Stack[this.SP]

	fmt.Printf("New PC: %d", this.PC)
	return nil
}

// Instruction 1nnn: Jump to location nnn.
// The VM sets the program counter to nnn.
func (this *VM) jump(nnn uint16) {
	fmt.Println("Instruction 1nnn: Jump to location nnn.")
	fmt.Printf("nnn: %d\n", nnn)

	this.PC = nnn

	fmt.Printf("New PC: %d\n", this.PC)
}

// Instruction 2nnn: Call subroutine at nnn.
// The VM increments the stack pointer, then puts the current PC on the top of the stack.
// The PC is then set to nnn.
func (this *VM) call(nnn uint16) error {
	fmt.Println("Instruction 2nnn: Call subroutine at nnn.")
	fmt.Printf("nnn: %d\n", nnn)

	this.Stack[this.SP] = this.PC
	this.PC = nnn

	// Increment stack pointer and error if it's above it's length
	if this.SP += 1; this.SP > uint16(len(this.Stack)) {
		fmt.Errorf("stack pointer out of points: %d", this.SP)
	}

	fmt.Printf("New Stack: %v\nnew SP: %d\tPC: %d\n", this.Stack, this.SP, this.PC)
	return nil
}

// Instruction 3xkk: Skip next instruction if Vx = kk.
// The VM compares register Vx to kk, and if they are equal,
// increments the program counter by 2.
func (this *VM) skipIf(vx byte, kk byte) {
	fmt.Println("Instruction 3xkk: Skip next instruction if Vx == kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	if this.V[vx] == kk {
		this.PC += 2
	}

	fmt.Printf("New PC: %d\n", this.PC)
}

// Instruction 4xkk: Skip next instruction if Vx != kk.
// The VM compares register Vx to kk, and if they are not equal,
// increments the program counter by 2.
func (this *VM) skipIfNot(vx byte, kk byte) {
	fmt.Println("Instruction 4xkk: Skip next instruction if Vx != kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	if this.V[vx] != kk {
		this.PC += 2
	}

	fmt.Printf("New PC: %d\n", this.PC)
}

// Instruction 5xy0: Skip next instruction if Vx = Vy.
// The VM compares register Vx to register Vy, and if they are equal,
// increments the program counter by 2.
func (this *VM) skipIfXY(vx byte, vy byte) {
	fmt.Println("Instruction 5xy0: Skip next isntruction if Vx = Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if vx == vy {
		this.PC += 2
	}

	fmt.Printf("New PC: %d\n", this.PC)
}

// Instruction 6xkk: Set Vx = kk.
// The VM puts the value kk into register Vx.
func (this *VM) load(vx byte, kk byte) {
	fmt.Println("Instruction 6xkk: Set Vx = kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	this.V[vx] = kk

	fmt.Printf("New V%X: %X\n", vx, this.V[vx])
}

// Instruction 7xkk: Set Vx = Vx + kk.
// Adds the value kk to the value of register Vx, then stores the result in Vx.
func (this *VM) add(vx byte, kk byte) {
	fmt.Println("Instruction 7xkk: Set Vx = Vx + kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	this.V[vx] += kk

	fmt.Printf("New V%X: %X\n", vx, this.V[vx])
}

// Instruction 8xy0: Set Vx = Vy.
// Stores the value of register Vy in register Vx.
func (this *VM) loadXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy0: Set Vx = Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	this.V[vx] = this.V[vy]

	fmt.Printf("New V%X: %X\n", vx, this.V[vx])
}

// Instruction 8xy1: Set Vx = Vx OR Vy.
// Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
// A bitwise OR compares the corrseponding bits from two values, and if either bit is 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (this *VM) orXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy1: Set Vx = Vx | Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	this.V[vx] = this.V[vx] | this.V[vy]

	fmt.Printf("New V%X: %X", vx, this.V[vx])
}

// Instruction 8xy2: Set Vx = Vx AND Vy.
// Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
// A bitwise AND compares the corrseponding bits from two values, and if both bits are 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (this *VM) andXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy2: Set Vx = Vx & Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	this.V[vx] = this.V[vx] & this.V[vy]

	fmt.Printf("New V%X: %X", vx, this.V[vx])
}

// Instruction 8xy3: Set Vx = Vx XOR Vy.
// Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx.
// An exclusive OR compares the corrseponding bits from two values,
// and if the bits are not both the same, then the corresponding bit in the result is set to 1.
// Otherwise, it is 0.
func (this *VM) xorXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy3: Set Vx = Vx ^ Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	this.V[vx] = this.V[vx] ^ this.V[vy]

	fmt.Printf("New V%X: %X", vx, this.V[vx])
}

// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
// VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
func (this *VM) addXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	this.V[vx] = this.V[vx] + this.V[vy]

	fmt.Printf("New V%X: %X", vx, this.V[vx])
}

// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
// and the results stored in Vx.
func (this *VM) subXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	this.V[vx] = this.V[vx] + this.V[vy]

	fmt.Printf("New V%X: %X", vx, this.V[vx])
}

// Instruction 8xy6: Set Vx = Vx SHR 1.
// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
// Then Vx is divided by 2.
func (this *VM) shiftRight(vx byte) {
	fmt.Println("Instruction 8xy6: Set Vx = Vx SHR 1.")
	fmt.Printf("Vx: %X\n", vx)

	this.V[0xF] = this.V[vx] & 0x1

	// Another way to divide by 2
	this.V[vx] = this.V[vx] >> 1

	fmt.Printf("New V%X: %X\tVF: %X", vx, this.V[vx], this.V[0xF])
}

// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
// If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy,
// and the results stored in Vx.
func (this *VM) subYX(vx byte, vy byte) {
	fmt.Println("Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if vy > vx {
		this.V[0xF] = 1
	} else {
		this.V[0xF] = 0
	}

	this.V[vx] = this.V[vy] - this.V[vx]

	fmt.Printf("New V%X: %d\tVF: %d\n", vx, this.V[vx], this.V[0xF])
}

// Instruction 8xyE: Set Vx = Vx SHL 1.
// If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
// Then Vx is multiplied by 2.
func (this *VM) shiftLeft(vx byte) {
	fmt.Println("Instruction 8xyE: Set Vx = Vx SHL 1.")
	fmt.Printf("Vx: %X\n", vx)

	// Get the most significant bit in a byte
	this.V[0xF] = this.V[vx] >> 7

	// Multiple by 2
	this.V[vx] = this.V[vx] << 1

	fmt.Printf("New V%X: %d\tVF: %d\n", vx, this.V[vx], this.V[0xF])
}

// Instruction 9xy0: Skip next instruction if Vx != Vy.
// The values of Vx and Vy are compared, and if they are not equal,
// the program counter is increased by 2.
func (this *VM) skipIfNotXY(vx byte, vy byte) {
	fmt.Println("Instruction 9xy0: Skip next instruction if Vx != Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if this.V[vx] != this.V[vy] {
		this.PC += 2
	}

	fmt.Printf("New PC: %d\n", this.PC)
}

// Instruction Annn: Set I = nnn.
// The value of register I is set to nnn.
func (this *VM) loadI(nnn uint16) {
	fmt.Println("Instruction Annn: Set I = nnnn.")
	fmt.Printf("nnn: %X\n", nnn)

	this.I = uint(nnn)

	fmt.Printf("New I: %X", this.I)
}

// Instruction Bnnn: Jump to location nnn + V0.
// The program counter is set to nnn plus the value of V0.
func (this *VM) jumpV0(nnn uint16) {
	fmt.Println("Instruction Bnnn: Jump to location nnn + V0.")
	fmt.Printf("nnn: %X\n", nnn)

	this.PC = uint16(this.V[0x0]) + nnn

	fmt.Printf("New PC: %d\n", this.PC)
}

// Instruction Cxkk: Set Vx = random byte AND kk.
// The VM generates a random number from 0 to 255,
// which is then ANDed with the value kk. The results are stored in Vx.
// See instruction 8xy2 for more information on AND.
func (this *VM) rand(vx byte, kk byte) {
	fmt.Println("Instruction Cxkk: Set Vx = random byte AND kk.")
	fmt.Printf("Vx: %X\n", vx)

	r := byte(rand.Intn(256))
	this.V[vx] = kk & r

	fmt.Printf("New V%X: %X", vx, this.V[vx])
}

// Instruction Dxyn: Display n-byte sprite starting at memory location I at (Vx, Vy),
// set VF = collision.
//
// The VM reads n bytes from memory, starting at the address stored in I.
// These bytes are then displayed as sprites on screen at coordinates (Vx, Vy).
// Sprites are XORed onto the existing screen. If this causes any pixels to be erased,
// VF is set to 1, otherwise it is set to 0. If the sprite is positioned so part of it
// is outside the coordinates of the display, it wraps around to the opposite side of the screen.
// See instruction 8xy3 for more information on XOR, and section 2.4, Display,
// for more information on the Chip-8 screen and sprites.
func (this *VM) draw(vx byte, vy byte, n byte) {
	fmt.Println("Instruction Dxyn: Display nbyte sprite starting at memory location I at (Vx, Vy), " +
		"set Vf = collusion.")
	fmt.Printf("Vx: %X\tVy: %X\tn: %X\n", vx, vy, n)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Ex9E: Skip next instruction if key with the value of Vx is pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the down position, PC is increased by 2.
func (this *VM) skipIfKey(vx byte) {
	fmt.Println("Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the up position, PC is increased by 2.
func (this *VM) skipIfKeyNot(vx byte) {
	fmt.Println("Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx07: Set Vx = delay timer value.
// The value of DT is placed into Vx.
func (this *VM) loadXDT(vx byte) {
	fmt.Println("Instruction Fx07: Set Vx = delay timer value.")
	fmt.Printf("Vx: %X\n", vx)

	this.V[vx] = this.DT
	fmt.Printf("New V%X: %X", vx, this.V[vx])
}

// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
// All execution stops until a key is pressed, then the value of that key is stored in Vx.
func (this *VM) loadKey(vx byte) {
	fmt.Println("Instruction Fx0A: Wait for a key press, store the value of the key in Vx.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx15: Set delay timer = Vx.
// DT is set equal to the value of Vx.
func (this *VM) loadDTX(vx byte) {
	fmt.Println("Instruction Fx15: Set delay timer = Vx.")
	fmt.Printf("Vx: %X\n", vx)

	this.DT = this.V[vx]

	fmt.Printf("New DT: %d", this.DT)
}

// Instruction Fx18: Set sound timer = Vx.
// ST is set equal to the value of Vx.
func (this *VM) loadSTX(vx byte) {
	fmt.Println("Instruction Fx18: Set sounder timer = Vx.")
	fmt.Printf("Vx: %X\n", vx)

	this.ST = this.V[vx]

	fmt.Printf("New ST: %d", this.ST)
}

// Instruction Fx1E: Set I = I + Vx.
// The values of I and Vx are added, and the results are stored in I.
func (this *VM) addIX(vx byte) {
	fmt.Println("Instruction Fx1E : Set I = I + Vx.")
	fmt.Printf("Vx: %X\n", vx)

	this.I = this.I + uint(this.V[vx])

	fmt.Printf("New I: %X", this.I)
}

// Instruction Fx29: Set I = location of sprite for digit Vx.
// The value of I is set to the location for the hexadecimal sprite corresponding
// to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
func (this *VM) loadIX(vx byte) {
	fmt.Println("Instruction Fx29: Set I = location of sprite for digit Vx.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, and I+2.
// The VM takes the decimal value of Vx, and places the hundreds digit in memory
// at location in I, the tens digit at location I+1, and the ones digit at location I+2.
func (this *VM) loadBCD(vx byte) {
	fmt.Println("Instruction Fx33: Store BCD represention of Vx in memory locations I, I+1, I+2.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
// The VM copies the values of registers V0 through Vx into memory,
// starting at the address in I.
func (this *VM) saveV(vx byte) {
	fmt.Println("Instruction Fx55: Store registers V0 through Vx in memory starting at location I.")
	fmt.Printf("Vx: %X\n", vx)

	for i := uint(0); i <= uint(vx); i++ {
		this.RAM[this.I+i] = this.V[i]
	}

	fmt.Printf("New ")
	for i := uint(0); i <= uint(vx); i++ {
		fmt.Printf("I+%d: %X", i, this.RAM[this.I+i])
	}
	fmt.Println()
}

// Instruction Fx65: Read registers V0 through Vx from memory starting at location I.
// The VM reads values from memory starting at location I into registers V0 through Vx.
func (this *VM) loadV(vx byte) {
	fmt.Println("Instruction Fx65: Read registers V0 through Vx in memory starting at location I.")
	fmt.Printf("Vx: %X\n", vx)

	for i := uint(0); i <= uint(vx); i++ {
		this.V[i] = this.RAM[this.I+i]
	}

	fmt.Printf("New ")
	for i := range this.V {
		fmt.Printf("V%X: %x\t", i, this.V[i])
	}
	fmt.Println()
}
