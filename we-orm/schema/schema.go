package schema

import (
	"go/ast"
	"reflect"
	"weorm/dialect"
)

type Field struct {
	Name string
	Type string
	Tag  string
}

type Schema struct {
	//Model 映射的对象
	Model interface{}
	//Name 表名
	Name string
	//Fields 字段
	Fields []*Field
	//FieldNames 字段名
	FieldNames []string
	//fieldMap 记录字段名与字段的映射关系
	fieldMap map[string]*Field
}

func (s *Schema) GetField(name string) *Field {
	return s.fieldMap[name]
}

func Parse(dest interface{}, d dialect.Dialect) *Schema {
	//获取结构体类型
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(), //将结构体名作为表名
		fieldMap: make(map[string]*Field),
	}
	//分析结构体内的属性
	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		if !p.Anonymous && ast.IsExported(p.Name) {
			//若该字段非匿名字段且被导出(访问权限为public,go中使用首字母大小写区分是否导出),则将该字段存入表结构
			field := &Field{
				Name: p.Name,
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
			}
			//分析字段所带注解tag
			if v, ok := p.Tag.Lookup("weorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}
