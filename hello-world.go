package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
)


func test() {

}

func main() {
    type Params []interface{}

    type Payload struct {
        Jsonrpc string        `json:"jsonrpc"`
        Method  string        `json:"method"`
        Params                 `json:"params"`
        ID      int           `json:"id"`
    }

    data := Payload{
        "2.0",
        "eth_getBlockByNumber",
        Params{"0x1b4", true},
        2,
    }

    payloadBytes, err := json.Marshal(data)
    if err != nil {
        return
    }
    //fmt.Println(data)
    body := bytes.NewReader(payloadBytes)
    //buf := new(bytes.Buffer)
    //buf.ReadFrom(body)
    //s := buf.String()

    //fmt.Println(s)

    req, err := http.NewRequest("POST", "http://localhost:8545", body)
    if err != nil {
        return 
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    fmt.Println("Resp:", resp)
    fmt.Println("Error:", err)

    if err != nil {
        return
    }

    defer resp.Body.Close()

    body1, err := ioutil.ReadAll(resp.Body)

    fmt.Println("Body1:", string(body1))
}