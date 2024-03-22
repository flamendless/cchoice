package templates

var SampleColumns map[string]*Column = map[string]*Column{
	"Product Name": {
		Index: -1,
		Required: true,
	},
	"Product Number": {
		Index: -1,
		Required: true,
	},
	"Description": {
		Index: -1,
		Required: true,
	},
	"Unit Price": {
		Index: -1,
		Required: true,
	},
	"Test": {
		Index: -1,
		Required: false,
	},
}
