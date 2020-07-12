package webapi

import (
	"log"
	"net/http"
	"net/url"
	"testing"
)

func noop(interface{}) {}

type mockResponseWriter struct {
	header http.Header
	data   []byte
	status int
}

func (w *mockResponseWriter) Status() int {
	return w.status
}

func (w *mockResponseWriter) Header() http.Header {
	return w.header
}

func (w *mockResponseWriter) Write(data []byte) (int, error) {
	w.data = append(w.data, data...)
	return len(data), nil
}

func (w *mockResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *mockResponseWriter) Data() []byte {
	return w.data
}

func (w *mockResponseWriter) String() string {
	return string(w.data)
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		header: http.Header{},
		data:   []byte{},
		status: http.StatusOK,
	}
}

func newMockRequest(options ...func(*ParsedRequest)) *ParsedRequest {
	ret := &ParsedRequest{
		request:    &http.Request{},
		pathParams: map[string]string{},
	}
	for _, opt := range options {
		opt(ret)
	}
	return ret
}

func TestSingleHandler(t *testing.T) {
	log.Println("Test Single Handler")
	mockData := "abcd"
	hnd := NewHandler(func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
		w.Write([]byte(mockData))
		return next()
	})

	w := newMockResponseWriter()
	hnd.HandleAll(w, newMockRequest())
	have, want := w.String(), mockData
	if have != want {
		t.Errorf("Handler did not return correct data, have %s want %s", have, want)
	}
}

func TestMultiHandler(t *testing.T) {
	log.Println("Test Multi Handler")
	mockData1 := "abcd"
	mockData2 := "efgh"
	hnd := NewHandler(func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
		w.Write([]byte(mockData1))
		return next()
	}, func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
		w.Write([]byte(mockData2))
		return next()
	})

	w := newMockResponseWriter()
	hnd.HandleAll(w, newMockRequest())
	have, want := w.String(), mockData1+mockData2
	if have != want {
		t.Errorf("Handler did not return correct data, have %s want %s", have, want)
	}
}

func TestErrorHandler(t *testing.T) {
	log.Println("Test Error Handler")
	mockData := "abcd"
	badRequestStr := "Bad Request"
	badRequest := NewErrorHandler(http.StatusBadRequest, badRequestStr)

	hnd := NewHandler(func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
		if r.request.Method == http.MethodPost {
			return badRequest
		}
		w.Write([]byte(mockData))
		return next()
	})

	// Good case
	{
		w := newMockResponseWriter()
		hnd.HandleAll(w, newMockRequest())
		have, want := w.status, http.StatusOK
		if have != want {
			t.Errorf("Good case: Handler did not return correct response code, have %d want %d", have, want)
		}
	}

	// Error Case
	{
		w := newMockResponseWriter()
		r := newMockRequest(func(req *ParsedRequest) {
			req.request.Method = http.MethodPost
		})
		hnd.HandleAll(w, r)
		have, want := w.status, http.StatusBadRequest
		if have != want {
			t.Errorf("Error case: Handler did not return correct response code, have %d want %d", have, want)
		}
	}
}

func TestRouter(t *testing.T) {
	log.Println("Test Router")
	fallback404 := NewErrorHandler(http.StatusNotFound, "404 not found")
	helloMock := "hello"
	byeMock := "bye"
	paramMock := "xyz"

	helloHandler := NewHandler(func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
		w.Write([]byte(helloMock))
		return next()
	})

	byeHandler := NewHandler(func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
		w.Write([]byte(byeMock))
		return next()
	})

	paramHandler := NewHandler(func(w http.ResponseWriter, r *ParsedRequest, next func() Handler) Handler {
		w.Write([]byte(r.pathParams["param"]))
		return next()
	})

	router := NewRouter(fallback404)
	router.Handle(http.MethodGet, "/hello", helloHandler)
	router.Handle(http.MethodGet, "/bye", byeHandler)
	router.Handle(http.MethodGet, "/param/:param", paramHandler)

	// Hello
	{
		w := newMockResponseWriter()
		r := newMockRequest(func(req *ParsedRequest) {
			req.request.Method = http.MethodGet
			req.request.URL = &url.URL{
				Path: "/hello",
			}
		})

		router.ServeHTTP(w, r.request)
		have, want := w.String(), helloMock
		if have != want {
			t.Errorf("Hello: Router did not return correct response code, have %s want %s", have, want)
		}
	}

	// Bye
	{
		w := newMockResponseWriter()
		r := newMockRequest(func(req *ParsedRequest) {
			req.request.Method = http.MethodGet
			req.request.URL = &url.URL{
				Path: "/bye",
			}
		})

		router.ServeHTTP(w, r.request)
		have, want := w.String(), byeMock
		if have != want {
			t.Errorf("Bye: Router did not return correct response code, have %s want %s", have, want)
		}
	}

	// Param
	{
		w := newMockResponseWriter()
		r := newMockRequest(func(req *ParsedRequest) {
			req.request.Method = http.MethodGet
			req.request.URL = &url.URL{
				Path: "/param/" + paramMock,
			}
		})

		router.ServeHTTP(w, r.request)
		have, want := w.String(), paramMock
		if have != want {
			t.Errorf("Param: Router did not return correct response code, have %s want %s", have, want)
		}
	}
}
