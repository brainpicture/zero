package zero

import "strings"

// RestAPI is general object used for REST api
type RestAPI struct {
	Path string
	http *HTTP
}

func (r *RestAPI) joinPath(path string) []string {
	mainParts := strings.Split(r.Path, "/")
	parts := strings.Split(path, "/")
	for _, p := range parts {
		mainParts = append(mainParts, p)
	}
	return mainParts
}

// GET handler for GET method
func (r *RestAPI) GET(path string, callback func(srv *Server)) {
	r.http.handlers.PushHandler(Methods["GET"], r.joinPath(path), callback, []string{})
}

// POST handler for POST method
func (r *RestAPI) POST(path string, callback func(srv *Server)) {
	r.http.handlers.PushHandler(Methods["POST"], r.joinPath(path), callback, []string{})
}

// PATCH handler for PATCH method
func (r *RestAPI) PATCH(path string, callback func(srv *Server)) {
	r.http.handlers.PushHandler(Methods["PATCH"], r.joinPath(path), callback, []string{})
}

// PUT handler for PUT method
func (r *RestAPI) PUT(path string, callback func(srv *Server)) {
	r.http.handlers.PushHandler(Methods["PUT"], r.joinPath(path), callback, []string{})
}

// DELETE handler for DELETE method
func (r *RestAPI) DELETE(path string, callback func(srv *Server)) {
	r.http.handlers.PushHandler(Methods["DELETE"], r.joinPath(path), callback, []string{})
}

// UPDATE handler for UPDATE method
func (r *RestAPI) UPDATE(path string, callback func(srv *Server)) {
	r.http.handlers.PushHandler(Methods["UPDATE"], r.joinPath(path), callback, []string{})
}

// OPTIONS handler for OPTIONS method
func (r *RestAPI) OPTIONS(path string, callback func(srv *Server)) {
	r.http.handlers.PushHandler(Methods["OPTIONS"], r.joinPath(path), callback, []string{})
}

// SetCORS setup cors header for the api
func (r *RestAPI) SetCORS(allow string) {
	r.http.CORS = allow
}

// Rest init function for rest object
func (h *HTTP) Rest(path string) *RestAPI {
	api := RestAPI{
		Path: path,
		http: h,
	}

	return &api
}
