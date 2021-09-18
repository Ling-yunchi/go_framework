package session

import (
	"database/sql"
	"strings"
	"weorm/dialect"
	"weorm/log"
	"weorm/schema"
)

type Session struct {
	db       *sql.DB
	dialect  dialect.Dialect //sql方言
	refTable *schema.Schema  //表结构
	sql      strings.Builder //拼接sql语句
	sqlVars  []interface{}   //sql语句占位符中的变量值
}

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
}

func (s *Session) DB() *sql.DB {
	return s.db
}

//Raw 改变sql语句与占位符对应的值
func (s *Session) Raw(sql string, values ...interface{}) *Session {
	s.sql.WriteString(sql)
	s.sql.WriteString(" ")
	s.sqlVars = append(s.sqlVars, values...)
	return s
}

//封装 Exec Query QueryRow
//	- 统一打印日志 (包括执行的 SQL语句 和 错误日志)
//	- 执行完成后,清空(s *Session).sql和(s *Session).sqlVars两个变量.这样Session可以复用,开启一次会话,可以执行多次SQL。

func (s *Session) Exec() (result sql.Result, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if result, err = s.DB().Exec(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}

func (s *Session) QueryRow() *sql.Row {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	return s.DB().QueryRow(s.sql.String(), s.sqlVars...)
}

func (s *Session) QueryRows() (rows *sql.Rows, err error) {
	defer s.Clear()
	log.Info(s.sql.String(), s.sqlVars)
	if rows, err = s.DB().Query(s.sql.String(), s.sqlVars...); err != nil {
		log.Error(err)
	}
	return
}
