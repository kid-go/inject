package inject

import (
	"fmt"
	"reflect"
	"unsafe"
)

/**
以容器为核心的依赖注入工具, 一个容器中可以写入任意个可注入对象.
调用注入时将会在这个容器及其父容器中查找 inject 标识需要注入的对象
*/

const tag = "inject"

type Injector interface {
	// Applicator 结构体对象注入
	Applicator
	// Invoker 函数调用注入
	Invoker
	// TypeMapper 类型映射
	TypeMapper
	// SetParent 给 Injector 设置一个父 Injector
	SetParent(Injector)
}

type Applicator interface {
	// Apply 将 inject 中存储的对象注入到传入的对象中
	Apply(interface{}) error
	ApplyAll() error
}

type Invoker interface {
	// Invoke 给函数注入入参
	Invoke(interface{}) ([]reflect.Value, error)
}

type TypeMapper interface {
	// Map 映射对象到可注入列表
	Map(interface{}) TypeMapper
	Maps(...interface{}) TypeMapper
	// MapTo 映射对象到可注入列表且指定其反射类型
	MapTo(interface{}, interface{}) TypeMapper
	// Set 映射一个对象到可注入列表且使用明确的反射类型表示
	Set(reflect.Type, reflect.Value) TypeMapper
	// Get 从可注入列表中获取一个指定类型的可注入对象
	Get(reflect.Type) reflect.Value
}

type injector struct {
	values map[reflect.Type]reflect.Value
	parent Injector
}

// InterfaceOf 获取一个任意类型对象的反射类型
func InterfaceOf(value interface{}) reflect.Type {
	t := reflect.TypeOf(value)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Interface {
		panic("Inject.InterfaceOf: with value not a pointer to an interface")
	}
	return t
}

func New() Injector {
	return &injector{
		values: make(map[reflect.Type]reflect.Value),
	}
}

func (inj *injector) Apply(val interface{}) error {
	return inj.apply(reflect.ValueOf(val))
}

func (inj *injector) ApplyAll() error {
	var err error
	for _, v := range inj.values {
		if err = inj.apply(v); err != nil {
			return err
		}
	}
	return nil
}

func (inj *injector) apply(v reflect.Value) error {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		structField := t.Field(i)
		if structField.Tag == "inject" ||
			structField.Tag == "inject:\"\"" ||
			structField.Tag.Get("inject") != "" {
			ft := f.Type()
			v := inj.Get(ft)
			if !f.CanSet() {
				nf := reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
				nf.Set(v)
				continue
			}
			if !v.IsValid() {
				return fmt.Errorf("value not found for type %v", ft)
			}
			f.Set(v)
		}
	}
	return nil
}

func (inj *injector) Invoke(f interface{}) ([]reflect.Value, error) {
	t := reflect.TypeOf(f)
	in := make([]reflect.Value, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		val := inj.Get(argType)
		if !val.IsValid() {
			return nil, fmt.Errorf("Value not found for type %v", argType)
		}
		in[i] = val
	}
	return reflect.ValueOf(f).Call(in), nil
}

func (inj *injector) Map(val interface{}) TypeMapper {
	inj.values[reflect.TypeOf(val)] = reflect.ValueOf(val)
	return inj
}

func (inj *injector) Maps(vals ...interface{}) TypeMapper {
	for _, val := range vals {
		inj.Map(val)
	}
	return inj
}

func (inj *injector) MapTo(val interface{}, ifacePtr interface{}) TypeMapper {
	inj.values[InterfaceOf(ifacePtr)] = reflect.ValueOf(val)
	return inj
}

func (inj *injector) Set(typ reflect.Type, val reflect.Value) TypeMapper {
	inj.values[typ] = val
	return inj
}

func (inj injector) Get(t reflect.Type) reflect.Value {
	val := inj.values[t]
	if val.IsValid() {
		return val
	}
	if t.Kind() == reflect.Interface {
		for k, v := range inj.values {
			if k.Implements(t) {
				val = v
				break
			}
		}
	}

	if !val.IsValid() && inj.parent != nil {
		val = inj.parent.Get(t)
	}
	return val
}

func (inj injector) SetParent(parent Injector) {
	inj.parent = parent
}
