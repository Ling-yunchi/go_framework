package weorm

import (
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"testing"
	"weorm/session"
)

func OpenDB(t *testing.T) *Engine {
	t.Helper()
	engine, err := NewEngine("mysql", "root:258963@tcp(localhost:3306)/we?charset=utf8&multiStatements=true")
	if err != nil {
		t.Fatal("failed to connect", err)
	}
	return engine
}

type User struct {
	Name string `weorm:"PRIMARY KEY"`
	Age  int
}

func TestEngine_Transaction(t *testing.T) {
	t.Run("rollback", func(t *testing.T) {
		transactionRollback(t)
	})
	t.Run("commit", func(t *testing.T) {
		transactionCommit(t)
	})
}

func transactionRollback(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_ = s.Model(&User{}).CreateTable()
	//mysql在事务中包含ddl语句时会导致事务回滚失败
	//应该严格地将DDL和DML完全分开，不能混合在一起执行
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_, err = s.Insert(&User{"Tom", 18})
		return nil, errors.New("error")
	})
	if err == nil {
		t.Fatal("failed to rollback")
	}
}

func transactionCommit(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_ = s.Model(&User{}).DropTable()
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		_ = s.Model(&User{}).CreateTable()
		_, err = s.Insert(&User{"Tom", 18})
		return
	})
	u := &User{}
	_ = s.First(u)
	if err != nil || u.Name != "Tom" {
		t.Fatal("failed to commit")
	}
}

func TestEngine_Migrate(t *testing.T) {
	engine := OpenDB(t)
	defer engine.Close()
	s := engine.NewSession()
	_, _ = s.Raw("DROP TABLE IF EXISTS user;").Exec()
	_, _ = s.Raw("CREATE TABLE user(name varchar(255) PRIMARY KEY, XXX integer);").Exec()
	_, _ = s.Raw("INSERT INTO user(`name`) values (?), (?)", "Tom", "Sam").Exec()
	engine.Migrate(&User{})

	rows, _ := s.Raw("SELECT * FROM user").QueryRows()
	columns, _ := rows.Columns()
	if !reflect.DeepEqual(columns, []string{"name", "age"}) {
		t.Fatal("Failed to migrate table user, got columns", columns)
	}
}
