package cmd

import (
	"fmt"
	"strconv"
	"bytes"
)


func TraceStateMachine(traces []byte, addressMap map[string]bool){
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

	// Get the block number
	blockNumStartIndex := bytes.Index(traces, []byte("blockNumber")) + 13
	blockNumEndIndex := blockNumStartIndex
	for j := blockNumStartIndex; j < len(traces); j++ {
		if traces[j] == comma {
			blockNumEndIndex = j
			break
		}
	}

	blockNumStr := leftPad(string(traces[blockNumStartIndex: blockNumEndIndex]), 9)
	

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
				//fmt.Println("From to:", string(traces[index + 4: index + 4 + 42]))
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
				transactionPositionStr := leftPad(string(traces[transactionPositionStart: transactionPositionEnd]), 5)
				//fmt.Println("Transaction Position:", transactionPositionStr)

				for j := 0; j < addressesIndex; j++ {
					if goodAddr(addressesInTrace[j]) {
						addressMap[addressesInTrace[j] + "\t" + blockNumStr + "\t" + transactionPositionStr] = true
					}
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
				//fmt.Println("From author:", string(traces[index + 8: index + 8 + 42]))
				addressMap[string(traces[index + 8: index + 8 + 42]) + "\t" + blockNumStr + "\t" + "99999"] = true
				// author
				// Read in address
				// move index forward
				state = STATE_START
			case STATE_O:
				//fmt.Println("Should be output:", string(traces[index + 8: index + 10]))
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
				//fmt.Println("Should be address:", string(traces[index + 7: index + 7 + 42]))
				addressesInTrace[addressesIndex] = string(traces[index + 7: index + 7 + 42])
				addressesIndex += 1
				state = STATE_START
			case STATE_F:
				//fmt.Println("Should be from:", string(traces[index + 6: index + 6 + 42]))
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
				//fmt.Println("Should be refundAddress:", string(traces[index + 15: index + 15 + 42]))
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
				//fmt.Println("Should be init:", string(traces[index + 5: index + 7]))
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
				//fmt.Println("Should be input:", string(traces[index + 6: index + 8]))
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
}



func LogStateMachine(logs []byte, addressMap map[string]bool) {
	// Declare all the states we need
	const (
		STATE_START = iota
		STATE_T
		STATE_O
		STATE_D
		STATE_A
		STATE_T_AFTER_A
		STATE_N
		STATE_I
	)

	state := STATE_START

	t := []byte("t")[0]
	o := []byte("o")[0]
	p := []byte("p")[0]
	d := []byte("d")[0]
	a := []byte("a")[0]
	n := []byte("n")[0]
	I := []byte("I")[0]
	comma := []byte(",")[0]
	quote := byte(34)
	//openBracket := byte(123) // byte value of {
	//closeBracket := byte(125) // byte value of }
	//openBracketStraight := byte(91) // byte value of [
	closeBracketStraight := byte(93) // byte value of ] TODO: this might be wrong!, but i don't really need it

	// Keep track of these indexes
	const MAX_ADDRESSES_IN_TRACE = 1000
	var addressesInTrace [5000]string
	addressesIndex := 0

	blockNumStartIndex := bytes.Index(traces, []byte("blockNumber")) + 13
	blockNumEndIndex := blockNumStartIndex
	for j := blockNumStartIndex; j < len(traces); j++ {
		if traces[j] == quote {
			blockNumEndIndex = j
			break
		}
	}

	blockNum, _ := strconv.ParseInt(string(logs[blockNumStartIndex: blockNumEndIndex]), 0, 64)				
	blockNumStr :=  leftPad(strconv.FormatInt(blockNum, 10), 5)


	for index := 0; index < len(logs); index++ {
		token := logs[index]

		switch token {
		case t:
			switch state {
			case STATE_START:
				state = STATE_T
			case STATE_A:
				state = STATE_T_AFTER_A
			default:
				state = STATE_START
			}
		case o:
			switch state {
			case STATE_T:
				state = STATE_O
			default:
				state = STATE_START
			}
		case p:
			switch state {
			case STATE_O:
				// Read in the topics
				startIndex := index + 7
				endIndex := index + 4
				for j := startIndex; j < len(logs); j++ {
					if logs[j] == closeBracketStraight {
						endIndex = j
						break
					}
				}
				//fmt.Println("TOPICS:", string(logs[startIndex: endIndex]))

				// jump by 69
				for j := startIndex; j <= endIndex ; j+= 69 {
					addr := string(logs[j + 1 + 2: j + 1 + 66])
					if potentialAddress(addr) {
						addr = "0x" + string(addr[24:])
						if goodAddr(addr) {
							addressesInTrace[addressesIndex] = addr
							addressesIndex += 1
						}
					}
				}

				state = STATE_START
			default:
				state = STATE_START
			}
		case d:
			switch state {
			case STATE_START:
				state = STATE_D
			default:
				state = STATE_START
			}
		case a:
			switch state {
			case STATE_D:
				state = STATE_A
			case STATE_T_AFTER_A:
				// Read in the input data!
				startIndex := index + 4
				endIndex := index + 4
				for j := startIndex; j < len(logs); j++ {
					if logs[j] == comma {
						endIndex = j - 1
						break
					}
				}
				//fmt.Println("DATA:", string(logs[startIndex: endIndex]))

				if startIndex + 2 <= endIndex {
					data := logs[startIndex + 2: endIndex]
					for i := 0; i < len(data) / 64; i++ {
						addr := string(data[i*64 : (i + 1) * 64])
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
			case STATE_START:
				state = STATE_N
			default:
				state = STATE_START
			}
		case I:
			switch state {
			case STATE_N:
				transactionPositionStart :=  index + 8
				transactionPositionEnd := index + 8
				for j := transactionPositionStart; j < len(logs); j++ {
					if logs[j] == comma {
						transactionPositionEnd = j - 1
						break
					}
				}

				txIdx, _ := strconv.ParseInt(string(logs[transactionPositionStart: transactionPositionEnd]), 0, 64)				

				// Write out addresses to map
				transactionPositionStr := leftPad(strconv.FormatInt(txIdx, 10), 5)
				//fmt.Println("Transaction Position:", transactionPositionStr)

				for j := 0; j < addressesIndex; j++ {
					if goodAddr(addressesInTrace[j]) {
						addressMap[addressesInTrace[j] + "\t" + blockNumStr + "\t" + transactionPositionStr] = true
					}
				}
				addressesIndex = 0

				state = STATE_START
			default:
				state = STATE_START
			}
		default:
			state = STATE_START
		}
	}
}
