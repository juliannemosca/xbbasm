
	.ORG 	$c000

	clear 	= $e544
	chrout	= $ffd2

jsr 	clear
ldx 	#$0

read	lda	msg,x
	jsr 	chrout
	cpx 	#21
	beq 	end
	inx
	jmp 	read
end 	rts
msg 	.text "COMMODORE 64 YEAH!!!!"
