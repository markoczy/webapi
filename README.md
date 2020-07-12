# Web API

Golang Web API Framework

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
router.Handle(http.MethodGet, "/hello/:name", helloHandler)
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