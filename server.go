package zero

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/reuseport"
)

//var httpHandlers = map[string]func(srv *Server){}

//var httpHandlers = routerTree{}

// HTTP main type for http server
type HTTP struct {
	handlers  routerTree
	OnError   func(srv *Server, name, text string)
	OnPanic   func(srv *Server, stackTrace string)
	OnRequest func(srv *Server)
	CORS      string
	server    *fasthttp.Server
	started   bool // true if server is started
	GZip      bool
	mux       sync.Mutex
}

// Server is an wrapper around fasthttp
type Server struct {
	Ctx         *fasthttp.RequestCtx
	Path        string
	PathParams  map[string]string
	ReferenceID int64 // used to point out userID during session
	http        *HTTP
	OnResponse  func(interface{})
	OnFail      func(int, string, interface{})
	supportGZip bool
}

func (srv *Server) Write(data []byte) {
	if srv.supportGZip {
		fasthttp.WriteGzip(srv.Ctx.Response.BodyWriter(), data)
		srv.Ctx.Response.Header.Add("Content-Encoding", "gzip")
	} else {
		srv.Ctx.Write(data)
	}
}

// WriteJSON will push JSON data to the user
func (srv *Server) WriteJSON(data interface{}) error {
	srv.Ctx.SetContentType("application/json; charset=utf8")
	writer := srv.Ctx.Response.BodyWriter()
	var encoder *json.Encoder
	if srv.supportGZip {
		w := gzip.NewWriter(writer)
		defer w.Close()
		encoder = json.NewEncoder(w)
		srv.Ctx.Response.Header.Add("Content-Encoding", "gzip")
	} else {
		encoder = json.NewEncoder(writer)
	}
	encoder.SetEscapeHTML(false)
	return encoder.Encode(data)
}

// Resp writes any data as JSON to HTTP stream
func (srv *Server) Resp(data interface{}) {
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}

	err := srv.WriteJSON(data)

	if err != nil {
		srv.Err("system", err)
	}

	if srv.OnResponse != nil {
		srv.OnResponse(data)
	}
}

// StreamBody will stream data to the client
func (srv *Server) StreamBody(reader io.Reader, contentLength int64, contentType string) {
	srv.Ctx.SetContentType(contentType + "; charset=utf8")
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}
	srv.Ctx.SetBodyStream(reader, int(contentLength))
}

// RespJSONP writes any data as JSONP to HTTP stream
func (srv *Server) RespJSONP(data interface{}) {
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}
	cbName := srv.GetParam("jsoncallback")
	jsonData, err := json.Marshal(data)
	if err != nil {
		srv.Err("system", err)
	}
	srv.Write([]byte(cbName + "(" + string(jsonData) + ")"))
	srv.Ctx.SetContentType("application/javascript; charset=utf8")
	if srv.OnResponse != nil {
		srv.OnResponse(data)
	}
}

// RespOk returns ok answer, means successfully performed action
func (srv *Server) RespOk() {
	srv.Resp(H{
		"ok": true,
	})
}

// HTML responds with HTML
func (srv *Server) HTML(data []byte) {
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}
	srv.Write(data)
	srv.Ctx.SetContentType("text/html; charset=utf8")
}

// JS responds with JS
func (srv *Server) JS(data []byte) {
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}
	srv.Write(data)
	srv.Ctx.SetContentType("application/javascript; charset=utf8")
}

// FileBlob responds with FileBlob
func (srv *Server) FileBlob(data []byte, contentType string) {
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}
	srv.Write(data)
	srv.Ctx.SetContentType(contentType)
}

// File responds with File
func (srv *Server) File(path string) {
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}
	srv.Ctx.SendFile(path)
}

