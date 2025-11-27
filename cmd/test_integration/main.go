package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "http://localhost:8080/api"

func main() {
	fmt.Println("Starting integration tests...")

	// 1. Truncate Database
	fmt.Println("\n[1] Truncating database...")
	if err := sendRequest("DELETE", "/truncate", nil); err != nil {
		fail(err)
	}

	// 2. Initialize Data
	fmt.Println("\n[2] Initializing data (1000 records)...")
	initReq := map[string]interface{}{
		"record_count": 1000,
		"content_size": "small",
	}
	if err := sendRequest("POST", "/initialize", initReq); err != nil {
		fail(err)
	}

	// 3. List Logs
	fmt.Println("\n[3] Listing logs...")
	logs, err := getLogs(nil)
	if err != nil {
		fail(err)
	}
	if len(logs) == 0 {
		fail(fmt.Errorf("expected logs, got 0"))
	}
	fmt.Printf("Got %d logs\n", len(logs))

	// 4. Search Logs (Full Text)
	// We'll search for a common term likely to be in the generated content
	searchTerm := "click"
	fmt.Printf("\n[4] Searching logs for '%s'...\n", searchTerm)
	searchLogs, err := getLogs(map[string]string{"content_like": searchTerm})
	if err != nil {
		fail(err)
	}
	fmt.Printf("Found %d logs matching '%s'\n", len(searchLogs), searchTerm)

	// 5. Filter by Domain
	if len(logs) > 0 {
		domain := logs[0]["domain"].(string)
		fmt.Printf("\n[5] Filtering by domain '%s'...\n", domain)
		domainLogs, err := getLogs(map[string]string{"domain": domain})
		if err != nil {
			fail(err)
		}
		fmt.Printf("Found %d logs for domain '%s'\n", len(domainLogs), domain)

		// Verify all results match the domain
		for _, log := range domainLogs {
			if log["domain"].(string) != domain {
				fail(fmt.Errorf("expected domain %s, got %s", domain, log["domain"]))
			}
		}
	}

	fmt.Println("\n✅ All integration tests passed successfully!")
}

func sendRequest(method, endpoint string, body interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func getLogs(params map[string]string) ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", baseURL+"/logs", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func fail(err error) {
	fmt.Printf("❌ Test failed: %v\n", err)
	os.Exit(1)
}
