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

//var httpHandlers = map[string]func(req *Request){}

//var httpHandlers = routerTree{}

// HTTP main type for http server
type HTTP struct {
	handlers  routerTree
	OnError   func(req *Request, name, text string)
	OnPanic   func(req *Request, stackTrace string)
	OnRequest func(req *Request)
	OnOptions func(req *Request)
	CORS      string
	server    *fasthttp.Server
	started   bool // true if server is started
	GZip      bool
	mux       sync.Mutex
}

// Request is an wrapper around fasthttp
type Request struct {
	Ctx         *fasthttp.RequestCtx
	Path        string
	PathParams  map[string]string
	ReferenceID int64 // used to point out userID during session
	http        *HTTP
	OnResponse  func(interface{})
	OnFail      func(int, string, interface{})
	supportGZip bool
}

func (req *Request) Write(data []byte) {
	if req.supportGZip {
		fasthttp.WriteGzip(req.Ctx.Response.BodyWriter(), data)
		req.Ctx.Response.Header.Add("Content-Encoding", "gzip")
	} else {
		req.Ctx.Write(data)
	}
}

// WriteJSON will push JSON data to the user
func (req *Request) WriteJSON(data interface{}) error {
	req.Ctx.SetContentType("application/json; charset=utf8")
	writer := req.Ctx.Response.BodyWriter()
	var encoder *json.Encoder
	if req.supportGZip {
		w := gzip.NewWriter(writer)
		defer w.Close()
		encoder = json.NewEncoder(w)
		req.Ctx.Response.Header.Add("Content-Encoding", "gzip")
	} else {
		encoder = json.NewEncoder(writer)
	}
	encoder.SetEscapeHTML(false)
	return encoder.Encode(data)
}

func (req *Request) writeCORSHeader() {
	if req.http.CORS != "" {
		req.Ctx.Response.Header.Set("Access-Control-Allow-Origin", req.http.CORS)
	}
}

// Resp writes any data as JSON to HTTP stream
func (req *Request) Resp(data interface{}) {
	req.writeCORSHeader()

	err := req.WriteJSON(data)

	if err != nil {
		req.Err("system", err)
	}

	if req.OnResponse != nil {
		req.OnResponse(data)
	}
}

// StreamBody will stream data to the client
func (req *Request) StreamBody(reader io.Reader, contentLength int64, contentType string) {
	req.Ctx.SetContentType(contentType + "; charset=utf8")
	req.writeCORSHeader()
	req.Ctx.SetBodyStream(reader, int(contentLength))
}

// RespJSONP writes any data as JSONP to HTTP stream
func (req *Request) RespJSONP(data interface{}) {
	req.writeCORSHeader()
	cbName := req.GetParam("jsoncallback")
	if cbName == "" {
		req.Resp(data)
		return
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		req.Err("system", err)
	}
	req.Write([]byte(cbName + "(" + string(jsonData) + ")"))
	req.Ctx.SetContentType("application/javascript; charset=utf8")
	if req.OnResponse != nil {
		req.OnResponse(data)
	}
}

// RespOk returns ok answer, means successfully performed action
func (req *Request) RespOk() {
	req.Resp(H{
		"ok": true,
	})
}

// HTML responds with HTML
func (req *Request) HTML(data []byte) {
	req.writeCORSHeader()
	req.Write(data)
	req.Ctx.SetContentType("text/html; charset=utf8")
}

// JS responds with JS
func (req *Request) JS(data []byte) {
	req.writeCORSHeader()
	req.Write(data)
	req.Ctx.SetContentType("application/javascript; charset=utf8")
}

// FileBlob responds with FileBlob
func (req *Request) FileBlob(data []byte, contentType string) {
	req.writeCORSHeader()
	req.Write(data)
	req.Ctx.SetContentType(contentType)
}

// File responds with File
func (req *Request) File(path string) {
	req.writeCORSHeader()
	req.Ctx.SendFile(path)
}

// GetParams return all params
func (req *Request) GetParams() map[string]string {
	args := req.Ctx.QueryArgs()
	res := map[string]string{}
	args.VisitAll(func(key []byte, value []byte) {
		res[string(key)] = string(value)
	})
	return res
}

// GetParamOpt fetches optional param as string
func (req *Request) GetParamOpt(key string) string {
	args := req.Ctx.QueryArgs()
	param := string(args.Peek(key))

	if param == "" {
		param = string(req.Ctx.PostArgs().Peek(key))
	}

	return param
}

// GetParam fetches required param as string
func (req *Request) GetParam(key string) string {
	param := req.GetParamOpt(key)
	if param == "" {
		req.Err("param", "param "+key+" is required")
	}
	return param
}

