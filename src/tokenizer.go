package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

type operand struct {
	addr     int
	label    string
	mode     string
	defBytes []byte
}

type tokenizedLine struct {
	label string
	opc   opcode
	opr   operand
}

type tokenizer struct {
	currFPath string
}

func (t tokenizer) tokenize(l string) (*tokenizedLine, error) {
	tl := tokenizedLine{}
	tokens := splitTokens(l)
	tnum := len(tokens)

	// handle case when an off line label
	// is appended to an alias label in the
	// next line
	var unhandled *tokenizedLine
	if tnum == 4 && tokens[2] == "=" {
		unhandled = &tokenizedLine{label: tokens[0]}
		tokens = tokens[1:]
		tnum--
	}

	if tnum > 3 || tnum == 0 { // == 0 should never happen anyway
		return nil, fmt.Errorf("Syntax error")
	} else if tnum == 1 {

		// non-inline labels have to end with `:`
		// (but inline labels don't!).
		if tokens[0][len(tokens[0])-1] == ':' {
			tl.label = tokens[0]
			return &tl, nil
		}

		// this is just a single opcode since currently
		// no single field pseudo opcodes are implemented
		opc, opcErr := readOpcode(tokens[0], IMP)
		if opcErr != nil {
			return nil, opcErr
		}
		tl.opc = opc

	} else if tnum == 2 {

		// check if it's an instruction (pseudo opcode)
		if inst, instOk, instErr := t.tryTokenizeInstruction(tokens, tnum); instErr != nil {
			return nil, instErr
		} else if instOk {
			return inst, nil
		}

		// this one is tricky, may be:
		// LABEL-OPCODE or OPCODE-OPERAND
		// so we'll try each one

		// first see if the last token is an operand
		if isOpcode(tokens[0]) {
			opr, oprErr := t.readOperand(tokens[1], tokens[0])
			if oprErr != nil {
				return nil, oprErr
			}
			tl.opr = *opr
			// if it is first token has to be an opcode
			opc, opcErr := readOpcode(tokens[0], tl.opr.mode)
			if opcErr != nil {
				return nil, opcErr
			}
			tl.opc = opc
		} else {
			// second token should be an opcode now
			opc, opcErr := readOpcode(tokens[1], IMP) // addr mode has to be IMPLIED
			if opcErr != nil {
				return nil, opcErr
			}
			tl.opc = opc
			tl.label = tokens[0]
		}

	} else if tnum == 3 {

		// check if it's an instruction (pseudo opcode)
		if inst, instOk, instErr := t.tryTokenizeInstruction(tokens, tnum); instErr != nil {
			return nil, instErr
		} else if instOk {
			return inst, nil
		}

		opr, oprErr := t.readOperand(tokens[2], tokens[1])
		if oprErr != nil {
			return nil, oprErr
		}
		tl.opr = *opr
		tl.label = tokens[0]

		// check if it's an alias, like for example
		// `SPRITE0 = $7F8`
		if tokens[1] == "=" {
			if tl.opr.label != "" {
				return nil, fmt.Errorf("Cannot use label as an alias")
			}
			// eagerly save symbols for aliases
			// and return any unhandled partial
			return unhandled, saveSymbol(tl.label, opr.addr)
		} else {
			opc, opcErr := readOpcode(tokens[1], tl.opr.mode)
			if opcErr != nil {
				return nil, opcErr
			}
			tl.opc = opc
		}

	}
	return &tl, nil
}

// To-Do: using `instruction` as a synonym for pseudoOpcode is innacurate, must rename.
//
func (t tokenizer) tryTokenizeInstruction(tokens []string, tnum int) (*tokenizedLine, bool, error) {

	var tl tokenizedLine
	var opc string
	var opr string

	switch tnum {
	case 1:
		// currently no instruction fits in 1 single token
		return nil, false, nil
	case 2:
		// only 2 fields possibilities now would be:
		// .TEXT {text},
		// .ORG {addr}
		// DFB {data...}
		// so it's OPCODE-OPERAND
		//
		// also ./bin goes in here and passed
		// like that all the way through
		// for the assembler to resolve
		opc = tokens[0]
		opr = tokens[1]
	case 3:
		tl.label = tokens[0]
		opc = tokens[1]
		opr = tokens[2]
	default:
		// not handle this
		return nil, false, nil
	}

	if poc := readPseudoOpcode(opc); poc != nil {
		tl.opc = *poc
		opr, oprErr := t.readPseudoOperand(opr, opc)
		if oprErr != nil {
			return nil, true, oprErr
		} else if opr == nil {
			return nil, true, nil
		}
		tl.opr = *opr
		return &tl, true, nil
	}
	return nil, false, nil
}

