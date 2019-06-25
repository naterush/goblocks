package main

import (
    "fmt"
    "net/http"
    "encoding/json"
    "bytes"
    "io/ioutil"
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
            Author     string `json:"author"`
			RewardType string `json:"rewardType"`
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

func getAddress(traces chan []byte, done chan int) {
    for blockTraces := range traces {
        var traces BlockTraces
        err := json.Unmarshal(blockTraces, &traces)
	    if err != nil {
	    	fmt.Println("error:", err)
        }
        addresses := make(map[string]bool)

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
            } else {
                fmt.Println("New trace type:", traces.Result[i].Type)
            }
        }

        // TODO: sort addresses

        // Write them to a file for a specific block

        for address := range addresses {
            fmt.Println(address)
        }
        done <- 1
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
        fmt.Println(string(body1))

        traces <- body1
    }
}

func main() {
    done := make(chan int)
    blocks := make(chan int)
    traces := make(chan []byte)

    // Only make one block processor, send it a block
    go getTrace(blocks, traces)
    blocks <- 2608034

    // Make a trace receiver, to process
    go getAddress(traces, done)
    <- done

}