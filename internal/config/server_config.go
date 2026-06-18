package config

type ServerConfiguration struct {
	Url    string `json:"url"`
	Weight int    `json:"weight"`
}
