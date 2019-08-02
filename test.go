package main

import "testing"

func TestAbs(t *testing.T) {
    Options.rpcProvider = "http://localhost:8545"
    Options.indexPath = "./"
    Options.ripePath = "./ripe/"
    Options.unripePath = "./unripe/"
    Options.startBlock = 0
    Options.nBlocks = 10000
    Options.nBlockProcs = 5
    Options.nAddrProcs = 25
    Options.ripeBlock = 8000000
    Options.dockerMode = false

    ProcessBlocks()
}