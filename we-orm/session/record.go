package session

import (
	"reflect"
	"weorm/clause"
	"weorm/log"
	"weorm/schema"
)

//Insert 插入多个对象,一次插入只允许插入相同类型的对象
func (s *Session) Insert(values ...interface{}) (int64, error) {
	//获取表结构
	var table *schema.Schema = s.Model(values[0]).RefTable()
	s.clause.Set(clause.INSERT, table.Name, table.FieldNames)

	recordValues := make([]interface{}, 0)
	for _, value := range values {
		if table != s.Model(value).RefTable() {
			log.Error("It is not allowed to insert different objects at a table")
		}
		recordValues = append(recordValues, table.RecordValues(value, s.dialect))
	}

	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)

	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

//Find 传入一个切片指针,查询结构保存在切片中
func (s *Session) Find(values interface{}) error {
	//通过反射获取到values的实例
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	//获得实例的type
	destType := destSlice.Type().Elem()
	//解析实例对应的表结构
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	//遍历每一行记录，利用反射创建 destType 的实例 dest,将 dest 的所有字段平铺开,构造切片 values
	for rows.Next() {
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, name := range table.FieldNames {
			values = append(values, dest.FieldByName(s.dialect.GoName(name)).Addr().Interface())
		}
		//调用 rows.Scan() 将该行记录每一列的值依次赋值给 values 中的每一个字段
		if err := rows.Scan(values...); err != nil {
			return err
		}
		//将 dest 添加到切片 destSlice 中。循环直到所有的记录都添加到切片 destSlice 中
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}
