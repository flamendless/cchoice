package utils

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/goccy/go-json"
)

func FormToStruct(r *http.Request, dst any) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	if reflect.TypeOf(dst).Kind() != reflect.Ptr {
		return errors.New("target must be a pointer")
	}
	elem := reflect.TypeOf(dst).Elem()
	if elem.Kind() != reflect.Struct {
		return errors.New("target must be a pointer to struct")
	}

	sliceFields := map[string]struct{}{}
	for i := 0; i < elem.NumField(); i++ {
		f := elem.Field(i)
		if f.Type.Kind() == reflect.Slice {
			if tag := f.Tag.Get("json"); tag != "" {
				sliceFields[tag] = struct{}{}
			}
		}
	}

	flat := make(map[string]any, len(r.Form))
	for k, v := range r.Form {
		if _, isSlice := sliceFields[k]; isSlice {
			sanitized := make([]string, len(v))
			for i, val := range v {
				sanitized[i] = SanitizeString(val)
			}
			flat[k] = sanitized
		} else {
			if len(v) > 0 {
				flat[k] = SanitizeString(v[0])
			} else {
				flat[k] = ""
			}
		}
	}

	b, err := json.Marshal(flat)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}
