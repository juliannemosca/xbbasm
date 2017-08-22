package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

// symbols table
var symbols map[string]int

func init() {
	resetSymbols()
}

func main() {

	// input and output filenames
	var input string
	var output *string

	output = flag.String("out", "a.prg", "output filename")
	flag.Parse()

	nonFlags := flag.Args()
	if len(nonFlags) == 0 {
		fail("error: must specify input file")
	} else {
		input = nonFlags[0]
	}

	// parse all files and tokenize
	p := beginParser(input)
	if p.fatal != nil {
		fail(p.fatal.Error())
	} else if len(p.errors) > 0 {
		for _, e := range p.errors {
			fmt.Fprintln(os.Stderr, e.Error())
		}
		os.Exit(1)
	}

	// assemble program
	if program, err := assemble(p.output); err != nil {
		fail(err.Error())
	} else if err := ioutil.WriteFile(*output, program, 0644); err != nil {
		fail(err.Error())
	} else {
		fmt.Println(fmt.Sprintf("%d bytes written to %s", len(program), *output))
	}

	return
}

func fail(errMsg string) {
	fmt.Fprintln(os.Stderr, fmt.Sprintf("error: %s", errMsg))
	os.Exit(1)
}
