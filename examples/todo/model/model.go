package model

import (
	"fmt"
	"time"
)

type NotificationModel struct {
	ID      int
	Title   string
	Content string
}

type UserModel struct {
	ID  int
	Tag *string
}

func (u *UserModel) Fullname() string {
	return fmt.Sprintf("user:%d", u.ID)
}

func (u *UserModel) CreatedAt() time.Time {
	return time.Now()
}

func (u *UserModel) UpdatedAt() time.Time {
	return time.Now()
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
	tag := "user"
	return &UserModel{ID: t.UserID, Tag: &tag}
}
