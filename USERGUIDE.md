PORTAL USER GUIDE
======================

Let [portal](https://github.com/iFaceless/portal) worry about trivial details, say goodbye to boilerplate code (our final goal)!

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

// Schema definition
type UserSchema struct {
	ID                   string        `json:"id,omitempty"`
	Name                 string        `json:"name,omitempty" portal:"attr:Fullname"`
}
```

### Load Data from Custom Method: `meth`
```go
type TaskSchema struct {
	Title       string      `json:"title,omitempty" portal:"meth:GetTitle"`
	Description string      `json:"description,omitempty" portal:"meth:GetDescription"`
}

func (ts *TaskSchema) GetTitle(ctx context.Context, model *model.TaskModel) string {
	// Accept extra context param.
	// TODO: Read info from the `ctx` here.
	return "Task Title"
}

func (ts *TaskSchema) GetDescription(model *model.TaskModel) string {
	// Here we ignore the first context param.
	return "Custom description"
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
