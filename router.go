package zero

import (
	"errors"
	"strings"
)

var Methods = map[string]int{
	"*":      0,
	"GET":    1,
	"POST":   2,
	"PATCH":  3,
	"UPDATE": 4,
	"DELETE": 5,
}

type routerTree struct {
	Tree    map[string]*routerTree
	Methods map[int]routerMethodHandler
	//Keys    []string
	//Handler map[int]func(srv *Server)
}

type routerMethodHandler struct {
	Keys   []string
	Handle func(srv *Server)
}

func (p *routerTree) PushHandler(method int, parts []string, handler func(srv *Server), keys []string) {
	if len(parts) == 0 {
		methodHandler := routerMethodHandler{
			Keys:   keys,
			Handle: handler,
		}
		if p.Methods == nil {
			p.Methods = map[int]routerMethodHandler{}
		}
		p.Methods[method] = methodHandler
		return
	}
	row := parts[0]
	if row == "" {
		p.PushHandler(method, parts[1:], handler, keys)
		return
	}
	if row[0:1] == ":" {
		keys = append(keys, row[1:])
		row = ":" // shorten variable
	}
	if p.Tree == nil {
		p.Tree = map[string]*routerTree{}
	}

	branch, ok := p.Tree[row]
	if !ok {
		branch = &routerTree{}
		p.Tree[row] = branch
	}
	branch.PushHandler(method, parts[1:], handler, keys)
}

func (p *routerTree) getHandler(method int, parts []string, values []string) (func(srv *Server), []string, []string, error) {
	if len(parts) == 0 {
		if p.Methods == nil {
			return nil, nil, nil, errors.New("This path is not supported")
		}
		if method == Methods["*"] {
			for _, m := range p.Methods {
				return m.Handle, m.Keys, values, nil
			}
		} else {
			m, ok := p.Methods[method]
			if ok {
				return m.Handle, m.Keys, values, nil
			} else {
				m, ok := p.Methods[Methods["*"]]
				if ok {
					return m.Handle, m.Keys, values, nil
				}
			}
		}
		supported := []string{}
		thisMethod := ""
		for methodID := range p.Methods {
			for k, v := range Methods {
				if v == method {
					thisMethod = k
				}
				if v == methodID {
					supported = append(supported, k)
				}
			}
		}
		errorText := thisMethod + " method is not supported for this path, use " + strings.Join(supported, " or ")
		return nil, nil, nil, errors.New(errorText)
	}
	row := parts[0]
	if row == "" {
		return p.getHandler(method, parts[1:], values)
	}
	if p.Tree == nil {
		p.Tree = map[string]*routerTree{}
	}
	branch, ok := p.Tree[row]
	if ok {
		return branch.getHandler(method, parts[1:], values)
	}
	branch, ok = p.Tree[":"]
	if ok {
		values = append(values, row)
		return branch.getHandler(method, parts[1:], values)
	}
	return nil, nil, nil, errors.New("This path is not supported")
}

// public methods here

func (p *routerTree) Handle(method int, path string, handler func(srv *Server)) {
	parts := strings.Split(path, "/")
	p.PushHandler(method, parts, handler, []string{})
}

func (p *routerTree) Route(method int, path string) (func(srv *Server), map[string]string, error) {
	parts := strings.Split(path, "/")
	cb, keys, values, err := p.getHandler(method, parts, []string{})
	params := map[string]string{}
	for n, key := range keys {
		if len(values) > n {
			params[key] = values[n]
		}
	}
	return cb, params, err
}
