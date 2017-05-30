# xbbasm: The seX-Bob-omB X-Assembler!

A cross-assembler targeting the 6502/6510 processor. Mainly written for learning purposes and to assist me in my experiments with assembly language programming for the Commodore 64. Its syntax, although very limited, borrows a bit from Merlin and from the C64 Macro Assembler Development System.

## Usage

Basic usage is extremely simple, just run:

    $ ./xbbasm program.asm

Optionally specify an output with:

    $ ./xbbasm -out b.prg program.asm

For maximum convenience **(!)** put the binary into your local `~/bin` and make sure it's in your `PATH`.

## Features

Some cool things you can do with it (a.k.a. _Do-s_ and _Don't-s_):

- Use include syntax anywhere in your file like:

    ./include screen.asm
    {... some code ...}
    ./include subs/math.asm
    {... some more code ...}
    ./include ../misc.asm

- Inline or off-line labels, just take into account that labels not on the same line need to end with `:`. For labels on the same line that is optional.

- It is allowed to enter label aliases (EQU in Merlin) between an off-line label and the next code line (see examples). This helps to put those below the subroutine name label but above the code and give it a more function-like look.

- For opcodes and operands syntax is case-insensitive.

## Notes

**About the examples:**

Example asm source files in the `/examples` directory were taken from different sources, credited when relevant, mostly to show what it can be done.

**A note on building from sources:**

For keeping the directory structure of the project, if you're using the default workspace in `$HOME/go` in Linux you can create a symbolic link like:

    ln -s /home/you/path/to/xbbasm/src/ /home/you/go/src/xbbasm




