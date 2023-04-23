package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/joho/godotenv"
)

const port = 3000

type Test struct {
	Value int `json:"value"`
}

type TestRequest struct {
	Secret string `json:"secret"`
	Tests  []Test `json:"tests"`
	Budget int    `json:"budget"`
}

type TestResponse struct {
	TestResults []Test  `json:"testResults"`
	Duration    float64 `json:"duration"`
}

func processTestData(tests []Test, budget int) []Test {
	sort.Slice(tests, func(i, j int) bool { return tests[i].Value < tests[j].Value })
	var results []Test
	for i := 0; i < len(tests) && tests[i].Value < budget; i++ {
		results = append(results, tests[i])
	}
	return results
}

type testHandler struct {
	apiKey string
}

func (th *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var testRequest TestRequest
	if err := json.NewDecoder(r.Body).Decode(&testRequest); err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if testRequest.Secret != th.apiKey {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	testResults := make(chan []Test)
	start := time.Now()
	go func() {
		testResults <- processTestData(testRequest.Tests, testRequest.Budget)
	}()

	select {
	case resp := <-testResults:
		duration := float64(time.Since(start).Nanoseconds()) / 1e9
		json.NewEncoder(w).Encode(TestResponse{TestResults: resp, Duration: duration})
	case <-time.After(time.Second * 5):
		http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mux := http.NewServeMux()
	th := &testHandler{os.Getenv("API_KEY")}
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
