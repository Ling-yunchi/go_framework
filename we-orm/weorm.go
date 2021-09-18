package weorm

import (
	"database/sql"
	"weorm/log"
	"weorm/session"
)

type Engine struct {
	db *sql.DB
}

func NewEngine(driver, source string) (e *Engine, err error) {
	var db *sql.DB
	db, err = sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	//发送ping请求保证数据库连接存活
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}
	e = &Engine{db: db}
	log.Info("Connect database success")
	return
}

func (e *Engine) Close() {
	if err := e.db.Close(); err != nil {
		log.Error("Fail to close database", err)
	}
	log.Info("Close database success")
}

func (e *Engine) NewSession() *session.Session {
	return session.New(e.db)
}
