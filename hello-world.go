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

func getBlock(block int) {
    defer waitgroup.Done()
    hexBlockNum := fmt.Sprintf("0x%x", block)

    data := Payload{
        "2.0",
        "eth_getBlockByNumber",
        Params{hexBlockNum, false},
        2,
    }

    payloadBytes, err := json.Marshal(data)
    if err != nil {
        fmt.Println(block)
        return
    }

    body := bytes.NewReader(payloadBytes)

    req, err := http.NewRequest("POST", "http://localhost:8545", body)
    if err != nil {
        fmt.Println(block)
        return 
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)

    if err != nil {
        fmt.Println(block)
        return
    }

    defer resp.Body.Close()

    body1, err := ioutil.ReadAll(resp.Body)

    fmt.Println(string(body1))
}

var waitgroup sync.WaitGroup


func main() {
    numBlocks := 1000

    start := time.Now()
    waitgroup.Add(numBlocks)
    for i := 5000000; i < 5000000 + numBlocks; i++ {
        go getBlock(i)    
    }
    waitgroup.Wait()
    elapsed := time.Since(start)

    fmt.Println("Took time:", elapsed)
}