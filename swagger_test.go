package httpSwagger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "github.com/swaggo/http-swagger/testdata/docs"
)

func TestWrapHandler(t *testing.T) {
	router := http.NewServeMux()

	router.Handle("/", WrapHandler)

	w1 := performRequest("GET", "/index.html", router)
	assert.Equal(t, 200, w1.Code)

	w2 := performRequest("GET", "/doc.json", router)
	assert.Equal(t, 200, w2.Code)
	assert.Equal(t, "application/json; charset=utf-8", w2.Header().Get("content-type"))

	w3 := performRequest("GET", "/favicon-16x16.png", router)
	assert.Equal(t, 200, w3.Code)

	w4 := performRequest("GET", "/notfound", router)
	assert.Equal(t, 404, w4.Code)

	w5 := performRequest("GET", "/", router)
	assert.Equal(t, 301, w5.Code)
}

func performRequest(method, target string, h http.Handler) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)
	return w
}
