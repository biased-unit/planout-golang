package planout

import (
	"reflect"
)

type TypedMap struct {
	data map[string]interface{}
}

func NewTypedMap(data map[string]interface{}) *TypedMap {
	return &TypedMap{data: data}
}

func (t *TypedMap) get(key string) (interface{}, bool) {
	value, exists := t.data[key]
	return value, exists
}

func (t *TypedMap) getString(key string) (string, bool) {
	value, exists := t.get(key)

	if !exists {
		return "", false
	}

	if reflect.TypeOf(value).String() != "string" {
		return "", false
	}

	return value.(string), true

}

func (t *TypedMap) getBool(key string) (bool, bool) {
	value, exists := t.get(key)

	if !exists {
		return false, false
	}

	if reflect.TypeOf(value).String() != "bool" {
		return false, false
	}

	return value.(bool), true

}

func (t *TypedMap) getInt(key string) (int, bool) {
	value, exists := t.get(key)

	if !exists {
		return 0, false
	}

	if reflect.TypeOf(value).String() != "int" {
		return 0, false
	}

	return value.(int), true

}

func (t *TypedMap) getInt64(key string) (int64, bool) {
	value, exists := t.get(key)

	if !exists {
		return 0, false
	}

	if reflect.TypeOf(value).String() != "int64" {
		return 0, false
	}

	return value.(int64), true

}

func (t *TypedMap) getFloat32(key string) (float32, bool) {
	value, exists := t.get(key)

	if !exists {
		return 0.0, false
	}

	if reflect.TypeOf(value).String() != "float32" {
		return 0.0, false
	}

	return value.(float32), true

}

func (t *TypedMap) getFloat64(key string) (float64, bool) {
	value, exists := t.get(key)

	if !exists {
		return 0.0, false
	}

	if reflect.TypeOf(value).String() != "float64" {
		return 0.0, false
	}

	return value.(float64), true

}

func (t *TypedMap) getMap(key string) (map[string]interface{}, bool) {
	value, exists := t.get(key)

	if !exists {
		return nil, false
	}

	if reflect.TypeOf(value).String() != "map[string]interface {}" {
		return nil, false
	}

	return value.(map[string]interface{}), true

}

func (t *TypedMap) getArray(key string) ([]interface{}, bool) {
	value, exists := t.get(key)

	if !exists {
		return nil, false
	}

	if reflect.TypeOf(value).String() != "[]interface {}" {
		return nil, false
	}

	return value.([]interface{}), true

}
