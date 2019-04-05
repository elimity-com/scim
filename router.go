package scim

import (
	"net/http"
	"net/url"
	"strings"
)

type handle func(http.ResponseWriter, *http.Request, url.Values)

type router struct {
	tree *node
	err  handle
}

func newRouter(err handle) router {
	return router{
		tree: &node{
			segment: "/",
			param:   false,
			methods: make(map[string]handle),
		},
		err: err,
	}
}

func (r router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	req.ParseForm()
	node, _ := r.tree.traverse(strings.Split(req.URL.Path, "/")[1:], req.Form)
	if handler := node.methods[req.Method]; handler != nil {
		handler(w, req, req.Form)
	} else {
		r.err(w, req, req.Form)
	}
}

func (r router) Handle(method, path string, handler handle) {
	if path[0] != '/' {
		panic("path has to start with a /")
	}
	r.tree.add(method, path, handler)
}

type node struct {
	children []*node
	segment  string
	param    bool
	methods  map[string]handle
}

func (n *node) add(method, path string, handler handle) {
	segments := strings.Split(path, "/")[1:]
	count := len(segments)

	for {
		// update
		n, segment := n.traverse(segments, nil)
		if n.segment == segment && count == 1 {
			n.methods[method] = handler
			return
		}

		new := node{
			segment: segment,
			param:   false,
			methods: make(map[string]handle),
		}

		// check param
		if strings.HasPrefix(segment, "{") &&
			strings.HasSuffix(segment, "}") {

			new.param = true
		}

		// last component
		if count == 1 {
			new.methods[method] = handler
		}
		n.children = append(n.children, &new)

		count--
		if count == 0 {
			break
		}
	}
}

func (n *node) traverse(segments []string, params url.Values) (*node, string) {
	segment := segments[0]
	if len(n.children) <= 0 {
		return n, segment
	}

	for _, child := range n.children {
		if segment != child.segment && !child.param {
			continue
		}

		if child.param && params != nil {
			params.Add(child.segment[1:len(child.segment)-1], segment)
		}

		next := segments[1:]
		if len(next) > 0 {
			return child.traverse(next, params)
		}
		return child, segment
	}
	return n, segment
}
