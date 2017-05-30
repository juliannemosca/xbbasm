package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func assemble(programData []tokenizedLine) ([]byte, error) {

	type assemblyLine struct {
		addr        int
		data        *tokenizedLine
		skipOperand bool
	}

	var partiallyAssembled []assemblyLine

	var p *tokenizedLine
	currentAddr := startAddr

	// first pass
	for i := 0; i < len(programData); i++ {
		p = &programData[i]
		l := assemblyLine{}
		l.addr = currentAddr

		// check for labels
		if p.label != "" {
			if symErr := saveSymbol(p.label, currentAddr); symErr != nil {
				return nil, symErr
			}
		}

		// check for opcodes .TEXT
		if p.opc.mnemonic == ".TEXT" || p.opc.mnemonic == "DFB" {

			// for each byte just use an opcode with the
			// corresponding hex value
			for _, b := range p.opr.defBytes {
				partiallyAssembled =
					append(
						partiallyAssembled,
						assemblyLine{
							addr:        currentAddr,
							data:        &tokenizedLine{opc: opcode{hex: b}},
							skipOperand: true,
						})
				currentAddr++
			}
			continue
		}

		if p.opr.mode == IMP || p.opr.mode == ACC || p.opr.mode == NOMODE {
			l.skipOperand = true
		}

		l.data = p

		partiallyAssembled = append(partiallyAssembled, l)
		currentAddr = currentAddr + p.opc.len
	}

	var program []byte

	var pa assemblyLine
	buffer := new(bytes.Buffer)

	// write the start address at the beginning
	if bWriteErr := binary.Write(buffer, binary.LittleEndian, uint16(startAddr)); bWriteErr != nil {
		return nil, bWriteErr
	}
	program = append(program, buffer.Bytes()...)

	// variable definitions for undefined modes lookup
	var opc opcode
	var opcFindErr error

	// second pass: resolve symbols and write hex values
	for i := 0; i < len(partiallyAssembled); i++ {
		pa = partiallyAssembled[i]

		// resolve symbols
		if pa.data.opr.label != "" {
			if v, lookupSymErr := lookupSymbol(pa.data.opr.label); lookupSymErr != nil {
				return nil, lookupSymErr
			} else {
				pa.data.opr.addr = v
			}
		}

		// look for the opcodes with undefined modes
		if pa.data.opc.mode == UNDEFINED || pa.data.opc.mode == UNDEFINED_X || pa.data.opc.mode == UNDEFINED_Y {
			if pa.data.opr.addr <= 0xFF {
				return nil, fmt.Errorf("Labels for Zero-Page addresing modes have to be defined first")
			}

			// now we know for sure it's absolute
			switch pa.data.opc.mode {
			case UNDEFINED:
				opc, opcFindErr = readOpcode(pa.data.opc.mnemonic, ABS)
			case UNDEFINED_X:
				opc, opcFindErr = readOpcode(pa.data.opc.mnemonic, ABSX)
			case UNDEFINED_Y:
				opc, opcFindErr = readOpcode(pa.data.opc.mnemonic, ABSY)
			default:
				panic("Cannot handle invalid mode")
			}
			if opcFindErr != nil {
				return nil, opcFindErr
			} else {
				// update opcode
				pa.data.opc = opc
			}
		}

		// write the opcode's hex value
		program = append(program, uint8(pa.data.opc.hex))

		// write the operand value(s)
		if !pa.skipOperand {

			// calculate offset for branch instructions
			if isBranchInstruction(pa.data.opc.mnemonic) {
				offset := calcBranchOffset(pa.addr, pa.data.opr.addr)
				if offset > 0xFF {
					return nil, fmt.Errorf("Branch too far: %d, offset: %d", pa.data.opr.addr, offset)
				}
				pa.data.opr.addr = offset
			}

			buffer.Reset()
			if bWriteErr := binary.Write(buffer, binary.LittleEndian, uint16(pa.data.opr.addr)); bWriteErr != nil {
				return nil, bWriteErr
			}
			if pa.data.opc.len == 2 { // on a 2 byte instruction the operand is 1 byte only
				program = append(program, buffer.Bytes()[0])
			} else {
				program = append(program, buffer.Bytes()...)
			}
		}
	}

	return program, nil
}

func saveSymbol(sym string, value int) error {
	if sym == "" {
		panic(value) // debug
	} else if sym[len(sym)-1] == ':' {
		// when it ends with ':' char remove it
		sym = sym[:len(sym)-1]
		if len(sym) == 0 {
			return fmt.Errorf("Invalid symbol definition (:)")
		}
	}
	if _, findErr := lookupSymbol(sym); findErr == nil {
		return fmt.Errorf("Symbol redefinition found for %s", sym)
	}
	symbols[sym] = value
	return nil
}

func lookupSymbol(sym string) (int, error) {
	if v, found := symbols[sym]; !found {
		return v, fmt.Errorf("Undefined symbol %s", sym)
	} else {
		return v, nil
	}
}
