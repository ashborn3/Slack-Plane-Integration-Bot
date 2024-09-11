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

func CategorizeIssues(issues []IssuesResponse) (openIssues, inProgressIssues, closedIssues, doneIssues, todoIssues []string) {
	for _, issue := range issues {
		for _, issueData := range issue.Results {
			stateName, err := FetchStateName(issueData.Project, issueData.State)
			fmt.Print("State Name: ", stateName, "\n")
			if err != nil {
				log.Printf("Error fetching state name: %v", err)
				continue
			}
			switch stateName {
			case "Backlog":
				openIssues = append(openIssues, issueData.Name)
			case "Todo":
				todoIssues = append(todoIssues, issueData.Name)
			case "In Progress":
				inProgressIssues = append(inProgressIssues, issueData.Name)
			case "Cancelled":
				closedIssues = append(closedIssues, issueData.Name)
			case "Done":
				doneIssues = append(doneIssues, issueData.Name)
			}
		}
	}
	return
}
