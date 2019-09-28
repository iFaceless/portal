# What's portal
![portal game](https://s2.ax1x.com/2019/09/28/u1TnEt.jpg)

It's a lightweight package which simplifies Go object serialization. Inspired heavily by [marshmallow](https://github.com/marshmallow-code/marshmallow), but with concurrency builtin for better performance.

[portal](https://github.com/iFaceless/portal/) can be used to:
- **Validate** input data.
- **Serialize** app-level objects to specified objects (schema structs). The serialized objects can be rendered to any standard formats like JSON for an HTTP API.

Most importantly, if some fileds of a schema have different data sources, portal could **spawn several goroutines to retrieve fields' data concurrently** if you prefer.

# Install

```
get get -u github.com/ifaceless/portal
```

# Quickstart

Full example can be found [here](./examples/todo).

## Model definitions

```go
type UserModel struct {
	ID int
}

func (u *UserModel) Fullname() string {
	// suppose we get user fullname from RPC
	return fmt.Sprintf("user:%d", u.ID)
}

type TaskModel struct {
	ID     int `gorm:"PRIMARY_KEY,AUTO_INCREMENT"`
	UserID int
	Title  string
}

// User returns a user object.
func (t *TaskModel) User() *UserModel {
	return &UserModel{t.UserID}
}
```

## Schema Definitions

```go
type UserSchema struct {
	ID   string `json:"id,omitempty"`
	// Get user name from `UserModel.Fullname()`
	Name string `json:"name,omitempty" portal:"attr:Fullname"`
}

type TaskSchema struct {
	ID          string      `json:"id,omitempty"`
	Title       string      `json:"title,omitempty"`
	// Get description from custom method `GetDescription()`
	Description string      `json:"description,omitempty" portal:"meth:GetDescription"`
	// UserSchema is nested to task schema
	User        *UserSchema `json:"user,omitempty" portal:"nested"`
}

func (ts *TaskSchema) GetDescription(model *model.TaskModel) string {
	return "Custom description"
}
```

## Serialize examples

```go
ctx := context.Background()
chell := portal.New()

// write to a specified task schema.
var taskSchema schema.TaskSchema
chell.Dump(ctx, &task, &taskSchema)
// {"id":"4096","title":"Finish your jobs.","description":"Custom description","user":{"id":"1024","name":"user:1024"}}
data, _ := json.Marshal(taskSchema)

// select specified fields
chell.Only("Title").Dump(ctx, &task, &taskSchema)
// {"title":"Finish your jobs."}
data, _ := json.Marshal(taskSchema)

// ignore specified fields
chell.Exclude("ID", "Description").Dump(ctx, &task, &taskSchema)
// {"title":"Finish your jobs.","user":{"id":"1024","name":"user:1024"}}
data, _ := json.Marshal(taskSchema)

// write to a slice of task schema
var taskSchemas []schema.TaskSchema
chell.Only("ID", "Title").DumpMany(context.Background(), tasks, &taskSchemas)
// [{"title":"Task #1"},{"id":"1","title":"Task #2"}]
data, _ := json.Marshal(taskSchema)
```

# License

[portal](https://github.com/iFaceless/portal) is licensed under the [MIT license](./LICENSE). Please feel free and have fun~
