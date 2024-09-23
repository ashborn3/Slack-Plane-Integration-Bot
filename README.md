# Slack-Plane-Integration-Bot

This project is a Slack bot designed to integrate with the Plane API to fetch project and issue data, and send daily overviews and issue details to Slack channels and users.

## Table of Contents
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)

## Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/ashborn3/Slack-Plane-Integration-Bot.git
    cd Slack-Plane-Integration-Bot
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Create an .env file with the following key value pairs:  
    ```
    SLACK_TOKEN=<Your Slack Token>  
    SLACK_SOCK_TOKEN=<Your Slack Socket Token>  
    PLANE_TOKEN=<Your Plane Token>  
    SLUG=<Your Workspace Slug>  
    SLACK_OVERVIEW_CHANNEL_ID=<Channel ID to Send Daily Overview in>  
    ```


4. Build the project:
    ```sh
    make build
    ```

## Configuration

1. Create a `.env` file in the root directory and add the following environment variables:
    ```env
    SLACK_BOT_TOKEN=your-slack-bot-token
    SLACK_APP_TOKEN=your-slack-app-token
    SLACK_OVERVIEW_CHANNEL_ID=your-slack-channel-id-to-recieve-overview-in
    PLANE_API_KEY=your-plane-api-key
    PLANE_WORKSPACE_SLUG=your-plane-workspace-slug
    ```

2. Ensure `user_mapping.csv` is present in the root directory to map user emails to Slack IDs.

## Usage

1. Run the project:
    ```sh
    make run
    ```

2. The bot will start and connect to Slack using Socket Mode. It will listen for slash commands and send daily overviews of issues to the specified Slack channel.
