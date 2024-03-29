package CHIP8

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"io/ioutil"
	"math/rand"
)

type CPU struct {
	RAM   [4096]byte   // CHIP-8 is capable of accessing 4KB (4,096 bytes) of RAM.
	GFX   [32][64]byte // CHIP-8 screen is 64x32 pixels.
	Stack [16]uint16   // 16 16-bit stack used for saving addresses before subroutines.

	V [16]byte // 16 8-bit Registers: V0 - VE are general registers and VF is a flag register.

	PC uint16 // 16-bit Program counter. All programs start at 0x200.
	SP uint16 // 16-bit Stack pointer
	I  uint   // Address register

	DT byte // Delay timer
	ST byte // Sound timer

	Key    [16]bool
	keypad map[sdl.Scancode]byte

	RS int  // ROM Size: length of CHIP-8 program byte array
	DF bool // Draw Flag
}

func (cpu *CPU) Init() {
	cpu.loadFont()

	cpu.keypad = map[sdl.Scancode]byte{
		sdl.SCANCODE_1: 0x1,
		sdl.SCANCODE_2: 0x2,
		sdl.SCANCODE_3: 0x3,
		sdl.SCANCODE_Q: 0x4,
		sdl.SCANCODE_W: 0x5,
		sdl.SCANCODE_E: 0x6,
		sdl.SCANCODE_A: 0x7,
		sdl.SCANCODE_S: 0x8,
		sdl.SCANCODE_D: 0x9,
		sdl.SCANCODE_X: 0x0,
		sdl.SCANCODE_Z: 0xA,
		sdl.SCANCODE_C: 0xB,
		sdl.SCANCODE_4: 0xC,
		sdl.SCANCODE_R: 0xD,
		sdl.SCANCODE_F: 0xE,
		sdl.SCANCODE_V: 0xF}
}

