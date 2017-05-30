;;; ****************************************************************************
;;; Eight bit multiplication and division routines
;;; Assembly Language Programming, by Marvin L. DeJong
;;;
;;; ****************************************************************************

	.org	$c000

;;; Multiply an eight bit number by another
;;; and get a max. 15 bit result (32767).

multiply:

	multiplier 	= $00fc
	multiplicand	= $00fd

	prodlo = $00fe
	prodhi = $00ff

	lda #00
	sta prodlo
	ldy #8
m_up	lsr multiplier
	bcc m_down
	clc
	adc multiplicand
m_down	lsr a
	ror prodlo
	dey
	bne m_up
	sta prodhi
	rts

;;; Divide two eight bit numbers

divide:

	dividend	= $00fc
	divisor 	= $00fd
	quotient	= $00fe

	lda #00
	ldy #8
d_up	asl dividend
	rol a
	cmp divisor
	bcc d_down
	sbc divisor
	d_down	rol quotient
	dey
	bne d_up
	rts
