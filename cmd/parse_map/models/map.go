package models

import (
	"cchoice/cmd/parse_map/enums"
	"fmt"
	"sort"
)

type Map struct {
	Parent   *Map        `json:"-"`
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Code     string      `json:"code"`
	Contents []*Map      `json:"contents"`
	Level    enums.Level `json:"level"`
}

func (m *Map) Stringer() string {
	return fmt.Sprintf("%s %s %s %s", m.ID, m.Name, m.Code, m.Level.String())
}

func traverseMap(m *Map, result *[]*Map, level enums.Level) {
	if m.Level == level {
		*result = append(*result, m)
	}
	for _, child := range m.Contents {
		traverseMap(child, result, level)
	}
}

func GetMapsByLevel(m []*Map, level enums.Level) []*Map {
	result := make([]*Map, 0, len(m))
	for _, n := range m {
		traverseMap(n, &result, level)
	}
	return result
}

func binarySearchMapByName(m []*Map, name string, level enums.Level) *Map {
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

func BinarySearchMapByName(m []*Map, name string, level enums.Level) *Map {
	if len(m) == 0 {
		return nil
	}

	if found := binarySearchMapByName(m, name, level); found != nil {
		return found
	}

	for i := range m {
		if found := BinarySearchMapByName(m[i].Contents, name, level); found != nil {
			return found
		}
	}

	return nil
}

func SortMap(m []*Map) {
	sort.Slice(m, func(i, j int) bool {
		return m[i].Name < m[j].Name
	})
	for i := range m {
		SortMap(m[i].Contents)
	}
}
