;;; Hex to ASCII conversion routine
HEXCII CMP #$0A	; is the bex digit >= 9 ?
	BCC AROUND	; no
	ADC #$06	; yes, add $07 (CARRY + $06)
AROUND ADC #$30	; add $30 to convert to ascii
	RTS

;;; ASCII to Hex conversion routine
ASCHEX cmp #$40
	BCC SKIP
	SBC #$07
SKIP	SEC
	SBC #$30
	RTS