// GetParams return all params
func (srv *Server) GetParams() map[string]string {
	args := srv.Ctx.QueryArgs()
	res := map[string]string{}
	args.VisitAll(func(key []byte, value []byte) {
		res[string(key)] = string(value)
	})
	return res
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

// GetParamPagination parse param for pagination
func (srv *Server) GetParamPagination(defCount int) *Pagination {
	result := Pagination{srv: srv}
	param := srv.GetParamOpt("from")
	count := srv.GetParamInt("count")
	result.Reverse = srv.GetParamBool("reverse")
	if count == 0 {
		count = defCount
	}
	result.Count = count
	if param == "" {
		return &result
	}
	result.from = param
	rows := strings.Split(param, ":")

	subrows := strings.Split(rows[0], ";")
	result.offset = I64(subrows[0])

	if len(rows) > 1 {
		objID, err := strconv.ParseInt(rows[1], 10, 64)
		if err != nil {
			srv.Err("param", "param from has invalid format for pagination")
		}
		result.ObjectID = objID
	}
	return &result
}

// GetParamInt request param converted to int
func (srv *Server) GetParamInt(key string) int {
	args := srv.Ctx.QueryArgs()
	param, err := args.GetUint(key)
	if err != nil {
		return 0
	}

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

// GetParamOptInt64 request param converted to int64
func (srv *Server) GetParamOptInt64(key string) (int64, bool) {
	args := srv.Ctx.QueryArgs()
	param := args.Peek(key)
	if len(param) == 0 {
		return 0, false
	}
	i, err := strconv.ParseInt(string(param), 10, 64)
	if err != nil {
		i = 0
	}
	return i, true
}

// GetParamFloat request param converted to float64
func (srv *Server) GetParamFloat(key string) float64 {
	args := srv.Ctx.QueryArgs()
	param, err := args.GetUfloat(key)
	if err != nil {
		return 0
	}

	return param
}

// GetParamBool request param converted to bool
func (srv *Server) GetParamBool(key string) bool {
	args := srv.Ctx.QueryArgs()
	param := string(args.Peek(key))
	if param == "1" || param == "true" || param == "y" {
		return true
	}
	return false
}

// GetBody return body of request
func (srv *Server) GetBody() []byte {
	return srv.Ctx.Request.Body()
}

// ErrAuth sends auth error to client
func (srv *Server) ErrAuth(code string, text interface{}) {
	srv.ErrCode(401, code, text)
}

// ErrForbidden this resource is prohibited for access
func (srv *Server) ErrForbidden(code string, text interface{}) {
	srv.ErrCode(403, code, text)
}

// ErrFlood too ofter, retry later
func (srv *Server) ErrFlood(code string, text interface{}) {
	srv.ErrCode(429, code, text)
}

// ErrNotFound this resource not found
func (srv *Server) ErrNotFound(code string, text interface{}) {
	srv.ErrCode(404, code, text)
}

// Err http api error
func (srv *Server) Err(code string, text interface{}) {
	srv.ErrCode(400, code, text)
}

// ErrServer meaning error doesnt depent on request and occur beacause of server error
func (srv *Server) ErrServer(code string, text interface{}) {
	srv.ErrCode(500, code, text)
}

// ErrMethod emmits when method not allowed
func (srv *Server) ErrMethod(code string, text interface{}) {
	srv.ErrCode(405, code, text)
}

// ErrCustom will send error with custom fields
func (srv *Server) ErrCustom(errCode int, code, desc string, data H) {
	srv.Ctx.SetStatusCode(errCode)
	srv.Ctx.SetContentType("application/json; charset=utf8")
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}

	encoder := json.NewEncoder(srv.Ctx.Response.BodyWriter())
	encoder.SetEscapeHTML(false)

	data["code"] = code
	data["desc"] = desc
	encoder.Encode(data)

	if srv.http.OnError != nil {
		srv.http.OnError(srv, code, desc)
	}
	if srv.OnFail != nil {
		srv.OnFail(errCode, code, desc)
	}
	panic("skip")
}

// ErrCode send error with code
func (srv *Server) ErrCode(httpCode int, code string, text interface{}) {
	srv.SendError(httpCode, code, text)
	if srv.http.OnError != nil {
		srv.http.OnError(srv, code, fmt.Sprintf("%s", text))
	}
	if srv.OnFail != nil {
		srv.OnFail(httpCode, code, text)
	}
	panic("skip")
}

// SendError http api error
func (srv *Server) SendError(httpCode int, code string, text interface{}) {
	srv.Ctx.SetStatusCode(httpCode)
	srv.Ctx.SetContentType("application/json; charset=utf8")
	if srv.http.CORS != "" {
		srv.Ctx.Response.Header.Set("Access-Control-Allow-Origin", srv.http.CORS)
	}

	dataError := struct {
		Code string `json:"code"`
		Desc string `json:"desc"`
	}{
		Code: code,
		Desc: fmt.Sprintf("%s", text),
	}

	srv.WriteJSON(dataError)
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

// Redirect will return server-side redirect
func (srv *Server) Redirect(uri string, statusCode int) {
	srv.Ctx.Redirect(uri, statusCode)
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

// GetFile fetches file from multipart
func (srv *Server) GetFile(name string) *File {
	if !srv.IsPost() {
		srv.Err("upload_file_error", "request method should be POST")
		return nil
	}
	fileHeadler, err := srv.Ctx.FormFile(name)
	if err != nil {
		srv.Err("upload_file_error", err)
		return nil
	}
	file := File{}
	file.SetMultipart(fileHeadler)
	return &file
}

// TryFile fetches file from multipart only if presented
func (srv *Server) TryFile(name string) *File {
	if !srv.IsPost() {
		return nil
	}
	fileHeadler, err := srv.Ctx.FormFile(name)
	if err != nil {
		return nil
	}
	file := File{}
	file.SetMultipart(fileHeadler)
	return &file
}

// GetPathParam fetches required param from path
func (srv *Server) GetPathParam(key string) string {
	param, ok := srv.PathParams[key]
	if !ok || param == "" {

		srv.Err("param", "param "+key+" should be presented in PATH: "+srv.Path)
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

// SetCookie will return cookie by name
func (srv *Server) SetCookie(key string, value string) {
	cookie := fasthttp.Cookie{}
	cookie.SetKey(key)
	cookie.SetValue(value)
	cookie.SetExpire(time.Now().Add(time.Hour * 24 * 365 * 2))
	cookie.SetPath("/")
	srv.Ctx.Response.Header.SetCookie(&cookie)
}

// GetHeader will return header by name
func (srv *Server) GetHeader(key string) string {
	return string(srv.Ctx.Request.Header.Peek(key))
}

// SetHeader will set header
func (srv *Server) SetHeader(key string, value string) {
	srv.Ctx.Response.Header.Set(key, value)
}

// GetLanguage will return short form of language browser use
func (srv *Server) GetLanguage() string {
	language := srv.GetHeader("Accept-Language")
	if len(language) < 2 {
		return "en"
	}
	langShort, _ := SplitDoubleString(language, "-")

	return strings.ToLower(langShort)
}

// GetUserAgent returns user agent header
func (srv *Server) GetUserAgent() string {
	ua := srv.GetHeader("X-User-Agent")
	if ua == "" {
		ua = srv.GetHeader("User-Agent")
	}
	return ua
}

// Method will return method from requiest header
func (srv *Server) Method() string {
	return string(srv.Ctx.Method())
}

// Check perform check and trigger srv.Err if err is not nil
func (srv *Server) Check(err error, text ...string) {
	if err != nil {
		if len(text) > 0 {
			if len(text) > 1 {
				srv.Err(text[0], text[1])
			} else {
				srv.Err(text[0], err)
			}
		} else {
			srv.ErrServer("request_failed", err)
		}
	}
}

// GetSessionID return Session-ID hearder int64
func (srv *Server) GetSessionID() int64 {
	sessionIDStr := srv.GetHeader("Session-ID")
	if sessionIDStr == "" {
		sessionIDStr = srv.GetParamOpt("session-id")
	}
	return I64(sessionIDStr)
}

// GetIP will return remove ip of request
func (srv *Server) GetIP() net.IP {
	return srv.Ctx.RemoteIP()
}

// EventSource starts an event server
func (srv *Server) EventSource(callback func(*ServerEvents)) {
	srv.Ctx.SetContentType("text/event-stream; charset=UTF-8")
	srv.Ctx.Response.Header.Set("Cache-Control", "no-cache")
	srv.Ctx.Response.Header.Set("Connection", "keep-alive")
	srv.Ctx.Response.Header.Set("Transfer-Encoding", "chunked")

	lastEventIDStr := srv.GetHeader("Last-Event-ID")
	if lastEventIDStr == "" {
		lastEventIDStr = srv.GetParamOpt("last-event-id")
	}
	sessionID := srv.GetSessionID()

	srv.Ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
		se := ServerEvents{
			Writer:    w,
			EventID:   I64(lastEventIDStr),
			SessionID: sessionID,
		}

		callback(&se)
	})
}

// Event sends event to the user
func (srv *Server) Event(data interface{}) {
	encoder := json.NewEncoder(srv.Ctx.Response.BodyWriter())
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(data)
	if err != nil {
		srv.Err("system", err)
	}
}

// ParseInt64List parses []int64 from json body
func (srv *Server) ParseInt64List() []int64 {
	input := []int64{}
	err := json.Unmarshal(srv.GetBody(), &input)
	if err != nil {
		srv.Err("user_invalid", "Body should be json list of strings")
	}
	return input
}

// ParseStrList parses []string from json body
func (srv *Server) ParseStrList() []string {
	input := []string{}
	err := json.Unmarshal(srv.GetBody(), &input)
	if err != nil {
		srv.Err("user_invalid", "Body should be json list of strings")
	}
	return input
}

// ParseBody return H object of input object
func (srv *Server) ParseBody() H {
	input := H{}
	err := json.Unmarshal(srv.GetBody(), &input)
	if err != nil {
		srv.Err("user_object", "Body should be json")
	}
	return input
}

// Env will return environment for
func (srv *Server) Env() *Environment {
	return Env(srv)
}

// GetRealIP will return ip address of user
func (srv *Server) GetRealIP() net.IP {
	ip := srv.GetHeader("X-Real-IP")
	if ip == "" {
		ipFWD := srv.GetHeader("X-Forwarded-For")
		fwdips := strings.Split(ipFWD, ", ")
		if len(fwdips) > 0 {
			ip = fwdips[0]
		}
	}
	IP := net.ParseIP(ip)
	return IP
}

// Shutdown will gracefully shutdown the app, stopping receiving new connections but continue receive old one
func (h *HTTP) Shutdown() error {
	if !h.IsStarted() {
		return nil
	}
	return h.server.Shutdown()
}

// IsStarted return true if server started
func (h *HTTP) IsStarted() bool {
	h.mux.Lock()
	isStarted := h.started
	h.mux.Unlock()
	return isStarted
}

// Serve start handling HTTP requests using fasthttp
func (h *HTTP) Serve(portHTTP string) {
	ctxHanler := func(ctx *fasthttp.RequestCtx) {
		srv := Server{
			Ctx:  ctx,
			Path: string(ctx.Path()),
			http: h,
		}
		if h.GZip {
			gzipHeader := srv.GetHeader("Accept-Encoding")
			if gzipHeader == "*" || strings.Contains(gzipHeader, "gzip") {
				srv.supportGZip = true
			}
		}
		defer func() {
			if r := recover(); r != nil {
				apiErrStr := fmt.Sprintf("%v", r)
				if apiErrStr != "skip" {
					if h.OnPanic != nil {
						h.OnPanic(&srv, apiErrStr+"\n\n"+string(debug.Stack()))
					} else {
						fmt.Println("UNCATCHED PANIC", r)
						debug.PrintStack()
					}
					// this error do not panic
					srv.SendError(500, "fatal", "runtime error")
				}
			}
		}()
		//start := time.Now()
		methodStr := srv.Method()
		if methodStr == "OPTIONS" {
			if srv.http.CORS == "*" {
				srv.SetHeader("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, UPDATE")
				srv.RespOk()
				return
			}
		}
		method, ok := Methods[methodStr]
		if !ok {
			srv.Err("invalid_request", "Unsupported method")
			return
		}

		cb, params, err := h.handlers.Route(method, srv.Path)
		if cb == nil {
			srv.Err("not_found", err)
			return
		}
		srv.PathParams = params
		if srv.http.OnRequest != nil {
			srv.http.OnRequest(&srv)
		}
		cb(&srv)
		//elapsed := time.Since(start)
	}

	//fasthttp.DialTimeout(addr, 24*time.Hour)
	log.Println("Server started, port", portHTTP)

	// NOTE: Package reuseport provides a TCP net.Listener with SO_REUSEPORT support.
	// SO_REUSEPORT allows linear scaling server performance on multi-CPU servers.
	ln, err := reuseport.Listen("tcp4", ":"+portHTTP)
	h.server = &fasthttp.Server{
		Handler:               ctxHanler,
		NoDefaultServerHeader: true,
	}
	if err == nil {
		err = h.server.Serve(ln)
	} else {
		log.Fatalf("error in reuseport listener: %s, fallback default", err)

		err = h.server.ListenAndServe(":" + portHTTP)
	}
	if err != nil {
		log.Fatalf("Error in start server: %s", err)
	} else {
		h.mux.Lock()
		h.started = true
		h.mux.Unlock()
	}
}

// Handle add callback to
func (h *HTTP) Handle(path string, callback func(srv *Server)) {
	h.handlers.Handle(Methods["*"], path, callback)
}

// SetCORS setup cors header for the api
func (h *HTTP) SetCORS(allow string) {
	h.CORS = allow
}
