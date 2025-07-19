package encode

type IEncode interface {
	Encode(int64) string
	Decode(string) int64
}
