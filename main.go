package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Counter struct {
	Counter   int64
	UpdatedAt time.Time
}

type ResponseStruct struct {
	Message  string
	LastTime time.Time
	Status   int64
}

var counter = &Counter{}
var counterMutex sync.Mutex

var Filename = "count.txt"

func main() {
	ReadDataFromFile()
	go func() {
		for range time.Tick(time.Second) {
			FlushDataInFile()
		}
	}()

	http.HandleFunc("/", RequestCounter)
	//for testing purpose
	http.HandleFunc("/test", concurrentGet)
	port := 8080
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func RequestCounter(response http.ResponseWriter, _ *http.Request) {
	counter = IncrementCounter()
	APIReturnResponse(counter, response)
	return
}

func FlushDataInFile() {
	counterMutex.Lock()
	defer counterMutex.Unlock()

	now := time.Now()
	if counter.UpdatedAt.IsZero() || now.Sub(counter.UpdatedAt) >= time.Minute {
		counter.Counter = 0
		counter.UpdatedAt = now
	}
	WriteIntoFile()

}

func ReadDataFromFile() {
	file, err := os.Open(Filename)
	defer file.Close()
	if err != nil {
		println(err, "while opening")
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(counter)
	if err != nil {
		println(err, "while opening and decoding")
	}
	return
}

func WriteIntoFile() {
	file, err := os.Create(Filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	now := time.Now()
	counter.UpdatedAt = now
	err = encoder.Encode(counter)
	if err != nil {
		log.Fatal(err)
	}
}

func IncrementCounter() *Counter {
	counterMutex.Lock()
	defer counterMutex.Unlock()
	counter.Counter++
	return counter
}

func APIReturnResponse(res *Counter, responseWrite http.ResponseWriter) {
	strNumber := strconv.FormatInt(res.Counter, 10)
	response := ResponseStruct{Message: "Number of Request in last are " + strNumber, Status: 200, LastTime: res.UpdatedAt}
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(responseWrite, "Error encoding JSON", http.StatusInternalServerError)
		return
	}
	responseWrite.Header().Set("Content-Type", "application/json")
	responseWrite.Write(jsonData)
}

func concurrentGet(response http.ResponseWriter, _ *http.Request) {
	numRequests := 99999
	url := "http://localhost:8080/"

	for i := 0; i < numRequests; i++ {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error in request %d: %v\n", i, err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("Request %d status: %s\n", i, resp.Status)
	}
}
