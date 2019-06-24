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

func getBlock(block int, res chan Result) {

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

    defer resp.Body.Close()

    body1, err := ioutil.ReadAll(resp.Body)
    res <- Result{block, string(body1)}

    //fmt.Println(string(body1))
}

func main() {
    start := time.Now()

    numBlocks := 1500
    res := make(chan Result)

    for i := 5000000; i < 5000000 + numBlocks; i++ {
        go getBlock(i, res)    
    }

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
    wg.Wait()

    elapsed := time.Since(start)
    fmt.Println("Concurrent took time:", elapsed)
    //fmt.Println(m)



    /*
    start1 := time.Now()
    for i := 5000000; i < 5000000 + numBlocks; i++ {
        getBlock(i, false)    
    }
    elapsed1 := time.Since(start1) */

    //fmt.Println("Nonconcurrent took time:", elapsed1)
}