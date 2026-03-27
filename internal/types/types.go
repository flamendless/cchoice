package types

type Location struct {
	Lat          float64 `json:"lat"`
	Lng          float64 `json:"lng"`
	RadiusMeters int     `json:"radius_meters"`
}

type UserAgentInfo struct {
	Browser        string
	BrowserVersion string
	OS             string
	Device         string
}

type ImageSize struct {
	Width  int
	Height int
}

var ImageSizes = []ImageSize{
	{640, 640},
	{1280, 1280},
}

type ThumbnailVariant struct {
	Size       string
	Path       string
	URL        string
	IsOriginal bool
}
