PORTAL USER GUIDE
======================

Let [portal](https://github.com/iFaceless/portal) worry about trivial details, say goodbye to boilerplate code (our final goal)!

## Options
### Specify fields to keep: `Only()`

```go
// keep field A only
c := New(Only("A")) 

// keep field B and C of the nested struct A
c := New("A[B,C]")
```

### Specify fields to exclude: `Exclude()`

```go
// exclude field A
c := New(Exclude("A")) 

// exclude field B and C of the nested struct A, but other fields of struct A are still selected.
c := New(Exclude("A[B,C]"))
```

### Set custom tag for each field in runtime: `CustomFieldTagMap()``.

It will override the default tag settings defined in your struct.

See example [here](https://github.com/iFaceless/portal/blob/65aaa0b537fd13607bd4d45c1016c1689dc53beb/_examples/todo/main.go#L36). 

## Special Tags
### Load Data from Model's Attribute: `attr`
```go
// Model definition
type UserModel struct {
	ID int
}

func (u *UserModel) Fullname() string {
	return fmt.Sprintf("user:%d", u.ID)
}

// Fullname2 'attribute' method can accept an ctx param.
func (u *UserModel) Fullname2(ctx context.Context) string {
	return fmt.Sprintf("user:%d", u.ID)
}

// Fullname3 'attribute' can return error too, portal will ignore the 
// result if error returned.
func (u *UserModel) Fullname3(ctx context.Context) (string, error) {
	return fmt.Sprintf("user:%d", u.ID)
}

type BadgeModel {
	Name string
}

func (u *UserModel) Badge(ctx context.Context) (*BadgeModel, error) {
	return &BadgeModel{
		Name: "Cool"
	}, nil
}

// Schema definition
type UserSchema struct {
	ID                   string        `json:"id,omitempty"`
	Name                 string        `json:"name,omitempty" portal:"attr:Fullname"`
	// Chaining accessing is also supported.
	// portal calls method `UserModel.Badge()`, then accesses Badge.Name field.
	BadgeName            string        `json:"badge_name,omitempty" portal:"attr:Badge.Name"`
}
```

### Load Data from Custom Method: `meth`
```go
type TaskSchema struct {
	Title       string      `json:"title,omitempty" portal:"meth:GetTitle"`
	Description string      `json:"description,omitempty" portal:"meth:GetDescription"`
	// Chaining accessing is also supported for method result.
	ScheduleAt          *field.Timestamp   `json:"schedule_at,omitempty" portal:"meth:FetchSchedule.At"`
	ScheduleDescription *string            `json:"schedule_description,omitempty" portal:"meth:FetchSchedule.Description"`
}

func (ts *TaskSchema) GetTitle(ctx context.Context, model *model.TaskModel) string {
	// Accept extra context param.
	// TODO: Read info from the `ctx` here.
	return "Task Title"
}

func (ts *TaskSchema) GetDescription(model *model.TaskModel) (string, error) {
	// Here we ignore the first context param.
	// If method returns an error, portal will ignore the result.
	return "Custom description", nil
}

type Schedule struct {
	At          time.Time
	Description string
}

func (ts *TaskSchema) FetchSchedule(model *model.TaskModel) *Schedule {
	return &Schedule{
		Description: "High priority",
		At:          time.Now(),
	}
}
```

### Load Data Asynchronously: `async`
```go
type TaskSchema struct {
	Title       string      `json:"title,omitempty" portal:"meth:GetTitle;async"`
	Description string      `json:"description,omitempty" portal:"meth:GetDescription;async"`
}
```

### Nested Schema: `nested`
```go
type UserSchema struct {
	ID          string        `json:"id,omitempty"`
}

type TaskSchema struct {
	User        *UserSchema `json:"user,omitempty" portal:"nested"``
}
```

### Field Filtering: `only` & `exclude`

```go
type NotiSchema struct {
	ID      string `json:"id,omitempty"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type UserSchema struct {
	Notifications        []*NotiSchema `json:"notifications,omitempty" portal:"nested;only:id,title"`
	AnotherNotifications []*NotiSchema `json:"another_notifications,omitempty" portal:"nested;attr:Notifications;exclude:content"`
}
```

### Set Const Value for Field: `const`
```go
type UserSchema struct {
	Type    string `json:"type" portal:"const:vip"`
}

```

### Disable Cache for Field: `disablecache`
```go
type Student struct {
    ID int
}

type info struct {
    Name   string
    Height int
}

func(s *Student) Info() info {
    return &info{Name: "name", Height: 180}
}

type StudentSchema struct {
    Name   string `json:"name" portal:"attr:Info.Name,disablecache"`
    Height int    `json:"height" portal:"attr:Info.Height,disablecache"`
}
```

### Set Default Value for Field: `default`

Only works for types: pointer/slice/map. For basic types (integer, string, bool), default value will be converted and set to field directly. For complex types (eg. map/slice/pointer to custom struct), set default to `AUTO_INIT`, portal will initialize field to its zero value. 

```go
type ContentSchema struct {
	BizID   *string        `json:"biz_id" portal:"default:100"`
	SkuID   *string        `json:"sku_id"`                             // -> json null
	Users   []*UserSchema  `json:"users" portal:"default:AUTO_INIT"`   // -> json []
	Members map[string]int `json:"members" portal:"default:AUTO_INIT"` // -> json {}
	User    *UserSchema    `json:"user" portal:"default:AUTO_INIT"`
}
```

## Embedding Schema
```go
type PersonSchema struct {
	ID  string `json:"id"`
	Age int    `json:"age"`
}

type UserSchema2 struct {
	PersonSchema // embedded schema
	Token string `json:"token"`
}
```

## Custom Field Type

Custom field type must implements the `Valuer` and `ValueSetter` interface defined in [types.go](./types.go).

```go
type Timestamp struct {
	tm time.Time
}

func (t *Timestamp) SetValue(v interface{}) error {
	switch timeValue := v.(type) {
	case time.Time:
		t.tm = timeValue
	case *time.Time:
		t.tm = *timeValue
	default:
		return fmt.Errorf("expect type `time.Time`, not `%T`", v)
	}
	return nil
}

func (t *Timestamp) Value() (interface{}, error) {
	return t.tm, nil
}

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.tm.Unix())
}

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var i int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	t.tm = time.Unix(i, 0)
	return nil
}
```
