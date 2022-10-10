package lightning

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Set admin.macaroon hex.
var (
	Macaroon = ""
	LNUrl    = GoDotEnvVariable("LND_HOST")
)

func loadMacaroon() (macaroon string) {
	file, err := os.Open(GoDotEnvVariable("MACAROON_LOCATION"))
	if err != nil {
		return GoDotEnvVariable("MACAROON")
	}

	defer file.Close()

	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)

	scanner.Split(bufio.ScanRunes)

	var finalResult []string
	var finalOriginal []string

	for scanner.Scan() {
		original := fmt.Sprintf("%s ", scanner.Text())

		finalOriginal = append(finalOriginal, original)

		hexstring := fmt.Sprintf("%x ", scanner.Text())

		finalResult = append(finalResult, hexstring)
	}

	return finalResult[0]
}

func sendGetRequest(endpoint string) (*http.Response, error) {
	myMacaroon := loadMacaroon()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", LNUrl+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", myMacaroon)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func sendPostRequestJSON(endpoint string, payload interface{}) (*http.Response, error) {
	myMacaroon := loadMacaroon()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	jsonStr, err := json.Marshal(payload)

	req, err := http.NewRequest("POST", LNUrl+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", myMacaroon)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return resp, nil
}

func sendPostRequest(endpoint string, payload string) (*http.Response, error) {
	myMacaroon := loadMacaroon()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
	}

	jsonStr := []byte(payload)

	req, err := http.NewRequest("POST", LNUrl+endpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Grpc-Metadata-macaroon", myMacaroon)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// use godot package to load/read the .env file and
// return the value of the key.
func GoDotEnvVariable(key string) string {
	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}
