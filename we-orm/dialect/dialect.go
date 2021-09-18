package dialect

import "reflect"

var dialectMap = map[string]Dialect{}

//Dialect 用于兼容不同数据库的方言
type Dialect interface {
	//DataTypeOf 将 Go 语言的类型转换为该数据库的数据类型
	DataTypeOf(typ reflect.Value) string
	//TableExistSQL 返回某个表是否存在的 SQL 语句
	TableExistSQL(tableName string) (string, []interface{})
}

func RegisterDialect(name string, dialect Dialect) {
	dialectMap[name] = dialect
}

func GetDialect(name string) (dialect Dialect, ok bool) {
	dialect, ok = dialectMap[name]
	return
}
