package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type IssuesResponse struct {
	Results []Issue `json:"results"`
}

type Issue struct {
	Name            string   `json:"name"`
	Priority        string   `json:"priority"`
	Labels          []string `json:"labels"`
	ID              string   `json:"id"`
	Project         string   `json:"project"`
	State           string   `json:"state"`
	DescriptionHTML string   `json:"description_html"`
	Assignees       []string `json:"assignees"`
	DueDate         string   `json:"target_date"`
}

func FetchIssues(projectIds []string) []IssuesResponse {
	var issues []IssuesResponse

	for _, projectId := range projectIds {
		url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/%s/issues/", slug, projectId)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
			return nil
		}

		req.Header.Add("x-api-key", planeToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
			return nil
		}

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
			return nil
		}

		var issueData IssuesResponse
		err = json.Unmarshal(body, &issueData)
		if err != nil {
			log.Fatalf("Error unmarshaling response body: %v", err)
			return nil
		}
		issues = append(issues, issueData)
	}

	return issues
}

func CategorizeIssues(issues []IssuesResponse) map[string][]Issue {
	stateMap, err := CreateStateIDToNameMap(FetchProjects())

	result := make(map[string][]Issue)

	if err != nil {
		log.Fatalf("Error creating state map: %v", err)
		return nil
	}

	for _, issue := range issues {
		for _, issueData := range issue.Results {
			stateName, exists := stateMap[issueData.State]
			if !exists {
				log.Printf("State ID %s not found in state map", issueData.State)
				continue
			}
			issueData.State = stateName
			result[stateName] = append(result[stateName], issueData)
		}
	}
	return result
}
