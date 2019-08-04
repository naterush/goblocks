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
		STATE_P_CAP
		STATE_O
		STATE_O_AFTER_P_CAP
		STATE_U_AFTER_O
		STATE_T_AFTER_OU
	)

	state := STATE_START

	t := []byte("t")[0]
	o := []byte("o")[0]
	a := []byte("a")[0]
	u := []byte("u")[0]
	d := []byte("d")[0]
	r := []byte("r")[0]
	e := []byte("e")[0]
	f := []byte("f")[0]
	i := []byte("i")[0]
	n := []byte("n")[0]
	p := []byte("p")[0]
	s := []byte("s")[0]
	P := []byte("P")[0]
	comma := []byte(",")[0]
	//openBracket := byte(123) // byte value of {
	//closeBracket := byte(125) // byte value of }
	//openBracketStraight := byte(91) // byte value of [
	//closeBracketStraight := byte(93) // byte value of ] TODO: this might be wrong!, but i don't really need it

	// Keep track of these indexes
	const MAX_ADDRESSES_IN_TRACE = 1000
	transactionPosition := 0
	var addressesInTrace [5000]string
	addressesIndex := 0

	addressesInTrace[1] = "hi"
	transactionPosition += 1
	addressesIndex += 1

	blockNumStr := "005000000"

	// Address + block + index store
	addressMap := make(map[string]bool)

	for index := 0; index < len(traces); index++ {
		token := traces[index]

		switch token {
		case t:
			switch state {
			case STATE_START:
				state = STATE_T
			case STATE_U_AFTER_O:
				state = STATE_T_AFTER_OU
			default:
				state = STATE_START
			}
		case o:
			switch state {
			case STATE_START:
				state = STATE_O
			case STATE_T:
				// READ IN "to"
				fmt.Println("From to:", string(traces[index + 4: index + 4 + 42]))
				addressesInTrace[addressesIndex] = string(traces[index + 4: index + 4 + 42])
				addressesIndex += 1

				state = STATE_START
			case STATE_P_CAP:
				state = STATE_O_AFTER_P_CAP
			default:
				state = STATE_START
			}
		case s:
			switch state {
			case STATE_O_AFTER_P_CAP:
				transactionPositionStart :=  index + 8
				transactionPositionEnd := index + 8
				for j := transactionPositionStart; j < len(traces); j++ {
					if traces[j] == comma {
						transactionPositionEnd = j
						break
					}
				}

				// Write out addresses to map
				transactionPositionStr := padLeft(string(traces[transactionPositionStart: transactionPositionEnd]), 5)
				fmt.Println("Transaction Position:", transactionPositionStr)

				for j := 0; j < addressesIndex; j++ {
					addressMap[addressesInTrace[j] + "\t" + blockNumStr + "\t" + transactionPositionStr] = true
				}
				addressesIndex = 0

				state = STATE_START
			default:
				state = STATE_START
			}
		case a:
			switch state {
			case STATE_START:
				state = STATE_A
			default:
				state = STATE_START
			}
		case u:
			switch state {
			case STATE_A:
				fmt.Println("From author:", string(traces[index + 8: index + 8 + 42]))
				// author
				// Read in address
				// move index forward
				state = STATE_START
			case STATE_O:
				fmt.Println("Should be output:", string(traces[index - 2: index + 4]))
				// output
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case d:
			switch state {
			case STATE_A:
				state = STATE_D
			case STATE_D:
				state = STATE_D_AFTER_D
			default:
				state = STATE_START
			}
		case r:
			switch state {
			case STATE_START:
				state = STATE_R
			case STATE_D_AFTER_D:
				fmt.Println("Should be address:", string(traces[index - 4: index + 3]))
				// address
				// Read in address
				// move index forward
				state = STATE_START
			case STATE_F:
				fmt.Println("Should be from:", string(traces[index - 2: index + 2]))
				// from
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case e:
			switch state {
			case STATE_R:
				state = STATE_E
			default:
				state = STATE_START
			}
		case f:
			switch state {
			case STATE_START:
				state = STATE_F
			case STATE_E:
				fmt.Println("Should be refundAddress:", string(traces[index - 3: index + 10]))
				// refundAddress
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case i:
			switch state {
			case STATE_START:
				state = STATE_I
			case STATE_N:
				fmt.Println("Should be init:", string(traces[index - 3: index + 11]))
				// init
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case n:
			switch state {
			case STATE_I:
				state = STATE_N
			default:
				state = STATE_START
			}
		case p:
			switch state {
			case STATE_N:
				fmt.Println("Should be input:", string(traces[index - 3: index + 2]))
				// input
				// Read in address
				// move index forward
				state = STATE_START
			default:
				state = STATE_START
			}
		case P:
			switch state {
			case STATE_START:
				state = STATE_P_CAP
			default:
				state = STATE_START
			}
		default:
			state = STATE_START
		}

	}
}

func padLeft(str string, totalLen int) string {
	if len(str) >= totalLen {
		return str
	}
	zeros := ""
	for i := 0; i < totalLen-len(str); i++ {
		zeros += "0"
	}
	return zeros + str
}