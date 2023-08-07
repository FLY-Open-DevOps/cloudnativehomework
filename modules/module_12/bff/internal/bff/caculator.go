package bff

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Caculator struct {
	baseUrl string
	client  *http.Client
}

type FiboReply struct {
	Result int `json:"result"`
}

func NewCaculator(baseUrl string) *Caculator {
	return &Caculator{
		baseUrl,
		http.DefaultClient,
	}
}

func (f *Caculator) Fibo(ctx context.Context, n int) (int, error) {
	path, err := url.JoinPath(f.baseUrl, "fibo")
	path += ("?n=" + strconv.Itoa(n))
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodGet, path, nil)
	req.Header = ctx.Value(headerKey).(http.Header)
	if err != nil {
		return 0, err
	}
	res, err := f.client.Do(req)
	if err != nil {
		return 0, err
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	r := &FiboReply{}
	if err := json.Unmarshal(data, r); err != nil {
		return 0, err
	}
	return r.Result, nil
}
