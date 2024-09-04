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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file!")
		return
	}

	slackClient = slack.New(os.Getenv("SLACK_TOKEN"))

	fetchProjects()

}

func fetchProjects() []string {
	var projects []map[string]interface{}

	url := fmt.Sprintf("https://api.plane.so/api/v1/workspaces/%s/projects/", os.Getenv("SLUG"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return
	}

	req.Header.Add("x-api-key", os.Getenv("PLANE_TOKEN"))

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

	err = json.Unmarshal(body, &projects)
	if err != nil {
		log.Fatalf("Error unmarshaling response body: %v", err)
		return
	}

	projectCount := int(projects["count"])
	projectResults := []string(projects["results"])

	projectId := make([]string, projectCount)

	for i := 0; i < projectCount; i++ {
		projectId[i] = projectResults[i]["id"]
	}

	return projectId
}
