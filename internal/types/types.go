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

