package main

import (
	"encoding/json"
	"fmt"

	"github.com/ifaceless/portal"

	"github.com/ifaceless/portal/examples/todo/model"
	"github.com/ifaceless/portal/examples/todo/schema"
)

func main() {
	//fmt.Println(portal.ParseFilterString("[A,B[C,D],E[F,G],H]"))
	portal.SetDebug(true)
	task := model.TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	// {"id":"1","title":"Finish your jobs.","description":"Custom description","user":{"id":"1","name":"user:1"}}
	//printFullFields(&task)
	// {"title":"Finish your jobs.","user":{"id":"1","name":"user:1"}}
	//printWithOnlyFields(&task, "ID", "User[ID,Notifications[ID,Title],AnotherNotifications[ID]]", "SimpleUser")
	// {"title":"Finish your jobs."}
	//printWithOnlyFields(&task, "User", "SimpleUser")
	//
	//printMany()
	printWithExcludeFields(&task, "Description", "ID", "User[Name,Notifications,AnotherNotifications[ID]]")
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

func printWithOnlyFields(task *model.TaskModel, fields ...string) {
	var taskSchema schema.TaskSchema
	err := portal.Dump(&taskSchema, task, portal.Only(fields...))
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchema)
	fmt.Println(string(data))
}

func printWithExcludeFields(task *model.TaskModel, fields ...string) {
	var taskSchema schema.TaskSchema
	err := portal.Dump(&taskSchema, task, portal.Exclude(fields...))
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchema)
	fmt.Println(string(data))
}

func printMany() {
	var taskSchemas []schema.TaskSchema

	tasks := make([]*model.TaskModel, 0)
	for i := 0; i < 2; i++ {
		tasks = append(tasks, &model.TaskModel{
			ID:     i,
			UserID: i + 100,
			Title:  fmt.Sprintf("Task #%d", i+1),
		})
	}

	err := portal.Dump(&taskSchemas, &tasks, portal.Only("ID", "Title"))
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchemas)
	fmt.Println(string(data))
}
