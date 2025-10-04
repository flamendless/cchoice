package errs

import "errors"

var (
	ErrJSONMarshal   = errors.New("[JSON]: Marshaling failed")
	ErrJSONUnmarshal = errors.New("[JSON]: Unmarshaling failed")
	ErrJSONDecode    = errors.New("[JSON]: Decoding failed")
	ErrJSONEncode    = errors.New("[JSON]: Encoding failed")

	ErrHTTPNewRequest    = errors.New("[HTTP]: Failed to create request")
	ErrHTTPDoRequest     = errors.New("[HTTP]: Failed to execute request")
	ErrHTTPReadResponse  = errors.New("[HTTP]: Failed to read response body")
	ErrHTTPCloseResponse = errors.New("[HTTP]: Failed to close response body")

	ErrIORead  = errors.New("[IO]: Read operation failed")
	ErrIOWrite = errors.New("[IO]: Write operation failed")
	ErrIOClose = errors.New("[IO]: Close operation failed")
	ErrIOCopy  = errors.New("[IO]: Copy operation failed")

	ErrStrConv      = errors.New("[STRCONV]: String conversion failed")
	ErrFormatString = errors.New("[FORMAT]: String formatting failed")
	ErrParseInt     = errors.New("[PARSE]: Integer parsing failed")
	ErrParseFloat   = errors.New("[PARSE]: Float parsing failed")

	ErrBufferWrite = errors.New("[BUFFER]: Buffer write failed")
	ErrBufferRead  = errors.New("[BUFFER]: Buffer read failed")
	ErrBytesOp     = errors.New("[BYTES]: Bytes operation failed")

	ErrFileOpen   = errors.New("[FILE]: Failed to open file")
	ErrFileRead   = errors.New("[FILE]: Failed to read file")
	ErrFileWrite  = errors.New("[FILE]: Failed to write file")
	ErrFileClose  = errors.New("[FILE]: Failed to close file")
	ErrFileCreate = errors.New("[FILE]: Failed to create file")

	ErrURLParse  = errors.New("[URL]: URL parsing failed")
	ErrPathJoin  = errors.New("[PATH]: Path join operation failed")
	ErrPathClean = errors.New("[PATH]: Path clean operation failed")

	ErrTimeParse  = errors.New("[TIME]: Time parsing failed")
	ErrTimeFormat = errors.New("[TIME]: Time formatting failed")

	ErrValidation    = errors.New("[VALIDATION]: Validation failed")
	ErrInvalidInput  = errors.New("[VALIDATION]: Invalid input")
	ErrMissingField  = errors.New("[VALIDATION]: Required field missing")
	ErrInvalidFormat = errors.New("[VALIDATION]: Invalid format")
)
