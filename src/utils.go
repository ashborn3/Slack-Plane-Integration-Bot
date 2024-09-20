package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type State struct {
	Name string `json:"name"`
}

type Result struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Response struct {
	Results []Result `json:"results"`
}

func CreateStateIDToNameMap(projectIDs []string) (map[string]string, error) {

	idToNameMap := make(map[string]string)
	var tempResult Response

	for _, projectID := range projectIDs {
		url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/%s/states/", slug, projectID)

		req, _ := http.NewRequest("GET", url, nil)

		req.Header.Add("x-api-key", planeToken)

		res, _ := http.DefaultClient.Do(req)

		defer res.Body.Close()
		body, _ := io.ReadAll(res.Body)

		err := json.Unmarshal(body, &tempResult)
		if err != nil {
			log.Print(err)
			continue
		}

		for _, res := range tempResult.Results {
			idToNameMap[res.ID] = res.Name
		}
	}

	return idToNameMap, nil
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

func getStateIDFromString(projectID, stateName string) (string, error) {
	url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/%s/states/", slug, projectID)

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("x-api-key", planeToken)

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var resp Response
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return "", err
	}

	for _, result := range resp.Results {
		if result.Name == stateName {
			return result.ID, nil
		}
	}

	return "", fmt.Errorf("ID for name %s not found", stateName)
}

func PreSetIssueState(issueID string) (string, error) {
	issueResps := FetchIssues(FetchProjects())

	for _, issueResp := range issueResps {
		for _, issue := range issueResp.Results {
			if issue.ID == issueID {
				return issue.Project, nil
			}
		}
	}
	return "", fmt.Errorf("Issue ID %s not found", issueID)
}

func SetIssueState(projectID, issueID, stateName string) error {
	url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/%s/issues/%s/", slug, projectID, issueID)

	stateID, err := getStateIDFromString(projectID, stateName)
	if err != nil {
		return err
	}

	log.Print(stateID)

	payload := map[string]string{
		"state": stateID,
	}

	// Marshal the map to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("PATCH", url, bytes.NewReader(jsonData))

	req.Header.Add("x-api-key", planeToken)
	req.Header.Add("Content-Type", "application/json")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	fmt.Println(res)
	fmt.Println(string(body))
	return nil
}
