package schema

import "github.com/ifaceless/portal/examples/todo/model"

type UserSchema struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty" portal:"attr:Fullname"`
}

type TaskSchema struct {
	ID          string      `json:"id,omitempty"`
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty" portal:"meth:GetDescription"`
	User        *UserSchema `json:"user,omitempty" portal:"nested"`
}

func (ts *TaskSchema) GetDescription(model *model.TaskModel) string {
	return "Custom description"
}
