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

func leftZero(str string, totalLen int) string {
    // Assume len(str) < totalLen
    zeros := ""
    for i :=0 ; i < totalLen - len(str); i++ {
        zeros += "0"
    }
    return zeros + str
}
/*
func searchForAddress(address string, blocks chan int, ) {
    for block := range blocks {
        // open block file
        // binary search it
        //


    }

} */

func isPotentialAddress(addr string) bool {
    small := "00000000000000ffffffffffffffffffffffffff"
    largePrefix := "000000000000000000000000"

    if addr <= small || !strings.HasPrefix(addr, largePrefix) {
        //fmt.Println("False bc large or small", addr <= small, addr >= large)
        return false
    }

    if strings.HasSuffix(addr, "00000000") {
        //fmt.Println("False bc has suffix")
        return false
    }

    //fmt.Println("True!")
    return true
}

func getAddress(traces chan []byte) {
    for blockTraces := range traces {
        var traces BlockTraces
        err := json.Unmarshal(blockTraces, &traces)
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
                    addr := string(inputData[24 + i * 64:(i + 1) * 64])
                    if isPotentialAddress(addr) {
                        addresses["0x" + addr + blockAndIdx] = true
                    }
                }
            }

            

            if traces.Result[i].Type == "call" {
                // If it's a call, get the to and from
                from := traces.Result[i].Action.From
                to := traces.Result[i].Action.To
                addresses[from + blockAndIdx] = true
                addresses[to + blockAndIdx] = true
            } else if traces.Result[i].Type == "reward" {
                // if it's a reward, add the miner
                author := traces.Result[i].Action.Author
                addresses[author + blockAndIdx] = true
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
            } else {
                fmt.Println("New trace type:", string(blockTraces))
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
        toWrite := []byte(strings.Join(addressArray[:], "\n"))

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


func getTrace(blocks chan int, traces chan []byte) {
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

        body1, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("Error", err)
        }
        resp.Body.Close()
        
        fmt.Println("Read in block and now sending", block)
        traces <- body1
    }
}

func main() {
    startBlock := 2608034
    numBlocks := 10000

    blocks := make(chan int)
    traces := make(chan []byte)

    // make a bunch of block trace getters
    for i := 0; i < 20; i++ {
        go getTrace(blocks, traces)
    }

    for i := 0; i < 100; i++ {
        go getAddress(traces)
    }

    for block := startBlock; block < startBlock + numBlocks; block++ {
        blocks <- block
    }
    
    // blah, just wait around for ever (have to manuall terminate the process...)
    done := make(chan int)
    <- done
}