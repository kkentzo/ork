package main

import (
	"fmt"
	"strings"
)

const DEFAULT_TASK_GROUP_SEP = "/"

type Inventory map[string]*Task

func (i Inventory) Populate(tasks []*Task) error {
	return i.populate(tasks, "")
}

func (i Inventory) populate(tasks []*Task, prefix string) error {
	for _, task := range tasks {
		if prefix != "" {
			task.Name = strings.Join([]string{prefix, task.Name}, DEFAULT_TASK_GROUP_SEP)
		}
		// add task
		if err := i.Add(task); err != nil {
			return err
		}
		// add nested tasks
		if err := i.populate(task.Tasks, task.Name); err != nil {
			return err
		}
	}
	return nil
}

func (i Inventory) Add(t *Task) error {
	if _, ok := i[t.Name]; ok {
		return fmt.Errorf("duplicate task: %s", t.Name)
	}
	i[t.Name] = t

	return nil
}

func (i Inventory) Find(label string) *Task {
	return i[label]
}

func (i Inventory) All() []*Task {
	tasks := make([]*Task, 0, len(i))
	for _, task := range i {
		tasks = append(tasks, task)
	}
	return tasks
}