// -----------------------------------------------------------------------------
// Read opcodes & pseudo-opcodes:
// -----------------------------------------------------------------------------

func readOpcode(oc, mode string) (opcode, error) {
	if mode == UNDEFINED || mode == UNDEFINED_X || mode == UNDEFINED_Y {
		// ok, we believe it's a valid opcode for now
		// but we won't support labels for zero page modes,
		// so assign a length of 3
		return opcode{mnemonic: oc, len: 3, mode: mode}, nil
	}
	lookupKey := fmt.Sprintf("%s-%s", strings.ToUpper(oc), mode)
	roc, found := opcodes[lookupKey]
	if !found {
		return opcode{}, fmt.Errorf("error: %s (mode: %s) is not a valid opcode", oc, mode)
	}
	return roc, nil
}

func readPseudoOpcode(poc string) *opcode {
	lookupKey := fmt.Sprintf(strings.ToUpper(poc))
	rpoc, found := pseudoOpcodes[lookupKey]
	if !found {
		return nil
	}
	return &rpoc
}

// -----------------------------------------------------------------------------
// Read operands:
// -----------------------------------------------------------------------------

func readAddress(a string) (int, string, error) {

	// Supported address formats:
	//
	// hex:	$FFFF
	// oct: 0177777
	// bin  b1111111111111111
	// dec:	65535
	//

	// gotcha: ParseInt always returns an int64
	// regardless the bitSize parameter
	var val int64
	var err error

	if a[0] == '$' {
		// since we can't specify 'unsigned' use 32 bit int
		val, err = strconv.ParseInt(a[1:], 16, 32)
	} else if a[0] == 'o' {
		val, err = strconv.ParseInt(a[1:], 8, 32)
	} else if a[0] == 'b' {
		val, err = strconv.ParseInt(a[1:], 2, 32)
	} else {
		val, err = strconv.ParseInt(a, 10, 32)
		if conversionError, ok := err.(*strconv.NumError); ok {
			if conversionError.Err == strconv.ErrSyntax {
				if isOpcode(a) {
					return 0, "", fmt.Errorf("Cannot use opcode %s as a label", a)
				}
				// it's a label or a formula
				return -1, a, nil
			}
		}
	}
	return int(val), "", err
}

func readIndirect(ro string) (*operand, error) {

	var opr operand
	var addrVal int
	var addrLabel string
	var addrErr error

	if strings.HasSuffix(ro, ")") {
		noParens := ro[1 : len(ro)-2]
		if sploper := strings.Split(noParens, ","); len(sploper) > 2 {
			return nil, fmt.Errorf("Syntax error in operand %s", ro)
		} else if len(sploper) == 2 {
			if strings.ToUpper(sploper[1]) != "X" {
				return nil, fmt.Errorf("Invalid register in operand %s. Expecting register X", ro)
			}
			// Indirect,X
			addrVal, addrLabel, addrErr = readAddress(sploper[0])
			opr.mode = IX
		} else {
			// Indirect
			addrVal, addrLabel, addrErr = readAddress(sploper[0])
			opr.mode = IND
		}
	} else {
		if sploper := strings.Split(ro, ","); len(sploper) != 2 {
			return nil, fmt.Errorf("Syntax error in operand %s", ro)
		} else if strings.ToUpper(sploper[1]) != "Y" {
			return nil, fmt.Errorf("Invalid register in operand %s. Expecting register Y", ro)
		} else {
			if !strings.HasSuffix(sploper[0], ")") {
				return nil, fmt.Errorf("Syntax error, missing ')' in operand %s", ro)
			}
			noParens := sploper[0][1 : len(sploper[0])-2]
			// Indirect,Y
			addrVal, addrLabel, addrErr = readAddress(noParens)
			opr.mode = IY
		}
	}

	if addrErr != nil {
		return nil, addrErr
	}

	opr.addr = addrVal
	opr.label = addrLabel
	return &opr, nil
}

