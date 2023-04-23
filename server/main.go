package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

const port = 3000

type Test struct {
	Value float64 `json:"x"`
	Cost  float64 `json:"y"`
	Name  string  `json:"name"`
}

type TestArray []Test

func (t TestArray) Len() int {
	return len(t)
}

func (t TestArray) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t TestArray) Less(i, j int) bool {
	return t[i].Value/t[i].Cost < t[j].Value/t[j].Cost
}

type testHandler struct{}

func processTestData(t TestArray, budget float64) TestArray {
	sort.Sort(t)
	var result TestArray
	var spent float64
	spent = 0
	i := 0
	for spent <= budget {
		result = append(result, t[i])
		i += 1
		spent += t[i].Cost
	}
	return result
}

func (th testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var testArray TestArray
	if err := json.NewDecoder(r.Body).Decode(&testArray); err != nil {
		log.Println(err)
	}

	testResults := make(chan TestArray)
	go processTestData(<-testResults, 100.0)

	select {
	case resp := <-testResults:
		json.NewEncoder(w).Encode(resp)
	case <-time.After(time.Second * 5):
		fmt.Fprintf(w, "timeout")
	}
}

func main() {
	mux := http.NewServeMux()
	th := testHandler{}

	mux.Handle("/runtest", th)
	listenAt := fmt.Sprintf(":%d", port)
	log.Printf("Listening on: http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(listenAt, mux))
}
