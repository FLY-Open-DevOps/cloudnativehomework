package bff

type Config struct {
	Env           string `json:"env,omitempty"`
	Port          int    `json:"port,omitempty"`
	CaculatorAddr string `json:"caculatoraddr,omitempty"`
}
