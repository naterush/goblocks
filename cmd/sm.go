package cmd

import (
	"strconv"
)


func TraceStateMachine(traces []byte, addressMap map[string]bool, blockNumStr string){
	// States for the state machine
	const (
		STATE_START = iota
		STATE_ZERO
		STATE_P_CAP
	)

	// Byte constants
	const (
		zero = byte(48)
		x = byte(120)
		P = byte(80)
	)

	state := STATE_START

	// We assume there are at most 5000 addresses in a single level of a trace
	var addressesInTrace [5000]string
	addressesIndex := 0

	for index := 0; index < len(traces); index++ {
		token := traces[index]

		switch token {
		case zero:
			switch state {
			case STATE_START:
				state = STATE_ZERO
			default:
				state = STATE_START
			}
		case x:
			switch state {
			case STATE_ZERO:
				/* 
				TODO: traces[index - 1:] starts with a 0x
				As such, it must be checked for addresses
				
				Current plan:
				- at the start of the state machine, create some number of "0xParsers"
				- whenever this state is reached, send this index to a 0x parser through a channel that accepts ints
				- when a 0x parser recieves the start index of a 0x string, it should try and extract addresses from it
				- and then send these addresses back through an "address channel" to an address writer
				- NOTE that before we have the address writer, we should do a first pass to find the transactionPositions.
				*/

				state = STATE_START
			default:
				state = STATE_START
			}
		default:
			state = STATE_START
		}
	}
}



func LogStateMachine(logs []byte, addressMap map[string]bool, blockNumStr string) {
	// States for the state machine
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

	// Byte constants
	const (
		// special chars
		comma = byte(44)
		closeBracketStraight = byte(93) // ]
		// capitol 
		I = byte(73)
		// lowercase
		a = byte(97)
		d = byte(100)
		n = byte(110)
		o = byte(111)
		p = byte(112)
		t = byte(116)
	)

	state := STATE_START

	// We assume there are at most 5000 addresses in a single level of a log
	var addressesInTrace [5000]string
	addressesIndex := 0

	for index := 0; index < len(logs); index++ {
		token := logs[index]

		switch token {
		case a:
			switch state {
			case STATE_D:
				state = STATE_A
			case STATE_T_AFTER_A:
				// Input
				startIndex := index + 4
				endIndex := index + 4
				for j := startIndex; j < len(logs); j++ {
					if logs[j] == comma {
						endIndex = j - 1
						break
					}
				}
				index = endIndex

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
		case d:
			switch state {
			case STATE_START:
				state = STATE_D
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
				// Topics
				startIndex := index + 7
				endIndex := index + 4
				for j := startIndex; j < len(logs); j++ {
					if logs[j] == closeBracketStraight {
						endIndex = j
						break
					}
				}
				index = endIndex

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
		case t:
			switch state {
			case STATE_START:
				state = STATE_T
			case STATE_A:
				state = STATE_T_AFTER_A
			default:
				state = STATE_START
			}
		case I:
			switch state {
			case STATE_N:
				// transactionIndex
				transactionPositionStart :=  index + 8
				transactionPositionEnd := index + 8
				for j := transactionPositionStart; j < len(logs); j++ {
					if logs[j] == comma {
						transactionPositionEnd = j - 1
						break
					}
				}
				index = transactionPositionEnd

				txIdx, _ := strconv.ParseInt(string(logs[transactionPositionStart: transactionPositionEnd]), 0, 64)
				transactionPositionStr := leftPad(strconv.FormatInt(txIdx, 10), 5)

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
