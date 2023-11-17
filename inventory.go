package main

import (
	"fmt"
	"strings"
)

const DEFAULT_TASK_GROUP_SEP = "."

type Inventory map[string]*LabeledTask

func (i Inventory) Populate(tasks []*Task) error {
	return i.populate(tasks, "")
}

func (i Inventory) populate(tasks []*Task, prefix string) error {
	for _, task := range tasks {
		taskName := task.Name
		if prefix != "" {
			taskName = strings.Join([]string{prefix, taskName}, DEFAULT_TASK_GROUP_SEP)
		}
		// add task
		if err := i.Add(taskName, task); err != nil {
			return err
		}
		// add generated tasks
		if err := i.populate(task.DynamicTasks, taskName); err != nil {
			return err
		}

		// add nested tasks
		if len(task.DynamicTasks) > 0 {
			// add all nested tasks under each dynamic task
			for _, dtask := range task.DynamicTasks {
				pref := strings.Join([]string{taskName, dtask.Name}, DEFAULT_TASK_GROUP_SEP)
				if err := i.populate(task.Tasks, pref); err != nil {
					return err
				}
			}
		} else {
			// no dynamic tasks -- just add the nested tasks under the current task
			if err := i.populate(task.Tasks, taskName); err != nil {
				return err
			}
		}
	}
	return nil
}

func (i Inventory) Add(name string, t *Task) error {
	if _, ok := i[name]; ok {
		return fmt.Errorf("duplicate task: %s", name)
	}
	i[name] = &LabeledTask{label: name, Task: t}

	return nil
}

func (i Inventory) Find(label string) *LabeledTask {
	return i[label]
}

func (i Inventory) Tasks(sel TaskSelector) []*LabeledTask {
	tasks := make([]*LabeledTask, 0, len(i))
	for _, task := range i {
		if sel(task) {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func (i Inventory) Labels(sel TaskSelector) []string {
	labels := make([]string, 0, len(i))
	for label, task := range i {
		if sel(task) {
			labels = append(labels, label)
		}
	}
	return labels
}
