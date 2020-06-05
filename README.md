[![Build Status](https://travis-ci.com/iFaceless/portal.svg?branch=master)](https://travis-ci.com/iFaceless/portal)
[![Coverage Status](https://coveralls.io/repos/github/iFaceless/portal/badge.svg?branch=master&branch=master)](https://coveralls.io/github/iFaceless/portal?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/iFaceless/portal)](https://goreportcard.com/report/github.com/iFaceless/portal)

[戳我看中文介绍](./README_CN.md)

# What's portal?
![portal game](https://s2.ax1x.com/2019/09/28/u1TnEt.jpg)

It's a lightweight package which simplifies Go object serialization. Inspired by [marshmallow](https://github.com/marshmallow-code/marshmallow), but with concurrency builtin for better performance.

[portal](https://github.com/iFaceless/portal/) can be used to serialize app-level objects to specified objects (schema structs). The serialized objects can be rendered to any standard formats like JSON for an HTTP API. Most importantly, if some fields of a schema have different data sources, portal could **spawn several goroutines to retrieve fields' data concurrently**.

*Note that unlike [marshmallow](https://github.com/marshmallow-code/marshmallow), [portal](https://github.com/iFaceless/portal/) only focuses on object serialization. So, if you want to validate struct fields, please refer to  [go-playground/validator](https://github.com/go-playground/validator) or [asaskevich/govalidator](https://github.com/asaskevich/govalidator).*

# Features

1. Can be configured to fill data into multiple fields concurrently.
1. Support flexible field filtering for any (nested) schemas.
1. Automatically convert field data to the expected field type, no more boilerplate code.
1. Simple and clean API.

# Install

```
get get -u github.com/ifaceless/portal
```

# Quickstart

Full example can be found [here](./examples/todo).

## Model Definitions

<details>
<summary>CLICK HERE | model.go</summary>

```go
type NotificationModel struct {
	ID      int
	Title   string
	Content string
}

type UserModel struct {
	ID int
}

func (u *UserModel) Fullname() string {
	return fmt.Sprintf("user:%d", u.ID)
}

func (u *UserModel) Notifications() (result []*NotificationModel) {
	for i := 0; i < 1; i++ {
		result = append(result, &NotificationModel{
			ID:      i,
			Title:   fmt.Sprintf("title_%d", i),
			Content: fmt.Sprintf("content_%d", i),
		})
	}
	return
}

type TaskModel struct {
	ID     int
	UserID int
	Title  string
}

func (t *TaskModel) User() *UserModel {
	return &UserModel{t.UserID}
}
```
    
</details>


## Schema Definitions

<details>
	<summary>CLICK HERE | schema.go</summary>
	
```go
type NotiSchema struct {
	ID      string `json:"id,omitempty"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type UserSchema struct {
	ID                   string        `json:"id,omitempty"`
	// Get user name from `UserModel.Fullname()`
	Name                 string        `json:"name,omitempty" portal:"attr:Fullname"`
	Notifications        []*NotiSchema `json:"notifications,omitempty" portal:"nested"`
	AnotherNotifications []*NotiSchema `json:"another_notifications,omitempty" portal:"nested;attr:Notifications"`
}

type TaskSchema struct {
	ID          string      `json:"id,omitempty"`
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty" portal:"meth:GetDescription"`
	// UserSchema is a nested schema
	User        *UserSchema `json:"user,omitempty" portal:"nested"`
	// We just want `Name` field for `SimpleUser`.
	// Besides, the data source is the same with `UserSchema`
	SimpleUser  *UserSchema `json:"simple_user,omitempty" portal:"nested;only:Name;attr:User"`
}

func (ts *TaskSchema) GetDescription(model *model.TaskModel) string {
	return "Custom description"
}
```

</details>


## Serialization Examples

```go
package main

import (
	"encoding/json"
	"github.com/ifaceless/portal"
)

func main() {
	// log debug info
	portal.SetDebug(true)
	// set max worker pool size
	portal.SetMaxPoolSize(1024)
	// make sure to clean up.
	defer portal.CleanUp()

	// write to a specified task schema
	var taskSchema schema.TaskSchema
	portal.Dump(&taskSchema, &taskModel)
	// data: {"id":"1","title":"Finish your jobs.","description":"Custom description","user":{"id":"1","name":"user:1","notifications":[{"id":"0","title":"title_0","content":"content_0"}],"another_notifications":[{"id":"0","title":"title_0","content":"content_0"}]},"simple_user":{"name":"user:1"}}
	data, _ := json.Marshal(taskSchema)

	// select specified fields
	portal.Dump(&taskSchema, &taskModel, portal.Only("Title", "SimpleUser"))
	// data: {"title":"Finish your jobs.","simple_user":{"name":"user:1"}}
	data, _ := json.Marshal(taskSchema)
	
	// select fields with alias defined in the json tag.
	// actually, the default alias tag is `json`, `portal.FieldAliasMapTagName("json")` is optional.
	portal.Dump(&taskSchema, &taskModel, portal.Only("title", "SimpleUser"), portal.FieldAliasMapTagName("json"))
	// data: {"title":"Finish your jobs.","simple_user":{"name":"user:1"}}
	data, _ := json.Marshal(taskSchema)

	// you can keep any fields for any nested schemas
	// multiple fields are separated with ','
	// nested fields are wrapped with '[' and ']'
	portal.Dump(&taskSchema, &taskModel, portal.Only("ID", "User[ID,Notifications[ID],AnotherNotifications[Title]]", "SimpleUser"))
	// data: {"id":"1","user":{"id":"1","notifications":[{"id":"0"}],"another_notifications":[{"title":"title_0"}]},"simple_user":{"name":"user:1"}}
	data, _ := json.Marshal(taskSchema)

	// ignore specified fields
	portal.Dump(&taskSchema, &taskModel, portal.Exclude("Description", "ID", "User[Name,Notifications[ID,Content],AnotherNotifications], SimpleUser"))
	// data: {"title":"Finish your jobs.","user":{"id":"1","notifications":[{"title":"title_0"}]}}
	data, _ := json.Marshal(taskSchema)

	// dump multiple tasks
	var taskSchemas []schema.TaskSchema
	portal.Dump(&taskSchemas, &taskModels, portal.Only("ID", "Title", "User[Name]"))
	// data: [{"id":"0","title":"Task #1","user":{"name":"user:100"}},{"id":"1","title":"Task #2","user":{"name":"user:101"}}]
	data, _ := json.Marshal(taskSchema)
}
```

To learn more about [portal](https://github.com/iFaceless/portal), please read the [User Guide](./USERGUIDE.md)~ 

# Concurrency Strategy

1. Any fields tagged with `portal:"async"` will be serialized asynchronously.
1. When dumping to multiple schemas, portal will do it concurrently if any fields in the schema are tagged with `protal:"async"`.
1. You can always disable concurrency strategy with option `portal.DisableConcurrency()`.

# Cache Strategy
1. Cache is implemented in the field level when `portal.SetCache(portal.DefaultCache)` is configured.
1. Cache will be disabled by tagging the fields with `portal:"disablecache"` or by defining a `PortalDisableCache() bool` meth for the schema struct, or by a `portal.DisableCache()` option setting while dumping.
1. Cache is available for one schema's one time dump, after the dump, the cache will be invalidated.


# Core APIs

```go
func New(opts ...Option) (*Chell, error)
func Dump(dst, src interface{}, opts ...Option) error 
func DumpWithContext(ctx context.Context, dst, src interface{}, opts ...Option)
func SetDebug(v bool)
func SetMaxPoolSize(size int)
func CleanUp()
```

# More Field Types
- `field.Timestamp`: `time.Time` <-> unix timestamp.
- `field.UpperString`: convert string to upper case.
- `field.LowerString`: convert string to lower case.

# License

[portal](https://github.com/iFaceless/portal) is licensed under the [MIT license](./LICENSE). Please feel free and have fun~
