package cmd

import (
	"fmt"
)


func TraceStateMachine(traces []byte) {
	// Declare all the states we need
	const (
		STATE_START = iota
		STATE_T
		STATE_A
		STATE_U_AFTER_A
		STATE_D
		STATE_D_AFTER_D
		STATE_R
		STATE_E
		STATE_F
		STATE_I
		STATE_N
		STATE_I_AFTER_N
		STATE_P
		STATE_O
		STATE_U_AFTER_O
		STATE_T_AFTER_OU
	)

	state := STATE_START

	t := []byte("t")
	o := []byte("o")
	a := []byte("a")
	u := []byte("u")
	d := []byte("d")
	r := []byte("r")
	e := []byte("e")
	f := []byte("f")
	i := []byte("i")
	n := []byte("n")
	p := []byte("p")
	openBracket := byte(123) // byte value of {
	closeBracket := byte(125) // byte value of }
	openBracketStraight := byte(91) // byte value of [
	closeBracketStraight := byte(93) // byte value of ] TODO: this might be wrong!, but i don't really need it

	for i := 0; i < len(traces); i++ {
		token := traces[i]

		switch token {
		case t:
			fmt.Println("t")
			switch state {
			case START_STATE:
				state = STATE_T
			case STATE_U_AFTER_O:
				state = STATE_T_AFTER_OU
			default:
				state = STATE_START
			}
		case o:
			fmt.Println("o")
			switch state {
			case START_STATE:
				state = STATE_O
			case STATE_T:
				fmt.Println("Should be to:", string(traces[index - 1, index + 1]))
				// to
				// Read in address
				// move index forward

				state = START_STATE
			default:
				state = STATE_START
			}
		case a:
			fmt.Println("a")
			switch state {
			case START_STATE:
				state = STATE_A
			default:
				state = STATE_START
			}
		case u:
			fmt.Println("u")
			switch state {
			case STATE_A:
				fmt.Println("Should be author:", string(traces[index - 2, index + 4]))
				// author
				// Read in address
				// move index forward
				state = STATE_START
			case STATE_O:
				fmt.Println("Should be output:", string(traces[index - 2, index + 4]))
				// output
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case d:
			fmt.Println("d")
			switch state {
			case STATE_A:
				state = STATE_D
			case STATE_D:
				state = STATE_D_AFTER_D
			default:
				state = STATE_START
			}
		case r:
			fmt.Println("r")
			switch state {
			case STATE_START:
				state = STATE_R
			case STATE_D_AFTER_D:
				fmt.Println("Should be address:", string(traces[index - 4, index + 3]))
				// address
				// Read in address
				// move index forward
				state = STATE_START
			case STATE_F:
				fmt.Println("Should be from:", string(traces[index - 2, index + 2]))
				// from
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case e:
			fmt.Println("e")
			switch state {
			case STATE_R:
				state = STATE_E
			default:
				state = STATE_START
			}
		case f:
			fmt.Println("f")
			switch state {
			case START_STATE:
				state = STATE_F
			case STATE_E:
				fmt.Println("Should be refundAddress:", string(traces[index - 3, index + 10]))
				// refundAddress
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case i:
			fmt.Println("i")
			switch state {
			case START_STATE:
				state = STATE_I
			case STATE_N:
				fmt.Println("Should be init:", string(traces[index - 3, index + 11]))
				// init
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case n:
			fmt.Println("n")
			switch state {
			case STATE_I:
				state = STATE_N
			default:
				state = STATE_START
			}
		case p:
			fmt.Println("p")
			switch state {
			case STATE_N:
				fmt.Println("Should be input:", string(traces[index - 3, index + 2]))
				// input
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		}
	}

}






