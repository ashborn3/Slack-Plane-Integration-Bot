package main

import "log"

func NotifyUsersDaily() {
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

	SendIssueDetailsToAssignees(issues)
}
