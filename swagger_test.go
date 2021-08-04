package httpSwagger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/swaggo/swag"
)

type mockedSwag struct{}

func (s *mockedSwag) ReadDoc() string {
	return `{
    "swagger": "2.0",
    "info": {
        "description": "This is a sample server Petstore server.",
        "title": "Swagger Example API",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "url": "http://www.swagger.io/support",
            "email": "support@swagger.io"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "1.0"
    },
    "host": "petstore.swagger.io",
    "basePath": "/v2",
    "paths": {}
}`
}

func TestWrapHandler(t *testing.T) {
	router := chi.NewRouter()

	router.Get("/*", Handler(DocExpansion("none"), DomID("#swagger-ui")))

	w1 := performRequest("GET", "/index.html", router)
	assert.Equal(t, 200, w1.Code)

	w2 := performRequest("GET", "/mockedSwag.json", router)
	assert.Equal(t, 500, w2.Code)

	swag.Register(swag.Name, &mockedSwag{})
	w2 = performRequest("GET", "/mockedSwag.json", router)
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
