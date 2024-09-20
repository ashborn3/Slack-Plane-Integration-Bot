package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

var (
	slug        string
	planeToken  string
	slackClient *slack.Client
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file!")
		return
	}
	slug = os.Getenv("SLUG")
	planeToken = os.Getenv("PLANE_TOKEN")

	slackClient = slack.New(os.Getenv("SLACK_TOKEN"), slack.OptionDebug(true), slack.OptionAppLevelToken(os.Getenv("SLACK_SOCK_TOKEN")))

	socketClient := socketmode.New(slackClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go StartSocketMode(ctx, slackClient, socketClient)

	// Set up the cron job
	c := cron.New(cron.WithLocation(time.FixedZone("IST", 5*60*60+30*60))) // IST is UTC+5:30
	_, err = c.AddFunc("0 9 * * *", NotifyUsersDaily)                      // Run at 9:00 AM IST every day
	if err != nil {
		log.Fatalf("Error setting up cron job: %v", err)
		return
	}
	_, err = c.AddFunc("0 9 * * *", SendDailyOverview) // Run at 9:00 AM IST every day
	if err != nil {
		log.Fatalf("Error setting up cron job: %v", err)
		return
	}
	c.Start()

	// Run the Socketmode client
	socketClient.Run()
}
