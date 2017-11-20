package main

import (
	"fmt"
	"io/ioutil"
)

type CHIP8 struct {
	// CHIP-8 is capable of accessing 4KB (4,096 bytes) of RAM.
	// Because the first 512 bytes are reserved for the interpreter, CHIP-8 programs will start at address 0x200 (512).
	RAM [4096]byte

	// CHIP-8 screen is 64x32 pixels.
	GFX [64 * 32]byte

	// 16 16-bit stack used for saving addresses before subroutines.
	Stack [16]uint

	// 16 8-bit Registers: V0 - VE are general registers and VF is a flag register.
	V [16]byte

	// 16-bit Program counter. All programs start at 0x200.
	PC uint

	// 8-bit Stack pointer
	SP uint

	// Address register
	I uint

	// Delay timer
	DT int64

	// Sound timer
	ST int64

	// Keys
	Key [16]uint

	// ROM Size: length of CHIP-8 program byte array
	RS int

	// Draw Flag
	DF uint
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

func (chip8 *CHIP8) getOpCode(PC uint) uint16 {
	// Each opcode is 2 bytes, but RAM is a byte array, so it must be accessed twice to create the opcode.
	//
	// Ex.
	// RAM[PC] = 0x01 (1 byte)
	// RAM[PC + 1] = 0xFE (1 byte)
	// opcode = RAM[PC] + RAM[PC + 1] = 0x01FE
	opCode1 := uint16(chip8.RAM[PC])
	opCode2 := uint16(chip8.RAM[PC+1])
	opCode := opCode1<<8 | opCode2

	fmt.Printf("1st OpCode: %X\n2nd OpCode: %X\nOpCode: %X\n", opCode1, opCode2, opCode)
	return opCode
}

func (chip8 *CHIP8) Cycle() {
	// Get opcode and increment PC twice
	opCode := chip8.getOpCode(chip8.PC)
	chip8.PC += 2

	// Execute code
	chip8.execute(opCode)
}

/*
	0nnn - SYS addr
	1nnn - JP addr
	2nnn - CALL addr
	3xkk - SE Vx, byte
	4xkk - SNE Vx, byte
	5xy0 - SE Vx, Vy
	6xkk - LD Vx, byte
	7xkk - ADD Vx, byte
	8xy0 - LD Vx, Vy
	8xy1 - OR Vx, Vy
	8xy2 - AND Vx, Vy
	8xy3 - XOR Vx, Vy
	8xy4 - ADD Vx, Vy
	8xy5 - SUB Vx, Vy
	8xy6 - SHR Vx {, Vy}
	8xy7 - SUBN Vx, Vy
	8xyE - SHL Vx {, Vy}
	9xy0 - SNE Vx, Vy
	Annn - LD I, addr
	Bnnn - JP V0, addr
	Cxkk - RND Vx, byte
	Dxyn - DRW Vx, Vy, nibble
	Ex9E - SKP Vx
	ExA1 - SKNP Vx
	Fx07 - LD Vx, DT
	Fx0A - LD Vx, K
	Fx15 - LD DT, Vx
	Fx18 - LD ST, Vx
	Fx1E - ADD I, Vx
	Fx29 - LD F, Vx
	Fx33 - LD B, Vx
	Fx55 - LD [I], Vx
	Fx65 - LD Vx, [I]
*/
func (chip8 *CHIP8) execute(opCode uint16) {
	if opCode == 0x00E0 {
		fmt.Println("Instruction 00E0: Clear the display")

	} else if opCode == 0x00EE {
		fmt.Println("Instruction 0x00E0: Return")

	} else if opCode&0xF000 == 0x1000 {
		fmt.Println("Instruction 1nnn: Jump")


	} else if opCode&0xF000 == 0x2000 {
		fmt.Println("Instruction 2nnn: Call subroutine")


	} else if opCode&0xF000 == 0x3000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0x4000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0x5000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0x6000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0x7000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0x8000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0x9000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0xA000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0xB000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0xC000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0xD000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0xE000 {
		fmt.Println("Instruction")


	} else if opCode&0xF000 == 0xF000 {
		fmt.Println("Instruction")


	} else {
		fmt.Printf("Unknown instruction: %X\n\n", opCode)
	}

}
