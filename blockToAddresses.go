package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
    "sort"
    "strings"
    "strconv"
)

type Params []interface{}

type Payload struct {
    Jsonrpc string        `json:"jsonrpc"`
    Method  string        `json:"method"`
    Params                `json:"params"`
    ID      int           `json:"id"`
}


type Result struct {
    block int
    body string
}

type BlockTraces struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  []struct {
		Action struct {
			CallType string `json:"callType"` // call
			From     string `json:"from"`
			Gas      string `json:"gas"`
			Input    string `json:"input"`
			To       string `json:"to"`
            Value    string `json:"value"`
            Author     string `json:"author"` // reward
            RewardType string `json:"rewardType"` 
            Address string `json:"address"` // suicide
            Balance string `json:"balance"` 
            RefundAddress string `json:"refundAddress"` 
            Init string `json:"init"` // create
		} `json:"action,omitempty"`
		BlockHash   string `json:"blockHash"`
		BlockNumber int    `json:"blockNumber"`
		Result      struct {
			GasUsed string `json:"gasUsed"` // call
            Output  string `json:"output"`
            Address string `json:"address"` // create
		} `json:"result"`
		Subtraces           int           `json:"subtraces"`
		TraceAddress        []interface{} `json:"traceAddress"`
		TransactionHash     string        `json:"transactionHash"`
		TransactionPosition int           `json:"transactionPosition"`
        Type                string        `json:"type"`
	} `json:"result"`
	ID int `json:"id"`
}

type BlockLogs struct {
	Jsonrpc string `json:"jsonrpc"`
	Result  []struct {
		Address             string   `json:"address"`
		BlockHash           string   `json:"blockHash"`
		BlockNumber         string   `json:"blockNumber"`
		Data                string   `json:"data"`
		LogIndex            string   `json:"logIndex"`
		Removed             bool     `json:"removed"`
		Topics              []string `json:"topics"`
		TransactionHash     string   `json:"transactionHash"`
		TransactionIndex    string   `json:"transactionIndex"`
		TransactionLogIndex string   `json:"transactionLogIndex"`
		Type                string   `json:"type"`
	} `json:"result"`
	ID int `json:"id"`
}

func leftZero(str string, totalLen int) string {
    // Assume len(str) < totalLen
    zeros := ""
    for i :=0 ; i < totalLen - len(str); i++ {
        zeros += "0"
    }
    return zeros + str
}

type AddrSighting struct {
    block int
    txIdx int
}


func searchForAddress(address string, fileNames chan string, sightings chan AddrSighting) {
    for fileName := range fileNames {
        data, err := ioutil.ReadFile("block/" + fileName)
        if err != nil {
            fmt.Println("Error:", err)
        }
        fmt.Print(string(data))
        sightings <- AddrSighting{0, 0}
    }
} 

func isPotentialAddress(addr string) bool {
    //fmt.Println("Checking for addr", addr)

    small := "00000000000000000000000000000000000000ffffffffffffffffffffffffff"
    largePrefix := "000000000000000000000000"

    if addr <= small || !strings.HasPrefix(addr, largePrefix) {
        //fmt.Println("False bc large or small", addr <= small, !strings.HasPrefix(addr, largePrefix))
        return false
    }

    if strings.HasSuffix(addr, "00000000") {
        //fmt.Println("False bc has suffix")
        return false
    }

    //fmt.Println("True!")
    return true
}

type TraceAndLogs struct {
    Traces []byte
    Logs  []byte
}

