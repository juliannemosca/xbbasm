package main

import (
	"testing"
)

func TestResolveFormula(t *testing.T) {
	tests := []struct {
		input  string
		result int
	}{{
		`[+ 3.4 2.6]`, 6,
	}, {
		`[- 2.6 0.6]`, 2,
	}, {
		`[+ 3.4 [- 2 [/ 0.8 2]]]`, 5,
	}, {
		`[/ 100 5]`, 20,
	}, {
		`[NOT b00000010]`, 253, // 253 = decimal for 1111 1101
	}, {
		`[NOT b10101010]`, 85, // 85 = decimal for 0101 0101
	}, {
		`[DIV 5 2]`, 2,
	}, {
		`[+ 0.5 [MOD 7 2]]`, 1,
	}, {
		`[>b 65535]`, 255,
	}, {
		`[<b 65535]`, 255,
	}, {
		`[AND 255 170]`, 170, // AND 1111 1111 WITH 1010 1010 = 1010 1010
	}}

	var r int
	var err error

	for _, test := range tests {
		r, err = resolveFormula(test.input)
		if err != nil {
			t.Errorf("failed on input %s with %s", test.input, err.Error())
		} else if uint8(r) != uint8(test.result) {
			t.Errorf(
				"%s: expected value %d (b %b) but got %d (b %b)",
				test.input,
				test.result, uint8(test.result),
				r, uint8(r))
		}
	}
}
