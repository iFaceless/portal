package model

import "fmt"

type UserModel struct {
	ID int
}

func (u *UserModel) Fullname() string {
	return fmt.Sprintf("user:%d", u.ID)
}

type TaskModel struct {
	ID     int
	UserID int
	Title  string
}

func (t *TaskModel) User() *UserModel {
	return &UserModel{t.UserID}
}
