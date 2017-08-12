package vm

import (
	"fmt"
	"reflect"
	"sync"
)

// Env 环境
type Env struct {
	// 包名
	//name string
	env    map[string]reflect.Value
	typ    map[string]reflect.Type
	parent *Env
	//interrupt *bool
	sync.RWMutex
}

// NewEnv 新的全局环境
func NewEnv() *Env {
	return &Env{
		env:    make(map[string]reflect.Value),
		typ:    make(map[string]reflect.Type),
		parent: nil,
	}
}

// NewEnv 新的局部环境
func (e *Env) NewEnv() *Env {
	return &Env{
		env:    make(map[string]reflect.Value),
		typ:    make(map[string]reflect.Type),
		parent: e,
	}
}

// Destroy 销毁
func (e *Env) Destroy() {
	e.Lock()
	defer e.Unlock()

	if e.parent == nil {
		return
	}
	for k, v := range e.parent.env {
		if v.IsValid() && v.Interface() == e {
			delete(e.parent.env, k)
		}
	}
	e.parent = nil
	e.env = nil
}

//// 包名
//func (e *Env) SetName(n string) {
//    e.Lock()
//    e.name = n
//    e.Unlock()
//}

// Type 返回类型
func (e *Env) Type(k string) (reflect.Type, error) {
	e.RLock()
	defer e.RUnlock()

	if v, ok := e.typ[k]; ok {
		return v, nil
	}
	if e.parent == nil {
		return NilType, fmt.Errorf("Undefined type '%s'", k)
	}
	return e.parent.Type(k)
}

// Get 取值
func (e *Env) Get(k string) (reflect.Value, error) {
	e.RLock()
	defer e.RUnlock()

	if v, ok := e.env[k]; ok {
		return v, nil
	}
	if e.parent == nil {
		return NilValue, fmt.Errorf("Undefined symbol '%s'", k)
	}
	return e.parent.Get(k)
}

// Set 设置值
func (e *Env) Set(k string, v interface{}) error {
	e.Lock()
	defer e.Unlock()

	if _, ok := e.env[k]; ok {
		val, ok := v.(reflect.Value)
		if !ok {
			val = reflect.ValueOf(v)
		}
		e.env[k] = val
		return nil
	}
	if e.parent == nil {
		return fmt.Errorf("Unknown symbol '%s'", k)

	}
	return e.parent.Set(k, v)
}

// Define 定义值
func (e *Env) Define(k string, v interface{}) error {
	//if strings.Contains(k, ".") {
	//    return fmt.Errorf("Unknown symbol '%s'", k)
	//}

	val, ok := v.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(v)
	}

	e.Lock()
	defer e.Unlock()

	e.env[k] = val

	return nil
}

// Dump 打印环境变量
func (e *Env) Dump() {
	e.RLock()
	defer e.RUnlock()

	for k, v := range e.env {
		fmt.Printf("%v = %#v\n", k, v)
	}
}
