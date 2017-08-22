;; *****************************************************************************
;; This small program is just example 001 with a Basic Loader, but it takes
;; a couple of additional jumps in the middle, to test assembling segments
;; in different order.
;; *****************************************************************************

;; Second part:
;; 
	.org $c100

second_part:
	jsr $e544
	jmp third_part

;;
;; The original test program to output X
;;
;; First part:

	.org $c000
	jmp second_part

;; *****************************************************************************
;; A loader for Basic at address 0x801
;; *****************************************************************************

	.org $0801

	dfb $0d,$08,$d9,$07,$9e,$20,$34,$39
	dfb $31,$35,$32,$00,$00,$00
;; *****************************************************************************

;; Third part:
;;

	.org $c200

third_part:

	lda #88
	jsr $e716
	rts

	