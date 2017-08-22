# xbbasm: The seX-Bob-omB X-Assembler!

A cross-assembler targeting the 6502/6510 processor. Mainly written for learning purposes and to assist me in my experiments with assembly language programming for the Commodore 64. Its syntax, although very limited, borrows a bit from Merlin and from the C64 Macro Assembler Development System.

![Alt text](https://cloud.githubusercontent.com/assets/19293817/26608976/c1e64982-459e-11e7-9cca-cc81a561ba5b.jpg "we're sex bob-omb!")

## Usage

Basic usage is extremely simple, just run:

    $ ./xbbasm program.asm

Optionally specify an output with:

    $ ./xbbasm -out b.prg program.asm

For maximum convenience **(!)** put the binary into your local `~/bin` and make sure it's in your `PATH`.

## Features

Some cool things you can do with it (a.k.a. _Do-s_ and _Don't-s_):

- Use LISP style S-Expressions in operands like:
```
lda [+ color $27]
ldy [- color 1],x
```
- Use include syntax anywhere in your file like:
```
./include screen.asm

{... some code ...}

./include subs/math.asm

{... some more code ...}

./include ../misc.asm
```
- Insert binary files with `./bin {filename}`

- Inline or off-line labels, just take into account that labels not on the same line need to end with `:`. For labels on the same line that is optional.

- It is allowed to enter label aliases (EQU in Merlin) between an off-line label and the next code line (see examples). This helps to put those below the subroutine name label but above the code and give it a more function-like look.

- For opcodes and operands syntax is case-insensitive.

## Notes

**About the examples:**

Example asm source files in the `/examples` directory were taken from different sources, credited when relevant, mostly to show what it can be done.

**A note on building from sources:**

For keeping the directory structure of the project, if you're using the default workspace in `$HOME/go` in Linux you can create a symbolic link like:

    $ ln -s /home/you/path/to/xbbasm/src/ /home/you/go/src/xbbasm

## To-Dos:

- Document all the supported syntax so a user doesn't have to go through the sources to figure things not in the examples.
- Add more unit tests since at the moment only the most basic ones exist (no tests for expected errors for example).
- Would be nice to extend the `./bin` instruction to support offset and length params ACME style.
