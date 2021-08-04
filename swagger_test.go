package httpSwagger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"
)

func TestWrapHandler(t *testing.T) {
	router := chi.NewRouter()

	router.Get("/*", WrapHandler)

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

func TestHandler(t *testing.T) {
	router := chi.NewRouter()
	router.Get("/*", Handler(DocExpansion("none"), DomID("#swagger-ui")))

	w1 := performRequest("GET", "/index.html", router)
	assert.Equal(t, 200, w1.Code)
	w2 := performRequest("GET", "/doc.json", router)

	assert.Equal(t, 200, w2.Code)
	assert.Equal(t, "application/json; charset=utf-8", w2.Header().Get("content-type"))
}

func performRequest(method, target string, h http.Handler) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	return w
}

func TestURL(t *testing.T) {
	expected := "https://github.com/swaggo/http-swagger"
	cfg := Config{}
	configFunc := URL(expected)
	configFunc(&cfg)
	assert.Equal(t, expected, cfg.URL)
}

func TestDeepLinking(t *testing.T) {
	expected := true
	cfg := Config{}
	configFunc := DeepLinking(expected)
	configFunc(&cfg)
	assert.Equal(t, expected, cfg.DeepLinking)
}

func TestDocExpansion(t *testing.T) {
	expected := "https://github.com/swaggo/docs"
	cfg := Config{}
	configFunc := DocExpansion(expected)
	configFunc(&cfg)
	assert.Equal(t, expected, cfg.DocExpansion)
}

func TestDomID(t *testing.T) {
	expected := "#swagger-ui"
	cfg := Config{}
	configFunc := DomID(expected)
	configFunc(&cfg)
	assert.Equal(t, expected, cfg.DomID)
}
