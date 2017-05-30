;;;
;;; PRBYTE: print the byte in the Accumulator
;;;

PRBYTE:
	STA TEMP		; temp store number.
	LSR A			; shift the num in the
	LSR A			; accumulator four bits to the right.
	LSR A			; most significant nibble is zero.
	LSR A			; number is now in low nibble.
	JSR HEXCII		; convert number to ascii.
	JSR CHROUT		; output its code to the screen.

	LDA TEMP		; get orig number back in a.
	AND #$0F		; mask the most significant nibble.
	JSR HEXCII		; convert second digit to ascii.
	JSR CHROUT		; output its code.

	LDA #$20		; output ascii space by sending.
	JSR CHROUT		; space code to ouput routine.
	LDA TEMP		; restore the accumulator.
	RTS

;;;
;;; GETBYT: get a byte of data from the keyboard:
;;; 
	
GETBYT:
	JSR GETIN		; read the keyboard
	BEQ GETBYT		; wait for a non-zero result
	TAX			; save char in X
	JSR CHROUT		; output the keyboard character
	TXA			; get the char back from X

	JSR ASCHEX		; convert ascii to hex
	ASL A			; shift digit to the most
	ASL A			; significant nibble using four
	ASL A			; shifts left in the
	ASL A			; accumulator addressing mode
	STA TEMP		; temporarily store this result

LOAF:
	JSR GETIN		; get another character code
	BEQ LOAF
	TAX			; save it in X
	JSR CHROUT		; output
	TXA			; put back into Acc


	JSR ASCHEX		; convert it to hex
	ORA TEMP		; combine with the first digit
	RTS
