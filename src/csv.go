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
	writer := csv.NewWriter(file)
	defer writer.Flush()
	if err := writer.Write(entry); err != nil {
		return err
	}

	return nil
}

func DeleteCSVEntry(planeID string) error {
	file, err := os.OpenFile("user_mapping.csv", os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("unable to read CSV file: %v", err)
	}

	var updatedRecords [][]string
	found := false
	for _, row := range records {
		if row[0] != planeID {
			updatedRecords = append(updatedRecords, row)
		} else {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("row with '%s' not found in the first column", planeID)
	}

	file, err = os.OpenFile("temp_user_mapping.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("unable to open file for writing: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	err = writer.WriteAll(updatedRecords)
	if err != nil {
		return fmt.Errorf("unable to write to file: %v", err)
	}
	writer.Flush()
	os.Rename("temp_user_mapping.csv", "user_mapping.csv")
	return nil
}

func ManageUserMapping(text string, slackID string) error {
	textSlice := strings.Split(text, " ")
	switch textSlice[0] {
	case "delete":
		if len(textSlice) == 2 {
			return DeleteCSVEntry(textSlice[1])
		} else {
			return fmt.Errorf("invalid argument count for delete")
		}
	case "add":
		if len(textSlice) == 2 {
			entry := []string{textSlice[1], slackID}
			return AddCSVEntry(entry)
		}
	default:
		return fmt.Errorf("%s is not supported", textSlice[0])
	}
	return fmt.Errorf("this should not have happened")
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
	for _, record := range records {
		planeUUID := record[0]
		slackUserID := record[1]
		userMapping[planeUUID] = slackUserID
	}

	return userMapping, nil
}
