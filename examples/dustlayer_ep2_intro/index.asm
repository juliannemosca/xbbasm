;============================================================
; Example Project for C64 Tutorials  
; Code by actraiser/Dustlayer
; Music: Ikari Intro by Laxity
;
; Simple Colorwash effect with a SID playing
;
; Tutorial: http://dustlayer.com/c64-coding-tutorials/2013/2/17/a-simple-c64-intro
; Dustlayer WHQ: http://dustlayer.com
;============================================================

;============================================================
; index file which loads all source code and resource files
;============================================================

;============================================================
; BASIC loader with start address $c000
;============================================================

.org $0801                            ; BASIC start address (#2049)
dfb $0d,$08,$dc,$07,$9e,$20,$34,$39   ; BASIC loader to start at $c000...
dfb $31,$35,$32,$00,$00,$00           ; puts BASIC line 2012 SYS 49152
.org $c000                            ; start address for 6502 code

;============================================================
;  Main routine with IRQ setup and custom IRQ routine
;============================================================

./include code/main.asm

;============================================================
;    setup and init symbols we use in the code
;============================================================

./include code/setup_symbols.asm
	
;============================================================
; tables and strings of data 
;============================================================

./include code/data_static_text.asm
./include code/data_colorwash.asm

;============================================================
; one-time initialization routines
;============================================================

./include code/init_clear_screen.asm
./include code/init_static_text.asm

;============================================================
;    subroutines called during custom IRQ
;============================================================

./include code/sub_colorwash.asm
./include code/sub_music.asm

;============================================================
; load resource files (for this small intro its just the sid)
;============================================================

./include code/load_resources.asm
