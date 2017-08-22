package main

import (
	"fmt"
	"strconv"
	"strings"
)

func resolveFormula(f string) (result int, err error) {
	_, expr, err := parseFormulaSubExpr(f)
	if err != nil {
		return result, err
	}

	var r interface{}
	r, err = evalFormula(expr)
	if err != nil {
		return result, err
	}

	if rUInt, ok := r.(uint); ok {
		result = int(rUInt)
	} else if rInt, ok := r.(int); ok && rInt >= 0 {
		result = int(rInt)
	} else if rFloat, ok := r.(float64); ok && rFloat >= 0 {
		result = int(rFloat)
	} else {
		err = fmt.Errorf("error: formula %s evaluates to unexpected result %v", f, r)
	}

	return result, err
}

// -----------------------------------------------------------------------------
// Evaluation
// -----------------------------------------------------------------------------

func evalFormula(expr []interface{}) (result interface{}, err error) {
	operation := expr[0].(string)
	arguments := expr[1:]

	for ind, arg := range arguments {
		if argarray, ok := arg.([]interface{}); ok {
			nestresult, err := evalFormula(argarray)
			if err != nil {
				fmt.Printf("error: Nested operation '%v' evaluation failed", arg)
			}
			arguments[ind] = nestresult
		}
	}

	operation = strings.ToUpper(operation)
	if formulaOperators[operation] != nil {
		return formulaOperators[operation](operation, arguments)
	}

	return nil, fmt.Errorf("error: operation '%s' is not defined", operation)
}

// -----------------------------------------------------------------------------
// Parser:
// -----------------------------------------------------------------------------

func parseFormula(input string) (unparsed string, expr interface{}, err error) {
	// remove all whitespace or newline at the beginning
	input = strings.TrimLeft(input, "\t\r\n ")

	if len(input) == 0 {
		// if nothing's left return empty unparsed
		return unparsed, nil, nil
	} else if input[0] == '[' {
		// if it opens a bracket parse the expression inside
		return parseFormulaSubExpr(input)
	} else {
		return parseFormulaAtom(input)
	}
}

func parseFormulaAtom(input string) (unparsed string, expr interface{}, err error) {
	var atom = ""
	for p := 0; p < len(input); p++ {
		if isWhitespace(input[p]) || isEndOfAtom(input[p]) {
			atom = input[:p]
			unparsed = input[p:]
			break
		}
	}

	if atom == "" {
		atom = input
	}

	// check if it is an operator
	if formulaOperators[atom] != nil {
		return unparsed, atom, nil
	}

	n, s, err := readAddress(atom)
	if err != nil {
		return unparsed, expr, err
	}

	if n < 0 {
		// check if it is a float before assuming it's a symbol
		var flVal float64
		flVal, err = strconv.ParseFloat(s, 64)
		if conversionError, ok := err.(*strconv.NumError); ok {
			if conversionError.Err == strconv.ErrSyntax {
				// it's a symbol
				return unparsed, s, nil
			}
		}
		return unparsed, flVal, err
	}

	return unparsed, n, err
}

func parseFormulaSubExpr(input string) (unparsed string, expr []interface{}, err error) {
	expr = []interface{}{}

	// skip opening '['
	input = input[1:]

	// parse input expr
	for p := 0; p < len(input); p++ {

		if isWhitespace(input[p]) {
			continue
		} else if input[p] == ']' {
			return input[p+1:], expr, err
		} else {
			remaining, parsedExpr, err := parseFormula(input[p:])
			if err != nil {
				return remaining, expr, err
			}

			expr = append(expr, parsedExpr)
			input = remaining
			p = -1
		}
	}

	return unparsed, expr, fmt.Errorf("syntax error, missing expected ']' before end of input")
}

func isEndOfAtom(c byte) bool {
	return c == ']'
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t'
}
