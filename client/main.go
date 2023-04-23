package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const serverURL = "http://localhost:3000/runtest"

type Test struct {
	Value float64 `json:"x"`
	Cost  float64 `json:"y"`
	Name  string  `json:"name"`
}

type TestArray []Test

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

func main() {
	client := http.DefaultClient

	testArray, err := readTestArrayFromFile("tests.json")
	if err != nil {
		log.Fatal(err)
	}

	requestBody, _ := json.Marshal(testArray)
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

	var testResults TestArray
	if err := json.Unmarshal(responseBody, &testResults); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Test results: %+v\n", testResults)
}
