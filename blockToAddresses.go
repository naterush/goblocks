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

func getAddress(traces chan []byte, done chan int) {
    for blockTraces := range traces {
        var traces BlockTraces
        err := json.Unmarshal(blockTraces, &traces)
	    if err != nil {
	    	fmt.Println("error:", err)
        }
        addresses := make(map[string]bool)
        fmt.Println("Now processing block", traces.Result[0].BlockNumber)

        for i :=0; i<len(traces.Result); i++ {
            if traces.Result[i].Type == "call" {
                // If it's a call, get the to and from
                from := traces.Result[i].Action.From
                to := traces.Result[i].Action.To
                addresses[from] = true
                addresses[to] = true
            } else if traces.Result[i].Type == "reward" {
                // if it's a reward, add the miner
                author := traces.Result[i].Action.Author
                addresses[author] = true
            } else if traces.Result[i].Type == "suicide" {
                // add the contract that died, and where it sent it's money
                address := traces.Result[i].Action.Address
                refundAddress := traces.Result[i].Action.RefundAddress
                addresses[address] = true
                addresses[refundAddress] = true
            } else if traces.Result[i].Type == "create" {
                // add the creator, and the new address name
                from := traces.Result[i].Action.From
                address := traces.Result[i].Result.Address
                addresses[from] = true
                addresses[address] = true
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

        // write this array to a file
        // at least one result (as the miner got a reward)
        fileName := "file" + strconv.Itoa(traces.Result[0].BlockNumber) + ".txt"
        err = ioutil.WriteFile(fileName, toWrite, 0777)
        if err != nil {
            fmt.Println("Error writing file:", err)
        }
        fmt.Println("Finished processing", traces.Result[0].BlockNumber)

    }
    done <- 1

}


func getTrace(blocks chan int, traces chan []byte, readDone chan int) {
    // Process blocks untill the blocks channel closes
    for block := range blocks {
        if block == -1 {
            readDone <- 1
        }
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
        fmt.Println(string(body1))
        
        fmt.Println("Read in block and now sending", block)
        traces <- body1
    }
}

func main() {
    readDone := make(chan int)
    processDone := make(chan int)
    blocks := make(chan int)
    traces := make(chan []byte)

    // make a bunch of block processors
    for i := 0; i < 25; i++ {
        go getTrace(blocks, traces, readDone)
    }

    for i := 0; i < 100; i++ {
        go getAddress(traces, processDone)
    }

    for block := 5000000; block < 5000000 + 1; block++ {
        blocks <- block
    }
    // when the reading is done
    <- readDone
    close(blocks)
    close(traces)
    // and then wait for the write to finish
    <- processDone
}