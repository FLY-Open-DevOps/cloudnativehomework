package caculator

type Config struct {
	Env      string `json:"env,omitempty"`
	Port     int    `json:"port,omitempty"`
	FiboAddr string `json:"fiboaddr,omitempty"`
}
