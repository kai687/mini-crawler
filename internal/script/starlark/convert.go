package starlark

import (
	"fmt"
	"reflect"

	"github.com/algolia/docs-crawler/internal/script"
	starlarkgo "go.starlark.net/starlark"
)

func docValue(doc script.Document) starlarkgo.Value {
	return newDocumentValue(doc)
}

func contextValue(ctx script.Context) starlarkgo.Value {
	return fromGoValue(map[string]any{
		"url":      ctx.URL,
		"position": ctx.Position,
		"metadata": ctx.Metadata,
	})
}

func fromGoValue(value any) starlarkgo.Value {
	if converted, ok := fromGoScalar(value); ok {
		return converted
	}

	switch value := value.(type) {
	case []any:
		return fromGoList(value)
	case map[string]any:
		return fromGoMap(value)
	default:
		return starlarkgo.None
	}
}

func fromGoScalar(value any) (starlarkgo.Value, bool) {
	if converted, ok := fromGoBasicScalar(value); ok {
		return converted, true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return starlarkgo.MakeInt64(reflected.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return starlarkgo.MakeUint64(reflected.Uint()), true
	default:
		return nil, false
	}
}

func fromGoBasicScalar(value any) (starlarkgo.Value, bool) {
	switch value := value.(type) {
	case nil:
		return starlarkgo.None, true
	case bool:
		return starlarkgo.Bool(value), true
	case string:
		return starlarkgo.String(value), true
	case float32:
		return starlarkgo.Float(value), true
	case float64:
		return starlarkgo.Float(value), true
	default:
		return nil, false
	}
}

func fromGoList(value []any) starlarkgo.Value {
	items := make([]starlarkgo.Value, 0, len(value))
	for _, item := range value {
		items = append(items, fromGoValue(item))
	}

	return starlarkgo.NewList(items)
}

func fromGoMap(value map[string]any) starlarkgo.Value {
	dict := starlarkgo.NewDict(len(value))
	for key, item := range value {
		_ = dict.SetKey(starlarkgo.String(key), fromGoValue(item))
	}

	return dict
}

func toGoValue(value starlarkgo.Value) (any, error) {
	if converted, ok := toGoScalar(value); ok {
		return converted, nil
	}

	switch value := value.(type) {
	case *starlarkgo.List:
		return toGoList(value)
	case starlarkgo.Tuple:
		return toGoTuple(value)
	case *starlarkgo.Dict:
		return toGoMap(value)
	default:
		return nil, fmt.Errorf("unsupported starlark value %s", value.Type())
	}
}

func toGoScalar(value starlarkgo.Value) (any, bool) {
	switch value := value.(type) {
	case starlarkgo.NoneType:
		return nil, true
	case starlarkgo.Bool:
		return bool(value), true
	case starlarkgo.String:
		return value.GoString(), true
	case starlarkgo.Float:
		return float64(value), true
	case starlarkgo.Int:
		if intValue, ok := value.Int64(); ok {
			return intValue, true
		}
	}

	return nil, false
}

func toGoList(value *starlarkgo.List) ([]any, error) {
	items := make([]any, 0, value.Len())

	iter := value.Iterate()
	defer iter.Done()

	var item starlarkgo.Value
	for iter.Next(&item) {
		goItem, err := toGoValue(item)
		if err != nil {
			return nil, err
		}

		items = append(items, goItem)
	}

	return items, nil
}

func toGoTuple(value starlarkgo.Tuple) ([]any, error) {
	items := make([]any, 0, value.Len())
	for i := 0; i < value.Len(); i++ {
		goItem, err := toGoValue(value.Index(i))
		if err != nil {
			return nil, err
		}

		items = append(items, goItem)
	}

	return items, nil
}

func toGoMap(value *starlarkgo.Dict) (map[string]any, error) {
	goMap := make(map[string]any, value.Len())
	for _, key := range value.Keys() {
		stringKey, ok := starlarkgo.AsString(key)
		if !ok {
			return nil, fmt.Errorf("dict key %s is not string", key.Type())
		}

		goItem, err := toGoDictItem(value, key)
		if err != nil {
			return nil, err
		}

		goMap[stringKey] = goItem
	}

	return goMap, nil
}

func toGoDictItem(dict *starlarkgo.Dict, key starlarkgo.Value) (any, error) {
	item, _, err := dict.Get(key)
	if err != nil {
		return nil, err
	}

	return toGoValue(item)
}

func toStringAnyMap(value starlarkgo.Value) (map[string]any, error) {
	goValue, err := toGoValue(value)
	if err != nil {
		return nil, err
	}

	goMap, ok := goValue.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("want dict, got %T", goValue)
	}

	return goMap, nil
}

func toRecords(value starlarkgo.Value) ([]map[string]any, error) {
	goValue, err := toGoValue(value)
	if err != nil {
		return nil, err
	}

	items, ok := goValue.([]any)
	if !ok {
		return nil, fmt.Errorf("want list, got %T", goValue)
	}

	return recordsFromItems(items)
}

func recordsFromItems(items []any) ([]map[string]any, error) {
	records := make([]map[string]any, 0, len(items))
	for i, item := range items {
		record, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("item %d: want dict, got %T", i, item)
		}

		records = append(records, record)
	}

	return records, nil
}
