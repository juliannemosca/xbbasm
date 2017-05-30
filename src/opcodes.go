package main

import (
	"strings"
)

type opcode struct {
	// Currently only the `hex` and `len` members
	// of this structure are used, the others are
	// included for possible future expansion.
	mnemonic    string
	mode        string
	hex         uint8
	len         int
	cycles      int
	crossesPage bool
}

// Addressing modes enum
const (
	ACC         = "Accumulator"
	RL          = "Relative"
	IMP         = "Implied"
	IM          = "Immediate"
	ZP          = "ZeroPage"
	ZPX         = "ZeroPage,X"
	ZPY         = "ZeroPage,Y"
	ABS         = "Absolute"
	ABSX        = "Absolute,X"
	ABSY        = "Absolute,Y"
	IND         = "Indirect"
	IX          = "Indirect,X"
	IY          = "Indirect,Y"
	NOMODE      = ""
	UNDEFINED   = "UNDEF"
	UNDEFINED_X = "UNDEF,X"
	UNDEFINED_Y = "UNDEF,Y"
)

var pseudoOpcodes map[string]opcode = map[string]opcode{

	// .ORG
	".ORG": opcode{mnemonic: ".ORG", mode: NOMODE},

	// .TEXT
	".TEXT": opcode{mnemonic: ".TEXT", mode: NOMODE},

	// DFB
	"DFB": opcode{mnemonic: "DFB", mode: NOMODE},
}

