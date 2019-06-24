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

func concurrentRPC() {
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

func sequentialRPC() {
    blocks := make(chan int)

    // Only make one block processor
    for i := 0; i < 1; i++ {
        go blockProcessor(blocks)
    }

    // Send the blocks to be processed
    for i := 0; i < 100000; i++ {
        blocks <- 5000000
    }
}


func main() {
    start := time.Now()
    concurrentRPC()
    elapsed := time.Since(start)
    fmt.Println("Concurrent RPC took:", elapsed)

    start = time.Now()
    sequentialRPC()
    elapsed = time.Since(start)
    fmt.Println("Sequential RPC took:", elapsed)
}