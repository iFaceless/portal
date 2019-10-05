package main

import (
	"encoding/json"
	"fmt"

	"github.com/ifaceless/portal"

	"github.com/ifaceless/portal/examples/todo/model"
	"github.com/ifaceless/portal/examples/todo/schema"
)

func main() {
	portal.SetDebug(false)
	task := model.TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	// {"id":"1","title":"Finish your jobs.","description":"Custom description","user":{"id":"1","name":"user:1","notifications":[{"id":"0","title":"title_0","content":"content_0"}],"another_notifications":[{"id":"0","title":"title_0","content":"content_0"}]},"simple_user":{"name":"user:1"}}
	printFullFields(&task)

	// {"title":"Finish your jobs.","simple_user":{"name":"user:1"}}
	printWithOnlyFields(&task, "Title", "SimpleUser")

	// {"id":"1","user":{"id":"1","notifications":[{"id":"0"}],"another_notifications":[{"title":"title_0"}]},"simple_user":{"name":"user:1"}}
	printWithOnlyFields(&task, "ID", "User[ID,Notifications[ID],AnotherNotifications[Title]]", "SimpleUser")

	// [{"id":"0","title":"Task #1","user":{"name":"user:100"}},{"id":"1","title":"Task #2","user":{"name":"user:101"}}]
	printMany()

	// {"title":"Finish your jobs.","user":{"id":"1","notifications":[{"title":"title_0"}]}}
	printWithExcludeFields(&task, "Description", "ID", "User[Name,Notifications[ID,Content],AnotherNotifications], SimpleUser")
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

	err := portal.Dump(&taskSchemas, &tasks, portal.Only("ID", "Title", "User[Name]"))
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchemas)
	fmt.Println(string(data))
}
