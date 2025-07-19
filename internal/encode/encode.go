package encode

type IEncode interface {
	Name() string
	Encode(int64) string
	Decode(string) int64
}
