package main

import "testing"
import "./cmd"

func TestAbs(t *testing.T) {
    cmd.ProcessBlocks("http://localhost:8545", 5, 25, 2400000, 63000, 8000000, "./unripe/", "./ripe/")
}