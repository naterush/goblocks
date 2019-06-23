package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
)


func test() {

}

func main() {
    type Payload struct {
        Jsonrpc string        `json:"jsonrpc"`
        Method  string        `json:"method"`
        Params  interface{} `json:"params"`
        ID      int           `json:"id"`
    }

    data := Payload{
        "2.0",
        "eth_getBlockByNumber",
        "0x1b4",
        1,
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
    fmt.Println("DONE")
}