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
	printFullFields(chell, t)
	printOnlyFields(chell, t, "User", "Title")
}

func printFullFields(chell *portal.Chell, t model.TaskModel) {
	var taskSchema schema.TaskSchema
	err := chell.Dump(context.Background(), &t, &taskSchema)
	if err != nil {
		panic(err)
	}
	data, _ := json.MarshalIndent(taskSchema, "", "  ")
	fmt.Println(string(data))
}

func printOnlyFields(chell *portal.Chell, t model.TaskModel, only ...string) {
	var taskSchema schema.TaskSchema
	err := chell.Only(only...).Dump(context.Background(), &t, &taskSchema)
	if err != nil {
		panic(err)
	}
	data, _ := json.MarshalIndent(taskSchema, "", "  ")
	fmt.Println(string(data))
}