// GetParamPagination parse param for pagination
func (req *Request) GetParamPagination(defCount int) *Pagination {
	result := Pagination{req: req}
	param := req.GetParamOpt("from")
	count := req.GetParamInt("count")
	result.Reverse = req.GetParamBool("reverse")
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
			req.Err("param", "param from has invalid format for pagination")
		}
		result.ObjectID = objID
	}
	return &result
}

// GetParamInt request param converted to int
func (req *Request) GetParamInt(key string) int {
	args := req.Ctx.QueryArgs()
	param, err := args.GetUint(key)
	if err != nil {
		return 0
	}

	return int(param)
}

// GetParamInt64 request param converted to int64
func (req *Request) GetParamInt64(key string) int64 {
	args := req.Ctx.QueryArgs()
	param := args.Peek(key)
	i, err := strconv.ParseInt(string(param), 10, 64)
	if err != nil {
		i = 0
	}
	return i
}

// GetParamOptInt64 request param converted to int64
func (req *Request) GetParamOptInt64(key string) (int64, bool) {
	args := req.Ctx.QueryArgs()
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
func (req *Request) GetParamFloat(key string) float64 {
	args := req.Ctx.QueryArgs()
	param, err := args.GetUfloat(key)
	if err != nil {
		return 0
	}

	return param
}

// GetParamBool request param converted to bool
func (req *Request) GetParamBool(key string) bool {
	args := req.Ctx.QueryArgs()
	param := string(args.Peek(key))
	if param == "1" || param == "true" || param == "y" {
		return true
	}
	return false
}

// GetBody return body of request
func (req *Request) GetBody() []byte {
	return req.Ctx.Request.Body()
}

// ErrAuth sends auth error to client
func (req *Request) ErrAuth(code string, text interface{}) {
	req.ErrCode(401, code, text)
}

// ErrForbidden this resource is prohibited for access
func (req *Request) ErrForbidden(code string, text interface{}) {
	req.ErrCode(403, code, text)
}

// ErrFlood too ofter, retry later
func (req *Request) ErrFlood(code string, text interface{}) {
	req.ErrCode(429, code, text)
}

// ErrNotFound this resource not found
func (req *Request) ErrNotFound(code string, text interface{}) {
	req.ErrCode(404, code, text)
}

// Err http api error
func (req *Request) Err(code string, text interface{}) {
	req.ErrCode(400, code, text)
}

// ErrServer meaning error doesnt depent on request and occur beacause of server error
func (req *Request) ErrServer(code string, text interface{}) {
	req.ErrCode(500, code, text)
}

// ErrMethod emmits when method not allowed
func (req *Request) ErrMethod(code string, text interface{}) {
	req.ErrCode(405, code, text)
}

// ErrCustom will send error with custom fields
func (req *Request) ErrCustom(errCode int, code, desc string, data H) {
	req.Ctx.SetStatusCode(errCode)
	req.Ctx.SetContentType("application/json; charset=utf8")
	req.writeCORSHeader()

	encoder := json.NewEncoder(req.Ctx.Response.BodyWriter())
	encoder.SetEscapeHTML(false)

	data["code"] = code
	data["desc"] = desc
	encoder.Encode(data)

	if req.http.OnError != nil {
		req.http.OnError(req, code, desc)
	}
	if req.OnFail != nil {
		req.OnFail(errCode, code, desc)
	}
	panic("skip")
}

// ErrCode send error with code
func (req *Request) ErrCode(httpCode int, code string, text interface{}) {
	req.SendError(httpCode, code, text)
	if req.http.OnError != nil {
		req.http.OnError(req, code, fmt.Sprintf("%s", text))
	}
	if req.OnFail != nil {
		req.OnFail(httpCode, code, text)
	}
	panic("skip")
}

// SendError http api error
func (req *Request) SendError(httpCode int, code string, text interface{}) {
	req.Ctx.SetStatusCode(httpCode)
	req.Ctx.SetContentType("application/json; charset=utf8")
	req.writeCORSHeader()

	dataError := struct {
		Code string `json:"code"`
		Desc string `json:"desc"`
	}{
		Code: code,
		Desc: fmt.Sprintf("%s", text),
	}

	req.WriteJSON(dataError)
}

// ErrJSONP http api error as JSONP
func (req *Request) ErrJSONP(code string, text interface{}) {
	desc := fmt.Sprintf("%s", text)
	req.RespJSONP(struct {
		Code  string `json:"code"`
		Error string `json:"error"`
	}{
		Code: code,
		Error: desc,
	})
	if req.http.OnError != nil {
		req.http.OnError(req, code, desc)
	}
	if req.OnFail != nil {
		req.OnFail(400, code, desc)
	}
	panic("skip")
}

// Redirect will return server-side redirect
func (req *Request) Redirect(uri string, statusCode int) {
	req.Ctx.Redirect(uri, statusCode)
}

// IsPost true if method is Post
func (req *Request) IsPost() bool {
	return req.Ctx.IsPost()
}

// IsGet true if method is Get
func (req *Request) IsGet() bool {
	return req.Ctx.IsGet()
}

// IsPut true if method is Put
func (req *Request) IsPut() bool {
	return req.Ctx.IsPut()
}

// IsPatch true if method is Patch
func (req *Request) IsPatch() bool {
	return bytes.Equal(req.Ctx.Method(), []byte("PATCH"))
}

// GetFile fetches file from multipart
func (req *Request) GetFile(name string) *File {
	if !req.IsPost() {
		req.Err("upload_file_error", "request method should be POST")
		return nil
	}
	fileHeadler, err := req.Ctx.FormFile(name)
	if err != nil {
		req.Err("upload_file_error", err)
		return nil
	}
	file := File{}
	file.SetMultipart(fileHeadler)
	return &file
}

// TryFile fetches file from multipart only if presented
func (req *Request) TryFile(name string) *File {
	if !req.IsPost() {
		return nil
	}
	fileHeadler, err := req.Ctx.FormFile(name)
	if err != nil {
		return nil
	}
	file := File{}
	file.SetMultipart(fileHeadler)
	return &file
}

// GetPathParam fetches required param from path
func (req *Request) GetPathParam(key string) string {
	param, ok := req.PathParams[key]
	if !ok || param == "" {

		req.Err("param", "param "+key+" should be presented in PATH: "+req.Path)
	}
	return param
}

// GetPathParamInt fetches required param from path and converts to Int
func (req *Request) GetPathParamInt(key string) int64 {
	param := req.GetPathParam(key)
	paramInt, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		req.Err("param", "param "+key+" presented in PATH should be int")
	}
	return paramInt
}

