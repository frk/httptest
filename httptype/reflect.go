package httptype

import (
	"reflect"
)

func newPtrFor(v any) any {
	rv := reflect.ValueOf(v)
	pv := reflect.New(rv.Type())
	pv.Elem().Set(rv)
	return pv.Interface()
}

func halfInit(v any) any {
	src := reflect.ValueOf(v)
	dst := reflect.New(src.Type()).Elem()
	if v := _halfInit(src); v.IsValid() {
		dst.Set(v)
	}
	return dst.Interface()
}

func _halfInit(src reflect.Value) (out reflect.Value) {
	// skip if zero or not valid
	if !src.IsValid() || src.IsZero() {
		return
	}

	switch src.Kind() {
	default:
		return // basic, channel, or func? nothing to do

	case reflect.Pointer:
		if v := _halfInit(src.Elem()); v.IsValid() {
			out = reflect.New(src.Type().Elem())
			out.Elem().Set(v)
		}
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			if v := _halfInit(src.Field(i)); v.IsValid() {
				if !out.IsValid() {
					out = reflect.New(src.Type()).Elem()
				}
				out.Field(i).Set(v)
			}
		}
	case reflect.Slice:
		for i := 0; i < src.Len(); i++ {
			if v := _halfInit(src.Index(i)); v.IsValid() {
				if !out.IsValid() {
					out = reflect.MakeSlice(src.Type(), src.Len(), src.Cap())
				}
				out.Index(i).Set(v)
			}
		}
	case reflect.Array:
		for i := 0; i < src.Len(); i++ {
			if v := _halfInit(src.Index(i)); v.IsValid() {
				if !out.IsValid() {
					out = reflect.New(src.Type()).Elem()
				}
				out.Index(i).Set(v)
			}
		}
	case reflect.Map:
		keys := src.MapKeys()
		for i := 0; i < len(keys); i++ {
			if v := _halfInit(src.MapIndex(keys[i])); v.IsValid() {
				if !out.IsValid() {
					out = reflect.MakeMap(src.Type())
				}
				out.SetMapIndex(keys[i], v)
			}
		}
	case reflect.Interface:
		out = reflect.New(src.Type()).Elem()
		v := _halfInit(src.Elem())
		if !v.IsValid() {
			v = reflect.New(src.Elem().Type()).Elem()
		}
		if v.Kind() == reflect.Ptr && v.IsNil() {
			v = reflect.New(v.Type().Elem())
		}
		out.Set(v)
	}

	return out
}
