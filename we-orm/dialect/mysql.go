package dialect

import (
	"bytes"
	"fmt"
	"reflect"
	"time"
)

type mysql struct{}

var _ Dialect = (*mysql)(nil)

func init() {
	RegisterDialect("mysql", &mysql{})
}

func (s *mysql) DataTypeOf(typ reflect.Value) string {
	switch typ.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.String:
		return "varchar(255)"
	case reflect.Int,
		reflect.Int32:
		return "integer"
	case reflect.Int8:
		return "tinyint"
	case reflect.Int16:
		return "smallint"
	case reflect.Int64:
		return "bigint"

	case reflect.Uint,
		reflect.Uint32:
		return "integer unsigned"
	case reflect.Uint8:
		return "tinyint unsigned"
	case reflect.Uint16:
		return "smallint unsigned"
	case reflect.Uint64:
		return "bigint unsigned"

	case reflect.Float32,
		reflect.Float64:
		return "double precision"

	case reflect.Struct:
		if _, ok := typ.Interface().(time.Time); ok {
			return "datetime"
		}
	}
	panic(fmt.Sprintf("invalid sql type %s (%s)", typ.Type().Name(), typ.Kind()))
}

//TableExistSQL for mysql 查询当前数据库下是否存在表
func (s *mysql) TableExistSQL(tableName string) (string, []interface{}) {
	args := []interface{}{tableName}
	return "SELECT t.TABLE_NAME FROM information_schema.`TABLES` t WHERE t.TABLE_NAME = ? AND t.TABLE_SCHEMA = (SELECT DATABASE())", args
}

func (s *mysql) DatabaseName(name string) string {
	if name == "" {
		return ""
	}
	str := make([]byte, 0, 32)
	i := 0
	if name[0] == '_' {
		str = append(str, 'X')
		i++
	}
	for ; i < len(name); i++ {
		c := name[i]
		if c == '_' &&
			i+1 < len(name) &&
			('A' <= name[i+1] && name[i+1] <= 'Z') {
			continue
		}
		if '0' <= c && c <= '9' {
			str = append(str, c)
			continue
		}

		if 'A' <= c && c <= 'Z' {
			c ^= ' '
		}
		str = append(str, c)

		for i+1 < len(name) && ('A' <= name[i+1] && name[i+1] <= 'Z') {
			i++
			str = append(str, '_')
			str = append(str, bytes.ToLower([]byte{name[i]})[0])
		}
	}
	return string(str)
}
