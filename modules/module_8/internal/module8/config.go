package module8

type Config struct {
	Env         string `json:"env,omitempty"`
	MaxSeq      int    `json:"maxseq,omitempty"`
	CacheResult bool   `json:"cacheresult,omitempty"`
	Port        int    `json:"port,omitempty"`
}