// GetCookie will return cookie by name
func (req *Request) GetCookie(key string) string {
	return string(req.Ctx.Request.Header.Cookie(key))
}

// SetCookie will return cookie by name
func (req *Request) SetCookie(key string, value string) {
	cookie := fasthttp.Cookie{}
	cookie.SetKey(key)
	cookie.SetValue(value)
	cookie.SetExpire(time.Now().Add(time.Hour * 24 * 365 * 2))
	cookie.SetPath("/")
	req.Ctx.Response.Header.SetCookie(&cookie)
}

// GetHeader will return header by name
func (req *Request) GetHeader(key string) string {
	return string(req.Ctx.Request.Header.Peek(key))
}

// SetHeader will set header
func (req *Request) SetHeader(key string, value string) {
	req.Ctx.Response.Header.Set(key, value)
}

// GetLanguage will return short form of language browser use
func (req *Request) GetLanguage() string {
	language := req.GetHeader("Accept-Language")
	if len(language) < 2 {
		return "en"
	}
	langShort, _ := SplitDoubleString(language, "-")

	return strings.ToLower(langShort)
}

// GetUserAgent returns user agent header
func (req *Request) GetUserAgent() string {
	ua := req.GetHeader("X-User-Agent")
	if ua == "" {
		ua = req.GetHeader("User-Agent")
	}
	return ua
}

// Method will return method from requiest header
func (req *Request) Method() string {
	return string(req.Ctx.Method())
}

// Check perform check and trigger req.Err if err is not nil
func (req *Request) Check(err error, text ...string) {
	if err != nil {
		if len(text) > 0 {
			if len(text) > 1 {
				req.Err(text[0], text[1])
			} else {
				req.Err(text[0], err)
			}
		} else {
			req.ErrServer("request_failed", err)
		}
	}
}

// GetSessionID return Session-ID hearder int64
func (req *Request) GetSessionID() int64 {
	sessionIDStr := req.GetHeader("Session-ID")
	if sessionIDStr == "" {
		sessionIDStr = req.GetParamOpt("session-id")
	}
	return I64(sessionIDStr)
}

// GetIP will return remove ip of request
func (req *Request) GetIP() net.IP {
	return req.Ctx.RemoteIP()
}

// Background will run anonymous goroutine in background with proper error catching
func (req *Request) Background(handler func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				apiErrStr := fmt.Sprintf("%v", r)
				if apiErrStr != "skip" {
					if req.http.OnPanic != nil {
						req.http.OnPanic(req, apiErrStr+"\n\n"+string(debug.Stack()))
					} else {
						fmt.Println("UNCATCHED PANIC", r)
						debug.PrintStack()
					}
				}
			}
		}()
		handler()
	}()
}

