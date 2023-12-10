package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"
)

type RequestBody struct {
	ToSort [][]int `json:"to_sort"`
}

type ResponseBody struct {
	SortedArrays [][]int `json:"sorted_arrays"`
	TimeNS       int64   `json:"time_ns"`
}

func processSingle(w http.ResponseWriter, r *http.Request) {
	var requestBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	sortedArrays := make([][]int, len(requestBody.ToSort))
	for i, subArray := range requestBody.ToSort {
		sort.Ints(subArray)
		sortedArrays[i] = subArray
	}

	timeTaken := time.Since(startTime).Nanoseconds()

	response := ResponseBody{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func processConcurrent(w http.ResponseWriter, r *http.Request) {
	var requestBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	wg.Add(len(requestBody.ToSort))

	resultCh := make(chan []int, len(requestBody.ToSort))

	for _, subArray := range requestBody.ToSort {
		go func(arr []int) {
			defer wg.Done()
			sort.Ints(arr)
			resultCh <- arr
		}(subArray)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	sortedArrays := make([][]int, 0, len(requestBody.ToSort))
	for result := range resultCh {
		sortedArrays = append(sortedArrays, result)
	}

	timeTaken := time.Since(startTime).Nanoseconds()

	response := ResponseBody{
		SortedArrays: sortedArrays,
		TimeNS:       timeTaken,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/process-single", processSingle)
	http.HandleFunc("/process-concurrent", processConcurrent)

	log.Fatal(http.ListenAndServe(":8000", nil))
}
