package zero

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Request request via method or smth
type Request struct {
	ID     int         `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	Write  chan []byte
}

type RespError struct {
	Code string `json:"code"`
	Desc string `json:"desc"`
}

// Resp will send some data socket
func (req *Request) Resp(data interface{}) {
	resp, _ := json.Marshal(struct {
		ID  int         `json:"request_id"`
		Res interface{} `json:"resp"`
	}{req.ID, data})
	req.Write <- resp
}

// Ok will send object {ok: true}
func (req *Request) Ok() {
	req.Resp(struct {
		Ok bool `json:"true"`
	}{true})
}

func (req *Request) WriteErr(code string, data interface{}) {
	resp, _ := json.Marshal(struct {
		ID    int       `json:"request_id"`
		Error RespError `json:"err"`
	}{req.ID, RespError{
		Code: code,
		Desc: fmt.Sprintf("%v", data),
	}})
	req.Write <- resp
}

func (req *Request) Err(code string, data interface{}) {
	panic(fmt.Sprintf("%s:%v", code, data))
}

func (req *Request) Param(name string) string {
	paramsRaw := req.Params.(map[string]interface{})
	if paramsRaw == nil {
		panic("param:" + name + " is undefined")
	}
	param, ok := paramsRaw[name]
	if ok {
		resp, okStr := param.(string)
		if okStr {
			if resp == "" {
				panic("param:" + name + " is empty")
			}
			return resp
		} else {
			panic("param:" + name + " is should be string")
		}

	} else {
		panic("param:" + name + " is undefined")
	}
}

// ParamFloat get param and check type for flost
func (req *Request) ParamFloat(name string) float64 {
	paramsRaw := req.Params.(map[string]interface{})
	if paramsRaw == nil {
		panic("param:" + name + " is undefined")
	}
	param, ok := paramsRaw[name]
	if ok {
		resp, okFloat := param.(float64)
		if okFloat {
			return resp
		}

		respStr, okStr := param.(string)
		if okStr {
			resp, err := strconv.ParseFloat(respStr, 64)
			if err == nil {
				return resp
			}
		}
		panic("param:" + name + " is not float")
	} else {
		panic("param:" + name + " is undefined")
	}
}

// ParamOpt fetches optional param
func (req *Request) ParamOpt(name string) string {
	paramsRaw := req.Params.(map[string]interface{})
	if paramsRaw != nil {
		param, ok := paramsRaw[name]
		if ok {
			resp, okStr := param.(string)
			if okStr {
				return resp
			}
		}
	}
	return ""
}

func (req *Request) ParamIntOpt(name string) int {
	paramsRaw := req.Params.(map[string]interface{})
	if paramsRaw != nil {
		param, ok := paramsRaw[name]
		if ok {
			resp, okInt := param.(int)
			if okInt {
				return resp
			}
			respStr, okStr := param.(string)
			if okStr {
				return I(respStr)
			}
		}
	}
	return 0
}

// ParamInt parse param return int
func (req *Request) ParamInt(name string) int {
	paramsRaw := req.Params.(map[string]interface{})
	if paramsRaw != nil {
		param, ok := paramsRaw[name]
		if ok {
			resp, okInt := param.(int)
			if okInt {
				return resp
			}
			respStr, okStr := param.(string)
			if okStr {
				return I(respStr)
			}
			panic("param:" + name + " should be int")
		} else {
			panic("param:" + name + " is undefined")
		}
	} else {
		panic("param:" + name + " is undefined")
	}
}
