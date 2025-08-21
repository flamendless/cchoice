package utils

import (
	"net/http"
	"reflect"

	"github.com/goccy/go-json"
)

func FormToStruct(r *http.Request, dst any) error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	t := reflect.TypeOf(dst).Elem()
	sliceFields := map[string]struct{}{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() == reflect.Slice {
			if tag := f.Tag.Get("json"); tag != "" {
				sliceFields[tag] = struct{}{}
			}
		}
	}

	flat := make(map[string]any, len(r.Form))
	for k, v := range r.Form {
		if _, isSlice := sliceFields[k]; isSlice {
			flat[k] = v
		} else if len(v) == 1 {
			flat[k] = v[0]
		} else {
			flat[k] = v
		}
	}

	b, err := json.Marshal(flat)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, dst)
}
