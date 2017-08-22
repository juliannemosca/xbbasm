package main

import (
	"fmt"
	"io"
	"os"
	"testing"
)

func TestEverything(t *testing.T) {
	tests := []struct {
		input  string
		output string
	}{{
		"../examples/001_printx.asm",
		"../test_binaries/001.prg",
	}, {
		"../examples/002_sprite_and_sound.asm",
		"../test_binaries/002.prg",
	}, {
		"../examples/003_colors.asm",
		"../test_binaries/003.prg",
	}, {
		"../examples/004_main.asm",
		"../test_binaries/004.prg",
	}, {
		"../examples/005_mult_and_div.asm",
		"../test_binaries/005.prg",
	}, {
		"../examples/006_text.asm",
		"../test_binaries/006.prg",
	}, {
		"../examples/007_loader.asm",
		"../test_binaries/007.prg",
	}, {
		"../examples/dustlayer_ep2_intro/index.asm",
		"../test_binaries/dustlayer_ep2_intro.prg",
	}}

	var apLen, exOLen, minLen int

	for _, test := range tests {

		resetSymbols()

		p := beginParser(test.input)
		if p.fatal != nil {
			t.Errorf(fmt.Sprintf("%s: %s", test.input, p.fatal))
			continue
		} else if len(p.errors) > 0 {
			for _, e := range p.errors {
				t.Errorf(fmt.Sprintf("%s: %s", test.input, e.Error()))
			}
			continue
		}

		program, err := assemble(p.output)
		if err != nil {
			t.Errorf("%s: %s", test.input, err.Error())
			continue
		}

		expected, err := readBinaryFile(test.output)
		if err != nil {
			t.Fatal(fmt.Sprintf("%s: %s", test.input, err.Error()))
		}

		minLen = 0
		apLen = len(program)
		exOLen = len(expected)

		if apLen != exOLen {
			t.Errorf("%s: assembled program size is %d but expected size is %d", test.input, apLen, exOLen)
			if apLen > exOLen {
				minLen = exOLen
			} else {
				minLen = apLen
			}
		} else {
			minLen = apLen
		}

		for i := 0; i < minLen; i++ {
			if program[i] != expected[i] {
				t.Errorf(
					"%s: expected value %x but got %x at position %d",
					test.input,
					expected[i],
					program[i],
					i)
			}
		}

	}
}

func readBinaryFile(filename string) (data []byte, binErr error) {
	data = []byte{}

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

		data = append(data, buffer[:n]...)
	}
	return data, binErr
}
