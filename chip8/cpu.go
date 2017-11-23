package CHIP8

import (
	"fmt"
	"io/ioutil"
	"math/rand"
)

type CPU struct {
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

func (self *CPU) Init() {
	self.loadFont()
}

func (self *CPU) loadFont() {
	fonts := [80]byte {0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
		0x20, 0x60, 0x20, 0x20, 0x70, // 1
		0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
		0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
		0xA0, 0xA0, 0xF0, 0x20, 0x20, // 4
		0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
		0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
		0xF0, 0x10, 0x20, 0x40, 0x40, // 7
		0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
		0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
		0xF0, 0x90, 0xF0, 0x90, 0x90, // A
		0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
		0xF0, 0x80, 0x80, 0x80, 0xF0, // C
		0xE0, 0x90, 0x90, 0x90, 0xE0, // D
		0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
		0xF0, 0x80, 0xF0, 0x80, 0x80} // F

		for i, font := range fonts {
			self.RAM[i] = font
		}
}

func (self *CPU) LoadROM(filename *string) error {
	// Read file into byte array
	rom, err := ioutil.ReadFile(*filename)
	if err != nil {
		return err
	}

	// Save ROM size
	self.RS = len(rom)

	// Move the PC to 0x200 (512 byte)
	self.PC = 0x200

	// Copy program byte array into RAM
	for i, b := range rom {
		self.RAM[self.PC+uint16(i)] = b
	}

	return nil
}

// Helpful for debugging
func (self *CPU) printRAM() {
	for i := 0; i < self.RS+512; i++ {
		if i % 10 == 0 {
			fmt.Printf("\n%d: %X", i, self.RAM[i])
		} else if self.RAM[i] & 0xF0 == 0{
			fmt.Printf("\t\t%d: 0%X", i, self.RAM[i])
		} else {
			fmt.Printf("\t\t%d: %X", i, self.RAM[i])
		}
	}

	fmt.Println()
}

// Helpful for debugging
func (self *CPU) printRegisters() {
	fmt.Printf("\nPC: %d     SP: %d     I: %d\n", self.PC, self.SP, self.I)
	fmt.Printf("Stack: %v\n", self.Stack)

	for i := range self.V {
		fmt.Printf("V%X: %x\t", i, self.V[i])
	}

	fmt.Println()
}

// Each opcode is 2 bytes, but RAM is a byte array, so it must be accessed twice to create the opcode.
//
// RAM[PC] = 0x01 (1 byte)
// RAM[PC + 1] = 0xFE (1 byte)
// opcode = RAM[PC] + RAM[PC + 1] = 0x01FE
func (self *CPU) getOpCode(PC uint16) uint16 {
	opCode1 := uint16(self.RAM[PC])
	opCode2 := uint16(self.RAM[PC+1])
	opCode := opCode1<<8 | opCode2

	//fmt.Printf("1st OpCode: %X\n2nd OpCode: %X\n", opCode1, opCode2)
	fmt.Printf("OpCode: %X\n", opCode)

	return opCode
}

func (self *CPU) Cycle() {
	// Debug
	self.printRegisters()

	// Get opcode
	opCode := self.getOpCode(self.PC)

	// Execute code
	self.execute(opCode)

	// Increment PC
	self.PC += 2
}

func (self *CPU) execute(opCode uint16) error {
	vx := byte((opCode & 0x0F00) >> 8)
	vy := byte((opCode & 0x00F0) >> 4)

	nnn := uint16(opCode & 0x0FFF)
	kk := byte(opCode & 0x00FF)
	n := byte(opCode & 0x000F)

	if opCode == 0x00E0 {
		// Instruction 00E0: Clear the display.
		self.clear()

	} else if opCode == 0x00EE {
		// Instruction 00EE: Return from a subroutine.
		return self.ret()

	} else if (opCode & 0xF000) == 0x1000 {
		// Instruction 1nnn: Jump to location nnn.
		self.jump(nnn)

	} else if (opCode & 0xF000) == 0x2000 {
		// Instruction 2nnn: Call subroutine at nnn.
		self.call(nnn)

	} else if (opCode & 0xF000) == 0x3000 {
		// Instruction 3xkk: Skip next instructionif Vx = kk.
		self.skipIf(vx, kk)

	} else if (opCode & 0xF000) == 0x4000 {
		// Instruction 4xkk: Skip next instruction if Vx != kk.
		self.skipIfNot(vx, kk)

	} else if (opCode & 0xF000) == 0x5000 {
		// Instruction 5xy0: Skip next isntruction if Vx = Vy.
		self.skipIfXY(vx, vy)

	} else if (opCode & 0xF000) == 0x6000 {
		// Instruction 6xkk: Set Vx = kk.
		self.load(vx, kk)

	} else if (opCode & 0xF000) == 0x7000 {
		// Instruction 7xkk: Set Vx = Vx + kk.
		self.add(vx, kk)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xy0: Set Vx = Vy.
		self.loadXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8001 {
		// Instruction 8xy1: Set Vx = Vx | Vy.
		self.orXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8002 {
		// Instruction 8xy2: Set Vx = Vx & Vy.
		self.andXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8003 {
		// Instruction 8xy3: Set Vx = Vx ^ Vy.
		self.xorXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8004 {
		// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
		self.addXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8005 {
		// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
		self.subXY(vx, vy)

	} else if (opCode & 0xF00F) == 0x8006 {
		// Instruction 8xy6: Set Vx = Vx SHR 1.
		self.shiftRight(vx)

	} else if (opCode & 0xF00F) == 0x8007 {
		// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
		self.subYX(vx, vy)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xyE: Set Vx = Vx SHL 1.
		self.shiftLeft(vx)

	} else if (opCode & 0xF00F) == 0x9000 {
		// Instruction 9xy0: Skip next instruction if Vx != Vy.
		self.skipIfNotXY(vx, vy)

	} else if (opCode & 0xF000) == 0xA000 {
		// Instruction Annn: Set I = nnnn.
		self.loadI(nnn)

	} else if (opCode & 0xF000) == 0xB000 {
		// Instruction Bnnn: Jump to location nnn + V0.
		self.jumpV0(nnn)

	} else if (opCode & 0xF000) == 0xC000 {
		// Instruction Cxkk: Set Vx = random byte AND kk.
		self.rand(vx, kk)

	} else if (opCode & 0xF000) == 0xD000 {
		// Instruction Dxyn: Display nbyte sprite starting at memory
		// location I at (Vx, Vy), set Vf = collusion.
		self.draw(vx, vy, n)

	} else if (opCode & 0xF0FF) == 0xE09E {
		// Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.
		self.skipIfKey(vx)

	} else if (opCode & 0xF0FF) == 0xE0A1 {
		// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
		self.skipIfKeyNot(vx)

	} else if (opCode & 0xF0FF) == 0xF007 {
		// Instruction Fx07: Set Vx = delay timer value.
		self.loadXDT(vx)

	} else if (opCode & 0xF0FF) == 0xF00A {
		// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
		self.loadKey(vx)

	} else if (opCode & 0xF0FF) == 0xF015 {
		// Instruction Fx15: Set delay timer = Vx.
		self.loadDTX(vx)

	} else if (opCode & 0xF0FF) == 0xF018 {
		// Instruction Fx18: Set sounder timer = Vx.
		self.loadSTX(vx)

	} else if (opCode & 0xF0FF) == 0xF01E {
		// Instruction Fx1E : Set I = I + Vx.
		self.addIX(vx)

	} else if (opCode & 0xF0FF) == 0xF029 {
		// Instruction Fx29: Set I = location of sprite for digit Vx.
		self.loadIX(vx)

	} else if (opCode & 0xF0FF) == 0xF033 {
		// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, I+2.
		self.loadBCD(vx)

	} else if (opCode & 0xF0FF) == 0xF055 {
		// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
		self.saveV(vx)

	} else if (opCode & 0xF0FF) == 0xF065 {
		// Instruction Fx65: Read registers V0 through Vx in memory starting at location I.
		self.loadV(vx)

	} else {
		fmt.Errorf("Unknown instruction: %X\n\n", opCode)
	}

	return nil
}

// Instruction 00E0: Clear the display.
func (self *CPU) clear() {
	fmt.Println("Instruction 00E0: Clear the display.")

	// Zero out gfx
	for i := range self.GFX {
		self.GFX[i] = 0
	}

	// Set draw flag
	self.DF = true
}

// Instruction 00EE: Return from a subroutine.
// The CPU sets the program counter to the address at the top of the stack,
// then subtracts 1 from the stack pointer.
func (self *CPU) ret() error {
	fmt.Println("Instruction 00EE: Return from a subroutine.")

	// Decrement stack pointer and error if it's below 0.
	if self.SP -= 1; self.SP < 0 {
		return fmt.Errorf("stack pointer out of bounds: %d", self.SP)
	}

	self.PC = self.Stack[self.SP]

	fmt.Printf("New PC: %d", self.PC)
	return nil
}

// Instruction 1nnn: Jump to location nnn.
// The CPU sets the program counter to nnn.
func (self *CPU) jump(nnn uint16) {
	fmt.Println("Instruction 1nnn: Jump to location nnn.")
	fmt.Printf("nnn: %d\n", nnn)

	self.PC = nnn

	fmt.Printf("New PC: %d\n", self.PC)
}

// Instruction 2nnn: Call subroutine at nnn.
// The CPU increments the stack pointer, then puts the current PC on the top of the stack.
// The PC is then set to nnn.
func (self *CPU) call(nnn uint16) error {
	fmt.Println("Instruction 2nnn: Call subroutine at nnn.")
	fmt.Printf("nnn: %d\n", nnn)

	self.Stack[self.SP] = self.PC
	self.PC = nnn

	// Increment stack pointer and error if it's above it's length
	if self.SP += 1; self.SP > uint16(len(self.Stack)) {
		fmt.Errorf("stack pointer out of points: %d", self.SP)
	}

	fmt.Printf("New Stack: %v\nnew SP: %d\tPC: %d\n", self.Stack, self.SP, self.PC)
	return nil
}

// Instruction 3xkk: Skip next instruction if Vx = kk.
// The CPU compares register Vx to kk, and if they are equal,
// increments the program counter by 2.
func (self *CPU) skipIf(vx byte, kk byte) {
	fmt.Println("Instruction 3xkk: Skip next instruction if Vx == kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	if self.V[vx] == kk {
		self.PC += 2
	}

	fmt.Printf("New PC: %d\n", self.PC)
}

// Instruction 4xkk: Skip next instruction if Vx != kk.
// The CPU compares register Vx to kk, and if they are not equal,
// increments the program counter by 2.
func (self *CPU) skipIfNot(vx byte, kk byte) {
	fmt.Println("Instruction 4xkk: Skip next instruction if Vx != kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	if self.V[vx] != kk {
		self.PC += 2
	}

	fmt.Printf("New PC: %d\n", self.PC)
}

// Instruction 5xy0: Skip next instruction if Vx = Vy.
// The CPU compares register Vx to register Vy, and if they are equal,
// increments the program counter by 2.
func (self *CPU) skipIfXY(vx byte, vy byte) {
	fmt.Println("Instruction 5xy0: Skip next isntruction if Vx = Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if vx == vy {
		self.PC += 2
	}

	fmt.Printf("New PC: %d\n", self.PC)
}

// Instruction 6xkk: Set Vx = kk.
// The CPU puts the value kk into register Vx.
func (self *CPU) load(vx byte, kk byte) {
	fmt.Println("Instruction 6xkk: Set Vx = kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	self.V[vx] = kk

	fmt.Printf("New V%X: %X\n", vx, self.V[vx])
}

// Instruction 7xkk: Set Vx = Vx + kk.
// Adds the value kk to the value of register Vx, then stores the result in Vx.
func (self *CPU) add(vx byte, kk byte) {
	fmt.Println("Instruction 7xkk: Set Vx = Vx + kk.")
	fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	self.V[vx] += kk

	fmt.Printf("New V%X: %X\n", vx, self.V[vx])
}

// Instruction 8xy0: Set Vx = Vy.
// Stores the value of register Vy in register Vx.
func (self *CPU) loadXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy0: Set Vx = Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	self.V[vx] = self.V[vy]

	fmt.Printf("New V%X: %X\n", vx, self.V[vx])
}

// Instruction 8xy1: Set Vx = Vx OR Vy.
// Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
// A bitwise OR compares the corrseponding bits from two values, and if either bit is 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (self *CPU) orXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy1: Set Vx = Vx | Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	self.V[vx] = self.V[vx] | self.V[vy]

	fmt.Printf("New V%X: %X", vx, self.V[vx])
}

// Instruction 8xy2: Set Vx = Vx AND Vy.
// Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
// A bitwise AND compares the corrseponding bits from two values, and if both bits are 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (self *CPU) andXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy2: Set Vx = Vx & Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	self.V[vx] = self.V[vx] & self.V[vy]

	fmt.Printf("New V%X: %X", vx, self.V[vx])
}

// Instruction 8xy3: Set Vx = Vx XOR Vy.
// Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx.
// An exclusive OR compares the corrseponding bits from two values,
// and if the bits are not both the same, then the corresponding bit in the result is set to 1.
// Otherwise, it is 0.
func (self *CPU) xorXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy3: Set Vx = Vx ^ Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	self.V[vx] = self.V[vx] ^ self.V[vy]

	fmt.Printf("New V%X: %X", vx, self.V[vx])
}

// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
// VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
func (self *CPU) addXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	self.V[vx] = self.V[vx] + self.V[vy]

	fmt.Printf("New V%X: %X", vx, self.V[vx])
}

// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
// and the results stored in Vx.
func (self *CPU) subXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	self.V[vx] = self.V[vx] + self.V[vy]

	fmt.Printf("New V%X: %X", vx, self.V[vx])
}

// Instruction 8xy6: Set Vx = Vx SHR 1.
// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
// Then Vx is divided by 2.
func (self *CPU) shiftRight(vx byte) {
	fmt.Println("Instruction 8xy6: Set Vx = Vx SHR 1.")
	fmt.Printf("Vx: %X\n", vx)

	self.V[0xF] = self.V[vx] & 0x1

	// Another way to divide by 2
	self.V[vx] = self.V[vx] >> 1

	fmt.Printf("New V%X: %X\tVF: %X", vx, self.V[vx], self.V[0xF])
}

// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
// If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy,
// and the results stored in Vx.
func (self *CPU) subYX(vx byte, vy byte) {
	fmt.Println("Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if vy > vx {
		self.V[0xF] = 1
	} else {
		self.V[0xF] = 0
	}

	self.V[vx] = self.V[vy] - self.V[vx]

	fmt.Printf("New V%X: %d\tVF: %d\n", vx, self.V[vx], self.V[0xF])
}

// Instruction 8xyE: Set Vx = Vx SHL 1.
// If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
// Then Vx is multiplied by 2.
func (self *CPU) shiftLeft(vx byte) {
	fmt.Println("Instruction 8xyE: Set Vx = Vx SHL 1.")
	fmt.Printf("Vx: %X\n", vx)

	// Get the most significant bit in a byte
	self.V[0xF] = self.V[vx] >> 7

	// Multiple by 2
	self.V[vx] = self.V[vx] << 1

	fmt.Printf("New V%X: %d\tVF: %d\n", vx, self.V[vx], self.V[0xF])
}

// Instruction 9xy0: Skip next instruction if Vx != Vy.
// The values of Vx and Vy are compared, and if they are not equal,
// the program counter is increased by 2.
func (self *CPU) skipIfNotXY(vx byte, vy byte) {
	fmt.Println("Instruction 9xy0: Skip next instruction if Vx != Vy.")
	fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if self.V[vx] != self.V[vy] {
		self.PC += 2
	}

	fmt.Printf("New PC: %d\n", self.PC)
}

// Instruction Annn: Set I = nnn.
// The value of register I is set to nnn.
func (self *CPU) loadI(nnn uint16) {
	fmt.Println("Instruction Annn: Set I = nnnn.")
	fmt.Printf("nnn: %X\n", nnn)

	self.I = uint(nnn)

	fmt.Printf("New I: %X", self.I)
}

// Instruction Bnnn: Jump to location nnn + V0.
// The program counter is set to nnn plus the value of V0.
func (self *CPU) jumpV0(nnn uint16) {
	fmt.Println("Instruction Bnnn: Jump to location nnn + V0.")
	fmt.Printf("nnn: %X\n", nnn)

	self.PC = uint16(self.V[0x0]) + nnn

	fmt.Printf("New PC: %d\n", self.PC)
}

// Instruction Cxkk: Set Vx = random byte AND kk.
// The CPU generates a random number from 0 to 255,
// which is then ANDed with the value kk. The results are stored in Vx.
// See instruction 8xy2 for more information on AND.
func (self *CPU) rand(vx byte, kk byte) {
	fmt.Println("Instruction Cxkk: Set Vx = random byte AND kk.")
	fmt.Printf("Vx: %X\n", vx)

	r := byte(rand.Intn(256))
	self.V[vx] = kk & r

	fmt.Printf("New V%X: %X", vx, self.V[vx])
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
func (self *CPU) draw(vx byte, vy byte, n byte) {
	fmt.Println("Instruction Dxyn: Display nbyte sprite starting at memory location I at (Vx, Vy), set Vf = collusion.")
	fmt.Printf("Vx: %X\tVy: %X\tn: %X\n", vx, vy, n)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Ex9E: Skip next instruction if key with the value of Vx is pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the down position, PC is increased by 2.
func (self *CPU) skipIfKey(vx byte) {
	fmt.Println("Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the up position, PC is increased by 2.
func (self *CPU) skipIfKeyNot(vx byte) {
	fmt.Println("Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx07: Set Vx = delay timer value.
// The value of DT is placed into Vx.
func (self *CPU) loadXDT(vx byte) {
	fmt.Println("Instruction Fx07: Set Vx = delay timer value.")
	fmt.Printf("Vx: %X\n", vx)

	self.V[vx] = self.DT
	fmt.Printf("New V%X: %X", vx, self.V[vx])
}

// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
// All execution stops until a key is pressed, then the value of that key is stored in Vx.
func (self *CPU) loadKey(vx byte) {
	fmt.Println("Instruction Fx0A: Wait for a key press, store the value of the key in Vx.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx15: Set delay timer = Vx.
// DT is set equal to the value of Vx.
func (self *CPU) loadDTX(vx byte) {
	fmt.Println("Instruction Fx15: Set delay timer = Vx.")
	fmt.Printf("Vx: %X\n", vx)

	self.DT = self.V[vx]

	fmt.Printf("New DT: %d", self.DT)
}

// Instruction Fx18: Set sound timer = Vx.
// ST is set equal to the value of Vx.
func (self *CPU) loadSTX(vx byte) {
	fmt.Println("Instruction Fx18: Set sounder timer = Vx.")
	fmt.Printf("Vx: %X\n", vx)

	self.ST = self.V[vx]

	fmt.Printf("New ST: %d", self.ST)
}

// Instruction Fx1E: Set I = I + Vx.
// The values of I and Vx are added, and the results are stored in I.
func (self *CPU) addIX(vx byte) {
	fmt.Println("Instruction Fx1E : Set I = I + Vx.")
	fmt.Printf("Vx: %X\n", vx)

	self.I = self.I + uint(self.V[vx])

	fmt.Printf("New I: %X", self.I)
}

// Instruction Fx29: Set I = location of sprite for digit Vx.
// The value of I is set to the location for the hexadecimal sprite corresponding
// to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
func (self *CPU) loadIX(vx byte) {
	fmt.Println("Instruction Fx29: Set I = location of sprite for digit Vx.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, and I+2.
// The CPU takes the decimal value of Vx, and places the hundreds digit in memory
// at location in I, the tens digit at location I+1, and the ones digit at location I+2.
func (self *CPU) loadBCD(vx byte) {
	fmt.Println("Instruction Fx33: Store BCD represention of Vx in memory locations I, I+1, I+2.")
	fmt.Printf("Vx: %X\n", vx)
	fmt.Println("NOT YET IMPLEMENTED")
}

// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
// The CPU copies the values of registers V0 through Vx into memory,
// starting at the address in I.
func (self *CPU) saveV(vx byte) {
	fmt.Println("Instruction Fx55: Store registers V0 through Vx in memory starting at location I.")
	fmt.Printf("Vx: %X\n", vx)

	for i := uint(0); i <= uint(vx); i++ {
		self.RAM[self.I+i] = self.V[i]
	}

	fmt.Printf("New ")
	for i := uint(0); i <= uint(vx); i++ {
		fmt.Printf("I+%d: %X", i, self.RAM[self.I+i])
	}
	fmt.Println()
}

// Instruction Fx65: Read registers V0 through Vx from memory starting at location I.
// The CPU reads values from memory starting at location I into registers V0 through Vx.
func (self *CPU) loadV(vx byte) {
	fmt.Println("Instruction Fx65: Read registers V0 through Vx in memory starting at location I.")
	fmt.Printf("Vx: %X\n", vx)

	for i := uint(0); i <= uint(vx); i++ {
		self.V[i] = self.RAM[self.I+i]
	}

	fmt.Printf("New ")
	for i := range self.V {
		fmt.Printf("V%X: %x\t", i, self.V[i])
	}
	fmt.Println()
}
