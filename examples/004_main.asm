;;; ****************************************************************************
;;; An example source split into multiple files, from the book
;;; Assembly Language Programming, by Marvin L. DeJong
;;;
;;; ****************************************************************************
./include 004_conversion.asm
./include 004_io.asm
	
	.org $c000

	GETIN = $ffe4
	CHROUT = 65490 		; ($FFD2)
	TEMP = $02

MAIN	JSR GETBYT		; get a number to put in A
	STA TEMP		; save it in temp
	LDA #$20		; output a space
	JSR CHROUT
	LDA TEMP		; get the number back from temp
	JSR PRBYTE		; output it as two hex digits
	CLV			; force a branch
	BVC MAIN		; stay in this loop forever
