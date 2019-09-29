package main

import (
	"encoding/json"
	"fmt"

	"github.com/ifaceless/portal"

	"github.com/ifaceless/portal/examples/todo/model"
	"github.com/ifaceless/portal/examples/todo/schema"
)

func main() {
	_ = model.TaskModel{
		ID:     4096,
		UserID: 1024,
		Title:  "Finish your jobs.",
	}

	// {"id":"4096","title":"Finish your jobs.","description":"Custom description","user":{"id":"1024","name":"user:1024"}}
	//printFullFields(&task)
	// {"title":"Finish your jobs.","user":{"id":"1024","name":"user:1024"}}
	//printOnlyFields(&task, "User", "Title")
	// {"title":"Finish your jobs."}
	//printOnlyFields(&task, "Title")

	printMany()
}

func printFullFields(task *model.TaskModel) {
	var taskSchema schema.TaskSchema
	err := portal.Dump(&taskSchema, task)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchema)
	fmt.Println(string(data))
}

func printOnlyFields(task *model.TaskModel, only ...string) {
	var taskSchema schema.TaskSchema
	err := portal.New().Only(only...).Dump(&taskSchema, task)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchema)
	fmt.Println(string(data))
}

func printMany() {
	var taskSchemas []schema.TaskSchema

	tasks := make([]*model.TaskModel, 0)
	for i := 0; i < 1; i++ {
		tasks = append(tasks, &model.TaskModel{
			ID:     i,
			UserID: i + 100,
			Title:  fmt.Sprintf("Task #%d", i+1),
		})
	}

	err := portal.New().Only("ID").Dump(&taskSchemas, tasks)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchemas)
	fmt.Println(string(data))
}