// EventSource starts an event server
func (req *Request) EventSource(callback func(*ServerEvents)) {
	req.Ctx.SetContentType("text/event-stream; charset=UTF-8")
	req.Ctx.Response.Header.Set("Cache-Control", "no-cache")
	req.Ctx.Response.Header.Set("Connection", "keep-alive")
	req.Ctx.Response.Header.Set("Transfer-Encoding", "chunked")
	req.writeCORSHeader()
	if req.http.CORS != "" {
		req.Ctx.Response.Header.Set("Access-Control-Expose-Headers", "*")
		req.Ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
	}

	lastEventIDStr := req.GetHeader("Last-Event-ID")
	if lastEventIDStr == "" {
		lastEventIDStr = req.GetParamOpt("last-event-id")
	}
	sessionID := req.GetSessionID()

	req.Ctx.SetBodyStreamWriter(func(w *bufio.Writer) {
		defer func() {
			if r := recover(); r != nil {
				apiErrStr := fmt.Sprintf("%v", r)
				if apiErrStr != "skip" {
					if req.http.OnPanic != nil {
						req.http.OnPanic(req, apiErrStr+"\n\n"+string(debug.Stack()))
					} else {
						fmt.Println("UNCATCHED PANIC", r)
						debug.PrintStack()
					}
				}
			}
		}()
		se := ServerEvents{
			Writer:    w,
			EventID:   I64(lastEventIDStr),
			SessionID: sessionID,
		}
		callback(&se)
	})
}

// Event sends event to the user
func (req *Request) Event(data interface{}) {
	encoder := json.NewEncoder(req.Ctx.Response.BodyWriter())
	encoder.SetEscapeHTML(false)

	err := encoder.Encode(data)
	if err != nil {
		req.Err("system", err)
	}
}

// ParseInt64List parses []int64 from json body
func (req *Request) ParseInt64List() []int64 {
	input := []int64{}
	err := json.Unmarshal(req.GetBody(), &input)
	if err != nil {
		req.Err("user_invalid", "Body should be json list of strings")
	}
	return input
}

// ParseStrList parses []string from json body
func (req *Request) ParseStrList() []string {
	input := []string{}
	err := json.Unmarshal(req.GetBody(), &input)
	if err != nil {
		req.Err("user_invalid", "Body should be json list of strings")
	}
	return input
}

// ParseBody return H object of input object
func (req *Request) ParseBody() H {
	input := H{}
	err := json.Unmarshal(req.GetBody(), &input)
	if err != nil {
		req.Err("user_object", "Body should be json")
	}
	return input
}

// FillBody allow to fill a struct with data from body
func (req *Request) FillBody(input interface{}) {
	err := json.Unmarshal(req.GetBody(), input)
	if err != nil {
		req.Err("user_object", "Body should be json")
	}
}

// Env will return environment for
func (req *Request) Env() *Environment {
	return Env(req)
}

// GetRealIP will return ip address of user
func (req *Request) GetRealIP() net.IP {
	ip := req.GetHeader("X-Real-IP")
	if ip == "" {
		ipFWD := req.GetHeader("X-Forwarded-For")
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
		req := Request{
			Ctx:  ctx,
			Path: string(ctx.Path()),
			http: h,
		}
		if h.GZip {
			gzipHeader := req.GetHeader("Accept-Encoding")
			if gzipHeader == "*" || strings.Contains(gzipHeader, "gzip") {
				req.supportGZip = true
			}
		}
		defer func() {
			if r := recover(); r != nil {
				apiErrStr := fmt.Sprintf("%v", r)
				if apiErrStr != "skip" {
					if h.OnPanic != nil {
						h.OnPanic(&req, apiErrStr+"\n\n"+string(debug.Stack()))
					} else {
						fmt.Println("UNCATCHED PANIC", r)
						debug.PrintStack()
					}
					// this error do not panic
					req.SendError(500, "fatal", "runtime error")
				}
			}
		}()
		methodStr := req.Method()
		if methodStr == "OPTIONS" {
			if h.OnOptions != nil {
				h.OnOptions(&req)
				return
			}
			if req.http.CORS != "" {
				req.SetHeader("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, UPDATE")
				req.SetHeader("Access-Control-Allow-Headers", "*")
				req.RespOk()
				return
			}
		}
		method, ok := Methods[methodStr]
		if !ok {
			req.Err("invalid_request", "Unsupported method")
			return
		}

		cb, params, err := h.handlers.Route(method, req.Path)
		if cb == nil {
			req.Err("not_found", err)
			return
		}
		req.PathParams = params
		if req.http.OnRequest != nil {
			req.http.OnRequest(&req)
		}
		cb(&req)
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
func (h *HTTP) Handle(path string, callback func(req *Request)) {
	h.handlers.Handle(Methods["*"], path, callback)
}

// SetCORS setup cors header for the api
func (h *HTTP) SetCORS(allow string) {
	h.CORS = allow
}
