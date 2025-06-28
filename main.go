package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Task represents a single todo item
type Task struct {
	ID          int       `json:"id"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TodoList manages a collection of tasks
type TodoList struct {
	Tasks    []Task `json:"tasks"`
	NextID   int    `json:"next_id"`
	filename string
}

// NewTodoList creates a new TodoList instance
func NewTodoList(filename string) *TodoList {
	return &TodoList{
		Tasks:    make([]Task, 0),
		NextID:   1,
		filename: filename,
	}
}

// LoadFromFile loads tasks from a JSON file
func (tl *TodoList) LoadFromFile() error {
	file, err := os.Open(tl.filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(tl)
}

// SaveToFile saves tasks to a JSON file
func (tl *TodoList) SaveToFile() error {
	file, err := os.Create(tl.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(tl)
}

// AddTask adds a new task to the list
func (tl *TodoList) AddTask(description string) {
	task := Task{
		ID:          tl.NextID,
		Description: description,
		Completed:   false,
		CreatedAt:   time.Now(),
	}
	tl.Tasks = append(tl.Tasks, task)
	tl.NextID++
}

// CompleteTask marks a task as completed
func (tl *TodoList) CompleteTask(id int) error {
	for i := range tl.Tasks {
		if tl.Tasks[i].ID == id {
			if tl.Tasks[i].Completed {
				return fmt.Errorf("task %d is already completed", id)
			}
			tl.Tasks[i].Completed = true
			now := time.Now()
			tl.Tasks[i].CompletedAt = &now
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

// DeleteTask removes a task from the list
func (tl *TodoList) DeleteTask(id int) error {
	for i, task := range tl.Tasks {
		if task.ID == id {
			tl.Tasks = append(tl.Tasks[:i], tl.Tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

// ListTasks displays all tasks
func (tl *TodoList) ListTasks(showCompleted bool) {
	if len(tl.Tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	fmt.Println("\n=== TODO LIST ===")
	for _, task := range tl.Tasks {
		if !showCompleted && task.Completed {
			continue
		}
		
		status := "[ ]"
		if task.Completed {
			status = "[✓]"
		}
		
		fmt.Printf("%s %d: %s\n", status, task.ID, task.Description)
		fmt.Printf("    Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04"))
		
		if task.Completed && task.CompletedAt != nil {
			fmt.Printf("    Completed: %s\n", task.CompletedAt.Format("2006-01-02 15:04"))
		}
		fmt.Println()
	}
}

// GetStats returns statistics about tasks
func (tl *TodoList) GetStats() (total, completed, pending int) {
	total = len(tl.Tasks)
	for _, task := range tl.Tasks {
		if task.Completed {
			completed++
		} else {
			pending++
		}
	}
	return
}

func main() {
	todoList := NewTodoList("todos.json")
	
	// Load existing tasks
	if err := todoList.LoadFromFile(); err != nil {
		fmt.Printf("Error loading tasks: %v\n", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		fmt.Println("\n=== TODO LIST MANAGER ===")
		total, completed, pending := todoList.GetStats()
		fmt.Printf("Stats: %d total, %d completed, %d pending\n", total, completed, pending)
		
		fmt.Println("\nCommands:")
		fmt.Println("1. add <description>    - Add a new task")
		fmt.Println("2. list                 - List pending tasks")
		fmt.Println("3. listall              - List all tasks")
		fmt.Println("4. complete <id>        - Mark task as completed")
		fmt.Println("5. delete <id>          - Delete a task")
		fmt.Println("6. stats                - Show statistics")
		fmt.Println("7. quit                 - Exit the application")
		
		fmt.Print("\nEnter command: ")
		
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		
		parts := strings.SplitN(input, " ", 2)
		command := strings.ToLower(parts[0])
		
		switch command {
		case "add":
			if len(parts) < 2 {
				fmt.Println("Please provide a task description.")
				continue
			}
			todoList.AddTask(parts[1])
			fmt.Printf("Task added successfully!\n")
			
		case "list":
			todoList.ListTasks(false)
			
		case "listall":
			todoList.ListTasks(true)
			
		case "complete":
			if len(parts) < 2 {
				fmt.Println("Please provide a task ID.")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid task ID. Please enter a number.")
				continue
			}
			if err := todoList.CompleteTask(id); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Task %d marked as completed!\n", id)
			}
			
		case "delete":
			if len(parts) < 2 {
				fmt.Println("Please provide a task ID.")
				continue
			}
			id, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println("Invalid task ID. Please enter a number.")
				continue
			}
			if err := todoList.DeleteTask(id); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Printf("Task %d deleted successfully!\n", id)
			}
			
		case "stats":
			total, completed, pending := todoList.GetStats()
			fmt.Printf("\n=== STATISTICS ===\n")
			fmt.Printf("Total tasks: %d\n", total)
			fmt.Printf("Completed: %d\n", completed)
			fmt.Printf("Pending: %d\n", pending)
			if total > 0 {
				completionRate := float64(completed) / float64(total) * 100
				fmt.Printf("Completion rate: %.1f%%\n", completionRate)
			}
			
		case "quit", "exit", "q":
			fmt.Println("Saving tasks...")
			if err := todoList.SaveToFile(); err != nil {
				fmt.Printf("Error saving tasks: %v\n", err)
			} else {
				fmt.Println("Tasks saved successfully!")
			}
			fmt.Println("Goodbye!")
			return
			
		default:
			fmt.Printf("Unknown command: %s\n", command)
		}
		
		// Auto-save after each operation (except quit)
		if err := todoList.SaveToFile(); err != nil {
			fmt.Printf("Warning: Could not save tasks: %v\n", err)
		}
	}
}