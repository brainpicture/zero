package zero

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"

	"github.com/valyala/fasthttp"
)

var httpHandlers map[string]func(srv *Server)
var methodHandlers map[string]func()

// Server is an wrapper around fasthttp
type Server struct {
	Ctx  *fasthttp.RequestCtx
	Path string
}

// Resp writes any data as JSON to HTTP stream
func (srv *Server) Resp(data interface{}) {
	srv.Ctx.SetContentType("application/json; charset=utf8")

	encoder := json.NewEncoder(srv.Ctx.Response.BodyWriter())
	encoder.SetEscapeHTML(false)

	dataWraped := struct {
		Resp interface{} `json:"resp"`
	}{data}
	err := encoder.Encode(dataWraped)
	if err != nil {
		srv.Err("system", err)
	}
}

// RespJSONP writes any data as JSONP to HTTP stream
func (srv *Server) RespJSONP(data interface{}) {
	cbName := srv.GetParam("jsoncallback")
	jsonData, err := json.Marshal(data)
	if err != nil {
		srv.Err("system", err)
	}
	srv.Ctx.Write([]byte(cbName + "(" + string(jsonData) + ")"))
	srv.Ctx.SetContentType("application/javascript; charset=utf8")
}

// HTML responds with HTML
func (srv *Server) HTML(data []byte) {
	srv.Ctx.Write(data)
	srv.Ctx.SetContentType("text/html; charset=utf8")
}

// JS responds with JS
func (srv *Server) JS(data []byte) {
	srv.Ctx.Write(data)
	srv.Ctx.SetContentType("application/javascript; charset=utf8")
}

// FileBlob responds with FileBlob
func (srv *Server) FileBlob(data []byte, contentType string) {
	srv.Ctx.Write(data)
	srv.Ctx.SetContentType(contentType)
}

// File responds with File
func (srv *Server) File(path string) {
	srv.Ctx.SendFile(path)
}

// GetParam request param as string
func (srv *Server) GetParam(key string) string {
	args := srv.Ctx.QueryArgs()
	param := string(args.Peek(key))

	if param == "" {
		param = string(srv.Ctx.PostArgs().Peek(key))
	}

	return param
}

// GetParamInt request param converted to int
func (srv *Server) GetParamInt(key string) int {
	args := srv.Ctx.QueryArgs()
	param, _ := args.GetUint(key)

	return int(param)
}

// GetParamInt64 request param converted to int64
func (srv *Server) GetParamInt64(key string) int64 {
	args := srv.Ctx.QueryArgs()
	param := args.Peek(key)
	i, err := strconv.ParseInt(string(param), 10, 64)
	if err != nil {
		i = 0
	}
	return i
}

// Err http api error
func (srv *Server) Err(code string, text interface{}) {
	srv.Ctx.SetStatusCode(400)
	srv.Ctx.SetContentType("application/json; charset=utf8")

	encoder := json.NewEncoder(srv.Ctx.Response.BodyWriter())
	encoder.SetEscapeHTML(false)

	data := struct {
		Code string `json:"code"`
		Desc string `json:"desc"`
	}{
		Code: code,
		Desc: fmt.Sprintf("%s", text),
	}

	dataWraped := struct {
		Error interface{} `json:"err"`
	}{data}
	encoder.Encode(dataWraped)
	panic("skip")
}

// ErrJSONP http api error as JSONP
func (srv *Server) ErrJSONP(text interface{}) {
	srv.RespJSONP(struct {
		Error string `json:"error"`
	}{
		Error: fmt.Sprintf("%s", text),
	})
	panic("skip")
}

func serverRequestHandler(ctx *fasthttp.RequestCtx) {
	defer func() {
		if r := recover(); r != nil {
			apiErrStr := fmt.Sprintf("%v", r)
			if apiErrStr != "skip" {
				fmt.Println("UNCATCHED PANIC", r)
				debug.PrintStack()
			}
		}
	}()
	srv := Server{Ctx: ctx, Path: string(ctx.Path())}

	path := srv.Path

	cb, ok := httpHandlers[path]
	if ok {
		cb(&srv)
	} else {
		srv.Err("Not found", "method is undefined")
	}
}

// Serve start handling HTTP requests using fasthttp
func Serve(portHTTP string) {
	h := serverRequestHandler

	log.Println("Server started, port", portHTTP)
	if err := fasthttp.ListenAndServe(":"+portHTTP, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

// Handle add callback to
func Handle(path string, callback func(srv *Server)) {
	httpHandlers[path] = callback
}
