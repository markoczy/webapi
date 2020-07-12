package webapi

import (
	"net/http"

	pathToRegexp "github.com/soongo/path-to-regexp"
)

// HandlerFunc is the definition of any Handler of this Web API framework
// the function `next()` returns the next handler to call, it is usually
// returned in any non-error case, otherwise an error handler should be
// returned instead.
type HandlerFunc func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler

// ParsedRequest is an enriched version of the native http.Request which
// contains the parsed PathParams.
type ParsedRequest struct {
	pathParams map[string]string
	request    *http.Request
}

// Handler is an interface that defines any request handler of this Framework.
// The handlers are used by the router to process HTTP Requests.
type Handler interface {
	// Handle calls the current handler.
	Handle(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler
	// HandleAll calls the current handler and all following handlers.
	HandleAll(w http.ResponseWriter, r *ParsedRequest)
	// Next returns the following handler.
	Next() Handler
}

// defaultHandler is the default implementation of the Handler interface.
type defaultHandler struct {
	fn   HandlerFunc
	next Handler
}

func (hnd *defaultHandler) Handle(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
	return hnd.fn(w, r, next)
}

func (hnd *defaultHandler) Next() Handler {
	return hnd.next
}

func (hnd *defaultHandler) HandleAll(w http.ResponseWriter, r *ParsedRequest) {
	next := hnd.Handle(w, r, hnd.Next)
	for next != nil {
		next = next.Handle(w, r, next.Next)
	}
}

// NewHandler creates a HTTP handler from default implementation. It takes
// one or more Handler Functions that act as middleware for the handler.
func NewHandler(firstHandler HandlerFunc, optionalHandlers ...HandlerFunc) Handler {
	first := &defaultHandler{
		fn:   firstHandler,
		next: nil,
	}
	cur := first

	for _, fn := range optionalHandlers {
		next := &defaultHandler{
			fn:   fn,
			next: nil,
		}
		cur.next = next
		cur = next
	}
	return first
}

// NewErrorHandler creates an error handler according to the net/http default
// implementation
func NewErrorHandler(code int, err string) Handler {
	return NewHandler(
		func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
			http.Error(w, err, code)
			return next()
		},
	)
}

// NewNativeHandler creates a handler from a net/http Handler.
func NewNativeHandler(handler http.Handler) Handler {
	return NewHandler(
		func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
			handler.ServeHTTP(w, r.request)
			return next()
		},
	)
}

// routeConig is an internal type that defines a Route Configuration.
type routeConfig struct {
	match   func(string) (*pathToRegexp.MatchResult, error)
	handler Handler
}

func (cfg *routeConfig) Match(route string) (bool, map[string]string) {
	res, err := cfg.match(route)
	if err != nil || res == nil {
		return false, nil
	}

	ret := make(map[string]string)
	for k, v := range res.Params {
		ret[k.(string)] = v.(string)
	}
	return true, ret
}

func (cfg *routeConfig) Handle(w http.ResponseWriter, r *ParsedRequest) {
	cfg.handler.HandleAll(w, r)
}

// Router is a type used to route HTTP Requests to a specific handler. The
// router is also capable of parsing path params if the routeConfig's regex
// supports named capture groups.
type Router struct {
	handlers map[string][]*routeConfig
	fallback Handler
}

// Handle registers a handler for a given request type.
func (router *Router) Handle(method, matcher string, handler Handler) {
	match := pathToRegexp.MustMatch(matcher, &pathToRegexp.Options{Decode: func(str string, token interface{}) (string, error) {
		return pathToRegexp.DecodeURIComponent(str)
	}}) // todo panics

	router.handlers[method] = append(router.handlers[method], &routeConfig{
		match: match,
	})
}

// ServeHTTP implements the net/http Handler interface so that the Router
// can be used as native net/http Handler.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if router.handlers[r.Method] == nil {
		router.fallback.HandleAll(w, &ParsedRequest{
			request: r,
		})
		return
	}
	for _, cfg := range router.handlers[r.Method] {
		matches, pathParams := cfg.Match(r.URL.Path)
		if matches {
			parsed := &ParsedRequest{
				pathParams: pathParams,
				request:    r,
			}
			cfg.handler.HandleAll(w, parsed)
			return
		}
	}
	router.fallback.HandleAll(w, &ParsedRequest{
		request: r,
	})
}

// NewRouter creates a new Router.
func NewRouter(fallback Handler) *Router {
	return &Router{
		handlers: make(map[string][]*routeConfig),
		fallback: fallback,
	}
}
