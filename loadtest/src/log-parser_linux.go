package main

import (
	"fmt"
)

func logLine(s string) {
	fmt.Println(s)
}

func init() {
	parser = &resultLogParser{}
}

func (lp *resultLogParser) Parse(log string) {
	return
}

func (lp *resultLogParser) Write(p []byte) (n int, err error) {
	return 0, nil
}

func (lp *resultLogParser) GetResults() *[]*result {
	return &[]*result{}
}
