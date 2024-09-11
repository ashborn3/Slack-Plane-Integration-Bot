package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type State struct {
	Name string `json:"name"`
}

func FetchStateName(projectID, stateID string) (string, error) {
	url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/%s/states/%s/", os.Getenv("SLUG"), projectID, stateID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("x-api-key", os.Getenv("PLANE_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var stateResponse State
	err = json.Unmarshal(body, &stateResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response body: %v", err)
	}

	return stateResponse.Name, nil
}

func SetIssueState(projectID, stateID, stateName string) error {
	url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/%s/states/%s/", slug, projectID, stateID)

	payload := strings.NewReader(fmt.Sprintf("{\n  \"name\": \"%s\"\n}", stateName))

	req, _ := http.NewRequest("PATCH", url, payload)

	req.Header.Add("x-api-key", planeToken)
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))
	return nil
}