func getAddress(traceAndLogs chan TraceAndLogs) {
    for blockTraceAndLog := range traceAndLogs {
        var traces BlockTraces
        err := json.Unmarshal(blockTraceAndLog.Traces, &traces)
	    if err != nil {
	    	fmt.Println("error:", err)
        }
        addresses := make(map[string]bool)
        fmt.Println("Now processing block", traces.Result[0].BlockNumber)

        // Format block number, so it's 9 digits total
        blockNum := leftZero(strconv.Itoa(traces.Result[0].BlockNumber), 9)
        for i :=0; i < len(traces.Result); i++ {
            idx := leftZero(strconv.Itoa(traces.Result[i].TransactionPosition), 5)
            blockAndIdx := "\t" + blockNum + "\t" + idx
            // Try to get addresses from the input data
            if len(traces.Result[i].Action.Input) > 10 {
                inputData := traces.Result[i].Action.Input[10:]
                //fmt.Println("Input data:", inputData, len(inputData))
                for i := 0; i < len(inputData) / 64; i++ {
                    addr := string(inputData[i * 64:(i + 1) * 64])
                    if isPotentialAddress(addr) {
                        addresses["0x" + string(addr[24:]) + blockAndIdx] = true
                    }
                }
            }
            //if idx == "00020" {
            //    fmt.Println("HEREHEHREHREH", string(blockTraceAndLog.Traces))
            //}

            if traces.Result[i].Type == "call" {
                // If it's a call, get the to and from
                from := traces.Result[i].Action.From
                to := traces.Result[i].Action.To
                addresses[from + blockAndIdx] = true
                addresses[to + blockAndIdx] = true
            } else if traces.Result[i].Type == "reward" {
                if traces.Result[i].Action.RewardType == "block" {
                    author := traces.Result[i].Action.Author
                    addresses[author + "\t" + blockNum + "\t" + "99999"] = true
                } else if traces.Result[i].Action.RewardType == "uncle" {

                    //author := traces.Result[i].Action.Author
                    //addresses[author + "\t" + blockNum + "\t" + "99998"] = true
                } else {
                    fmt.Println("New type of reward", traces.Result[i].Action.RewardType)
                }
            } else if traces.Result[i].Type == "suicide" {
                // add the contract that died, and where it sent it's money
                address := traces.Result[i].Action.Address
                refundAddress := traces.Result[i].Action.RefundAddress
                addresses[address + blockAndIdx] = true
                addresses[refundAddress + blockAndIdx] = true
            } else if traces.Result[i].Type == "create" {
                // add the creator, and the new address name
                from := traces.Result[i].Action.From
                address := traces.Result[i].Result.Address
                addresses[from + blockAndIdx] = true
                addresses[address + blockAndIdx] = true

                if len(traces.Result[i].Action.Init) > 2 {
                    initData := traces.Result[i].Action.Init[2:]
                    //fmt.Println("Input data:", inputData, len(inputData))
                    for i := 0; i < len(initData) / 64; i++ {
                        addr := string(initData[i * 64:(i + 1) * 64])
                        if isPotentialAddress(addr) {
                            addresses["0x" + string(addr[24:]) + blockAndIdx] = true
                        }
                    }
                }



            } else {
                fmt.Println("New trace type:", string(blockTraceAndLog.Traces))
            }

            // Parse output of trace
            if len(traces.Result[i].Result.Output) > 2 {
                outputData := traces.Result[i].Result.Output[2:]
                //fmt.Println("Input data:", inputData, len(inputData))
                for i := 0; i < len(outputData) / 64; i++ {
                    addr := string(outputData[i * 64:(i + 1) * 64])
                    if isPotentialAddress(addr) {
                        addresses["0x" + string(addr[24:]) + blockAndIdx] = true
                    }
                }
            }

        }

        // Now, parse log data
        var logs BlockLogs
        err = json.Unmarshal(blockTraceAndLog.Logs, &logs)
	    if err != nil {
	    	fmt.Println("error:", err)
        }

        for i :=0; i < len(logs.Result); i++ {
            idxInt, err := strconv.ParseInt(logs.Result[i].TransactionIndex, 0, 32)
            if err != nil {
                fmt.Println("Error:", err)
            }
            idx := leftZero(strconv.FormatInt(idxInt, 10), 5)
            blockAndIdx := "\t" + blockNum + "\t" + idx
            
            for j := 0 ; j < len(logs.Result[i].Topics); j++ {
                addr := string(logs.Result[i].Topics[j][2:])
                if (isPotentialAddress(addr)) {
                    addresses["0x" + string(addr[24:]) + blockAndIdx] = true
                    //fmt.Println("Adding address", addr)
                }
            }

            if len(logs.Result[i].Data) > 2 {
                inputData := logs.Result[i].Data[2:]
                //fmt.Println("Input data:", inputData, len(inputData))
                for i := 0; i < len(inputData) / 64; i++ {
                    addr := string(inputData[i * 64:(i + 1) * 64])
                    if isPotentialAddress(addr) {
                        addresses["0x" + string(addr[24:]) + blockAndIdx] = true
                    }
                }
            }
        }

        // create an array with all the addresses, and sort
        addressArray := make([]string, len(addresses))
        idx := 0
        for address := range addresses {
            addressArray[idx] = address
            idx++
        }
        sort.Strings(addressArray)
        toWrite := []byte(strings.Join(addressArray[:], "\n") + "\n")

        // TODO: make this a seperate process

        // write this array to a file
        // at least one result (as the miner got a reward)
        fileName := "block/" + blockNum + ".txt"
        err = ioutil.WriteFile(fileName, toWrite, 0777)
        if err != nil {
            fmt.Println("Error writing file:", err)
        }
        fmt.Println("Finished processing", traces.Result[0].BlockNumber)

    }
}

