package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

var slackClient *slack.Client
var slug string
var planeToken string

type ProjectsResponse struct {
	Count   int       `json:"count"`
	Results []Project `json:"results"`
}

type Project struct {
	ID string `json:"id"`
}

type IssuesResponse struct {
	GroupedBy       interface{} `json:"grouped_by"`
	SubGroupedBy    interface{} `json:"sub_grouped_by"`
	TotalCount      int         `json:"total_count"`
	NextCursor      string      `json:"next_cursor"`
	PrevCursor      string      `json:"prev_cursor"`
	NextPageResults bool        `json:"next_page_results"`
	PrevPageResults bool        `json:"prev_page_results"`Response
	Count           int         `json:"count"`
	TotalPages      int         `json:"total_pages"`
	TotalResults    int         `json:"total_results"`
	ExtraStats      interface{} `json:"extra_stats"`
	Results         []Issue     `json:"results"`
}

type Issue struct {
	ID                string      `json:"id"`
	TypeID            interface{} `json:"type_id"`
	CreatedAt         string      `json:"created_at"`
	UpdatedAt         string      `json:"updated_at"`
	DeletedAt         interface{} `json:"deleted_at"`
	Point             interface{} `json:"point"`
	Name              string      `json:"name"`
	DescriptionHTML   string      `json:"description_html"`
	DescriptionBinary interface{} `json:"description_binary"`
	Priority          string      `json:"priority"`
	StartDate         interface{} `json:"start_date"`
	TargetDate        interface{} `json:"target_date"`
	SequenceID        int         `json:"sequence_id"`
	SortOrder         float64     `json:"sort_order"`
	CompletedAt       interface{} `json:"completed_at"`
	ArchivedAt        interface{} `json:"archived_at"`
	IsDraft           bool        `json:"is_draft"`
	ExternalSource    interface{} `json:"external_source"`
	ExternalID        interface{} `json:"external_id"`
	CreatedBy         string      `json:"created_by"`
	UpdatedBy         string      `json:"updated_by"`
	Project           string      `json:"project"`
	Workspace         string      `json:"workspace"`
	Parent            interface{} `json:"parent"`
	State             string      `json:"state"`
	EstimatePoint     interface{} `json:"estimate_point"`
	Type              interface{} `json:"type"`
	Assignees         []string    `json:"assignees"`
	Labels            []string    `json:"labels"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file!")
		return
	}
	slug = os.Getenv("SLUG")
	planeToken = os.Getenv("PLANE_TOKEN")

	slackClient = slack.New(os.Getenv("SLACK_TOKEN"))

	fetchIssues(fetchProjects())

}

func fetchProjects() []string {
	var projects ProjectsResponse

	url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/", os.Getenv("SLUG"))

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

	err = json.Unmarshal(body, &projects)
	if err != nil {
		log.Fatalf("Error unmarshaling response body: %v", err)
		return nil
	}

	projectCount := int(projects.Count)
	projectResults := projects.Results

	projectId := make([]string, projectCount)

	for i := 0; i < projectCount; i++ {
		projectId[i] = projectResults[i].ID
	}

	return projectId
}

func fetchIssues(projectIds []string) {
	for _, projectId := range projectIds {
		url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/%s/issues/", slug, projectId)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
			return
		}

		req.Header.Add("x-api-key", planeToken)

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatalf("Error sending request: %v", err)
			return
		}

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
			return
		}

		// Write response to JSON file
		fileName := fmt.Sprintf("project_%s_issues.json", projectId)
		err = os.WriteFile(fileName, body, 0644)
		if err != nil {
			log.Fatalf("Error writing response to file: %v", err)
			return
		}

		fmt.Printf("Issues for project %s fetched and written to %s\n", projectId, fileName)
	}
}
