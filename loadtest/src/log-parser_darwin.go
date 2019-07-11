package main

import (
	"os"
)

func logLine(s string) {
	parser.Write([]byte(s + "\n"))
}

func init() {
	parser = makeResultLogParser(resultsBuffer, os.Stdout)
}

func (lp *resultLogParser) Parse(log string) {
	matches := r.FindAllString(log, -1)
	if len(matches) > 0 {
		newRes := *lp.results
		for _, match := range matches {
			result := decodeResult(match)

			newRes = append(newRes, result)
		}

		lp.results = &newRes
	}
}

func (lp *resultLogParser) Write(p []byte) (n int, err error) {
	stringRep := string(p)
	lp.Parse(stringRep)
	return lp.writer.Write(p)
}

func (lp *resultLogParser) GetResults() *[]*result {
	return lp.results
}
