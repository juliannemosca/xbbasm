package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type assemblyLine struct {
	addr        int
	data        *tokenizedLine
	skipOperand bool
}

type segment struct {
	partiallyAssembled []assemblyLine
}

func assemble(programData []tokenizedLine) ([]byte, error) {

	var programSegments []segment
	var currentSegment segment
	var p *tokenizedLine

	startAddr := -1
	currentAddr := -1

	// first pass
	for i := 0; i < len(programData); i++ {
		p = &programData[i]

		// labels
		if p.label != "" {
			if currentAddr < 0 {
				// FIX-ME: if a label is added after the first
				//         .org instruction it will error here and
				//         it wouldn't be so clear for the user :/
				return nil, fmt.Errorf("No starting address found")
			} else if symErr := saveSymbol(p.label, currentAddr); symErr != nil {
				return nil, symErr
			}
		}

		// segment origin
		if p.opc.mnemonic == ".ORG" {
			if p.opr.label != "" {
				if addr, lookupErr := lookupSymbol(p.opr.label); lookupErr != nil {
					return nil, lookupErr
				} else {
					currentAddr = addr
				}
			} else {
				currentAddr = p.opr.addr
			}
			// set start address for program
			if startAddr == -1 {
				startAddr = currentAddr
			} else if startAddr > currentAddr {
				startAddr = currentAddr
			}
			// append last segment to program if contains any data
			if len(currentSegment.partiallyAssembled) > 0 {
				programSegments = append(programSegments, currentSegment)
			}
			currentSegment = segment{partiallyAssembled: []assemblyLine{}}
			continue
		}

		if currentAddr < 0 {
			return nil, fmt.Errorf("No starting address found")
		}

		// ./bin include command
		if p.opc.mnemonic == "./BIN" {
			data, binErr := binInclude(p.opr.label, &currentAddr)
			if binErr != nil {
				return nil, fmt.Errorf("Error when attempting to read %s : %s", p.opr.label, binErr)
			}
			currentSegment.partiallyAssembled = append(currentSegment.partiallyAssembled, data...)
			continue
		}

		// .TEXT and .DFB instructions
		if p.opc.mnemonic == ".TEXT" || p.opc.mnemonic == "DFB" {

			// for each byte just use an opcode with the
			// corresponding hex value
			for _, b := range p.opr.defBytes {
				if p.opc.mnemonic == ".TEXT" {
					b = encodeForC64Screen(b)
				}
				currentSegment.partiallyAssembled =
					append(
						currentSegment.partiallyAssembled,
						assemblyLine{
							addr:        currentAddr,
							data:        &tokenizedLine{opc: opcode{hex: b}},
							skipOperand: true,
						})
				currentAddr++
			}
			continue
		}

		// All other instructions:

		// create a new assembly line
		l := assemblyLine{}
		l.addr = currentAddr

		// flag operand skip according to mode.
		if p.opr.mode == IMP || p.opr.mode == ACC || p.opr.mode == NOMODE {
			l.skipOperand = true
		}

		// copy program data
		l.data = p

		// append line
		currentSegment.partiallyAssembled = append(currentSegment.partiallyAssembled, l)

		// add opcode len to current address
		currentAddr = currentAddr + p.opc.len
	}

	// add last segment

	if len(currentSegment.partiallyAssembled) > 0 {
		programSegments = append(programSegments, currentSegment)
	}

	// sort and flatten segments into partially assembled result

	pas, sortSegErr := sortSegments(startAddr, programSegments)
	if sortSegErr != nil {
		return nil, sortSegErr
	}

	// prepare buffer

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
	for i := 0; i < len(pas); i++ {
		pa = pas[i]

		// resolve symbols
		if pa.data.opr.label != "" {
			if pa.data.opr.label[0] == '[' {
				if v, formulaErr := resolveFormula(pa.data.opr.label); formulaErr != nil {
					return nil, formulaErr
				} else {
					pa.data.opr.addr = v
				}
			} else if v, lookupSymErr := lookupSymbol(pa.data.opr.label); lookupSymErr != nil {
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

func sortSegments(start int, segs []segment) ([]assemblyLine, error) {

	// This function puts all written memory addresses in a map
	// and adds padding as necessary between them

	// Since an `assemblyLine` already contains the values of 1 or more
	// memory addresses the memPholder struct has a flag that indicates
	// the type of value is holding (an instr or a placeholder).

	// When the complete result is partially assembled to be returned
	// here we skip the placeholders since the complete byte sequence
	// for all instructions is going to be written in the main assembly
	// process's second pass

	const padValue = 0x00
	const (
		instr = iota
		phold // placeholder for the whole length of an instruction
	)

	type memPholder struct {
		al     *assemblyLine
		holdAs int
	}

	psize := 0
	written := map[int]memPholder{}

	isWritten := func(addr int) bool {
		_, isW := written[addr]
		return isW
	}

	for i := 0; i < len(segs); i++ {
		psize += len(segs[i].partiallyAssembled)
		for iseg := 0; iseg < len(segs[i].partiallyAssembled); iseg++ {
			pa := segs[i].partiallyAssembled[iseg]
			if isWritten(pa.addr) {
				return nil, fmt.Errorf("Segments overlap at %d", pa.addr)
			}
			written[pa.addr] = memPholder{al: &pa, holdAs: instr}
			for iw := 1; iw < pa.data.opc.len; iw++ {
				// for each addr to be padded
				// insert a dummy placeholder
				written[pa.addr+iw] = memPholder{holdAs: phold}
			}
		}
	}

	result := []assemblyLine{}
	for ires := start; psize > 0; ires++ {
		if isWritten(ires) && written[ires].holdAs == instr {
			result = append(result, *written[ires].al)
			psize--
		} else if isWritten(ires) && written[ires].holdAs == phold {
			continue
		} else {
			result =
				append(
					result,
					assemblyLine{
						addr:        ires,
						data:        &tokenizedLine{opc: opcode{hex: padValue}},
						skipOperand: true,
					})
		}
	}

	return result, nil
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

func resetSymbols() {
	symbols = map[string]int{}
}

func binInclude(filename string, currentAddr *int) (data []assemblyLine, binErr error) {
	data = []assemblyLine{}

	bfile, binErr := os.Open(filename)
	if binErr != nil {
		return data, binErr
	}

	buffer := make([]byte, 1024)

	var readErr error
	var n int

	for true {
		if readErr == io.EOF {
			break
		}

		n, readErr = bfile.Read(buffer)
		if readErr != nil && readErr != io.EOF {
			return data, readErr
		}

		for i := 0; i < n; i++ {
			data = append(
				data,
				assemblyLine{
					addr:        *currentAddr,
					data:        &tokenizedLine{opc: opcode{hex: buffer[i]}},
					skipOperand: true,
				})
			*currentAddr++
		}
	}
	return data, binErr
}

// this function is exactly as the
// encoder_scr function from ACME to
// convert raw to C64 screencode
func encodeForC64Screen(b byte) byte {
	if (b >= 'a') && (b <= 'z') {
		// shift uppercase down
		return (b - 96)
	} else if (b >= '[') && (b <= '_') {
		// shift [\]^_ down
		return (b - 64)
	} else if b == '`' {
		// shift ` down
		return 64
	} else if b == '@' {
		// shift @ down
		return 0
	} else {
		return b
	}
}
