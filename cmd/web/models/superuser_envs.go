package models

type EnvVar struct {
	Key   string
	Value string
}

type EnvsData struct {
	Vars []EnvVar
}
