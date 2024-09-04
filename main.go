package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

var (
	slug        string
	planeToken  string
	slackClient *slack.Client
)

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
	PrevPageResults bool        `json:"prev_page_results"`
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

type StateResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func fetchStateName(projectID, stateID string) (string, error) {
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

	var stateResponse StateResponse
	err = json.Unmarshal(body, &stateResponse)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling response body: %v", err)
	}

	return stateResponse.Name, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file!")
		return
	}
	slug = os.Getenv("SLUG")
	planeToken = os.Getenv("PLANE_TOKEN")

	slackClient = slack.New(os.Getenv("SLACK_TOKEN"), slack.OptionDebug(true), slack.OptionAppLevelToken(os.Getenv("SLACK_SOCK_TOKEN")))

	socketClient := socketmode.New(slackClient, socketmode.OptionDebug(true))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go startSocketMode(ctx, slackClient, socketClient)

	sendDailyOverview()
	fmt.Print("sendDailyOverview() executed\n")
	notifyUsersDaily()
	fmt.Print("notifyUsersDaily() executed\n")

	// Set up the cron job
	c := cron.New(cron.WithLocation(time.FixedZone("IST", 5*60*60+30*60))) // IST is UTC+5:30
	_, err = c.AddFunc("0 9 * * *", notifyUsersDaily)                      // Run at 9:00 AM IST every day
	if err != nil {
		log.Fatalf("Error setting up cron job: %v", err)
		return
	}
	c.Start()

	// Run the Socketmode client
	socketClient.Run()
}

func handleEvent(event slackevents.EventsAPIEvent) {
	// Add your event handling logic here
	log.Printf("Received event: %v", event)
}

func startSocketMode(ctx context.Context, slackClient *slack.Client, sockClient *socketmode.Client) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down socketmode listener")
			return
		case event := <-sockClient.Events:
			switch event.Type {
			case socketmode.EventTypeEventsAPI:
				eventsAPI, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Could not type cast the event to the EventsAPI: %v\n", event)
					continue
				}
				sockClient.Ack(*event.Request)
				log.Println(eventsAPI)
				// Handle the event here
				handleEvent(eventsAPI)
			}
		}
	}
}

func loadUserMapping(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	userMapping := make(map[string]string)
	for _, record := range records[1:] { // Skip header row
		planeUUID := record[0]
		slackUserID := record[1]
		userMapping[planeUUID] = slackUserID
	}

	return userMapping, nil
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

func fetchIssues(projectIds []string) []IssuesResponse {
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

func sendIssueDetailsToAssignees(issues []IssuesResponse) int {
	var message string
	userMapping, _ := loadUserMapping("user_mapping.csv")

	for _, issue := range issues {
		for _, issueData := range issue.Results {
			message = fmt.Sprintf("Issue: %s\nPriority: %s\nTarget Date: %s\nDescription: %s\n", issueData.Name, issueData.Priority, issueData.TargetDate, issueData.DescriptionHTML[3:len(issueData.DescriptionHTML)-4])
			for _, assignee := range issueData.Assignees {
				slackUserID, exists := userMapping[assignee]
				if !exists {
					log.Printf("No Slack user ID found for Plane UUID: %s", assignee)
					continue
				}
				fmt.Print("Sending message to: ", slackUserID, "\n")
				channel, _, _, err := slackClient.OpenConversation(&slack.OpenConversationParameters{
					Users: []string{"U07K0L9DNUX"},
				})
				slackClient.PostMessage(channel.ID, slack.MsgOptionText(message, false))
				if err != nil {
					log.Fatalf("Error sending message: %v", err)
					return -1
				}
			}
		}
	}

	return 0
}

func notifyUsersDaily() {
	// Fetch projects
	projectIds := fetchProjects()
	if projectIds == nil {
		log.Println("No projects found")
		return
	}

	// Fetch issues
	issues := fetchIssues(projectIds)
	if issues == nil {
		log.Println("No issues found")
		return
	}

	// Send notifications
	sendIssueDetailsToAssignees(issues)
}

func categorizeIssues(issues []IssuesResponse) (openIssues, inProgressIssues, closedIssues, doneIssues, todoIssues []string) {
	for _, issue := range issues {
		for _, issueData := range issue.Results {
			stateName, err := fetchStateName(issueData.Project, issueData.State)
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

func sendDailyOverview() {
	// Fetch projects
	projectIds := fetchProjects()
	if projectIds == nil {
		log.Println("No projects found")
		return
	}

	// Fetch issues
	issues := fetchIssues(projectIds)
	if issues == nil {
		log.Println("No issues found")
		return
	}

	// Categorize issues
	openIssues, inProgressIssues, closedIssues, doneIssues, todoIssues := categorizeIssues(issues)

	fmt.Print("All Issues: ", openIssues, inProgressIssues, closedIssues, doneIssues, todoIssues, "\n")

	// Create the summary message
	message := "Daily Overview of Issues:\n\n"
	message += "*Open Issues:*\n"
	for _, issue := range openIssues {
		message += issue + "\n"
	}
	message += "\n*In Progress Issues:*\n"
	for _, issue := range inProgressIssues {
		message += issue + "\n"
	}
	message += "\n*Closed Issues:*\n"
	for _, issue := range closedIssues {
		message += issue + "\n"
	}
	message += "\n*Done Issues:*\n"
	for _, issue := range doneIssues {
		message += issue + "\n"
	}
	message += "\n*Todo Issues:*\n"
	for _, issue := range todoIssues {
		message += issue + "\n"
	}

	// Send the message to a Slack channel
	channelID := os.Getenv("SLACK_OVERVIEW_CHANNEL_ID")
	_, _, err := slackClient.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
}
