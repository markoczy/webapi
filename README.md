# Web API

Golang Web API Framework

## Purpose

The Project "webapi" is a Go Module with the purpose of creating feature rich
web apis with little effort. It extends the functionality of the `net/http`
package and adds a Router that implements the `net/http.Handler` interface.
The Router can be used in the context of a `net/http.Server` to route 
HTTP(S) requests to a specific handler by using regex matchers that are defined
in the powerful, yet easy to understand 
[Path-To-Regexp](https://github.com/soongo/path-to-regexp) Syntax. With
Path-To-Regexp, the Router is also capable of parsing Path Parameters that are
defined in the route's matcher string, (see example or test cases). The path
parameters will be passed as a `map[string]string` to any of the given request
handlers. The framework defines new Handler functions that are more 
flexible than `net/http.HandlerFunc`'s and allow the programmer to create
chained handlers with customizable alternative paths (e.g. Error Handlers).
See examples for more info...

## Usage

Here's a minimalistic example usage:

```go
// Create Fallback handler (for unmatched requests)
fallback404 := webapi.NewErrorHandler(http.StatusNotFound, "404 not found")
// Create Router
router := webapi.NewRouter(fallback404)
// Create Handler by chaining Handler functions
handler := webapi.NewHandler(
    // Handler function 1
    func(w http.ResponseWriter, r *webapi.ParsedRequest, next func() webapi.Handler) webapi.Handler {
        // (optional) Handle errors
        if r.PathParams["name"] == "abc" {
            return webapi.NewErrorHandler(http.StatusBadRequest, "Your name is not 'abc' ;-)") 
        }
        // Write data (or do other stuff)
        w.Write([]byte("Hello, " + r.PathParams["name"]))
        return next()
    }, 
    // Handler function 2
    func(w http.ResponseWriter, r *webapi.ParsedRequest, next func() webapi.Handler) webapi.Handler {
        // do stuff for handler 2 (e.g. Fill Headers)
        return next()
    },
)
// Register Handler
router.Handle(http.MethodGet, "/hello/:name", handler)
// Serve HTTP
http.ListenAndServe(":7890", router)
//
// cURL Results:
// `curl localhost:7890/hello/John+Doe` returns Hello, John Doe
// `curl localhost:7890/hello/abc` returns Bad Request
```

> For Path matching syntax, check `github.com/soongo/path-to-regexp`, for
more examples, check the test cases. 

## Dependencies

- PathToRegexp: `github.com/soongo/path-to-regexp` (License: MIT)