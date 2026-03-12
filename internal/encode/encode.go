package encode

const INVALID int64 = -1

type IEncode interface {
	Name() string
	Encode(int64) string
	Decode(string) int64
}
