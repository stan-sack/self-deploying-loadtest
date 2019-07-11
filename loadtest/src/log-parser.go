package main

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type result struct {
	hashDurationMillis  int
	success             bool
	totalDurationMillis int
}

type LogParser interface {
	Parse(string)
	Write(p []byte) (n int, err error)
	GetResults() *[]*result
}

type resultLogParser struct {
	results *[]*result
	writer  io.Writer
}

func makeResultLogParser(results *[]*result, writer io.Writer) *resultLogParser {
	return &resultLogParser{
		results: results,
		writer:  writer,
	}
}

const startResultTag = "-~:"
const endResultTag = ":~-"
const delimiter = ":"

func encodeResult(res *result) string {
	if !res.success {
		return fmt.Sprintf(
			"%s%d%s%s%s%d%s",
			startResultTag,
			0, delimiter,
			"false",
			delimiter,
			0,
			endResultTag)
	}
	return fmt.Sprintf(
		"%s%d%s%s%s%d%s",
		startResultTag,
		res.hashDurationMillis,
		delimiter,
		"true",
		delimiter,
		res.totalDurationMillis,
		endResultTag)
}

func decodeResult(s string) *result {
	stripped := strings.Replace(s, startResultTag, "", 1)
	stripped = strings.Replace(stripped, endResultTag, "", 1)
	parts := strings.Split(stripped, delimiter)
	hashDurationMillis, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}
	success, err := strconv.ParseBool(parts[1])
	if err != nil {
		panic(err)
	}
	totalDurationMillis, err := strconv.Atoi(parts[2])
	if err != nil {
		panic(err)
	}

	return &result{
		hashDurationMillis:  hashDurationMillis,
		success:             success,
		totalDurationMillis: totalDurationMillis,
	}

}

var r = regexp.MustCompile(fmt.Sprintf(`%s(.*?)%s`, startResultTag, endResultTag))