func (t tokenizer) readOperand(rawoper, opc string) (*operand, error) {

	// Syntax examples for the addressing modes:
	//
	// Implied       {opcode}
	// Relative      {opcode} $44
	// Accumulator   {opcode} A
	// Immediate     {opcode} #$44
	// Zero Page     {opcode} $44
	// Zero Page,X   {opcode} $44,X
	// Zero Page,Y   {opcode} $44,Y
	// Absolute      {opcode} $4400
	// Absolute,X    {opcode} $4400,X
	// Absolute,Y    {opcode} $4400,Y
	// Indirect      {opcode} ($5597)
	// Indirect,X    {opcode} ($44,X)
	// Indirect,Y    {opcode} ($44),Y

	if rawoper == "" {
		// Implied
		return &operand{addr: 0, mode: IMP}, nil
	} else if strings.ToUpper(rawoper) == "A" {
		// Accumulator
		return &operand{addr: 0, mode: ACC}, nil
	} else if rawoper[0] == '#' {
		if addrVal, addrLabel, addrErr := readAddress(rawoper[1:]); addrErr != nil {
			return nil, addrErr
		} else if addrVal > 0xFF {
			return nil, fmt.Errorf("Out of range value in operand %s", rawoper)
		} else {
			// Immediate
			return &operand{addr: addrVal, label: addrLabel, mode: IM}, nil
		}
	} else if rawoper[0] == '(' {
		return readIndirect(rawoper[1:])
	} else if sploper := strings.Split(rawoper, ","); len(sploper) > 2 {
		return nil, fmt.Errorf("Syntax error in operand %s", rawoper)
	} else if len(sploper) == 2 {
		if register := strings.ToUpper(sploper[1]); register != "X" && register != "Y" {
			return nil, fmt.Errorf("Invalid register in operand %s. Expecting registers X or Y", rawoper)
		} else if addrVal, addrLabel, addrErr := readAddress(sploper[0]); addrErr != nil {
			return nil, addrErr
		} else if addrLabel != "" {
			// same as further below, try to handle symbols for aliases here
			// or leave undefined. later if undefined turns out to be any of the
			// zero page modes it will error in the assembler, since we reserve
			// space for absolute modes (3 byte instructions)
			symAddr, lookupSymErr := lookupSymbol(addrLabel)
			if lookupSymErr != nil {
				if register == "X" {
					return &operand{addr: addrVal, label: addrLabel, mode: UNDEFINED_X}, nil
				} else {
					return &operand{addr: addrVal, label: addrLabel, mode: UNDEFINED_Y}, nil
				}
			} else if symAddr <= 0xFF {
				if register == "X" {
					// Zero Page,X
					return &operand{addr: symAddr, label: addrLabel, mode: ZPX}, nil
				} else {
					// Zero Page,Y
					return &operand{addr: symAddr, label: addrLabel, mode: ZPY}, nil

				}
			} else if symAddr <= 0xFFFF {
				if register == "X" {
					// Absolute,X
					return &operand{addr: symAddr, label: addrLabel, mode: ABSX}, nil
				} else {
					// Absolute,Y
					return &operand{addr: symAddr, label: addrLabel, mode: ABSY}, nil
				}
			} else {
				return nil, fmt.Errorf("Out of range value in operand %s", rawoper)
			}
		} else if addrVal <= 0xFF {
			if register == "X" {
				// Zero Page,X
				return &operand{addr: addrVal, label: addrLabel, mode: ZPX}, nil
			} else {
				// Zero Page,Y
				return &operand{addr: addrVal, label: addrLabel, mode: ZPY}, nil
			}
		} else if addrVal <= 0xFFFF {
			if register == "X" {
				// Absolute,X
				return &operand{addr: addrVal, label: addrLabel, mode: ABSX}, nil
			} else {
				// Absolute,Y
				return &operand{addr: addrVal, label: addrLabel, mode: ABSY}, nil
			}
		} else {
			return nil, fmt.Errorf("Out of range value in operand %s", rawoper)
		}
	} else {
		if addrVal, addrLabel, addrErr := readAddress(sploper[0]); addrErr != nil {
			return nil, addrErr
		} else if opc != "" && isBranchInstruction(opc) {
			// Relative
			return &operand{addr: addrVal, label: addrLabel, mode: RL}, nil
		} else if addrLabel != "" {
			// for labels in this modes try to handle zero page symbol references
			// now, if not it should be absolute. if mode is zero page and the label is
			// not yet defined it will fail properly when assembling
			symAddr, lookupSymErr := lookupSymbol(addrLabel)
			if lookupSymErr != nil {
				return &operand{addr: addrVal, label: addrLabel, mode: UNDEFINED}, nil
			} else if symAddr <= 0xFF {
				// Zero Page
				return &operand{addr: symAddr, label: addrLabel, mode: ZP}, nil
			} else if symAddr <= 0xFFFF {
				// Absolute
				return &operand{addr: symAddr, label: addrLabel, mode: ABS}, nil
			} else {
				return nil, fmt.Errorf("Out of range value in operand %s", rawoper)
			}
		} else if addrVal <= 0xFF {
			// Zero Page
			return &operand{addr: addrVal, label: addrLabel, mode: ZP}, nil
		} else if addrVal <= 0xFFFF {
			// Absolute
			return &operand{addr: addrVal, label: addrLabel, mode: ABS}, nil
		} else {
			return nil, fmt.Errorf("Out of range value in operand %s", rawoper)
		}
	}
	return nil, fmt.Errorf("Syntax error in operand %s", rawoper)
}