func (cpu *CPU) loadFont() {
	fonts := [80]byte{0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
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

	copy(cpu.RAM[:], fonts[:])
}

func (cpu *CPU) LoadROM(filename *string) error {
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
func (cpu *CPU) printRAM() {
	for i := 0; i < cpu.RS+512; i++ {
		if (i % 10) == 0 {
			fmt.Printf("\n%d: %X", i, cpu.RAM[i])
		} else if cpu.RAM[i]&0xF0 == 0 {
			fmt.Printf("\t\t%d: 0%X", i, cpu.RAM[i])
		} else {
			fmt.Printf("\t\t%d: %X", i, cpu.RAM[i])
		}
	}

	fmt.Println()
}

// Helpful for debugging
func (cpu *CPU) printRegisters() {
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
func (cpu *CPU) getOpCode(PC uint16) uint16 {
	opCode1 := uint16(cpu.RAM[PC])
	opCode2 := uint16(cpu.RAM[PC+1])
	opCode := opCode1<<8 | opCode2

	//fmt.Printf("1st OpCode: %X\t2nd OpCode: %X\t", opCode1, opCode2)
	if opCode != 0 {
		cpu.printRegisters()
		fmt.Printf("PC: %d\tOpCode: %X\n", cpu.PC, opCode)
	}

	return opCode
}

func (cpu *CPU) Cycle() error {
	// Debug
	//cpu.printRegisters()
	if cpu.PC < 4094 {
		// Get opcode
		opCode := cpu.getOpCode(cpu.PC)

		// Execute code
		if err := cpu.execute(opCode); err != nil {
			return err
		}

		// Increment PC and decrement DT & ST

		if cpu.DT > 0 {
			cpu.DT -= 1
		}

		if cpu.ST > 0 {
			cpu.ST -= 1
		}
	}

	return nil
}

func (cpu *CPU) execute(opCode uint16) error {
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
		return cpu.jump(nnn)

	} else if (opCode & 0xF000) == 0x2000 {
		// Instruction 2nnn: Call subroutine at nnn.
		return cpu.call(nnn)

	} else if (opCode & 0xF000) == 0x3000 {
		// Instruction 3xkk: Skip next instruction if Vx = kk.
		cpu.skipIf(vx, kk)

	} else if (opCode & 0xF000) == 0x4000 {
		// Instruction 4xkk: Skip next instruction if Vx != kk.
		cpu.skipIfNot(vx, kk)

	} else if (opCode & 0xF00F) == 0x5000 {
		// Instruction 5xy0: Skip next instruction if Vx = Vy.
		cpu.skipIfXY(vx, vy)

	} else if (opCode & 0xF000) == 0x6000 {
		// Instruction 6xkk: Set Vx = kk.
		cpu.load(vx, kk)

	} else if (opCode & 0xF000) == 0x7000 {
		// Instruction 7xkk: Set Vx = Vx + kk.
		cpu.add(vx, kk)

	} else if (opCode & 0xF00F) == 0x8000 {
		// Instruction 8xy0: Set Vx = Vy.
		fmt.Printf("UHM 8X000: %X\n", opCode)
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

	} else if (opCode & 0xF00F) == 0x800E {
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
		return cpu.draw(vx, vy, n)

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
		fmt.Printf("Unknown instruction: %X\n", opCode)
	}

	return nil
}

// Instruction 00E0: Clear the display.
func (cpu *CPU) clear() {
	fmt.Println("Instruction 00E0: Clear the display.")

	// Zero out gfx
	cpu.GFX = [32][64]byte{}

	// Set draw flag
	cpu.DF = true

	// Increment PC counter
	cpu.PC += 2
}

// Instruction 00EE: Return from a subroutine.
// The CPU sets the program counter to the address at the top of the stack,
// then subtracts 1 from the stack pointer.
func (cpu *CPU) ret() error {
	fmt.Println("Instruction 00EE: Return from a subroutine.")

	// Decrement stack pointer and error if it's below 0.
	if cpu.SP -= 1; cpu.SP < 0 {
		return fmt.Errorf("stack pointer out of bound: %d", cpu.SP)
	}

	cpu.PC = cpu.Stack[cpu.SP]
	cpu.PC += 2

	return nil
}

// Instruction 1nnn: Jump to location nnn.
// The CPU sets the program counter to nnn.
func (cpu *CPU) jump(nnn uint16) error {
	fmt.Println("Instruction 1nnn: Jump to location nnn.")
	//fmt.Printf("nnn: %d\n", nnn)

	// Set PC to nnn. Error if it accesses invalid memory.
	if cpu.PC = nnn; cpu.PC > 4028 {
		return fmt.Errorf("jump: program counter out of bound: %d", nnn)
	}

	//fmt.Printf("New PC: %d\n", cpu.PC)
	return nil
}

// Instruction 2nnn: Call subroutine at nnn.
// The CPU increments the stack pointer, then puts the current PC on the top of the stack.
// The PC is then set to nnn.
func (cpu *CPU) call(nnn uint16) error {
	fmt.Println("Instruction 2nnn: Call subroutine at nnn.")
	//fmt.Printf("nnn: %d\n", nnn)

	cpu.Stack[cpu.SP] = cpu.PC

	// Set PC to nnn. Error if it accesses invalid memory.
	if cpu.PC = nnn; cpu.PC > 4028 {
		return fmt.Errorf("call: program counter out of bound: %d", nnn)
	}

	// Increment stack pointer and error if it's above it's length
	if cpu.SP += 1; cpu.SP > uint16(len(cpu.Stack)) {
		return fmt.Errorf("call: stack pointer out of bound: %d", cpu.SP)
	}

	//fmt.Printf("New Stack: %v\nnew SP: %d\tPC: %d\n", cpu.Stack, cpu.SP, cpu.PC)
	return nil
}

// Instruction 3xkk: Skip next instruction if Vx = kk.
// The CPU compares register Vx to kk, and if they are equal,
// increments the program counter by 2.
func (cpu *CPU) skipIf(vx byte, kk byte) {
	fmt.Println("Instruction 3xkk: Skip next instruction if Vx == kk.")
	//fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	if cpu.V[vx] == kk {
		cpu.PC += 2
	}

	//fmt.Printf("New PC: %d\n", cpu.PC)
	cpu.PC += 2
}

// Instruction 4xkk: Skip next instruction if Vx != kk.
// The CPU compares register Vx to kk, and if they are not equal,
// increments the program counter by 2.
func (cpu *CPU) skipIfNot(vx byte, kk byte) {
	fmt.Println("Instruction 4xkk: Skip next instruction if Vx != kk.")
	//fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	if cpu.V[vx] != kk {
		cpu.PC += 2
	}

	//fmt.Printf("New PC: %d\n", cpu.PC)
	cpu.PC += 2
}

