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

type Project struct {
	ID string `json:"id"`
}

type Response struct {
	Count   int       `json:"count"`
	Results []Project `json:"results"`
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

	fetchProjects()

}

func fetchProjects() []string {
	var projects Response

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
	// you can fetch issues for each project
}
