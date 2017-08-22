package main

import (
	"fmt"
	"math"
)

type operationImpl func(operator string, arguments []interface{}) (interface{}, error)

var formulaOperators map[string]operationImpl = map[string]operationImpl{

	// Basic Arithmetic and Math
	"*":   basicArithmetic,
	"/":   basicArithmetic,
	"DIV": basicArithmetic,
	"+":   basicArithmetic,
	"-":   basicArithmetic,
	"^":   exponent,
	"%":   modulo,
	"MOD": modulo,

	// Shifts
	"ASL": shiftLeft,
	"LSL": shiftLeft,
	"<<":  shiftLeft,

	"ASR": shiftRight,
	"LSR": shiftRight,
	">>":  shiftRight,
	">>>": shiftRight,

	// Byte masks
	"<B": lowbyte,
	">B": highbyte,
	"^B": bankbyte,

	// Comparison
	"<=": eqLower,
	"<":  lower,
	">=": eqHigher,
	">":  higher,
	"!=": neq,
	"<>": neq,
	"><": neq,
	"=":  eq,

	// Bit-wise logical
	"&":   and,
	"AND": and,
	"|":   or,
	"OR":  or,
	"XOR": xor,
	"EOR": xor,
	"!":   complement,
	"NOT": complement,
}

func frmOperCheckArgs(operation string, arguments []interface{}, n int) error {
	if len(arguments) != n {
		return fmt.Errorf("invalid number of arguments for %s: %T. Expected %d args. and got %d", operation, arguments, n, len(arguments))
	}
	return nil
}

func frmUInt(n interface{}) (uint, error) {
	if uintVal, ok := n.(uint); ok {
		return uintVal, nil
	} else if intVal, ok := n.(int); ok && intVal >= 0 {
		return uint(intVal), nil
	} else if symbol, ok := n.(string); ok {
		intVal, err := lookupSymbol(symbol)
		if err != nil {
			return 0, err
		} else {
			return frmUInt(intVal)
		}
	}
	return 0, fmt.Errorf("%v is not an unsigned integer", n)
}

func frmInt(n interface{}) (int, error) {
	if intVal, ok := n.(int); ok {
		return intVal, nil
	} else if symbol, ok := n.(string); ok {
		return lookupSymbol(symbol)
	}
	return 0, fmt.Errorf("%v is not an integer", n)
}

func frmFloat(f interface{}) (float64, error) {
	if flVal, ok := f.(float64); ok {
		return flVal, nil
	} else if intVal, ok := f.(int); ok {
		return float64(intVal), nil
	} else if symbol, ok := f.(string); ok {
		intVal, err := lookupSymbol(symbol)
		if err != nil {
			return 0, err
		} else {
			return float64(intVal), nil
		}
	}
	return 0, fmt.Errorf("%v is not a numeric value", f)
}

// -----------------------------------------------------------------------------
// Operators
// -----------------------------------------------------------------------------

func exponent(operation string, arguments []interface{}) (result interface{}, err error) {
	var val1, val2 float64
	err = frmOperCheckArgs(operation, arguments, 2)
	if err != nil {
		return 0, err
	}
	val1, err = frmFloat(arguments[0])
	if err != nil {
		return 0, err
	}
	val2, err = frmFloat(arguments[1])
	if err != nil {
		return 0, err
	}
	return math.Pow(val1, val2), nil
}

func modulo(operation string, arguments []interface{}) (result interface{}, err error) {
	// Modulus division is only defined for integers
	var val1, val2 int
	err = frmOperCheckArgs(operation, arguments, 2)
	if err != nil {
		return 0, err
	}
	// Modulus division is only defined for integers
	val1, err = frmInt(arguments[0])
	if err != nil {
		return 0, err
	}
	val2, err = frmInt(arguments[1])
	if err != nil {
		return 0, err
	}
	return (val1 % val2), nil
}

func basicArithmetic(operation string, arguments []interface{}) (result interface{}, err error) {
	terms := []float64{}
	for _, arg := range arguments {
		if t, terr := frmFloat(arg); terr != nil {
			return terms, terr
		} else {
			terms = append(terms, t)
		}
	}
	if err != nil {
		return 0, err
	}

	var res float64

	switch operation {
	// Multiply
	case "*":
		res = 1.0
		for _, t := range terms {
			res = res * t
		}
		break
	// Divide
	case "/":
		fallthrough
	// Integer Divide
	case "DIV":
		res = terms[0]
		for i := 1; i < len(terms); i++ {
			res = res / terms[i]
		}

		if operation == "DIV" {
			return int(res), nil
		}
		break
	// Add
	case "+":
		res = 0.0
		for _, t := range terms {
			res = res + t
		}
		break

	// Negate or substract
	case "-":
		res = terms[0]
		for i := 1; i < len(terms); i++ {
			res = res - terms[i]
		}
		break

	default:
		panic("invalud operation passed as argument")
	}
	return res, nil
}

