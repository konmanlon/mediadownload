package aria2

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Config struct {
	Api    string         `json:"api"`
	Dir    string         `json:"dir"`
	Params map[string]any `json:"params"`
}

type jsonRpc struct {
	Jsonrpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []any  `json:"params"`
	Id      int    `json:"id"`
}

func (j *jsonRpc) setId(id int) {
	j.Id = id
}

func (j *jsonRpc) addUri(uri []string) {
	j.Params = append(j.Params, uri)
}

func (j *jsonRpc) addParams(params map[string]any) {
	j.Params = append(j.Params, params)
}

func newJsonRpc() *jsonRpc {
	return &jsonRpc{
		Jsonrpc: "2.0",
		Method:  "aria2.addUri",
	}
}

type Options struct {
	Uri  []string
	Dir  string
	Id   int
	Conf *Config
}

func (o *Options) SetUri(uri []string) {
	o.Uri = uri
}

func (o *Options) SetDir(dir string) {
	o.Dir = dir
}
func (o *Options) SetId(id int) {
	o.Id = id
}

func NewOptions(conf *Config) *Options {
	return &Options{
		Conf: conf,
	}
}

type RpcResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   *RpcError   `json:"error"`
	Id      int         `json:"id"`
}

type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func Download(opts *Options) (rpcResp *RpcResponse, err error) {
	params := opts.Conf.Params

	if params == nil {
		params = make(map[string]any)
	}

	params["dir"] = opts.Conf.Dir + opts.Dir

	payload := newJsonRpc()
	payload.setId(opts.Id)
	payload.addUri(opts.Uri)
	payload.addParams(params)

	data, err := json.Marshal(payload)
	if err != nil {
		return
	}

	httpReq, err := http.NewRequest("POST", opts.Conf.Api, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	rpcResp = &RpcResponse{}
	err = json.NewDecoder(resp.Body).Decode(rpcResp)
	if err != nil {
		return
	}

	return
}
