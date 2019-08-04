package cmd

import (
	"fmt"
)


func TraceStateMachine(traces []byte) map[string]bool {
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
	var addressesInTrace [5000]string
	addressesIndex := 0

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
				addressMap[string(traces[index + 8: index + 8 + 42]) + "\t" + blockNumStr + "\t" + "99999"] = true
				// author
				// Read in address
				// move index forward
				state = STATE_START
			case STATE_O:
				fmt.Println("Should be output:", string(traces[index + 8: index + 10]))
				startIndex := index + 8
				endIndex := index + 8
				for j := startIndex; j < len(traces); j++ {
					if traces[j] == comma {
						endIndex = j
						break
					}
				}

				if startIndex + 10 < endIndex {
					data := traces[startIndex + 10: endIndex]
					for i := 0; i < len(data) / 64; i++ {
						addr := string(data[i * 64 : (i + 1) * 64])
						if potentialAddress(addr) {
							addr = "0x" + string(addr[24:])
							if goodAddr(addr) {
								addressesInTrace[addressesIndex] = addr
								addressesIndex += 1
							}
						}
					}
				}
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
				fmt.Println("Should be address:", string(traces[index + 7: index + 7 + 42]))
				addressesInTrace[addressesIndex] = string(traces[index + 7: index + 7 + 42])
				addressesIndex += 1
				state = STATE_START
			case STATE_F:
				fmt.Println("Should be from:", string(traces[index + 6: index + 6 + 42]))
				addressesInTrace[addressesIndex] = string(traces[index + 6: index + 6 + 42])
				addressesIndex += 1
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
				fmt.Println("Should be refundAddress:", string(traces[index + 15: index + 15 + 42]))
				addressesInTrace[addressesIndex] = string(traces[index + 15: index + 15 + 42])
				addressesIndex += 1
				state = STATE_START
			default:
				state = STATE_START
			}
		case i:
			switch state {
			case STATE_START:
				state = STATE_I
			case STATE_N:
				fmt.Println("Should be init:", string(traces[index + 5: index + 7]))
				startIndex := index + 5
				endIndex := index + 5
				for j := startIndex; j < len(traces); j++ {
					if traces[j] == comma {
						endIndex = j
						break
					}
				}

				if startIndex + 10 < endIndex {
					data := traces[startIndex + 10: endIndex]
					for i := 0; i < len(data) / 64; i++ {
						addr := string(data[i * 64 : (i + 1) * 64])
						if potentialAddress(addr) {
							addr = "0x" + string(addr[24:])
							if goodAddr(addr) {
								addressesInTrace[addressesIndex] = addr
								addressesIndex += 1
							}
						}
					}
				}

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
				fmt.Println("Should be input:", string(traces[index + 6: index + 8]))
				startIndex := index + 6
				endIndex := index + 6
				for j := startIndex; j < len(traces); j++ {
					if traces[j] == comma {
						endIndex = j
						break
					}
				}

				if startIndex + 10 < endIndex {
					data := traces[startIndex + 10: endIndex]
					for i := 0; i < len(data) / 64; i++ {
						addr := string(data[i * 64 : (i + 1) * 64])
						if potentialAddress(addr) {
							addr = "0x" + string(addr[24:])
							if goodAddr(addr) {
								addressesInTrace[addressesIndex] = addr
								addressesIndex += 1
							}
						}
					}
				}
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
	return addressMap
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


// goodAddr Returns true if the address is not a precompile and not zero
func goodAddr(addr string) bool {
	// As per EIP 1352, all addresses less than the following value are reserved
	// for pre-compiles. We don't index precompiles.
	if addr < "0x000000000000000000000000000000000000ffff" {
		return false
	}
	return true
}

// potentialAddress Processing 'input' value, 'output' value or event 'data' value
// we do our best, but we don't include everything we could. We do the best we can
func potentialAddress(addr string) bool {
	// Any address smaller than this we call a 'baddress' and do not index
	small := "00000000000000000000000000000000000000ffffffffffffffffffffffffff"
	//        -------+-------+-------+-------+-------+-------+-------+-------+
	if addr <= small {
		return false
	}

	// Any address with less than this many leading zeros is not an left-padded 20-byte address
	largePrefix := "000000000000000000000000"
	//              -------+-------+-------+
	if !strings.HasPrefix(addr, largePrefix) {
		return false
	}

	if strings.HasSuffix(addr, "00000000") {
		return false
	}
	return true
}