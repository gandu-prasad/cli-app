package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

type Task struct {
	ID          int32     `json:"id"`
	Description string    `json:"escription"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func writeJson(filename string, task Task) error {

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)

	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(task)
}

func readJson(filename string) ([]Task, error) {

	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var tasks []Task
	decoder := json.NewDecoder(file)

	for {
		var task Task
		if err := decoder.Decode(&task); err != nil {
			if err.Error() == "EOF" {
				break //end of file
			}
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil

}

func main() {

	jsonfile := "task.json"

	task := Task{ID: 1, Description: "it works1", Status: "status1", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	err := writeJson(jsonfile, task)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	fmt.Printf("file created : %s\n", jsonfile)

	tasks, err := readJson(jsonfile)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	for _, t := range tasks {
		fmt.Printf("Parsed :\n ID=%d,\n Description=%s,\n Status=%s,\n CreatedAt=%v,\n UpdatedAt=%v\n", t.ID, t.Description, t.Status, t.CreatedAt, t.UpdatedAt)

	}

}
