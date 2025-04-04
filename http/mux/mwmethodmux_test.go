package http

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const helloWorld = "Hello, World!"

func hwHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, helloWorld)
	}
}

func TestMWMethodMux(t *testing.T) {
	mux := NewMWMethodMux(http.NewServeMux())

	mux.Handle("/foo", hwHandler(), "GET")
	mux.Handle("/bar", hwHandler())

	tests := []struct {
		method     string
		path       string
		statusCode int
		body       string
	}{
		{"GET", "/foo", 200, helloWorld},
		{"POST", "/foo", 405, ""},
		{"GET", "/bar", 200, helloWorld},
		{"POST", "/bar", 200, helloWorld},
		{"GET", "/baz", 404, ""},
	}

	for _, test := range tests {
		t.Run(strings.Join([]string{
			test.method, strings.Trim(test.path, "/"), fmt.Sprintf("%d", test.statusCode),
		}, "-"), func(t *testing.T) {
			req, err := http.NewRequest(test.method, test.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != test.statusCode {
				t.Errorf("expected status code %d, got %d", test.statusCode, w.Code)
			}

			if test.body != "" {
				if !bytes.Contains(w.Body.Bytes(), []byte(test.body)) {
					t.Errorf("expected body to contain '%s', got '%s'", test.body, w.Body.String())
				}
			}
		})

	}

	t.Run("already-registered-main", func(t *testing.T) {
		var panicString string

		func() {
			defer func() {
				if r := recover(); r != nil {
					panicString, _ = r.(string)
				}
			}()

			mux.Handle("/foo", hwHandler())
		}()

		if !strings.Contains(panicString, "multiple registrations") {
			t.Errorf("unexpected panic string: %v", panicString)
		}
	})

	t.Run("already-registered-method", func(t *testing.T) {
		var panicString string

		func() {
			defer func() {
				if r := recover(); r != nil {
					panicString, _ = r.(string)
				}
			}()

			mux.Handle("/foo", hwHandler(), "GET")
		}()

		if !strings.Contains(panicString, "multiple registrations") {
			t.Errorf("unexpected panic string: %v", panicString)
		}
	})

}
