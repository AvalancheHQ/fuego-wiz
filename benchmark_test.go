package fuego

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Benchmark request handling
func BenchmarkSimpleGetRequest(b *testing.B) {
	s := NewServer(WithoutLogger())
	Get(s, "/hello", func(c ContextNoBody) (string, error) {
		return "Hello, World!", nil
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark JSON serialization
func BenchmarkJSONResponse(b *testing.B) {
	type Response struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
		Data    string `json:"data"`
	}

	s := NewServer(WithoutLogger())
	Get(s, "/json", func(c ContextNoBody) (Response, error) {
		return Response{
			Message: "Success",
			Status:  200,
			Data:    "Some sample data",
		}, nil
	})

	req := httptest.NewRequest("GET", "/json", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark POST with body deserialization
func BenchmarkPostWithBody(b *testing.B) {
	type Request struct {
		Name  string `json:"name" validate:"required"`
		Email string `json:"email" validate:"required,email"`
	}

	type Response struct {
		Message string `json:"message"`
	}

	s := NewServer(WithoutLogger())
	Post(s, "/user", func(c ContextWithBody[Request]) (Response, error) {
		body, err := c.Body()
		if err != nil {
			return Response{}, err
		}
		return Response{Message: "Hello, " + body.Name}, nil
	})

	jsonBody := []byte(`{"name":"John Doe","email":"john@example.com"}`)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/user", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark path parameters
func BenchmarkPathParams(b *testing.B) {
	s := NewServer(WithoutLogger())
	Get(s, "/user/{id}", func(c ContextNoBody) (string, error) {
		id := c.PathParam("id")
		return "User ID: " + id, nil
	})

	req := httptest.NewRequest("GET", "/user/12345", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark query parameters
func BenchmarkQueryParams(b *testing.B) {
	s := NewServer(WithoutLogger())
	Get(s, "/search", func(c ContextNoBody) (string, error) {
		query := c.QueryParam("q")
		return "Search: " + query, nil
	})

	req := httptest.NewRequest("GET", "/search?q=test", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark with middleware
func BenchmarkWithMiddleware(b *testing.B) {
	s := NewServer(WithoutLogger())

	// Add a simple middleware
	Use(s, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Custom-Header", "value")
			next.ServeHTTP(w, r)
		})
	})

	Get(s, "/hello", func(c ContextNoBody) (string, error) {
		return "Hello, World!", nil
	})

	req := httptest.NewRequest("GET", "/hello", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark error handling
func BenchmarkErrorHandling(b *testing.B) {
	s := NewServer(WithoutLogger())
	Get(s, "/error", func(c ContextNoBody) (string, error) {
		return "", HTTPError{
			Err:    http.ErrAbortHandler,
			Status: http.StatusBadRequest,
		}
	})

	req := httptest.NewRequest("GET", "/error", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark context operations
func BenchmarkContextOperations(b *testing.B) {
	s := NewServer(WithoutLogger())
	Get(s, "/context", func(c ContextNoBody) (string, error) {
		ctx := c.Context()
		_ = ctx.Value("test")
		req := c.Request()
		_ = req.Header.Get("User-Agent")
		return "OK", nil
	})

	req := httptest.NewRequest("GET", "/context", nil)
	req.Header.Set("User-Agent", "test-agent")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Types for transformation benchmark
type transformRequest struct {
	Name string `json:"name"`
}

func (r *transformRequest) InTransform(ctx context.Context) error {
	r.Name = "Transformed: " + r.Name
	return nil
}

type transformResponse struct {
	Message string `json:"message"`
}

func (r *transformResponse) OutTransform(ctx context.Context) error {
	r.Message = "Transformed: " + r.Message
	return nil
}

// Benchmark with transformation
func BenchmarkWithTransformation(b *testing.B) {
	s := NewServer(WithoutLogger())
	Post(s, "/transform", func(c ContextWithBody[transformRequest]) (transformResponse, error) {
		body, err := c.Body()
		if err != nil {
			return transformResponse{}, err
		}
		return transformResponse{Message: body.Name}, nil
	})

	jsonBody := []byte(`{"name":"Test"}`)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/transform", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Type for rendering benchmark
type benchmarkRenderer struct{}

func (t benchmarkRenderer) Render(w io.Writer) error {
	w.Write([]byte("<html><body>Hello World</body></html>"))
	return nil
}

// Benchmark rendering
func BenchmarkRendering(b *testing.B) {
	s := NewServer(WithoutLogger())
	Get(s, "/render", func(c ContextNoBody) (Renderer, error) {
		return benchmarkRenderer{}, nil
	})

	req := httptest.NewRequest("GET", "/render", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark multiple routes
func BenchmarkMultipleRoutes(b *testing.B) {
	s := NewServer(WithoutLogger())

	// Register multiple routes
	Get(s, "/route1", func(c ContextNoBody) (string, error) {
		return "Route 1", nil
	})
	Get(s, "/route2", func(c ContextNoBody) (string, error) {
		return "Route 2", nil
	})
	Get(s, "/route3", func(c ContextNoBody) (string, error) {
		return "Route 3", nil
	})
	Post(s, "/route4", func(c ContextNoBody) (string, error) {
		return "Route 4", nil
	})

	routes := []string{"/route1", "/route2", "/route3"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		routeIndex := i % len(routes)
		req := httptest.NewRequest("GET", routes[routeIndex], nil)
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}

// Benchmark header operations
func BenchmarkHeaderOperations(b *testing.B) {
	s := NewServer(WithoutLogger())
	Get(s, "/headers", func(c ContextNoBody) (string, error) {
		c.Response().Header().Set("X-Custom-1", "value1")
		c.Response().Header().Set("X-Custom-2", "value2")
		c.Response().Header().Set("X-Custom-3", "value3")
		return "OK", nil
	})

	req := httptest.NewRequest("GET", "/headers", nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		s.Mux.ServeHTTP(w, req)
	}
}
