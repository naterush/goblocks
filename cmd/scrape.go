package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

// Params Parameters used during calls to the RPC.
type Params []interface{}

// RPCPayload Data structure used during calls to the RPC.
type RPCPayload struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  `json:"params"`
	ID      int `json:"id"`
}

// Filter Used by the getLogs RPC call to identify the block range.
type Filter struct {
	Fromblock string `json:"fromBlock"`
	Toblock   string `json:"toBlock"`
}

type Range struct {
	StartIdx int 
	EndIdx int
}

// BlockInternals - carries both the traces and the logs for a block
type BlockInternals struct {
	BlockNumber int
	Traces []byte
	Logs   []byte
}

// toScreen Sends a prompt and a value to the screen (adjusts spacing if running from docker)
func toScreen(dockerMode bool, prompt string, value string, newLine bool) {
	space1 := "\t"
	if dockerMode {
		space1 = "   "
	}
	fmt.Print(space1, prompt, "\t", value)
	if newLine {
		fmt.Println("")
	}
}

// getBlockHeader Returns the block header for a given block.
func getBlockHeader(rpcProvider string, blockNum int) ([]byte, error) {
	payloadBytes, err := json.Marshal(RPCPayload{"2.0", "parity_getBlockHeaderByNumber", Params{fmt.Sprintf("0x%x", blockNum)}, 2})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", rpcProvider, body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	blockHeaderBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()
	return blockHeaderBody, nil
}

// getTracesFromBlock Returns all traces for a given block.
func getTracesFromBlock(rpcProvider string, blockNum int) ([]byte, error) {
	payloadBytes, err := json.Marshal(RPCPayload{"2.0", "trace_block", Params{fmt.Sprintf("0x%x", blockNum)}, 2})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", rpcProvider, body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	tracesBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()
	return tracesBody, nil
}

// getLogsFromBlock Returns all logs for a given block.
func getLogsFromBlock(rpcProvider string, blockNum int) ([]byte, error) {
	payloadBytes, err := json.Marshal(RPCPayload{"2.0", "eth_getLogs", Params{Filter{fmt.Sprintf("0x%x", blockNum), fmt.Sprintf("0x%x", blockNum)}}, 2})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", rpcProvider, body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	logsBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()
	return logsBody, nil
}