var opcodes map[string]opcode = map[string]opcode{

	//
	// ADC (ADd with Carry)
	//
	"ADC-" + IM:   opcode{"ADC", IM, 0x69, 2, 2, false},
	"ADC-" + ZP:   opcode{"ADC", ZP, 0x65, 2, 3, false},
	"ADC-" + ZPX:  opcode{"ADC", ZPX, 0x75, 2, 4, false},
	"ADC-" + ABS:  opcode{"ADC", ABS, 0x6D, 3, 4, false},
	"ADC-" + ABSX: opcode{"ADC", ABSX, 0x7D, 3, 4, true},
	"ADC-" + ABSY: opcode{"ADC", ABSY, 0x79, 3, 4, true},
	"ADC-" + IX:   opcode{"ADC", IX, 0x61, 2, 6, false},
	"ADC-" + IY:   opcode{"ADC", IY, 0x71, 2, 5, true},

	//
	// AND (bitwise AND with accumulator)
	//
	"AND-" + IM:   opcode{"AND", IM, 0x29, 2, 2, false},
	"AND-" + ZP:   opcode{"AND", ZP, 0x25, 2, 3, false},
	"AND-" + ZPX:  opcode{"AND", ZPX, 0x35, 2, 4, false},
	"AND-" + ABS:  opcode{"AND", ABS, 0x2D, 3, 4, false},
	"AND-" + ABSX: opcode{"AND", ABSX, 0x3D, 3, 4, true},
	"AND-" + ABSY: opcode{"AND", ABSY, 0x39, 3, 4, true},
	"AND-" + IX:   opcode{"AND", IX, 0x21, 2, 6, false},
	"AND-" + IY:   opcode{"AND", IY, 0x31, 2, 5, true},

	//
	// ASL (Arithmetic Shift Left)
	//
	"ASL-" + ACC:  opcode{"ASL", ACC, 0x0A, 1, 2, false},
	"ASL-" + ZP:   opcode{"ASL", ZP, 0x06, 2, 5, false},
	"ASL-" + ZPX:  opcode{"ASL", ZPX, 0x16, 2, 6, false},
	"ASL-" + ABS:  opcode{"ASL", ABS, 0x0E, 3, 6, false},
	"ASL-" + ABSX: opcode{"ASL", ABSX, 0x1E, 3, 7, false},

	//
	// BIT (test BITs)
	//
	"BIT-" + ZP:  opcode{"BIT", ZP, 0x24, 2, 3, false},
	"BIT-" + ABS: opcode{"BIT", ABS, 0x2C, 3, 4, false},

	//
	// Branch instructions
	//
	// Note on cycles:
	//   A branch not taken requires two machine cycles.
	//   Add one if the branch is taken.
	//   Add one more if the branch crosses a page boundary.
	"BPL-" + RL: opcode{"BPL", RL, 0x10, 2, 2, true},
	"BMI-" + RL: opcode{"BMI", RL, 0x30, 2, 2, true},
	"BVC-" + RL: opcode{"BVC", RL, 0x50, 2, 2, true},
	"BVS-" + RL: opcode{"BVS", RL, 0x70, 2, 2, true},
	"BCC-" + RL: opcode{"BCC", RL, 0x90, 2, 2, true},
	"BCS-" + RL: opcode{"BCS", RL, 0x80, 2, 2, true},
	"BNE-" + RL: opcode{"BNE", RL, 0xD0, 2, 2, true},
	"BEQ-" + RL: opcode{"BEQ", RL, 0xF0, 2, 2, true},

	//
	// BRK (BReaK)
	//
	"BRK-" + IMP: opcode{"BRK", IMP, 0x00, 1, 7, false},

	//
	// CMP (CoMPare accumulator)
	//
	"CMP-" + IM:   opcode{"CMP", IM, 0xC9, 2, 2, false},
	"CMP-" + ZP:   opcode{"CMP", ZP, 0xC5, 2, 3, false},
	"CMP-" + ZPX:  opcode{"CMP", ZPX, 0xD5, 2, 4, false},
	"CMP-" + ABS:  opcode{"CMP", ABS, 0xCD, 3, 4, false},
	"CMP-" + ABSX: opcode{"CMP", ABSX, 0xDD, 3, 4, true},
	"CMP-" + ABSY: opcode{"CMP", ABSY, 0xD9, 3, 4, true},
	"CMP-" + IX:   opcode{"CMP", IX, 0xC1, 2, 6, false},
	"CMP-" + IY:   opcode{"CMP", IY, 0xD1, 2, 5, true},

	//
	// CPX (ComPare X register)
	//
	"CPX-" + IM:  opcode{"CPX", IM, 0xE0, 2, 2, false},
	"CPX-" + ZP:  opcode{"CPX", ZP, 0xE4, 2, 3, false},
	"CPX-" + ABS: opcode{"CPX", ABS, 0xEC, 3, 4, false},

	//
	// CPY (ComPare Y register)
	//
	"CPY-" + IM:  opcode{"CPY", IM, 0xC0, 2, 2, false},
	"CPY-" + ZP:  opcode{"CPY", ZP, 0xC4, 2, 3, false},
	"CPY-" + ABS: opcode{"CPY", ABS, 0xCC, 3, 4, false},

	//
	// DEC (DECrement memory)
	//
	"DEC-" + ZP:   opcode{"DEC", ZP, 0xC6, 2, 5, false},
	"DEC-" + ZPX:  opcode{"DEC", ZPX, 0xD6, 2, 6, false},
	"DEC-" + ABS:  opcode{"DEC", ABS, 0xCE, 3, 6, false},
	"DEC-" + ABSX: opcode{"DEC", ABSX, 0xDE, 3, 7, false},

	//
	// EOR (bitwise Exclusive OR)
	//
	"EOR-" + IM:   opcode{"EOR", IM, 0x49, 2, 2, false},
	"EOR-" + ZP:   opcode{"EOR", ZP, 0x45, 2, 3, false},
	"EOR-" + ZPX:  opcode{"EOR", ZPX, 0x55, 2, 4, false},
	"EOR-" + ABS:  opcode{"EOR", ABS, 0x4D, 3, 4, false},
	"EOR-" + ABSX: opcode{"EOR", ABSX, 0x5D, 3, 4, true},
	"EOR-" + ABSY: opcode{"EOR", ABSY, 0x59, 3, 4, true},
	"EOR-" + IX:   opcode{"EOR", IX, 0x41, 2, 6, false},
	"EOR-" + IY:   opcode{"EOR", IY, 0x51, 2, 5, true},

	//
	// Flag (Processor Status) Instructions
	//
	"CLC-" + IMP: opcode{"CLC", IY, 0x18, 1, 2, true}, // CLear Carry
	"SEC-" + IMP: opcode{"SEC", IY, 0x38, 1, 2, true}, // SEt Carry
	"CLI-" + IMP: opcode{"CLI", IY, 0x58, 1, 2, true}, // CLear Interrupt
	"SEI-" + IMP: opcode{"SEI", IY, 0x78, 1, 2, true}, // SEt Interrupt
	"CLV-" + IMP: opcode{"CLV", IY, 0xB8, 1, 2, true}, // CLear oVerflow
	"CLD-" + IMP: opcode{"CLD", IY, 0xD8, 1, 2, true}, // CLear Decimal
	"SED-" + IMP: opcode{"SED", IY, 0xF8, 1, 2, true}, // SEt Decimal

	//
	// INC (INCrement memory)
	//
	"INC-" + ZP:   opcode{"INC", ZP, 0xE6, 2, 5, false},
	"INC-" + ZPX:  opcode{"INC", ZPX, 0xF6, 2, 5, false},
	"INC-" + ABS:  opcode{"INC", ABS, 0xEE, 3, 6, false},
	"INC-" + ABSX: opcode{"INC", ABSX, 0xFE, 3, 7, false},

	//
	// JMP (JuMP)
	//
	"JMP-" + ABS: opcode{"JMP", ABS, 0x4C, 3, 3, false},
	"JMP-" + IND: opcode{"JMP", IND, 0x6C, 3, 5, false},

	//
	// JSR (Jump to SubRoutine)
	//
	"JSR-" + ABS: opcode{"JSR", ABS, 0x20, 3, 6, false},

	//
	// LDA (LoaD Accumulator)
	//
	"LDA-" + IM:   opcode{"LDA", IM, 0xA9, 2, 2, false},
	"LDA-" + ZP:   opcode{"LDA", ZP, 0xA5, 2, 3, false},
	"LDA-" + ZPX:  opcode{"LDA", ZPX, 0xB5, 2, 4, false},
	"LDA-" + ABS:  opcode{"LDA", ABS, 0xAD, 3, 4, false},
	"LDA-" + ABSX: opcode{"LDA", ABSX, 0xBD, 3, 4, true},
	"LDA-" + ABSY: opcode{"LDA", ABSY, 0xB9, 3, 4, true},
	"LDA-" + IX:   opcode{"LDA", IX, 0xA1, 2, 6, false},
	"LDA-" + IY:   opcode{"LDA", IY, 0xB1, 2, 5, true},

	//
	// LDX (LoaD X register)
	//
	"LDX-" + IM:   opcode{"LDX", IM, 0xA2, 2, 2, false},
	"LDX-" + ZP:   opcode{"LDX", ZP, 0xA6, 2, 3, false},
	"LDX-" + ZPY:  opcode{"LDX", ZPY, 0xB6, 2, 4, false},
	"LDX-" + ABS:  opcode{"LDX", ABS, 0xAE, 3, 4, false},
	"LDX-" + ABSY: opcode{"LDX", ABSY, 0xBE, 3, 4, true},

	//
	// LDY (LoaD Y register)
	//
	"LDY-" + IM:   opcode{"LDY", IM, 0xA0, 2, 2, false},
	"LDY-" + ZP:   opcode{"LDY", ZP, 0xA4, 2, 3, false},
	"LDY-" + ZPX:  opcode{"LDY", ZPY, 0xB4, 2, 4, false},
	"LDY-" + ABS:  opcode{"LDY", ABS, 0xAC, 3, 4, false},
	"LDY-" + ABSX: opcode{"LDY", ABSY, 0xBC, 3, 4, true},

	//
	// LSR (Logical Shift Right)
	//
	"LSR-" + ACC:  opcode{"LSR", ACC, 0x4A, 1, 2, false},
	"LSR-" + ZP:   opcode{"LSR", ZP, 0x46, 2, 5, false},
	"LSR-" + ZPX:  opcode{"LSR", ZPY, 0x56, 2, 6, false},
	"LSR-" + ABS:  opcode{"LSR", ABS, 0x4E, 3, 6, false},
	"LSR-" + ABSX: opcode{"LSR", ABSY, 0x5E, 3, 7, false},

	// NOP (No OPeration)
	"NOP-" + IMP: opcode{"NOP", ABSY, 0xEA, 1, 2, false},

	//
	// ORA (bitwise OR with Accumulator)
	//
	"ORA-" + IM:   opcode{"ORA", IM, 0x09, 2, 2, false},
	"ORA-" + ZP:   opcode{"ORA", ZP, 0x05, 2, 3, false},
	"ORA-" + ZPX:  opcode{"ORA", ZPX, 0x15, 2, 4, false},
	"ORA-" + ABS:  opcode{"ORA", ABS, 0x0D, 3, 4, false},
	"ORA-" + ABSX: opcode{"ORA", ABSX, 0x1D, 3, 4, true},
	"ORA-" + ABSY: opcode{"ORA", ABSY, 0x19, 3, 4, true},
	"ORA-" + IX:   opcode{"ORA", IX, 0x01, 2, 6, false},
	"ORA-" + IY:   opcode{"ORA", IY, 0x11, 2, 5, true},

	//
	// Register Instructions
	//
	"TAX-" + IMP: opcode{"TAX", IMP, 0xAA, 1, 2, true}, // Transfer A to X
	"TXA-" + IMP: opcode{"TXA", IMP, 0x8A, 1, 2, true}, // Transfer X to A
	"DEX-" + IMP: opcode{"DEX", IMP, 0xCA, 1, 2, true}, // DEcrement X
	"INX-" + IMP: opcode{"INX", IMP, 0xE8, 1, 2, true}, // INcrement X
	"TAY-" + IMP: opcode{"TAY", IMP, 0xA8, 1, 2, true}, // Transfer A to Y
	"TYA-" + IMP: opcode{"TYA", IMP, 0x98, 1, 2, true}, // Transfer Y to A
	"DEY-" + IMP: opcode{"DEY", IMP, 0x88, 1, 2, true}, // DEcrement Y
	"INY-" + IMP: opcode{"INY", IMP, 0xC8, 1, 2, true}, // INcrement Y

	//
	// ROL (ROtate Left)
	//
	"ROL-" + ACC:  opcode{"ROL", ACC, 0x2A, 1, 2, false},
	"ROL-" + ZP:   opcode{"ROL", ZP, 0x26, 2, 5, false},
	"ROL-" + ZPX:  opcode{"ROL", ZPY, 0x36, 2, 6, false},
	"ROL-" + ABS:  opcode{"ROL", ABS, 0x2E, 3, 6, false},
	"ROL-" + ABSX: opcode{"ROL", ABSY, 0x3E, 3, 7, false},

	//
	// ROR (ROtate Right)
	//
	"ROR-" + ACC:  opcode{"ROR", ACC, 0x6A, 1, 2, false},
	"ROR-" + ZP:   opcode{"ROR", ZP, 0x66, 2, 5, false},
	"ROR-" + ZPX:  opcode{"ROR", ZPY, 0x76, 2, 6, false},
	"ROR-" + ABS:  opcode{"ROR", ABS, 0x6E, 3, 6, false},
	"ROR-" + ABSX: opcode{"ROR", ABSY, 0x7E, 3, 7, false},

	//
	// RTI (ReTurn from Interrupt)
	//
	"RTI-" + IMP: opcode{"RTI", IMP, 0x40, 1, 6, false},

	//
	// RTS (ReTurn from Subroutine)
	//
	"RTS-" + IMP: opcode{"RTS", IMP, 0x60, 1, 6, false},

	//
	// SBC (SuBtract with Carry)
	//
	"SBC-" + IM:   opcode{"SBC", IM, 0xE9, 2, 2, false},
	"SBC-" + ZP:   opcode{"SBC", ZP, 0xE5, 2, 3, false},
	"SBC-" + ZPX:  opcode{"SBC", ZPX, 0xF5, 2, 4, false},
	"SBC-" + ABS:  opcode{"SBC", ABS, 0xED, 3, 4, false},
	"SBC-" + ABSX: opcode{"SBC", ABSX, 0xFD, 3, 4, true},
	"SBC-" + ABSY: opcode{"SBC", ABSY, 0xF9, 3, 4, true},
	"SBC-" + IX:   opcode{"SBC", IX, 0xE1, 2, 6, false},
	"SBC-" + IY:   opcode{"SBC", IY, 0xF1, 2, 5, true},

	//
	// STA (STore Accumulator)
	//
	"STA-" + ZP:   opcode{"STA", ZP, 0x85, 2, 3, false},
	"STA-" + ZPX:  opcode{"STA", ZPX, 0x95, 2, 4, false},
	"STA-" + ABS:  opcode{"STA", ABS, 0x8D, 3, 4, false},
	"STA-" + ABSX: opcode{"STA", ABSX, 0x9D, 3, 5, false},
	"STA-" + ABSY: opcode{"STA", ABSY, 0x99, 3, 5, false},
	"STA-" + IX:   opcode{"STA", IX, 0x81, 2, 6, false},
	"STA-" + IY:   opcode{"STA", IY, 0x91, 2, 6, false},

	//
	// Stack Instructions
	//
	"TXS-" + IMP: opcode{"TXS", IMP, 0x9A, 1, 2, true}, // Transfer X to Stack ptr
	"TSX-" + IMP: opcode{"TSX", IMP, 0xBA, 1, 2, true}, // Transfer Stack ptr to X
	"PHA-" + IMP: opcode{"PHA", IMP, 0x48, 1, 3, true}, // PusH Accumulator
	"PLA-" + IMP: opcode{"PLA", IMP, 0x68, 1, 4, true}, // PuLl Accumulator
	"PHP-" + IMP: opcode{"PHP", IMP, 0x08, 1, 3, true}, // PusH Processor status
	"PLP-" + IMP: opcode{"PLP", IMP, 0x28, 1, 4, true}, // PuLl Processor status

	//
	// STX (STore X register)
	//
	"STX-" + ZP:  opcode{"STX", ZP, 0x86, 2, 3, false},
	"STX-" + ZPY: opcode{"STX", ZPX, 0x96, 2, 4, false},
	"STX-" + ABS: opcode{"STX", ABS, 0x8E, 3, 4, false},

	//
	// STY (STore Y register)
	//
	"STY-" + ZP:  opcode{"STY", ZP, 0x84, 2, 3, false},
	"STY-" + ZPY: opcode{"STY", ZPX, 0x94, 2, 4, false},
	"STY-" + ABS: opcode{"STY", ABS, 0x8C, 3, 4, false},
}

//
// Misc. helpers
//
func isBranchInstruction(oc string) bool {
	if len(oc) != 3 {
		return false
	}
	upoc := strings.ToUpper(oc)
	if upoc[0] == 'B' && upoc != "BIT" && upoc != "BRK" {
		return true
	}
	return false
}

func calcBranchOffset(instructionAddress, branchToAddress int) int {
	if instructionAddress > branchToAddress {
		return 254 - (instructionAddress - branchToAddress)
	} else {
		return (branchToAddress - instructionAddress) - 2
	}
}
