package session

import (
	"reflect"
	"weorm/log"
)

// Hooks const
const (
	BQuery  = "BeforeQuery"
	AQuery  = "AfterQuery"
	BUpdate = "BeforeUpdate"
	AUpdate = "AfterUpdate"
	BDelete = "BeforeDelete"
	ADelete = "AfterDelete"
	BInsert = "BeforeInsert"
	AInsert = "AfterInsert"
)

// Hooks interface
type (
	BeforeQuery interface {
		BeforeQuery(s *Session) error
	}
	AfterQuery interface {
		AfterQuery(s *Session) error
	}
	BeforeUpdate interface {
		BeforeUpdate(s *Session) error
	}
	AfterUpdate interface {
		AfterUpdate(s *Session) error
	}
	BeforeDelete interface {
		BeforeDelete(s *Session) error
	}
	AfterDelete interface {
		AfterDelete(s *Session) error
	}
	BeforeInsert interface {
		BeforeInsert(s *Session) error
	}
	AfterInsert interface {
		AfterInsert(s *Session) error
	}
)

//CallMethod 请求已注册的钩子
func (s *Session) CallMethod(method string, value interface{}) {
	//使用方法名实现
	//fm := reflect.ValueOf(s.RefTable().Model).MethodByName(method)
	//if value != nil {
	//	fm = reflect.ValueOf(value).MethodByName(method)
	//}
	//param := []reflect.Value{reflect.ValueOf(s)}
	//if fm.IsValid() {
	//	if v := fm.Call(param); len(v) > 0 {
	//		if err, ok := v[0].Interface().(error); ok {
	//			log.Error(err)
	//		}
	//	}
	//}

	//使用接口实现
	var dest interface{}
	//传入value则调用value对应对象的hooks,用于before,此时s.RefTable()还未赋值
	//未传入value则调用s.RefTable().Model的hooks,用于after,此时s.RefTable()已经有值
	if value != nil {
		dest = reflect.ValueOf(value).Interface()
	} else {
		dest = reflect.ValueOf(s.RefTable().Model).Interface()
	}
	//通过method调用对应hooks
	var err error
	switch method {
	case BQuery:
		if i, ok := dest.(BeforeQuery); ok {
			err = i.BeforeQuery(s)
		}
	case AQuery:
		if i, ok := dest.(AfterQuery); ok {
			err = i.AfterQuery(s)
		}
	case BDelete:
		if i, ok := dest.(BeforeDelete); ok {
			err = i.BeforeDelete(s)
		}
	case ADelete:
		if i, ok := dest.(AfterDelete); ok {
			err = i.AfterDelete(s)
		}
	case BInsert:
		if i, ok := dest.(BeforeInsert); ok {
			err = i.BeforeInsert(s)
		}
	case AInsert:
		if i, ok := dest.(AfterInsert); ok {
			err = i.AfterInsert(s)
		}
	case BUpdate:
		if i, ok := dest.(BeforeUpdate); ok {
			err = i.BeforeUpdate(s)
		}
	case AUpdate:
		if i, ok := dest.(AfterUpdate); ok {
			err = i.AfterUpdate(s)
		}
	default:
		log.Error("Hooks method error")
	}
	if err != nil {
		log.Error(err)
	}
}
