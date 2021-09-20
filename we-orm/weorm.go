package weorm

import (
	"database/sql"
	"fmt"
	"strings"
	"weorm/dialect"
	"weorm/log"
	"weorm/session"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
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
	//确保sql方言存在
	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Error("Dialect %s Not Found", driver)
	}

	e = &Engine{
		db:      db,
		dialect: dial,
	}
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
	return session.New(e.db, e.dialect)
}

//TxFunc 将在 tx.Begin() 和 tx.Commit() 之间调用
type TxFunc func(*session.Session) (interface{}, error)

//Transaction 事务执行包装在事务中的sql，如果没有发生错误，则自动提交
//	txFunc中包含事务的所有操作
func (e *Engine) Transaction(txFunc TxFunc) (result interface{}, err error) {
	s := e.NewSession()
	if err = s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			log.Error(err)
			_ = s.Rollback() //// err is non-nil; don't change it
		} else {
			defer func() {
				if err != nil {
					_ = s.Rollback()
				}
			}()
			err = s.Commit() // err is nil; if Commit returns error update err
		}
	}()
	return txFunc(s)
}

// difference returns a - b
func difference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

func (e *Engine) Migrate(value interface{}) error {
	//开始事务
	_, err := e.Transaction(func(s *session.Session) (result interface{}, err error) {
		//判断表是否存在
		if !s.Model(value).HasTable() {
			log.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		//获取表结构
		table := s.RefTable()
		//查询一条记录获得列名
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns()
		_ = rows.Close() //mysql为了保证事务的顺序执行 连接池里面只有一个连接，在Query()操作之后，rows获得了这个链接，要断开这个链接，才可以继续Exec别的SQL语句
		//获取数据库字段名与对象字段名差异
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		log.Infof("added cols %v, deleted cols %v", addCols, delCols)

		//循环添加字段
		for _, col := range addCols {
			f := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		//删除字段时需要新建一个临时表将原表数据拷贝过去,将原表删除后再将临时表重命名
		if len(delCols) == 0 {
			return
		}
		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s from %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}
