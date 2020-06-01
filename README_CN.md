[![Build Status](https://travis-ci.com/iFaceless/portal.svg?branch=master)](https://travis-ci.com/iFaceless/portal)
[![Coverage Status](https://coveralls.io/repos/github/iFaceless/portal/badge.svg?branch=master&branch=master)](https://coveralls.io/github/iFaceless/portal?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/iFaceless/portal)](https://goreportcard.com/report/github.com/iFaceless/portal)

# PORTAL 介绍
![portal game](https://s2.ax1x.com/2019/09/28/u1TnEt.jpg)

[portal](https://github.com/iFaceless/portal/) 是一个专注于 Go 语言中对象序列化的辅助框架，接口设计上深受 Python 社区中的 [marshmallow](https://github.com/marshmallow-code/marshmallow) 框架启发，但同时内建了并发支持，期望能够提高接口响应速度。

总体而言，它可以用来将应用层的数据模型对象（数据源可以是数据库、缓存、RPC 等）序列化成指定的 API Schema 结构体。然后用户可选择将序列化后的结构转换成 JSON 或者其它的格式供 HTTP API 返回。

*需要注意的是，[marshmallow](https://github.com/marshmallow-code/marshmallow) 框架除了提供对象序列化的功能外，还支持非常灵活的表单字段校验。但是 [portal](https://github.com/iFaceless/portal/) 只关注核心的序列化功能，对于结构体字段校验，可以使用 [go-playground/validator](https://github.com/go-playground/validator) 或者 [asaskevich/govalidator](https://github.com/asaskevich/govalidator) 框架。*

# 核心功能

1. 可选择异步填充 Schema 结构体字段值的填充；
1. 支持非常灵活的字段过滤功能；
1. 自动尝试完成来源数据类型到目标数据类型的转换，无需更多的样板代码；
1. 简洁易用的 API。

# 安装

```
get get -u github.com/ifaceless/portal
```

# 快速入门

完整的示例可参考 [这里](./examples/todo).

## Model 定义

<details>
<summary>点击此处展开 | model.go</summary>

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


## Schema 定义

<details>
	<summary>点击此处展开 | schema.go</summary>
	
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


## 序列化示例

```go
package main

import (
	"encoding/json"
	"github.com/ifaceless/portal"
)

func main() {
	// 设置日志等级为调试模式，方便查看和排错
	portal.SetDebug(true)
	// 设置全局 goroutine 池的最大值，避免启动过多 goroutine
	portal.SetMaxPoolSize(1024)
	// 最后要清理 goroutine 池，通知其退出
	defer portal.CleanUp()

	// 填充所有字段值
	var taskSchema schema.TaskSchema
	portal.Dump(&taskSchema, &taskModel)
	// data: {"id":"1","title":"Finish your jobs.","description":"Custom description","user":{"id":"1","name":"user:1","notifications":[{"id":"0","title":"title_0","content":"content_0"}],"another_notifications":[{"id":"0","title":"title_0","content":"content_0"}]},"simple_user":{"name":"user:1"}}
	data, _ := json.Marshal(taskSchema)

	// 选择填充部分字段值
	portal.Dump(&taskSchema, &taskModel, portal.Only("Title", "SimpleUser"))
	// data: {"title":"Finish your jobs.","simple_user":{"name":"user:1"}}
	data, _ := json.Marshal(taskSchema)
	
	// 可使用字段的 JSON 别名来选择，效果和 Schema Struct Name 一样
	// 当然，默认就是 JSON；也可以选择如 `yaml` 等。
	portal.Dump(&taskSchema, &taskModel, portal.Only("title", "SimpleUser"), portal.FieldAliasMapTagName("json"))
	// data: {"title":"Finish your jobs.","simple_user":{"name":"user:1"}}
	data, _ := json.Marshal(taskSchema)

	// 可以选择保留嵌套的 Schema 字段值
	// 多个字段请用逗号 `,` 分隔；嵌套字段包含在 `[]` 中
	portal.Dump(&taskSchema, &taskModel, portal.Only("ID", "User[ID,Notifications[ID],AnotherNotifications[Title]]", "SimpleUser"))
	// data: {"id":"1","user":{"id":"1","notifications":[{"id":"0"}],"another_notifications":[{"title":"title_0"}]},"simple_user":{"name":"user:1"}}
	data, _ := json.Marshal(taskSchema)

	// 忽略指定的字段（依然支持非常灵活的字段过滤规则）
	portal.Dump(&taskSchema, &taskModel, portal.Exclude("Description", "ID", "User[Name,Notifications[ID,Content],AnotherNotifications], SimpleUser"))
	// data: {"title":"Finish your jobs.","user":{"id":"1","notifications":[{"title":"title_0"}]}}
	data, _ := json.Marshal(taskSchema)

	// 填充多个 Schema，会在多个 goroutine 中并发完成
	var taskSchemas []schema.TaskSchema
	portal.Dump(&taskSchemas, &taskModels, portal.Only("ID", "Title", "User[Name]"))
	// data: [{"id":"0","title":"Task #1","user":{"name":"user:100"}},{"id":"1","title":"Task #2","user":{"name":"user:101"}}]
	data, _ := json.Marshal(taskSchema)
}

```

更多使用细节，请参考 [使用指南](./USERGUIDE.md)~ 

# 并发策略控制

1. 当某个 Schema 结构体字段标记了 `portal:"async"` 标签时会异步填充字段值；
1. 当序列化 Schema 列表时，会分析 Schema 中有无标记了 `async` 的字段，如果存在的话，则使用并发填充策略；否则只在当前 goroutine 中完成序列化；
1. 可以在 Dump 时添加 `portal.DisableConcurrency()` 禁用并发序列化的功能。

# 缓存策略控制
1. 当 `portal.SetCache(portal.DefaultCache)` 被设置之后，字段维度的缓存会被开启；
1. 以下情况下缓存会被禁用。Schema 字段中标记了 `portal:"diablecache"` 的 Tag； 被序列化的 Schema 定义了 `DisableCache() bool` 方法；序列化时设置了 `portal.DisableCache()` 选项；

# 核心 APIs

```go
func New(opts ...Option) (*Chell, error)
func Dump(dst, src interface{}, opts ...Option) error 
func DumpWithContext(ctx context.Context, dst, src interface{}, opts ...Option)
func SetDebug(v bool)
func SetMaxPoolSize(size int)
func CleanUp()
```
# License

[portal](https://github.com/iFaceless/portal) 采用 [MIT LICENSE](./LICENSE)。