func shiftLeft(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := shiftOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 << val2), nil
}

func shiftRight(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := shiftOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 >> val2), nil
}

func lowbyte(operation string, arguments []interface{}) (result interface{}, err error) {
	val, err := byteMaskOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	// mask against 255 to clear the high byte
	return (val & 0xff), nil
}

func highbyte(operation string, arguments []interface{}) (result interface{}, err error) {
	val, err := byteMaskOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	// mask against 255 to clear the low byte
	return (val >> 8), nil
}

func bankbyte(operation string, arguments []interface{}) (result interface{}, err error) {
	val, err := byteMaskOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	// shift 16 places to get bits 16 to 23
	return (val >> 16), nil
}

func eq(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := comparisonOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 == val2), nil
}

func eqHigher(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := comparisonOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 >= val2), nil
}

func eqLower(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := comparisonOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 <= val2), nil
}

func higher(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := comparisonOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 > val2), nil
}

func lower(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := comparisonOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 < val2), nil
}

func neq(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := comparisonOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 != val2), nil
}

func and(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := bitwiseOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 & val2), nil
}

func or(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := bitwiseOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 | val2), nil
}

func xor(operation string, arguments []interface{}) (result interface{}, err error) {
	val1, val2, err := bitwiseOperationArgs(operation, arguments)
	if err != nil {
		return 0, err
	}
	return (val1 ^ val2), nil
}

func complement(operation string, arguments []interface{}) (result interface{}, err error) {
	err = frmOperCheckArgs(operation, arguments, 1)
	if err != nil {
		return 0, err
	}
	val, err := frmUInt(arguments[0])
	if err != nil {
		return 0, err
	}
	return (^val), nil
}

// -----------------------------------------------------------------------------
// General operation argument helpers:
// -----------------------------------------------------------------------------

func shiftOperationArgs(operation string, arguments []interface{}) (val1 int, val2 uint, err error) {
	err = frmOperCheckArgs(operation, arguments, 2)
	if err != nil {
		return 0, 0, err
	}
	// The signed/unsignedness of the left side
	// arg. will determine if the shift is
	// arithmetic (for signed) of logical (for unsigned)
	val1, err = frmInt(arguments[0])
	if err != nil {
		return 0, 0, err
	}
	val2, err = frmUInt(arguments[1])
	if err != nil {
		return 0, 0, err
	}
	return val1, val2, nil
}

func byteMaskOperationArgs(operation string, arguments []interface{}) (val uint, err error) {
	err = frmOperCheckArgs(operation, arguments, 1)
	if err != nil {
		return 0, err
	}
	val, err = frmUInt(arguments[0])
	if err != nil {
		return 0, err
	}
	if operation == "<B" && val > 0xffff {
		return 0, fmt.Errorf("Cannot get the low byte of %#x", val)
	} else if operation == ">B" && val > 0xffff {
		return 0, fmt.Errorf("Cannot get the high byte of %#x", val)
	}
	return val, err
}

func bitwiseOperationArgs(operation string, arguments []interface{}) (val1, val2 int, err error) {
	err = frmOperCheckArgs(operation, arguments, 2)
	if err != nil {
		return 0, 0, err
	}
	val1, err = frmInt(arguments[0])
	if err != nil {
		return 0, 0, err
	}
	val2, err = frmInt(arguments[1])
	if err != nil {
		return 0, 0, err
	}
	return val1, val2, nil
}

func comparisonOperationArgs(operation string, arguments []interface{}) (val1, val2 float64, err error) {
	err = frmOperCheckArgs(operation, arguments, 2)
	if err != nil {
		return 0, 0, err
	}
	val1, err = frmFloat(arguments[0])
	if err != nil {
		return 0, 0, err
	}
	val2, err = frmFloat(arguments[1])
	if err != nil {
		return 0, 0, err
	}
	return val1, val2, nil
}
