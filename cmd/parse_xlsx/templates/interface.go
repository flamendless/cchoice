package templates

type ITemplate interface {
	Print()
	AlignRow([]string) []string
	ValidateColumns() bool
}
