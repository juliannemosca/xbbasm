; color data table
; first 9 rows (40 bytes) are used for the color washer
; on start the gradient is done by byte 40 is mirroed in byte 1, byte 39 in byte 2 etc... 

color        dfb $09,$09,$02,$02,$08 
             dfb $08,$0a,$0a,$0f,$0f 
             dfb $07,$07,$01,$01,$01 
             dfb $01,$01,$01,$01,$01 
             dfb $01,$01,$01,$01,$01 
             dfb $01,$01,$01,$07,$07 
             dfb $0f,$0f,$0a,$0a,$08 
             dfb $08,$02,$02,$09,$09 

color2       dfb $09,$09,$02,$02,$08 
             dfb $08,$0a,$0a,$0f,$0f 
             dfb $07,$07,$01,$01,$01 
             dfb $01,$01,$01,$01,$01 
             dfb $01,$01,$01,$01,$01 
             dfb $01,$01,$01,$07,$07 
             dfb $0f,$0f,$0a,$0a,$08 
             dfb $08,$02,$02,$09,$09 
