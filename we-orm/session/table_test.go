package session

import "testing"

type User struct {
	Name string `weorm:"PRIMARY KEY"`
	Age  int
}

func TestSession_CreateTable(t *testing.T) {
	s := NewSession().Model(&User{})
	_ = s.DropTable()
	_ = s.CreateTable()
	//TODO go字段名与mysql字段名转换
	if !s.HasTable() {
		t.Fatal("Failed to create table User")
	}
}
