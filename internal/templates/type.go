package templates

import "fmt"


type Column struct {
	Index int
	Required bool
}

type Template struct {
	Columns map[string]*Column
}

func CreateTemplate(kind TemplateKind) *Template {
	switch kind {
	case Undefined:
		panic("Can't use undefined template")
	case Sample:
		return &Template{
			Columns: SampleColumns,
		}
	}
	return nil
}

func (tpl *Template) Print() {
	for k, v := range tpl.Columns {
		fmt.Printf("%s = %d\n", k, v.Index)
	}
}
