# What's portal
![portal game](https://s2.ax1x.com/2019/09/28/u1TnEt.jpg)

It's a lightweight package which simplifies Go object serialization. Inspired heavily by [marshmallow](https://github.com/marshmallow-code/marshmallow), but with concurrency builtin for better performance.

[portal](https://github.com/iFaceless/portal/) can be used to:
- **Validate** input data.
- **Serialize** app-level objects to specified objects (schema structs). The serialized objects can be rendered to any standard formats like JSON for an HTTP API.

Most importantly, if some fileds of a schema have different data sources (which means multiple network connections maybe), portal could **spawn several goroutines to retrieve fields' data concurrently** if you prefer.

# Install

```
get get -u github.com/iFaceless/portal
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

// marshal to JSON
data, _ := json.Marshal(&dest)
```

Dump to many:

```golang
var dest []SubscriptionSchema

models := []SubscriptionModel{}{...}
chell.DumpMany(ctx, models, &dest)

// marshal to JSON
data, _ := json.Marshal(&dest)
```

# License

[portal](https://github.com/iFaceless/portal) is licensed under the [MIT license](./LICENSE). Please feel free and have fun~
