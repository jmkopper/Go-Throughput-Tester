package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const serverURL = "http://localhost:3000/runtest"

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

type CollectedResults struct {
	ServerTimes []float64 `json:"serverTimes"`
	ClientTimes []float64 `json:"clientTimes"`
}

func readTestArrayFromFile(filename string) ([]Test, error) {
	fileBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var testArray []Test
	err = json.Unmarshal(fileBytes, &testArray)
	if err != nil {
		return nil, err
	}

	return testArray, nil
}

func runTests(testArray *[]Test, apiKey string, numRuns int) CollectedResults {
	testRequest := TestRequest{Secret: apiKey, Tests: *testArray, Budget: 500}
	client := http.DefaultClient
	var collectedResults CollectedResults
	requestBody, _ := json.Marshal(testRequest)
	for i := 0; i < numRuns; i++ {
		start := time.Now()
		request, _ := http.NewRequest(http.MethodPost, serverURL, bytes.NewBuffer(requestBody))
		request.Header.Set("Content-Type", "application/json")

		response, err := client.Do(request)
		if err != nil {
			log.Fatal(err)
		}
		defer response.Body.Close()

		responseBody, _ := ioutil.ReadAll(response.Body)

		if response.StatusCode != http.StatusOK {
			log.Fatalf("Unexpected response code %d: %s", response.StatusCode, string(responseBody))
		}

		var testResponse TestResponse
		if err := json.Unmarshal(responseBody, &testResponse); err != nil {
			log.Fatal(err)
		}

		duration := float64(time.Since(start).Nanoseconds()) / 1e9
		collectedResults.ServerTimes = append(collectedResults.ServerTimes, testResponse.Duration)
		collectedResults.ClientTimes = append(collectedResults.ClientTimes, duration)
	}

	return collectedResults
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	testArray, err := readTestArrayFromFile("tests.json")
	if err != nil {
		log.Fatal(err)
	}

	apiKey := os.Getenv("API_KEY")

	collectedResults := runTests(&testArray, apiKey, 100)

	file, err := os.Create("results.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(collectedResults)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

}