// Instruction 5xy0: Skip next instruction if Vx = Vy.
// The CPU compares register Vx to register Vy, and if they are equal,
// increments the program counter by 2.
func (cpu *CPU) skipIfXY(vx byte, vy byte) {
	fmt.Println("Instruction 5xy0: Skip next isntruction if Vx = Vy.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if cpu.V[vx] == cpu.V[vy] {
		cpu.PC += 2
	}

	//fmt.Printf("New PC: %d\n", cpu.PC)
	cpu.PC += 2
}

// Instruction 6xkk: Set Vx = kk.
// The CPU puts the value kk into register Vx.
func (cpu *CPU) load(vx byte, kk byte) {
	fmt.Println("Instruction 6xkk: Set Vx = kk.")
	//fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	cpu.V[vx] = kk

	//fmt.Printf("New V%X: %X\n", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction 7xkk: Set Vx = Vx + kk.
// Adds the value kk to the value of register Vx, then stores the result in Vx.
func (cpu *CPU) add(vx byte, kk byte) {
	fmt.Println("Instruction 7xkk: Set Vx = Vx + kk.")
	//fmt.Printf("Vx: %X\tkk: %X\n", vx, kk)

	cpu.V[vx] += kk

	//fmt.Printf("New V%X: %X\n", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction 8xy0: Set Vx = Vy.
// Stores the value of register Vy in register Vx.
func (cpu *CPU) loadXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy0: Set Vx = Vy.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] = cpu.V[vy]

	//fmt.Printf("New V%X: %X\n", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction 8xy1: Set Vx = Vx OR Vy.
// Performs a bitwise OR on the values of Vx and Vy, then stores the result in Vx.
// A bitwise OR compares the corresponding bits from two values, and if either bit is 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (cpu *CPU) orXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy1: Set Vx = Vx | Vy.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] |= cpu.V[vy]

	//fmt.Printf("New V%X: %X", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction 8xy2: Set Vx = Vx AND Vy.
// Performs a bitwise AND on the values of Vx and Vy, then stores the result in Vx.
// A bitwise AND compares the corresponding bits from two values, and if both bits are 1,
// then the same bit in the result is also 1. Otherwise, it is 0.
func (cpu *CPU) andXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy2: Set Vx = Vx & Vy.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] &= cpu.V[vy]

	//fmt.Printf("New V%X: %X", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction 8xy3: Set Vx = Vx XOR Vy.
// Performs a bitwise exclusive OR on the values of Vx and Vy, then stores the result in Vx.
// An exclusive OR compares the corrseponding bits from two values,
// and if the bits are not both the same, then the corresponding bit in the result is set to 1.
// Otherwise, it is 0.
func (cpu *CPU) xorXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy3: Set Vx = Vx ^ Vy.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	cpu.V[vx] ^= cpu.V[vy]

	//fmt.Printf("New V%X: %X", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.
// The values of Vx and Vy are added together. If the result is greater than 8 bits (i.e., > 255,)
// VF is set to 1, otherwise 0. Only the lowest 8 bits of the result are kept, and stored in Vx.
func (cpu *CPU) addXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy4: Set Vx = Vx + Vy, set VF = carry.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	num := uint(cpu.V[vx]) + uint(cpu.V[vy])

	if num > 255 {
		cpu.V[0xF] = 1
	} else {
		cpu.V[0xF] = 0
	}

	cpu.V[vx] = byte(num)

	//fmt.Printf("New V%X: %X", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.
