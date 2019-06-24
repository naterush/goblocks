package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
    "time"
    "sync"
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

func blockProcessor(blocks chan int, res chan Result) {
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
            res <- Result{block, ""}
            fmt.Println("Error:", err)
            return
        }
    
        body := bytes.NewReader(payloadBytes)
    
        req, err := http.NewRequest("POST", "http://localhost:8545", body)
        if err != nil {
            res <- Result{block, ""}
            fmt.Println("Error:", err)
            return 
        }
        req.Header.Set("Content-Type", "application/json")
    
        resp, err := http.DefaultClient.Do(req)
    
        if err != nil {
            res <- Result{block, ""}
            fmt.Println("Error:", err)
            return
        }
    
        defer resp.Body.Close()
    
        body1, err := ioutil.ReadAll(resp.Body)
        res <- Result{block, string(body1)}
    }
}

func main() {
    start := time.Now()

    // Number of blocks to process
    numBlocks := 1000

    var blocks chan int
    blocks = make(chan int)

    res := make(chan Result)

    // Make 500 dedicated block processors
    for i := 0; i < 100; i++ {
        go blockProcessor(blocks, res)
    }

    // Set up the receiver
    var m map[int]string
    m = make(map[int]string)
    var wg sync.WaitGroup
    wg.Add(1)
    go func () {
        numReceived := 0
    
        for res := range res {
            m[res.block] = res.body
    
            numReceived = numReceived + 1
            if numReceived == numBlocks {
                wg.Done()
            }
        }
    }()

    for i := 5000000; i < 5000000 + numBlocks; i++ {
        blocks <- i
    }

    wg.Wait()

    elapsed := time.Since(start)
    fmt.Println("Concurrent took time:", elapsed)



    // Have a channel we send blocks down -- that closes when all blocks have been sent
    // Have a finite number of processes (about 1000)
        // Gets the block from the RPC, and sends to the result channel
        // Have a result reciever always chillin
    
    // How do we get out of a receiver into a map








    //fmt.Println(m)



    /*
    start1 := time.Now()
    for i := 5000000; i < 5000000 + numBlocks; i++ {
        getBlock(i, false)    
    }
    elapsed1 := time.Since(start1) */

    //fmt.Println("Nonconcurrent took time:", elapsed1)
}