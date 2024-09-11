package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

func AddCSVEntry(entry []string) error {
	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile("user_mapping.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write the entry to the CSV file
	if err := writer.Write(entry); err != nil {
		return err
	}

	return nil
}

func AddUserMapping(input string) error {
	// Split the input string into two parts
	parts := strings.SplitN(input, " ", 2)
	if len(parts) != 2 {
		return fmt.Errorf("input must contain exactly two space-separated strings")
	}

	// Create the entry to be added
	entry := []string{parts[0], parts[1]}

	// Add the entry to the CSV file
	return AddCSVEntry(entry)
}

func LoadUserMapping(filePath string) (map[string]string, error) {
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