func (t tokenizer) readPseudoOperand(rawoper, opc string) (*operand, error) {
	switch strings.ToUpper(opc) {
	case ".ORG":
		addrVal, addrLabel, addrErr := readAddress(rawoper)
		if addrErr != nil {
			return nil, fmt.Errorf("Could not read start address %s", rawoper)
		} else {
			return &operand{addr: addrVal, label: addrLabel, mode: NOMODE}, nil
		}
	case ".TEXT":
		if !isAsciiString(rawoper) {
			return nil, fmt.Errorf("%s is not a valid ASCII text", rawoper)
		}
		return &operand{defBytes: []byte(rawoper), mode: NOMODE}, nil
	case "DFB":
		chunks := strings.Split(rawoper, ",")
		dfbValues := []byte{}
		for _, c := range chunks {
			if c == "" {
				continue
			}
			value, text, dataErr := readAddress(c)
			if dataErr != nil {
				return nil, fmt.Errorf("Syntax error in %s on DFB instruction", c)
			} else if text != "" {
				return nil, fmt.Errorf("Syntax error in %s on DFB instruction", c)
			} else if value > 0xFF {
				return nil,
					fmt.Errorf(
						"Value %s is out of range. Enter only values valid for a byte's range", c)
			} else {
				dfbValues = append(dfbValues, byte(value))
			}
		}
		if len(dfbValues) == 0 {
			return nil, fmt.Errorf("No valid data found for DFB instruction")
		}
		return &operand{defBytes: dfbValues, mode: NOMODE}, nil
	case "./BIN":
		return &operand{label: fmt.Sprintf("%s%c%s", t.currFPath, filepath.Separator, rawoper), mode: NOMODE}, nil
	default:
		// this should never be reached
		panic(fmt.Sprintf("Unrecognized pseudo-opcode %s", opc))
	}
}

// -----------------------------------------------------------------------------
// Misc. helpers:
// -----------------------------------------------------------------------------

func splitTokens(line string) []string {

	// This takes the line with the spaces
	// before and after already trimmed so we only
	// have to worry about the spaces in between and the
	// double quoted text for .TEXT instructions

	var p byte
	var i int
	var readingFormula int

	toks := []string{}
	tok := ""

	for i < len(line) {
		p = line[i]
		if readingFormula == 0 && isWhitespaceOrTab(p) && tok == "" {
			i++
			continue
		} else if readingFormula == 0 && isWhitespaceOrTab(p) {
			toks = append(toks, tok)
			tok = ""
			i++
			continue
		} else if p == '"' {
			// if we find a double quote take all the rest
			// as text and also remove any double quote
			// at the end

			if tok != "" {
				toks = append(toks, tok)
			}

			tok = line[i+1:]
			i += len(line) - i

			// note: this hacky solution has the pretty funny
			// effect that you could totally ommit the " at the end
			//
			// 	¯\_(ツ)_/¯
			//
			if tok[len(tok)-1] == '"' {
				tok = tok[:len(tok)-1]
			}

			toks = append(toks, tok)
			tok = ""
			continue
		} else if p == '[' {
			readingFormula++
		} else if p == ']' {
			readingFormula--
		}

		tok += string(p)
		i++
	}

	if tok != "" {
		toks = append(toks, tok)
	}

	return toks
}

func isWhitespaceOrTab(b byte) bool {
	return b == ' ' || b == '\t'
}

func isAsciiString(s string) bool {
	bytes := []byte(s)
	for _, b := range bytes {
		if b > 127 {
			return false
		}
	}
	return true
}

// Not the most efficient or elegant solution here:
//
var cachedOpcodes map[string]bool

func isOpcode(oc string) bool {
	if len(cachedOpcodes) == 0 {
		cachedOpcodes = map[string]bool{}
		for _, oc := range opcodes {
			if _, wasAdded := cachedOpcodes[strings.ToUpper(oc.mnemonic)]; !wasAdded {
				cachedOpcodes[oc.mnemonic] = false
			}
		}
	}
	_, ok := cachedOpcodes[strings.ToUpper(oc)]
	return ok
}
