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
	Value float64 `json:"x"`
	Cost  float64 `json:"y"`
}

type TestArray []Test

type TestRequest struct {
	Secret string `json:"secret"`
	Tests  TestArray
}

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

	log.Printf("server secret: %s client secret: %s", th.apiKey, testRequest.Secret)

	if testRequest.Secret != th.apiKey {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}

	testResults := make(chan TestArray)
	go func() {
		testResults <- processTestData(testRequest.Tests, 100.0)
	}()

	select {
	case resp := <-testResults:
		json.NewEncoder(w).Encode(resp)
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
