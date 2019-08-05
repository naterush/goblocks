package cmd

import (
	"strconv"
	"sync"
)


func TraceStateMachine(traces []byte, rangeChannel chan Range, addressChannel chan string, blockNumStr string, traceWG *sync.WaitGroup){
	// States for the state machine
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

	// Byte constants
	const (
		// special chars
		comma = byte(44)
		// capitol 
		P = byte(80)
		// lowercase
		a = byte(97)
		d = byte(100)
		e = byte(101)
		f = byte(102)
		i = byte(105)
		n = byte(110)
		o = byte(111)
		p = byte(112)
		r = byte(114)
		s = byte(115)
		t = byte(116)
		u = byte(117)
	)

	state := STATE_START

	// We assume there are at most 5000 addresses in a single level of a trace
	var addressesInTrace [5000]string
	addressesIndex := 0

	for r := range rangeChannel {
		for index := r.StartIdx; index < r.EndIdx; index++ {
			token := traces[index]
	
			switch token {
			case a:
				switch state {
				case STATE_START:
					state = STATE_A
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
					// refundAddress
					addressesInTrace[addressesIndex] = string(traces[index + 15: index + 15 + 42])
					addressesIndex += 1
					index = index + 15 + 42
					state = STATE_START
				default:
					state = STATE_START
				}
			case i:
				switch state {
				case STATE_START:
					state = STATE_I
				case STATE_N:
					// init
					startIndex := index + 5
					endIndex := index + 5
					for j := startIndex; j < len(traces); j++ {
						if traces[j] == comma {
							endIndex = j
							break
						}
					}
					index = endIndex
	
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
			case o:
				switch state {
				case STATE_START:
					state = STATE_O
				case STATE_T:
					//fmt.Println("From to:", string(traces[index + 4: index + 4 + 42]))
					addressesInTrace[addressesIndex] = string(traces[index + 4: index + 4 + 42])
					addressesIndex += 1
					index = index + 4 + 42
					state = STATE_START
				case STATE_P_CAP:
					state = STATE_O_AFTER_P_CAP
				default:
					state = STATE_START
				}
			case p:
				switch state {
				case STATE_N:
					// Input
					startIndex := index + 6
					endIndex := index + 6
					for j := startIndex; j < len(traces); j++ {
						if traces[j] == comma {
							endIndex = j
							break
						}
					}
					index = endIndex
	
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
			case r:
				switch state {
				case STATE_START:
					state = STATE_R
				case STATE_D_AFTER_D:
					// Address
					addressesInTrace[addressesIndex] = string(traces[index + 7: index + 7 + 42])
					addressesIndex += 1
					index = index + 7 + 42
					state = STATE_START
				case STATE_F:
					// From
					addressesInTrace[addressesIndex] = string(traces[index + 6: index + 6 + 42])
					addressesIndex += 1
					index = index + 6 + 42
					state = STATE_START
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
					index = transactionPositionEnd
					transactionPositionStr := leftPad(string(traces[transactionPositionStart: transactionPositionEnd]), 5)
					blockAndIndex := "\t" + blockNumStr + "\t" + transactionPositionStr
	
					// Write out addresses to map
					for j := 0; j < addressesIndex; j++ {
						if goodAddr(addressesInTrace[j]) {
							addressChannel <- addressesInTrace[j] + blockAndIndex
						}
					}
					addressesIndex = 0
	
					state = STATE_START
				default:
					state = STATE_START
				}
			case t:
				switch state {
				case STATE_START:
					state = STATE_T
				case STATE_U_AFTER_O:
					state = STATE_T_AFTER_OU
				default:
					state = STATE_START
				}
			case u:
				switch state {
				case STATE_A:
					// Author
					addressChannel <- string(traces[index + 8: index + 8 + 42]) + "\t" + blockNumStr + "\t" + "99999"
					state = STATE_START
				case STATE_O:
					// Output
					startIndex := index + 8
					endIndex := index + 8
					for j := startIndex; j < len(traces); j++ {
						if traces[j] == comma {
							endIndex = j
							break
						}
					}
					index = endIndex
	
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

	traceWG.Done()
}



func LogStateMachine(logs []byte, rangeChannel chan Range, addressChannel chan string, blockNumStr string, logWG *sync.WaitGroup) {
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

	for r := range rangeChannel {
		for index := r.StartIdx; index < r.EndIdx; index++ {
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
							addressChannel <- addressesInTrace[j] + "\t" + blockNumStr + "\t" + transactionPositionStr
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
	logWG.Done()
}
