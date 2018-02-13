package main

import (
	bt "go.opencensus.io/benchmarks/trace"
)

func main() {
	bt.Generate(".")
}
