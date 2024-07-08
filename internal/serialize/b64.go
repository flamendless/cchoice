package serialize

import "encoding/base64"

func toBase64(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func PNGEncode(data []byte) string {
	res := "data:image/png;base64,"
	res += toBase64(data)
	return res
}
