package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"weorm"
)

func main() {
	engine, _ := weorm.NewEngine("mysql", "root:258963@tcp(localhost:3306)/we?charset=utf8")
	defer engine.Close()
	s := engine.NewSession()
	result, _ := s.Raw("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam").Exec()
	count, _ := result.RowsAffected()
	fmt.Printf("Exec success, %d affected\n", count)
}
