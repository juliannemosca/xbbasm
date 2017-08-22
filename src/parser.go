package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type parser struct {
	input     []string
	output    []tokenizedLine
	errors    []error
	fatal     error
	currFPath string
	tk        tokenizer
}

func beginParser(mainInput string) *parser {
	p := parser{}
	p.inputPush(mainInput)
	p.parse()
	return &p
}

func (p *parser) parse() {

	if p.fatal != nil {
		return
	}

	if input := p.inputPop(); input == nil {
		return
	} else {
		file, err := os.Open(*input)
		if err != nil {
			p.fatal = err
			return
		}

		p.currFPath = filepath.Dir(*input)

		// set the current filepath in the tokenizer
		// for including bin files
		p.tk.currFPath = p.currFPath

		fsc := bufio.NewScanner(file)

		var rawline string
		var partial string
		var lnum int

		for fsc.Scan() {
			lnum++
			rawline = fsc.Text()

			// discard comments and trim spaces
			cl := strings.TrimSpace(strings.Split(rawline, ";")[0])
			if len(cl) == 0 {
				continue
			}

			// parse line
			if strings.HasPrefix(strings.ToLower(cl), "./include ") {
				p.parseIncludeLine(cl, *input, lnum)
			} else {
				cl = fmt.Sprintf("%s%s", partial, cl)
				partial = p.parseCodeLine(cl, *input, lnum)
			}
		}

		file.Close()
	}

	p.parse()
	return
}

func (p *parser) parseIncludeLine(l, f string, lnum int) {
	// Note: this function has not been tested very carefully, should
	//       malfunction in file includes occur suspect this part of code 😬
	filename := l[(len("./include ") - 1):]
	if filename == "" || filename == "." || filename == ".." {
		p.errors = append(p.errors, fmt.Errorf("Invalid include statement in file %s line %d", f, lnum))
	} else {
		filename = strings.TrimSpace(filename)
		filename := fmt.Sprintf("%s%c%s", p.currFPath, filepath.Separator, filename)
		p.inputPush(filepath.Clean(filename))
	}
	return
}

func (p *parser) parseCodeLine(l, f string, lnum int) string {
	if tl, err := p.tk.tokenize(l); err != nil {
		p.errors = append(p.errors, fmt.Errorf("%s:%d:%s", f, lnum, err.Error()))
	} else if tl != nil {
		if tl.label != "" && tl.opc.mnemonic == "" {
			return tl.label + " "
		}
		p.outputPush(*tl)
	}
	return ""
}

func (p *parser) inputPush(i string) {
	p.input = append(p.input, i)
}

func (p *parser) inputPop() *string {
	if len(p.input) == 0 {
		return nil
	}
	i := p.input[0]
	p.input = p.input[1:]
	return &i
}

func (p *parser) outputPush(tl tokenizedLine) {
	p.output = append(p.output, tl)
}
