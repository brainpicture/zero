package zero

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"

	"github.com/valyala/fasthttp"
)

//var httpHandlers = map[string]func(srv *Server){}

var httpHandlers = routerTree{}

// Server is an wrapper around fasthttp
type Server struct {
	Ctx        *fasthttp.RequestCtx
	Path       string
	PathParams map[string]string
}

// Resp writes any data as JSON to HTTP stream
func (srv *Server) Resp(data interface{}) {
	srv.Ctx.SetContentType("application/json; charset=utf8")

	encoder := json.NewEncoder(srv.Ctx.Response.BodyWriter())
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(data)
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

// RespOk returns ok answer, means successfully performed action
func (srv *Server) RespOk() {
	srv.Resp(H{
		"ok": true,
	})
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

// GetParamOpt fetches optional param as string
func (srv *Server) GetParamOpt(key string) string {
	args := srv.Ctx.QueryArgs()
	param := string(args.Peek(key))

	if param == "" {
		param = string(srv.Ctx.PostArgs().Peek(key))
	}

	return param
}

// GetParam fetches required param as string
func (srv *Server) GetParam(key string) string {
	param := srv.GetParamOpt(key)
	if param == "" {
		srv.Err("param", "param "+key+" is required")
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

// GetBody return body of request
func (srv *Server) GetBody() []byte {
	return srv.Ctx.Request.Body()
}

// ErrAuth sends auth error to client
func (srv *Server) ErrAuth(code string, text interface{}) {
	srv.SendError(401, code, text)
}

// ErrForbidden this resource is prohibited for access
func (srv *Server) ErrForbidden(code string, text interface{}) {
	srv.SendError(403, code, text)
}

// ErrNotFound this resource not found
func (srv *Server) ErrNotFound(code string, text interface{}) {
	srv.SendError(404, code, text)
}

// Err http api error
func (srv *Server) Err(code string, text interface{}) {
	srv.SendError(400, code, text)
}

// ErrServer meaning error doesnt depent on request and occur beacause of server error
func (srv *Server) ErrServer(code string, text interface{}) {
	srv.SendError(500, code, text)
}

// ErrMethod emmits when method not allowed
func (srv *Server) ErrMethod(code string, text interface{}) {
	srv.SendError(405, code, text)
}

// SendError http api error
func (srv *Server) SendError(httpCode int, code string, text interface{}) {
	srv.Ctx.SetStatusCode(httpCode)
	srv.Ctx.SetContentType("application/json; charset=utf8")

	encoder := json.NewEncoder(srv.Ctx.Response.BodyWriter())
	encoder.SetEscapeHTML(false)

	dataError := struct {
		Code string `json:"code"`
		Desc string `json:"desc"`
	}{
		Code: code,
		Desc: fmt.Sprintf("%s", text),
	}

	encoder.Encode(dataError)
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

// IsPost true if method is Post
func (srv *Server) IsPost() bool {
	return srv.Ctx.IsPost()
}

// IsGet true if method is Get
func (srv *Server) IsGet() bool {
	return srv.Ctx.IsGet()
}

// IsPut true if method is Put
func (srv *Server) IsPut() bool {
	return srv.Ctx.IsPut()
}

// IsPatch true if method is Patch
func (srv *Server) IsPatch() bool {
	return bytes.Equal(srv.Ctx.Method(), []byte("PATCH"))
}

// GetPathParam fetches required param from path
func (srv *Server) GetPathParam(key string) string {
	param, ok := srv.PathParams[key]
	if !ok || param == "" {
		srv.Err("param", "param "+key+" should be presented in PATH")
	}
	return param
}

// GetPathParamInt fetches required param from path and converts to Int
func (srv *Server) GetPathParamInt(key string) int64 {
	param := srv.GetPathParam(key)
	paramInt, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		srv.Err("param", "param "+key+" presented in PATH should be int")
	}
	return paramInt
}

// GetCookie will return cookie by name
func (srv *Server) GetCookie(key string) string {
	return string(srv.Ctx.Request.Header.Cookie(key))
}

// GetHeader will return header by name
func (srv *Server) GetHeader(key string) string {
	return string(srv.Ctx.Request.Header.Peek(key))
}

// GetHeader will return header by name
func (srv *Server) Method() string {
	return string(srv.Ctx.Method())
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
	method, ok := Methods[srv.Method()]
	if !ok {
		srv.Err("invalid_request", "Unsupported method")
		return
	}

	cb, params, err := httpHandlers.Route(method, srv.Path)
	if cb == nil {
		srv.Err("not_found", err)
		return
	}
	srv.PathParams = params
	cb(&srv)
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
	httpHandlers.Handle(Methods["*"], path, callback)
}
