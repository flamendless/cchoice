package models

import (
	"cchoice/cmd/parse_map/enums"
	"fmt"
)

type Map struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Code     string      `json:"code"`
	Level    enums.Level `json:"level"`
	Contents []*Map      `json:"contents"`
	Parent   *Map        `json:"-"`
}

func (m *Map) Stringer() string {
	return fmt.Sprintf("%s %s %s %s", m.ID, m.Name, m.Code, m.Level.String())
}

func BinarySearchMap(m []*Map, name string, level enums.Level) *Map {
	low, high := 0, len(m)-1
	for low <= high {
		pivot := (low + high) / 2
		if m[pivot].Name == name {
			if m[pivot].Level == level {
				return m[pivot]
			}
			return nil
		}
		if m[pivot].Name < name {
			low = pivot + 1
		} else {
			high = pivot - 1
		}
	}
	return nil
}
