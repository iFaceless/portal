package schema

import (
	"time"

	"github.com/ifaceless/portal/field"

	"github.com/ifaceless/portal/_examples/todo/model"
)

type NotiSchema struct {
	Type    string `json:"type" portal:"const:vip"`
	ID      string `json:"id,omitempty"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type UserSchema struct {
	ID                   string           `json:"id,omitempty"`
	Tag                  *string          `json:"tag,omitempty"`
	Name                 string           `json:"name,omitempty" portal:"attr:Fullname"`
	Notifications        []*NotiSchema    `json:"notifications,omitempty" portal:"nested"`
	AnotherNotifications []*NotiSchema    `json:"another_notifications,omitempty" portal:"nested;attr:Notifications"`
	CreatedAt            *field.Timestamp `json:"created_at,omitempty"`
	UpdatedAt            *field.Timestamp `json:"updated_at,omitempty"`
}

type TaskSchema struct {
	ID           string      `json:"id,omitempty"`
	Title        string      `json:"title,omitempty"`
	Description  string      `json:"description,omitempty" portal:"meth:GetDescription;async"`
	Description1 string      `json:"description1,omitempty" portal:"meth:GetDescription;async"`
	Description2 string      `json:"description2,omitempty" portal:"meth:GetDescription;async"`
	Description3 string      `json:"description3,omitempty" portal:"meth:GetDescription;async"`
	User         *UserSchema `json:"user,omitempty" portal:"nested"`
	SimpleUser   *UserSchema `json:"simple_user,omitempty" portal:"nested;only:Name;attr:User"`
	Unknown      string      `json:"unknown"`
}

func (ts *TaskSchema) GetDescription(model *model.TaskModel) (string, error) {
	time.Sleep(1 * time.Second)
	return "Custom description", nil
}
