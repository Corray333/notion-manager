package mindmap

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
)

type Task struct {
	Title    string  `json:"Title"`
	Link     string  `json:"Link"`
	Hours    float64 `json:"Hours"`
	Subtasks []Task  `json:"Subtasks"`
}

var linkRegex = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`) // Regex для поиска ссылок

func ParseMarkdownTasks(data string) (string, []Task, error) {
	var tasks []Task
	var stack []*Task
	var projectName string

	scanner := bufio.NewScanner(strings.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		level, content := detectLevel(line)

		if level == 0 {
			projectName = content
			continue
		}

		title, link := extractLink(content)
		if num, err := strconv.ParseFloat(title, 64); err == nil {
			// Если строка - это число, добавляем в часы последней задачи
			if len(stack) > 0 {
				stack[len(stack)-1].Hours += num
			}
			continue
		}

		task := Task{Title: title, Link: link}

		// Найдем родительский уровень
		for len(stack) > 0 && len(stack) > level-1 {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			tasks = append(tasks, task)
			stack = append(stack, &tasks[len(tasks)-1])
		} else {
			parent := stack[len(stack)-1]
			parent.Subtasks = append(parent.Subtasks, task)
			stack = append(stack, &parent.Subtasks[len(parent.Subtasks)-1])
		}
	}

	if err := scanner.Err(); err != nil {
		return "", nil, err
	}

	// Подсчитываем сумму часов для каждой задачи
	for i := range tasks {
		SumTaskHours(&tasks[i])
	}

	if len(tasks) == 1 {
		return tasks[0].Title, tasks[0].Subtasks, nil
	} else {
		return projectName, tasks, nil
	}
}

func SumTaskHours(task *Task) float64 {
	sum := task.Hours
	for i := range task.Subtasks {
		sum += SumTaskHours(&task.Subtasks[i])
	}
	task.Hours = sum
	return sum
}

func detectLevel(line string) (int, string) {
	level := 0
	tabCount := 0

	for strings.HasPrefix(line, "    ") {
		tabCount++
		line = strings.TrimPrefix(line, "    ")
	}

	for strings.HasPrefix(line, "\t") {
		tabCount++
		line = strings.TrimPrefix(line, "\t")
	}

	for strings.HasPrefix(line, "#") {
		level++
		line = strings.TrimPrefix(line, "#")
	}

	if strings.HasPrefix(line, "-") {
		level = tabCount + 4
		line = strings.TrimPrefix(line, "-")
	}

	return level, strings.TrimSpace(line)
}

func extractLink(text string) (string, string) {
	match := linkRegex.FindStringSubmatch(text)
	if match != nil {
		return match[1], match[2]
	}
	return text, ""
}
