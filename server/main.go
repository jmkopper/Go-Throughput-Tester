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

func processTestData(tests TestArray, budget float64) TestArray {
	log.Printf("Running process datas")
	sort.Slice(tests, func(i, j int) bool { return tests[i].Value/tests[i].Cost < tests[j].Value/tests[j].Cost })
	var spent float64
	var results TestArray
	spent = 0
	for i := 0; i < len(tests) && spent <= budget; i++ {
		results = append(results, tests[i])
		spent += tests[i].Cost
	}
	return results
}

type testHandler struct{}

func (th *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var testArray TestArray
	if err := json.NewDecoder(r.Body).Decode(&testArray); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	testResults := make(chan TestArray)
	go func() {
		testResults <- processTestData(testArray, 100.0)
	}()

	select {
	case resp := <-testResults:
		json.NewEncoder(w).Encode(resp)
	case <-time.After(time.Second * 5):
		http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
	}
}

func main() {
	mux := http.NewServeMux()
	th := &testHandler{}
	mux.Handle("/runtest", th)

	listenAddr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}

	log.Printf("Listening on http://localhost%s\n", listenAddr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to serve: %v", err)
	}
}
