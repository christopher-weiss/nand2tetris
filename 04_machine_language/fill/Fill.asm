// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

(LOOP)
    @SCREEN
    D=A
    @currpixelgrp
    M=D         // currpixelgrp = *SCREEN; top left 16 pixels
    @KBD
    D=M         // D = *KBD
    @NOKEYPRESS
    D;JEQ       // if (*KBD == 0) GOTO NOKEYPRESS
    @KEYPRESS
    0;JEQ       // if (*KBD != 0) GOTO KEYPRESS
(DRAW)
    @currpixelgrp
    D=M         // D = *currpixelgrp
    @KBD
    D=D-A       // D -= KBD
    @LOOP
    D;JEQ       // GOTO LOOP
    @pixels
    D=M         // D = *pixels
    @currpixelgrp
    A=M
    M=D         // *currpixelgrp = *pixels
    @currpixelgrp
    M=M+1       // currpixelgrp++
    @DRAW
    0;JMP       // GOTO DRAW
(NOKEYPRESS)
    @pixels
    M=0         // *pixels = 0; (BIN(000.0))
    @DRAW
    0;JMP       // GOTO DRAW
(KEYPRESS)
   @pixels
   M=-1         // *pixels = -1; (BIN(111.1))
   @DRAW
   0;JMP        // GOTO DRAW
