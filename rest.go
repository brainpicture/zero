package zero

import "strings"

// RestAPI is general object used for REST api
type RestAPI struct {
	Path string
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
	httpHandlers.PushHandler(Methods["GET"], r.joinPath(path), callback, []string{})
}

// POST handler for POST method
func (r *RestAPI) POST(path string, callback func(srv *Server)) {
	httpHandlers.PushHandler(Methods["POST"], r.joinPath(path), callback, []string{})
}

// PATCH handler for PATCH method
func (r *RestAPI) PATCH(path string, callback func(srv *Server)) {
	httpHandlers.PushHandler(Methods["PATCH"], r.joinPath(path), callback, []string{})
}

// DELETE handler for DELETE method
func (r *RestAPI) DELETE(path string, callback func(srv *Server)) {
	httpHandlers.PushHandler(Methods["DELETE"], r.joinPath(path), callback, []string{})
}

// UPDATE handler for UPDATE method
func (r *RestAPI) UPDATE(path string, callback func(srv *Server)) {
	httpHandlers.PushHandler(Methods["UPDATE"], r.joinPath(path), callback, []string{})
}

// Rest init function for rest object
func Rest(path string) *RestAPI {
	api := RestAPI{
		Path: path,
	}

	return &api
}
