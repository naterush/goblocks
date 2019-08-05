package main

import "testing"
import "./cmd"

func TestAbs(t *testing.T) {
    cmd.ProcessBlocks("http://localhost:8545", 5, 25, 5000000, 250000, 8000000, "/unripe/", "/ripe/")
}