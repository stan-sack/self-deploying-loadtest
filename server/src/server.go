package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"strconv"
	"time"
)

func main() {
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":80", nil); err != nil {
		panic(err)
	}
}

func timeTrack(start time.Time, name string) time.Duration {
	elapsed := time.Since(start)
	log.Printf("function %s took %s", name, elapsed)
	return elapsed

}

func handler(w http.ResponseWriter, r *http.Request) {

	//timer starts
	startTime := time.Now()

	//loop multiple times to show significant amount of time
	for i := 1; i <= 100000; i++ {

		//a time consuming calulation
		//like calculate the hash of the request origin
		hash(r.Header.Get("X-Forwarded-For"))
	}

	// how much time has elapsed since we started the timer
	duration := (timeTrack(startTime, "handler"))

	//calculating duration from nano second to milliseconds
	durationMilli := int(time.Duration(duration) / time.Millisecond)

	//printing in terminal logs
	fmt.Println("milliseconds:", int64(durationMilli))

	// convert the integer number of milliseconds to string:
	durationString := strconv.Itoa(durationMilli)

	//return that string
	w.Write([]byte(durationString))

}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
