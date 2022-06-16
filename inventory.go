package main

import "fmt"

type Inventory map[string]*Task

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