// If Vx > Vy, then VF is set to 1, otherwise 0. Then Vy is subtracted from Vx,
// and the results stored in Vx.
func (cpu *CPU) subXY(vx byte, vy byte) {
	fmt.Println("Instruction 8xy5: Set Vx = Vx - Vy, set VF = NOT borrow.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if cpu.V[vx] > cpu.V[vy] {
		cpu.V[0xF] = 1
	} else {
		cpu.V[0xF] = 0
	}

	cpu.V[vx] = cpu.V[vx] - cpu.V[vy]

	//fmt.Printf("New V%X: %X", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction 8xy6: Set Vx = Vx SHR 1.
// If the least-significant bit of Vx is 1, then VF is set to 1, otherwise 0.
// Then Vx is divided by 2.
func (cpu *CPU) shiftRight(vx byte) {
	fmt.Println("Instruction 8xy6: Set Vx = Vx SHR 1.")
	//fmt.Printf("Vx: %X\n", vx)

	cpu.V[0xF] = cpu.V[vx] & 0x1

	// Divide by 2
	cpu.V[vx] = cpu.V[vx] >> 1

	//fmt.Printf("New V%X: %X\tVF: %X", vx, cpu.V[vx], cpu.V[0xF])
	cpu.PC += 2
}

// Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.
// If Vy > Vx, then VF is set to 1, otherwise 0. Then Vx is subtracted from Vy,
// and the results stored in Vx.
func (cpu *CPU) subYX(vx byte, vy byte) {
	fmt.Println("Instruction 8xy7: Set Vx = Vy - Vx, set VF = NOT borrow.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if cpu.V[vy] > cpu.V[vx] {
		cpu.V[0xF] = 1
	} else {
		cpu.V[0xF] = 0
	}

	cpu.V[vx] = cpu.V[vy] - cpu.V[vx]

	//fmt.Printf("New V%X: %d\tVF: %d\n", vx, cpu.V[vx], cpu.V[0xF])
	cpu.PC += 2
}

// Instruction 8xyE: Set Vx = Vx SHL 1.
// If the most-significant bit of Vx is 1, then VF is set to 1, otherwise to 0.
// Then Vx is multiplied by 2.
func (cpu *CPU) shiftLeft(vx byte) {
	fmt.Println("Instruction 8xyE: Set Vx = Vx SHL 1.")
	//fmt.Printf("VX: %X\n", cpu.V[vx])

	// Get the most significant bit in a byte
	cpu.V[0xF] = (cpu.V[vx] >> 7) & 0x1

	// Multiple by 2
	cpu.V[vx] = cpu.V[vx] << 1

	//fmt.Printf("New V%X: %d\tVF: %d\n", vx, cpu.V[vx], cpu.V[0xF])
	cpu.PC += 2
}

// Instruction 9xy0: Skip next instruction if Vx != Vy.
// The values of Vx and Vy are compared, and if they are not equal,
// the program counter is increased by 2.
func (cpu *CPU) skipIfNotXY(vx byte, vy byte) {
	fmt.Println("Instruction 9xy0: Skip next instruction if Vx != Vy.")
	//fmt.Printf("Vx: %X\tVy: %X\n", vx, vy)

	if cpu.V[vx] != cpu.V[vy] {
		cpu.PC += 2
	}

	//fmt.Printf("New PC: %d\n", cpu.PC)
	cpu.PC += 2
}

// Instruction Annn: Set I = nnn.
// The value of register I is set to nnn.
func (cpu *CPU) loadI(nnn uint16) {
	fmt.Println("Instruction Annn: Set I = nnn.")
	//fmt.Printf("nnn: %X\n", nnn)

	cpu.I = uint(nnn)

	//fmt.Printf("New I: %X", cpu.I)
	cpu.PC += 2
}

// Instruction Bnnn: Jump to location nnn + V0.
// The program counter is set to nnn plus the value of V0.
func (cpu *CPU) jumpV0(nnn uint16) {
	fmt.Println("Instruction Bnnn: Jump to location nnn + V0.")
	//fmt.Printf("nnn: %X\n", nnn)

	cpu.PC = uint16(cpu.V[0x0]) + nnn

	//fmt.Printf("New PC: %d\n", cpu.PC)
}

// Instruction Cxkk: Set Vx = random byte AND kk.
// The CPU generates a random number from 0 to 255,
// which is then ANDed with the value kk. The results are stored in Vx.
// See instruction 8xy2 for more information on AND.
func (cpu *CPU) rand(vx byte, kk byte) {
	fmt.Println("Instruction Cxkk: Set Vx = random byte AND kk.")
	//fmt.Printf("Vx: %X\n", vx)

	r := byte(rand.Intn(0xFF))
	cpu.V[vx] = kk & r

	//fmt.Printf("New V%X: %X", vx, cpu.V[vx])
	cpu.PC += 2
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
func (cpu *CPU) draw(vx byte, vy byte, n byte) error {
	fmt.Println("Instruction Dxyn: Display nbyte sprite starting at memory location I at (Vx, Vy), set Vf = collusion.")
	//fmt.Printf("Vx: %X\tVy: %X\tn: %X\n", vx, vy, n)

	x := cpu.V[vx]
	y := cpu.V[vy]

	fmt.Printf("Coordinates: (%d, %d)\n", x, y)
	for i := uint(0); i < uint(n); i++ {
		if (y + byte(i)) > 32 {
			return fmt.Errorf("draw: Y out of bounds: %d", y+byte(i))
		}

		value := cpu.RAM[cpu.I+i]

		for j := uint(0); j < 8; j++ {
			//fmt.Printf("%b", )
			if (y + byte(i)) > 64 {
				return fmt.Errorf("draw: X out of bounds: %d", x+byte(j))
			}

			if (value & (0x80>>j)) != 0 {
				//fmt.Printf("(%d,%d): %b\t", x+byte(j), y+byte(i), 1)
				//fmt.Printf("%d", 1)
				if cpu.GFX[y+byte(i)][x+byte(j)] == 1 {
					cpu.V[0xF] = 1
				}

				cpu.GFX[y+byte(i)][x+byte(j)] ^= 1
			} else {
				//fmt.Printf("(%d,%d): %b\t", x+byte(i), y+byte(j), 0)
				//fmt.Printf(" ")
			}
		}
		//fmt.Println()
	}

	//fmt.Print(cpu.GFX)
	//fmt.Println("GFX\n\n\n\n\n")
	//for i := 0; i < 32; i++ {
	//	for j := 0; j < 64; j++ {
	//		fmt.Printf("%d", cpu.GFX[i][j])
	//	}
	//	fmt.Println()
	//}
	//
	//fmt.Printf("%b\n", mem)
	cpu.DF = true
	cpu.PC += 2

	return nil
}

// Instruction Ex9E: Skip next instruction if key with the value of Vx is pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the down position, PC is increased by 2.
func (cpu *CPU) skipIfKey(vx byte) {
	fmt.Println("Instruction Ex9E: Skip instruction if key with the value of Vx is pressed.")
	//fmt.Printf("Vx: %X\n", vx)

	// If the key is pressed
	if cpu.Key[cpu.V[vx]] {
		cpu.PC += 2
	}

	//fmt.Printf("New PC: %d\tKey: %d\tPressed: %t\n", cpu.PC, cpu.V[vx], cpu.Key[cpu.V[vx]])
	cpu.PC += 2
}

// Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.
// Checks the keyboard, and if the key corresponding to the value of Vx is currently
// in the up position, PC is increased by 2.
func (cpu *CPU) skipIfKeyNot(vx byte) {
	fmt.Println("Instruction ExA1: Skip next instruction if key with the value of Vx is not pressed.")
	//fmt.Printf("Vx: %X\n", vx)

	// If the key isn't pressed
	if !cpu.Key[cpu.V[vx]] {
		cpu.PC += 2
	}

	//fmt.Printf("New PC: %d\tKey: %d\tNot Pressed: %t\n", cpu.PC, cpu.V[vx], cpu.Key[cpu.V[vx]])
	cpu.PC += 2
}

// Instruction Fx07: Set Vx = delay timer value.
// The value of DT is placed into Vx.
func (cpu *CPU) loadXDT(vx byte) {
	fmt.Println("Instruction Fx07: Set Vx = delay timer value.")
	//fmt.Printf("Vx: %X\n", vx)

	cpu.V[vx] = cpu.DT
	//fmt.Printf("New V%X: %X", vx, cpu.V[vx])
	cpu.PC += 2
}

// Instruction Fx0A: Wait for a key press, store the value of the key in Vx.
// All execution stops until a key is pressed, then the value of that key is stored in Vx.
func (cpu *CPU) loadKey(vx byte) {
	fmt.Println("Instruction Fx0A: Wait for a key press, store the value of the key in Vx.")
	//fmt.Printf("Vx: %X\n", vx)

	wait := true

	for wait {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch eventType := event.(type) {
			case *sdl.KeyDownEvent:
				if _, ok := cpu.keypad[eventType.Keysym.Scancode]; ok {
					cpu.V[vx] = cpu.keypad[eventType.Keysym.Scancode]
					wait = false
				}
			}
		}
	}

	cpu.PC += 2
}

// Instruction Fx15: Set delay timer = Vx.
// DT is set equal to the value of Vx.
func (cpu *CPU) loadDTX(vx byte) {
	fmt.Println("Instruction Fx15: Set delay timer = Vx.")
	//fmt.Printf("Vx: %X\n", vx)

	cpu.DT = cpu.V[vx]

	//fmt.Printf("New DT: %d", cpu.DT)
	cpu.PC += 2
}

// Instruction Fx18: Set sound timer = Vx.
// ST is set equal to the value of Vx.
func (cpu *CPU) loadSTX(vx byte) {
	fmt.Println("Instruction Fx18: Set sounder timer = Vx.")
	//fmt.Printf("Vx: %X\n", vx)

	cpu.ST = cpu.V[vx]

	//fmt.Printf("New ST: %d", cpu.ST)
	cpu.PC += 2
}

// Instruction Fx1E: Set I = I + Vx.
// The values of I and Vx are added, and the results are stored in I.
func (cpu *CPU) addIX(vx byte) {
	fmt.Println("Instruction Fx1E : Set I = I + Vx.")
	//fmt.Printf("Vx: %X\n", vx)

	cpu.I = cpu.I + uint(cpu.V[vx])

	//fmt.Printf("New I: %X", cpu.I)
	cpu.PC += 2
}

// Instruction Fx29: Set I = location of sprite for digit Vx.
// The value of I is set to the location for the hexadecimal sprite corresponding
// to the value of Vx. See section 2.4, Display, for more information on the Chip-8 hexadecimal font.
func (cpu *CPU) loadIX(vx byte) {
	fmt.Println("Instruction Fx29: Set I = location of sprite for digit Vx.")
	//fmt.Printf("V%X: %X\tI: %X\n", vx, cpu.V[vx], cpu.I)

	cpu.I = uint(cpu.V[vx]) * 5

	//fmt.Printf("New I: %X\n\n", cpu.I)
	cpu.PC += 2
}

// Instruction Fx33: Store BCD representation of Vx in memory locations I, I+1, and I+2.
// The CPU takes the decimal value of Vx, and places the hundreds digit in memory
// at location in I, the tens digit at location I+1, and the ones digit at location I+2.
func (cpu *CPU) loadBCD(vx byte) {
	fmt.Println("Instruction Fx33: Store BCD represention of Vx in memory locations I, I+1, I+2.")
	//fmt.Printf("Vx: %X\n", vx)

	dec := cpu.V[vx]

	for i := 2; i >= 0; i-- {
		cpu.RAM[cpu.I+uint(i)] = byte(dec % 10)
		dec /= 10
	}

	//fmt.Printf("Num: %d\tI: %d\tI+1: %d\tI+2: %d\n", cpu.V[vx], cpu.RAM[cpu.I], cpu.RAM[cpu.I+1], cpu.RAM[cpu.I+2])
	cpu.PC += 2
}

// Instruction Fx55: Store registers V0 through Vx in memory starting at location I.
// The CPU copies the values of registers V0 through Vx into memory,
// starting at the address in I.
func (cpu *CPU) saveV(vx byte) {
	fmt.Println("Instruction Fx55: Store registers V0 through Vx in memory starting at location I.")
	//fmt.Printf("Vx: %X\n", vx)

	for i := uint(0); i <= uint(vx); i++ {
		cpu.RAM[cpu.I+i] = cpu.V[i]
	}

	//fmt.Printf("New ")
	//for i := uint(0); i <= uint(vx); i++ {
	//fmt.Printf("I+%d: %X", i, cpu.RAM[cpu.I+i])
	//}
	//fmt.Println()
	cpu.PC += 2
}

// Instruction Fx65: Read registers V0 through Vx from memory starting at location I.
// The CPU reads values from memory starting at location I into registers V0 through Vx.
func (cpu *CPU) loadV(vx byte) {
	fmt.Println("Instruction Fx65: Read registers V0 through Vx in memory starting at location I.")
	//fmt.Printf("Vx: %X\n", vx)

	for i := uint(0); i <= uint(vx); i++ {
		cpu.V[i] = cpu.RAM[cpu.I+i]
	}

	//fmt.Printf("New ")
	//for i := range cpu.V {
	//	fmt.Printf("V%X: %x\t", i, cpu.V[i])
	//}
	//fmt.Println()
	cpu.PC += 2
}
