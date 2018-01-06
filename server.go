package server

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

type Server struct {
	Ctx  *fasthttp.RequestCtx
	Path string
}

var handler func(srv *Server)
var handlerSocket func(soc *Socket)

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

func (srv *Server) RespJSONP(data interface{}) {
	cbName := srv.GetParam("jsoncallback")
	jsonData, err := json.Marshal(data)
	if err != nil {
		srv.Err("system", err)
	}
	srv.Ctx.Write([]byte(cbName + "(" + string(jsonData) + ")"))
	srv.Ctx.SetContentType("application/javascript; charset=utf8")
}

func (srv *Server) HTML(data []byte) {
	srv.Ctx.Write(data)
	srv.Ctx.SetContentType("text/html; charset=utf8")
}

func (srv *Server) JS(data []byte) {
	srv.Ctx.Write(data)
	srv.Ctx.SetContentType("application/javascript; charset=utf8")
}

func (srv *Server) FileBlob(data []byte, contentType string) {
	srv.Ctx.Write(data)
	srv.Ctx.SetContentType(contentType)
}

func (srv *Server) File(path string) {
	srv.Ctx.SendFile(path)
}

func (srv *Server) GetParam(key string) string {
	args := srv.Ctx.QueryArgs()
	param := string(args.Peek(key))

	if param == "" {
		param = string(srv.Ctx.PostArgs().Peek(key))
	}

	return param
}

func (srv *Server) GetParamInt(key string) int {
	args := srv.Ctx.QueryArgs()
	param, _ := args.GetUint(key)

	return int(param)
}

func (srv *Server) GetParamInt64(key string) int64 {
	args := srv.Ctx.QueryArgs()
	param := args.Peek(key)
	i, err := strconv.ParseInt(string(param), 10, 64)
	if err != nil {
		i = 0
	}
	return i
}

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
}
func (srv *Server) ErrJSONP(text interface{}) {
	srv.RespJSONP(struct {
		Error string `json:"error"`
	}{
		Error: fmt.Sprintf("%s", text),
	})
}

func serverRequestHandler(ctx *fasthttp.RequestCtx) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("UNCATCHED PANIC", r)
			debug.PrintStack()
		}
	}()
	srv := Server{Ctx: ctx, Path: string(ctx.Path())}

	path := srv.Path
	parts := strings.Split(path, "/")

	if len(parts) > 1 && parts[1] == "ws" {
		soc := Socket{
			Write: make(chan []byte),
			Read:  make(chan []byte),
		}
		soc.Token = srv.GetParam("token")
		soc.SessionID = srv.GetParamInt64("session_id")
		soc.LastEventID = srv.GetParamInt("last_event_id")
		fmt.Println("AUTH: " + soc.Token)
		srv.UpgradeWS(&soc)
		return
	}

	handler(&srv)
}

// StartServe serving files for jarvis-backend
func StartServe(port string, handlerVal func(srv *Server), handlerSocketVal func(soc *Socket)) {
	h := serverRequestHandler

	handler = handlerVal
	handlerSocket = handlerSocketVal

	if err := fasthttp.ListenAndServe(port, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
