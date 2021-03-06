package session

import (
	"database/sql"
	"strings"
	"weorm/clause"
	"weorm/dialect"
	"weorm/log"
	"weorm/schema"
)

type Session struct {
	db       *sql.DB
	dialect  dialect.Dialect //sql方言
	tx       *sql.Tx         //支持事务
	refTable *schema.Schema  //表结构
	clause   clause.Clause   //用于生成sql语句
	sql      strings.Builder //拼接sql语句
	sqlVars  []interface{}   //sql语句占位符中的变量值
}

//CommonDB 是DataBase的最小函数集合
type CommonDB interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
}

//检查sql.DB与sql.Tx是否实现CommonDB
var _ CommonDB = (*sql.DB)(nil)
var _ CommonDB = (*sql.Tx)(nil)

func New(db *sql.DB, dialect dialect.Dialect) *Session {
	return &Session{
		db:      db,
		dialect: dialect,
	}
}

func (s *Session) Clear() {
	s.sql.Reset()
	s.sqlVars = nil
	//清空sql构造器
	s.clause = clause.Clause{}
}

//DB 返回 *sql.DB,如果事务已经开始,则返回 *sql.Tx
func (s *Session) DB() CommonDB {
	if s.tx != nil {
		return s.tx
	}
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
