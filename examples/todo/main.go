package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ifaceless/portal"

	"github.com/ifaceless/portal/examples/todo/model"
	"github.com/ifaceless/portal/examples/todo/schema"
)

func main() {
	start := time.Now()
	defer portal.CleanUp()

	portal.SetMaxPoolSize(1024)
	portal.SetDebug(true)

	task := model.TaskModel{
		ID:     1,
		UserID: 1,
		Title:  "Finish your jobs.",
	}

	printFullFields(&task)
	//printWithOnlyFields(&task, "Title", "SimpleUser")
	//printWithOnlyFields(&task, "ID", "User[ID,Notifications[ID],AnotherNotifications[Title]]", "SimpleUser")
	//printMany()
	//printWithExcludeFields(&task, "Description", "ID", "User[Name,Notifications[ID,Content],AnotherNotifications], SimpleUser")
	fmt.Printf("elapsed: %.1f ms\n", time.Since(start).Seconds()*1000)
}

func printFullFields(task *model.TaskModel) {
	var taskSchema *schema.TaskSchema
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := portal.DumpWithContext(ctx, &taskSchema, task)
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
	for i := 0; i < 10; i++ {
		tasks = append(tasks, &model.TaskModel{
			ID:     i,
			UserID: i + 100,
			Title:  fmt.Sprintf("Task #%d", i+1),
		})
	}

	err := portal.Dump(&taskSchemas, &tasks, portal.Only("ID", "Title", "User[Name]", "Description"))
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(taskSchemas)
	fmt.Println(string(data))
}
