package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
    //"io/ioutil"
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
    hexBlockNum := fmt.Sprintf("0x%x", block)

    data := Payload{
        "2.0",
        "eth_getBlockByNumber",
        Params{hexBlockNum, false},
        2,
    }

    payloadBytes, err := json.Marshal(data)
    if err != nil {
        return
    }

    body := bytes.NewReader(payloadBytes)

    req, err := http.NewRequest("POST", "http://localhost:8545", body)
    if err != nil {
        return 
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)

    if err != nil {
        return
    }

    defer resp.Body.Close()

    //body1, err := ioutil.ReadAll(resp.Body)

    //fmt.Println(string(body1))
    fmt.Println(block)
    waitgroup.Done()
}

var waitgroup sync.WaitGroup


func main() {
    start := time.Now()
    waitgroup.Add(5005000 - 5000000 - 1)
    for i := 5000000; i < 5005000; i++ {
        go getBlock(i)    
    }
    elapsed := time.Since(start)
    waitgroup.Wait()

    fmt.Println("Took time:", elapsed)
}