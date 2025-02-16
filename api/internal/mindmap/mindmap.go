package mindmap

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type Task struct {
	Title     string
	Link      string
	Hours     float64
	Subpoints []string
	Subtasks  []Task
}

var linkRegex = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`) // Regex для поиска ссылок

func ParseMarkdownTasks(data string) (string, []Task, error) {

	var tasks []Task
	var currentTask *Task
	var currentSubtask *Task
	var projectName string

	scanner := bufio.NewScanner(strings.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "# ") {
			projectName = line[2:]
		} else if strings.HasPrefix(line, "## ") {
			title, link := extractLink(line[3:])
			task := Task{Title: title, Link: link}
			tasks = append(tasks, task)
			currentTask = &tasks[len(tasks)-1]
			currentSubtask = nil
		} else if strings.HasPrefix(line, "### ") {
			title, link := extractLink(line[4:])
			if currentTask != nil {
				subtask := Task{Title: title, Link: link}
				currentTask.Subtasks = append(currentTask.Subtasks, subtask)
				currentSubtask = &currentTask.Subtasks[len(currentTask.Subtasks)-1]
			}
		} else if strings.HasPrefix(line, "- ") {
			if currentSubtask != nil {
				currentSubtask.Subpoints = append(currentSubtask.Subpoints, line[2:])
			} else if currentTask != nil {
				currentTask.Subpoints = append(currentTask.Subpoints, line[2:])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	return projectName, tasks, nil
}

func extractLink(text string) (string, string) {
	match := linkRegex.FindStringSubmatch(text)
	if match != nil {
		return match[1], match[2]
	}
	return text, ""
}

func PrintTasks(projectName string, tasks []Task, level int) {
	fmt.Println("Project:", projectName)
	prefix := strings.Repeat("  ", level)
	for _, task := range tasks {
		fmt.Println(prefix+"- "+task.Title, "(Link:", task.Link, ")")
		for _, subpoint := range task.Subpoints {
			fmt.Println(prefix+"  *", subpoint)
		}
		PrintTasks(projectName, task.Subtasks, level+1)
	}
}
