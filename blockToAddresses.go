package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
    "time"
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
			CallType string `json:"callType"`
			From     string `json:"from"`
			Gas      string `json:"gas"`
			Input    string `json:"input"`
			To       string `json:"to"`
			Value    string `json:"value"`
		} `json:"action,omitempty"`
		BlockHash   string `json:"blockHash"`
		BlockNumber int    `json:"blockNumber"`
		Result      struct {
			GasUsed string `json:"gasUsed"`
			Output  string `json:"output"`
		} `json:"result"`
		Subtraces           int           `json:"subtraces"`
		TraceAddress        []interface{} `json:"traceAddress"`
		TransactionHash     string        `json:"transactionHash"`
		TransactionPosition int           `json:"transactionPosition"`
		Type                string        `json:"type"`
	} `json:"result"`
	ID int `json:"id"`
}


func traceProcessor(blocks chan int) {
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
        fmt.Println(string(body1))

        var traces BlockTraces
        err = json.Unmarshal(body1, &traces)
	    if err != nil {
	    	fmt.Println("error:", err)
	    }
	    fmt.Printf("%+v", traces)
    }
}


func blockProcessor(blocks chan int) {
    // Process blocks untill the blocks channel closes
    for block := range blocks {
        hexBlockNum := fmt.Sprintf("0x%x", block)

        data := Payload{
            "2.0",
            "eth_getBlockByNumber",
            Params{hexBlockNum, false},
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

        _, err = ioutil.ReadAll(resp.Body)
        if err != nil {
            fmt.Printf("Error", err)
        }
        resp.Body.Close()
    }
}

func concurrentHTTP() {
    blocks := make(chan int)

    // Make 250 block processors
    for i := 0; i < 250; i++ {
        go blockProcessor(blocks)
    }

    // Send the blocks to be processed
    for i := 0; i < 100000; i++ {
        blocks <- 5000000
    }
}

func sequentialHTTP() {
    blocks := make(chan int)

    // Only make one block processor
    for i := 0; i < 1; i++ {
        go traceProcessor(blocks)
    }

    // Send the blocks to be processed
    for i := 0; i < 2; i++ {
        blocks <- 7223970
    }
}


func main() {
    //start := time.Now()
    //concurrentHTTP()
    //elapsed := time.Since(start)
    //fmt.Println("Concurrent http took:", elapsed)

    start := time.Now()
    sequentialHTTP()
    elapsed := time.Since(start)
    fmt.Println("Sequential http took:", elapsed)
}

// Make 100 trace processors

// Send each processor a range of blocks to request the traces for
// Request that range of traces
    // Get all the accounts out of there
    // But also get all the transaction hashes, and then request those



// Make 500 "processors," each that watch a block channel

// When they receive a block:
    // Get the number of transactions in that block
    // Get the 