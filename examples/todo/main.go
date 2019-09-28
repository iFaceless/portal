package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ifaceless/portal"
	"github.com/ifaceless/portal/examples/todo/model"
	"github.com/ifaceless/portal/examples/todo/schema"
)

func main() {
	t := model.TaskModel{
		ID:     4096,
		UserID: 1024,
		Title:  "Finish your jobs.",
	}

	chell := portal.New()
	// {"id":"4096","title":"Finish your jobs.","description":"Custom description","user":{"id":"1024","name":"user:1024"}}
	printFullFields(chell, t)
	// {"title":"Finish your jobs.","user":{"id":"1024","name":"user:1024"}}
	printOnlyFields(chell, t, "User", "Title")
	// {"title":"Finish your jobs."}
	printOnlyFields(chell, t, "Title")

	printMany(chell)
}

func printFullFields(chell *portal.Chell, t model.TaskModel) {
	var taskSchema schema.TaskSchema
	err := chell.Dump(context.Background(), &t, &taskSchema)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchema)
	fmt.Println(string(data))
}

func printOnlyFields(chell *portal.Chell, t model.TaskModel, only ...string) {
	var taskSchema schema.TaskSchema
	err := chell.Only(only...).Dump(context.Background(), &t, &taskSchema)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchema)
	fmt.Println(string(data))
}

func printMany(chell *portal.Chell) {
	var taskSchemas []schema.TaskSchema

	tasks := make([]*model.TaskModel, 0)
	for i := 0; i < 2; i++ {
		tasks = append(tasks, &model.TaskModel{
			ID:     i,
			UserID: i + 100,
			Title:  fmt.Sprintf("Task #%d", i+1),
		})
	}

	err := chell.Only("ID", "Title").DumpMany(context.Background(), tasks, &taskSchemas)
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchemas)
	fmt.Println(string(data))
}