type Filter struct {
    Fromblock string        `json:"fromBlock"`
    Toblock  string        `json:"toBlock"`
}


func getTrace(blocks chan int, traceAndLogs chan TraceAndLogs) {
    // Process blocks untill the blocks channel closes
    for block := range blocks {
        hexBlockNum := fmt.Sprintf("0x%x", block)
        data := Payload{
            "2.0",
            "trace_block",
            Params{hexBlockNum},
            2,
        }
    
        payloadBytes, err := json.Marshal(data)
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
    
        body := bytes.NewReader(payloadBytes)
    
        req, err := http.NewRequest("POST", "http://localhost:8545", body)
        if err != nil {
            fmt.Println("Error:", err)
            return 
        }
        req.Header.Set("Content-Type", "application/json")
    
        resp, err := http.DefaultClient.Do(req)
    
        if err != nil {
            fmt.Println("Error:", err)
            return
        }

        traceBody, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("Error", err)
        }
        resp.Body.Close()
        
        fmt.Println("Read in block and now sending", block)
        
        // Now, get the logs!
        data = Payload{
            "2.0",
            "eth_getLogs",
            Params{Filter{hexBlockNum, hexBlockNum}},
            2,
        }

        payloadBytes, err = json.Marshal(data)
        if err != nil {
            fmt.Println("Error:", err)
            return
        }
    
        body = bytes.NewReader(payloadBytes)
    
        req, err = http.NewRequest("POST", "http://localhost:8545", body)
        if err != nil {
            fmt.Println("Error:", err)
            return 
        }
        req.Header.Set("Content-Type", "application/json")
    
        resp, err = http.DefaultClient.Do(req)
    
        if err != nil {
            fmt.Println("Error:", err)
            return
        }

        logBody, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("Error", err)
        }
        resp.Body.Close()
        
        fmt.Println("Read in block and now sending", block)

        traceAndLogs <- TraceAndLogs{traceBody, logBody}
    }
}

func testSearch() {
    fileNames := make(chan string)
    sightings := make(chan AddrSighting)

    for i := 0; i < 10; i++ {
        go searchForAddress("0xe3e1d847f4d369faa89b01393b34a8193da6dead", fileNames, sightings)
    }

    for i := 6000000; i < 6000000 + 10000; i++ {
        fileName := leftZero(strconv.Itoa(i), 9) + ".txt"
        fileNames <- fileName
    }

    done := make(chan int)
    <- done
}

func main() {
    //testSearch()
    
    startBlock := 71003624//7100251
    numBlocks := 1//7101521 - 7100251

    blocks := make(chan int)
    traceAndLogs := make(chan TraceAndLogs)

    // make a bunch of block trace getters
    for i := 0; i < 20; i++ {
        go getTrace(blocks, traceAndLogs)
    }

    for i := 0; i < 100; i++ {
        go getAddress(traceAndLogs)
    }

    for block := startBlock; block < startBlock + numBlocks; block++ {
        blocks <- block
    }
    
    // blah, just wait around for ever (have to manuall terminate the process...)
    done := make(chan int)
    <- done 
}