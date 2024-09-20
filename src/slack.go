package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
)

func HandleSlashCommand(cmd slack.SlashCommand, slackClient *slack.Client) {
	response := fmt.Sprintf("You invoked the slash command: %s with text: %s", cmd.Command, cmd.Text)
	if cmd.Command == "/register" {
		err := ManageUserMapping(cmd.Text, cmd.UserID)
		if err != nil {
			response = fmt.Sprintf("Error adding user mapping: %v", err)
		} else {
			response = "User mapping added successfully"
		}
	} else {
		if cmd.Command == "/issueupdate" {
			parts := strings.SplitN(cmd.Text, " ", 2)
			if len(parts) != 2 {
				response = "Invalid input. Please provide issue ID and state name."
			} else {
				pID, err := PreSetIssueState(parts[0])
				if err != nil {
					response = fmt.Sprintf("Error updating issue state: %v", err)
				} else {
					err = SetIssueState(pID, parts[0], parts[1])
					if err != nil {
						response = fmt.Sprintf("Error updating issue state: %v", err)
					} else {
						response = "Issue state updated successfully for issue ID: " + parts[0]
					}
				}
			}
		}
	}
	// Send the response as an ephemeral message
	_, _, err := slackClient.PostMessage(
		cmd.ChannelID,
		slack.MsgOptionText(response, false),
		slack.MsgOptionResponseURL(cmd.ResponseURL, slack.ResponseTypeEphemeral),
	)
	if err != nil {
		log.Printf("Error responding to slash command: %v", err)
	}
}

func StartSocketMode(ctx context.Context, slackClient *slack.Client, sockClient *socketmode.Client) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down socketmode listener")
			return
		case event := <-sockClient.Events:
			switch event.Type {
			case socketmode.EventTypeSlashCommand:
				cmd, ok := event.Data.(slack.SlashCommand)
				if !ok {
					log.Printf("Could not type cast the event to SlashCommand: %v\n", event)
					continue
				}
				sockClient.Ack(*event.Request)
				HandleSlashCommand(cmd, slackClient)
			case socketmode.EventTypeEventsAPI:
				eventsAPI, ok := event.Data.(slackevents.EventsAPIEvent)
				if !ok {
					log.Printf("Could not type cast the event to the EventsAPI: %v\n", event)
					continue
				}
				sockClient.Ack(*event.Request)
				log.Println(eventsAPI)
			}
		}
	}
}

func SendIssueDetailsToAssignees(issues []IssuesResponse) int {
	var message string
	userMapping, _ := LoadUserMapping("user_mapping.csv")
	for _, issue := range issues {
		for _, issueData := range issue.Results {

			message = fmt.Sprintf(
				"Issue: %s\nIssue ID: %s\nDue Date: %s\nPriority: %s\nLink: <%s|Plane>\nDescription: %s\n",
				issueData.Name,
				issueData.ID,
				issueData.DueDate,
				issueData.Priority,
				fmt.Sprintf("https://app.plane.so/%s/projects/%s/issues/%s", slug, issueData.Project, issueData.ID),
				issueData.DescriptionHTML[3:len(issueData.DescriptionHTML)-4],
			)

			for _, assignee := range issueData.Assignees {
				slackUserID, exists := userMapping[assignee]
				if !exists {
					log.Printf("No Slack user ID found for Plane UUID: %s", assignee)
					continue
				}
				fmt.Print("Sending message to: ", slackUserID, "\n")
				channel, _, _, err := slackClient.OpenConversation(&slack.OpenConversationParameters{
					Users: []string{slackUserID},
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

func SendDailyOverview() {
	projectIds := FetchProjects()
	if projectIds == nil {
		log.Println("No projects found")
		return
	}

	issues := FetchIssues(projectIds)
	if issues == nil {
		log.Println("No issues found")
		return
	}

	categorizedIssues := CategorizeIssues(issues)

	var message string = fmt.Sprintf("*Here's the daily overview of issues for %s:*\n\n", time.Now().Format("2006-01-02"))
	for state, issues := range categorizedIssues {
		message += fmt.Sprintf("*%s*\n", state)
		for _, issue := range issues {
			message += fmt.Sprintf("• <%s|%s>\n", fmt.Sprintf("https://app.plane.so/%s/projects/%s/issues/%s", slug, issue.Project, issue.ID), issue.Name)
		}
		message += "\n"
	}

	channelID := os.Getenv("SLACK_OVERVIEW_CHANNEL_ID")
	_, _, err := slackClient.PostMessage(channelID, slack.MsgOptionText(message, false))
	if err != nil {
		log.Fatalf("Error sending message: %v", err)
	}
}