// getTransactionReceipt Returns recipt for a given transaction -- only used in errored contract creations
func getTransactionReceipt(rpcProvider string, hash string) ([]byte, error) {
	payloadBytes, err := json.Marshal(RPCPayload{"2.0", "eth_getTransactionReceipt", Params{hash}, 2})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequest("POST", rpcProvider, body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	receiptBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()
	return receiptBody, nil
}

// getTracesAndLogs Process the block channel and for each block query the node for both traces and logs. Send results to addressChannel
func getTracesAndLogs(rpcProvider string, blockChannel chan int, addressChannel chan BlockInternals, blockWG *sync.WaitGroup) {

	for blockNum := range blockChannel {
		traces, err := getTracesFromBlock(rpcProvider, blockNum)
		if err != nil {
			fmt.Println(err)
			os.Exit(1) // caller will start over if this process exits with non-zero value
		}
		logs, err := getLogsFromBlock(rpcProvider, blockNum)
		if err != nil {
			fmt.Println(err)
			os.Exit(1) // caller will start over if this process exits with non-zero value
		}
		addressChannel <- BlockInternals{blockNum, traces, logs}
	}
	blockWG.Done()
}


func recieveAddresses(addressChannel chan string, recieveWG *sync.WaitGroup, blockNumberStr string, nBlocks int, ripeBlock int, unripePath string, ripePath string) {
	addressMap := make(map[string]bool)
	for addressSighting := range addressChannel {
		addressMap[addressSighting] = true
	}
	writeAddresses("SM" + blockNumberStr, addressMap, nBlocks, ripeBlock, unripePath, ripePath)
	recieveWG.Done()
}

const (
	straightCloseBracket = byte(93)
	curlyCloseBracket = byte(125)
	y = byte(121)
)

func extractAddresses(rpcProvider string, addressChannel chan BlockInternals, addressWG *sync.WaitGroup, nBlocks int, ripeBlock int, unripePath string, ripePath string) {

	for blockTraceAndLog := range addressChannel {
		blockNumberStr := leftPad(strconv.Itoa(blockTraceAndLog.BlockNumber), 9)
		addressChannel := make(chan string)

		var recieveWG sync.WaitGroup
		recieveWG.Add(1)
		go recieveAddresses(addressChannel, &recieveWG, blockNumberStr, nBlocks, ripeBlock, unripePath, ripePath)

		var traceWG sync.WaitGroup
		traceWG.Add(20)
		rangeChannelTraces := make(chan Range)
		for i := 0; i < 20; i ++ {
			go TraceStateMachine(blockTraceAndLog.Traces[:], rangeChannelTraces, addressChannel, blockNumberStr, &traceWG)
		}
		fmt.Println("\n\n\n")

		// Send the approprate ranges to the TraceStateMachines
		chunkSize := len(blockTraceAndLog.Traces) / 20 // amount each jawn processes
		startIdx := 0
		for i := 0; i < 20; i ++ {
			endIdx := startIdx + chunkSize
			if endIdx >= len(blockTraceAndLog.Logs) {
				endIdx = len(blockTraceAndLog.Logs) - 1
			}
			// move the end of the chunk to a "safe location"
			// here, we move to 
			for j := endIdx; j < len(blockTraceAndLog.Traces); j++ {
				if blockTraceAndLog.Traces[j] == straightCloseBracket || blockTraceAndLog.Traces[j] == curlyCloseBracket {
					endIdx = j
					break
				}
			}
			fmt.Println(string(blockTraceAndLog.Traces[startIdx:endIdx]))
			rangeChannelTraces <- Range{startIdx, endIdx}
			startIdx = endIdx
		}
		close(rangeChannelTraces)

		var logWG sync.WaitGroup
		logWG.Add(20)
		rangeChannelLogs := make(chan Range)
		for i := 0; i < 20; i ++ {
			go LogStateMachine(blockTraceAndLog.Logs[:], rangeChannelLogs, addressChannel, blockNumberStr, &logWG)
		}
		fmt.Println("\n\n")

		// Send the approprate ranges to the LogStateMachines
		chunkSize = len(blockTraceAndLog.Logs) / 20 // amount each jawn processes
		startIdx = 0
		for i := 0; i < 20; i ++ {
			endIdx := startIdx + chunkSize
			if endIdx >= len(blockTraceAndLog.Logs) {
				endIdx = len(blockTraceAndLog.Logs) - 1
			}
			// move the end of the chunk to a "safe location"
			// here, we move to the location of "y" in type, which
			// is the last field in a single object in result array
			for j := endIdx; j < len(blockTraceAndLog.Logs); j++ {
				if blockTraceAndLog.Logs[j] == y {
					endIdx = j
					break
				}
			}
			rangeChannelLogs <- Range{startIdx, endIdx}
			fmt.Println(string(blockTraceAndLog.Logs[startIdx:endIdx]))
			startIdx = endIdx
		}
		close(rangeChannelLogs)

		// Wait for the various state machines to finish
		traceWG.Wait()
		logWG.Wait()
		// Close the address channel, to notify the address receiver all have been sent
		close(addressChannel)
		// wait for the address receiver to write out the file
		recieveWG.Wait()
	}
	addressWG.Done()
}

var counter = 0

func writeAddresses(blockNum string, addressMap map[string]bool, nBlocks int, ripeBlock int, unripePath string, ripePath string) {

	addressArray := make([]string, len(addressMap))
	idx := 0
	for address := range addressMap {
		addressArray[idx] = address
		idx++
	}
	sort.Strings(addressArray)
	toWrite := []byte(strings.Join(addressArray[:], "\n") + "\n")

	bn, _ := strconv.Atoi(blockNum)
	// This code (disabled) tried to extract timestamp while we're scraping. It didn't work, so
	// commented out. We were trying to attach timestamp to the ripe block's filename
	//blockHeaderBytes, err := getBlockHeader(bn)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1) // caller will start over if this process exits with non-zero value
	//}
	//
	//var header BlockHeader
	//err = json.Unmarshal(blockHeaderBytes, &header)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1) // caller will start over if this process exits with non-zero value
	//}
	//
	//fileName := Options.ripePath + blockNum + "_ts" + header.Result.Timestamp + ".txt"
	fileName := ripePath + blockNum + ".txt"
	if bn > ripeBlock {
		fileName = unripePath + blockNum + ".txt"
	}

	err := ioutil.WriteFile(fileName, toWrite, 0744)
	if err != nil {
		fmt.Println(err)
		os.Exit(1) // caller will start over if this process exits with non-zero value
	}
	// Show twenty-five dots no matter how many blocks we're scraping
	skip := nBlocks / 50
	if skip < 1 {
		skip = 1
	}
	counter++
	if counter%skip == 0 {
		fmt.Print(".")
	}
}

func ProcessBlocks(rpcProvider string, nBlockProcs int, nAddrProcs int, startBlock int, nBlocks int, ripeBlock int, unripePath string, ripePath string) {

	blockChannel := make(chan int)
	addressChannel := make(chan BlockInternals)

	var blockWG sync.WaitGroup
	blockWG.Add(nBlockProcs)
	for i := 0; i < nBlockProcs; i++ {
		go getTracesAndLogs(rpcProvider, blockChannel, addressChannel, &blockWG)
	}

	var addressWG sync.WaitGroup
	addressWG.Add(nAddrProcs)
	for i := 0; i < nAddrProcs; i++ {
		go extractAddresses(rpcProvider, addressChannel, &addressWG, nBlocks, ripeBlock, unripePath, ripePath)
	}

	for block := startBlock; block < startBlock + nBlocks; block++ {
		blockChannel <- block
	}

	close(blockChannel)
	blockWG.Wait()

	close(addressChannel)
	addressWG.Wait()
}

func leftPad(str string, totalLen int) string {
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

var scrapeCmd = &cobra.Command{
	Use:   "scrape",
	Short: "Freshen the index to the front of the chain",
	Long: `
Description:

  The 'scrape' subcommand freshens the TrueBlocks index, picking up where it last
  left off. 'Scrape' visits every block, queries that block's traces and logs
  looking for addresses, and writes an index of those addresses per transaction.`,
	Run: func(cmd *cobra.Command, args []string) {
		toScreen(Options.dockerMode, "  options:", strconv.Itoa(Options.startBlock) + "/"+strconv.Itoa(Options.nBlocks)+"/"+strconv.Itoa(Options.ripeBlock), true)
		toScreen(Options.dockerMode, "  processes:", strconv.Itoa(Options.nBlockProcs)+"/"+strconv.Itoa(Options.nAddrProcs), true)
		toScreen(Options.dockerMode, "  rpcProvider:", Options.rpcProvider, true)
		toScreen(Options.dockerMode, "  indexPath:", Options.indexPath, true)
		toScreen(Options.dockerMode, "  ripePath:", Options.ripePath, true)
		toScreen(Options.dockerMode, "  unripePath:", Options.unripePath, true)
		toScreen(Options.dockerMode, "  dockerMode:", strconv.FormatBool(Options.dockerMode), true)
		toScreen(Options.dockerMode, "  scraping:", "", false)
		ProcessBlocks(
			Options.rpcProvider,
			Options.nBlockProcs, 
			Options.nAddrProcs, 
			Options.startBlock, 
			Options.nBlocks, 
			Options.ripeBlock, 
			Options.unripePath, 
			Options.ripePath)
		fmt.Println("")
	},
}

func init() {
	rootCmd.AddCommand(scrapeCmd)
}
