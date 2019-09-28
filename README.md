# What's portal
It's a lightweight package which simplifies Go object serialization. Inspired heavily by [marshmallow](https://github.com/marshmallow-code/marshmallow), but with concurrency builtin for better performance.

portal can be used to:
- **Validate** input data.
- **Serialize** app-level objects to specified structs. The serialized objects can be rendered to any standard formats like JSON for a HTTP API.

# Install

```
get get -u github.com/ifaceless/portal
```

# Quickstart

```go
type SubscriptionSchema struct {
	ID          int64  `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

// SubscriptionModel defines database table
type SubscriptionModel struct {
	ID          int64 `gorm:"PRIMARY_KEY"`
	Title       string
	Description string
}
```

Dump to one:

```golang
chell := portal.New()

model := &SubscriptionModel{...}    // Suppose data is loaded
var dest SubscriptionSchema
chell.Dump(ctx, model, &dest)
```

Dump to many:

```golang
var dest []SubscriptionSchema

models := []SubscriptionModel{}{...}    // 数据库中加载了一列
chell.DumpMany(ctx, models, &dest)
```
