package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"
)

type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status:pending"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type App struct {
	Tasks  map[int]*Task // key : task ID
	nextID int
	path   string
}

// creates an empty app bound to json file path
func NewApp(path string) *App {
	return &App{
		Tasks:  make(map[int]*Task),
		nextID: 1,
		path:   path,
	}
}

// load the json data into map for editing
func (a *App) Load() error {
	data, err := os.ReadFile(a.path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var tasks []Task

	if err := json.Unmarshal(data, &tasks); err != nil {
		return fmt.Errorf("corrupted %s : %w", a.path, err)
	}

	for i := range tasks {
		t := tasks[i]
		a.Tasks[t.ID] = &t
		if t.ID >= a.nextID {
			a.nextID = t.ID + 1
		}
	}
	return nil
}

// save the data
func (a *App) Save() error {
	ids := make([]int, 0, len(a.Tasks))
	for id := range a.Tasks {
		ids = append(ids, id)
	}

	sort.Ints(ids)

	tasks := make([]Task, 0, len(ids))

	for _, id := range ids {
		tasks = append(tasks, *a.Tasks[id])
	}

	data, err := json.MarshalIndent(tasks, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(a.path, data, 0o644)
}

// add task
func (a *App) AddTask(description string) (*Task, error) {
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}

	now := time.Now()

	t := &Task{
		ID:          a.nextID,
		Description: description,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	a.Tasks[t.ID] = t
	a.nextID++

	if err := a.Save(); err != nil {
		return nil, err
	}

	return t, nil
}

// edit task
func (a *App) EditTask(id int, description string) (*Task, error) {
	t, ok := a.Tasks[id]

	if !ok {
		return nil, fmt.Errorf("task %d not found", id)
	}

	if description == "" {
		return nil, errors.New("description cannot be empty")
	}

	t.Description = description
	t.UpdatedAt = time.Now()

	if err := a.Save(); err != nil {
		return nil, err
	}

	return t, nil
}

// status update
func (a *App) CompleteTask(id int) (*Task, error) {

	t, ok := a.Tasks[id]

	if !ok {
		return nil, fmt.Errorf("task %d not found", id)
	}

	t.Status = "done"
	t.UpdatedAt = time.Now()

	if err := a.Save(); err != nil {
		return nil, err
	}
	return t, nil
}

// deleteTask
func (a *App) deleteTask(id int) error {

	_, ok := a.Tasks[id]

	if !ok {
		return fmt.Errorf("task %d not found", id)
	}

	delete(a.Tasks, id)
	return a.Save()

}

// List the tasks
func (a *App) List() {

	ids := make([]int, 0, len(a.Tasks))
	for id := range a.Tasks {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	if len(ids) == 0 {
		fmt.Println("No tasks yet try: main add \"<any-task>\"")
		return
	}

	for _, id := range ids {
		t := a.Tasks[id]
		fmt.Printf("[%d] %-6s->%s\n", t.ID, t.Status, t.Description)
	}
}

// main
func main() {

	path := "task.json"
	app := NewApp(path)

	if err := app.Load(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		app.List()
		return
	}

	var err error

	switch os.Args[1] {
	case "add":
		if len(os.Args) < 3 {
			err = errors.New("usage: main add \"description\"")
			break
		}
		var t *Task
		t, err = app.AddTask(os.Args[2])
		if err == nil {
			fmt.Printf("Added task [%d] %s\n", t.ID, t.Description)
		}
	case "list":
		app.List()
	case "done":
		var t *Task
		t, err = parseAndRun(os.Args, app.CompleteTask)
		if err == nil {
			fmt.Printf("Done : [%d] %s\n", t.ID, t.Description)
		}
	case "edit":
		if len(os.Args) < 4 {
			err = errors.New("usage: main edit <id> \"new description\"")
			break
		}
		var id int
		id, err = strconv.Atoi(os.Args[2])
		if err != nil {
			break
		}
		var t *Task
		t, err := app.EditTask(id, os.Args[3])
		if err == nil {
			fmt.Printf("Edited: [%d] %s\n", t.ID, t.Description)
		}
	case "delete":
		var t *Task

		t, err = parseAndRun(os.Args, func(id int) (*Task, error) {
			if err := app.deleteTask(id); err != nil {
				return nil, err
			}
			return &Task{ID: id}, nil
		})
		if err == nil {
			fmt.Printf("Deleted task [%d]\n", t.ID)
		}
	default:
		err = fmt.Errorf("unknown command %q (try add, list, done, edit, delete)", os.Args[1])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// parses "<cmd> <id>" args and runs func(id)

func parseAndRun(args []string, fn func(int) (*Task, error)) (*Task, error) {
	if len(args) < 3 {
		return nil, errors.New("usage: main " + args[1] + "<id>")
	}
	id, err := strconv.Atoi(args[2])

	if err != nil {
		return nil, fmt.Errorf("invalid id %q: %w", args[2], err)
	}
	return fn(id)
}
