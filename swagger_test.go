package httpSwagger

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	tests := []struct {
		RootFolder   string
		InstanceName string
	}{
		{
			RootFolder:   "/",
			InstanceName: "default",
		},

		{
			RootFolder:   "/swagger/",
			InstanceName: "swagger",
		},
		{
			RootFolder:   "/custom/",
			InstanceName: "custom",
		},
	}
	for _, test := range tests {
		router := http.NewServeMux()
		router.Handle(test.RootFolder, Handler(
			DocExpansion("none"),
			DomID("swagger-ui"),
			InstanceName(test.InstanceName),
		))

		w1 := performRequest(http.MethodGet, test.RootFolder+"index.html", router)
		assert.Equal(t, http.StatusOK, w1.Code)
		assert.Equal(t, w1.Header()["Content-Type"][0], "text/html; charset=utf-8")

		assert.Equal(t, http.StatusInternalServerError, performRequest(http.MethodGet, test.RootFolder+"doc.json", router).Code)

		doc := &mockedSwag{}
		swag.Register(test.InstanceName, doc)
		w2 := performRequest(http.MethodGet, test.RootFolder+"doc.json", router)
		assert.Equal(t, http.StatusOK, w2.Code)
		assert.Equal(t, "application/json; charset=utf-8", w2.Header().Get("content-type"))

		// Perform body rendering validation
		w2Body, err := io.ReadAll(w2.Body)
		assert.NoError(t, err)
		assert.Equal(t, doc.ReadDoc(), string(w2Body))

		w3 := performRequest(http.MethodGet, test.RootFolder+"favicon-16x16.png", router)
		assert.Equal(t, http.StatusOK, w3.Code)
		assert.Equal(t, w3.Header()["Content-Type"][0], "image/png")

		w4 := performRequest(http.MethodGet, test.RootFolder+"swagger-ui.css", router)
		assert.Equal(t, http.StatusOK, w4.Code)
		assert.Equal(t, w4.Header()["Content-Type"][0], "text/css; charset=utf-8")

		w5 := performRequest(http.MethodGet, test.RootFolder+"swagger-ui-bundle.js", router)
		assert.Equal(t, http.StatusOK, w5.Code)
		assert.Equal(t, w5.Header()["Content-Type"][0], "application/javascript")

		w6 := performRequest(http.MethodGet, test.RootFolder+"oauth2-redirect.html?state=0&session_state=1&code=2", router)
		assert.Equal(t, http.StatusOK, w6.Code)
		assert.Equal(t, w6.Header()["Content-Type"][0], "text/html; charset=utf-8")

		assert.Equal(t, http.StatusNotFound, performRequest(http.MethodGet, test.RootFolder+"notfound", router).Code)

		assert.Equal(t, 301, performRequest(http.MethodGet, test.RootFolder, router).Code)

		assert.Equal(t, http.StatusMethodNotAllowed, performRequest(http.MethodPost, test.RootFolder+"index.html", router).Code)

		assert.Equal(t, http.StatusMethodNotAllowed, performRequest(http.MethodPut, test.RootFolder+"index.html", router).Code)
	}
}

func performRequest(method, target string, h http.Handler) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, target, nil)
	w := httptest.NewRecorder()

	h.ServeHTTP(w, r)

	return w
}

func TestURL(t *testing.T) {
	var cfg *Config

	expected := "https://github.com/swaggo/http-swagger"
	cfg = newConfig(URL(expected))
	assert.Equal(t, expected, cfg.URL)
}

func TestDeepLinking(t *testing.T) {
	var cfg Config
	// Default value
	assert.Equal(t, false, cfg.DeepLinking)

	// Set true
	DeepLinking(true)(&cfg)
	assert.Equal(t, true, cfg.DeepLinking)

	// Set false
	DeepLinking(false)(&cfg)
	assert.Equal(t, false, cfg.DeepLinking)
}

func TestLayout(t *testing.T) {
	var cfg *Config

	cfg = newConfig()
	assert.Equal(t, StandaloneLayout, cfg.Layout)

	cfg = newConfig(Layout(BaseLayout))
	assert.Equal(t, BaseLayout, cfg.Layout)
}

func TestDocExpansion(t *testing.T) {
	var cfg *Config

	expected := "https://github.com/swaggo/docs"
	cfg = newConfig(DocExpansion(expected))
	assert.Equal(t, expected, cfg.DocExpansion)
}

func TestDomID(t *testing.T) {
	var cfg *Config

	expected := "swagger-ui"
	cfg = newConfig(DomID(expected))
	assert.Equal(t, expected, cfg.DomID)
}

func TestInstanceName(t *testing.T) {
	var cfg *Config

	expected := swag.Name
	cfg = newConfig(InstanceName(expected))
	assert.Equal(t, expected, cfg.InstanceName)

	expected = "custom_name"
	cfg = newConfig(InstanceName(expected))
	assert.Equal(t, expected, cfg.InstanceName)

	cfg = newConfig(InstanceName(""))
	assert.Equal(t, swag.Name, cfg.InstanceName)
}

func TestPersistAuthorization(t *testing.T) {
	var cfg *Config

	// Default value
	cfg = newConfig()
	assert.Equal(t, false, cfg.PersistAuthorization)

	// Set true
	cfg = newConfig(PersistAuthorization(true))
	assert.Equal(t, true, cfg.PersistAuthorization)

	// Set false
	cfg = newConfig(PersistAuthorization(false))
	assert.Equal(t, false, cfg.PersistAuthorization)
}

func TestConfigURL(t *testing.T) {

	type fixture struct {
		desc  string
		cfgfn func(c *Config)
		exp   *Config
	}

	fixtures := []fixture{
		{
			desc: "configure URL",
			exp: &Config{
				URL: "https://example.org/doc.json",
			},
			cfgfn: URL("https://example.org/doc.json"),
		},
		{
			desc: "configure DeepLinking",
			exp: &Config{
				DeepLinking: true,
			},
			cfgfn: DeepLinking(true),
		},
		{
			desc: "configure DocExpansion",
			exp: &Config{
				DocExpansion: "none",
			},
			cfgfn: DocExpansion("none"),
		},
		{
			desc: "configure DomID",
			exp: &Config{
				DomID: "#swagger-ui",
			},
			cfgfn: DomID("#swagger-ui"),
		},
		{
			desc: "configure Plugins",
			exp: &Config{
				Plugins: []template.JS{
					"SomePlugin",
					"AnotherPlugin",
				},
			},
			cfgfn: Plugins([]string{
				"SomePlugin",
				"AnotherPlugin",
			}),
		},
		{
			desc: "configure UIConfig",
			exp: &Config{
				UIConfig: map[template.JS]template.JS{
					"urls": `["https://example.org/doc1.json","https://example.org/doc1.json"],`,
				},
			},
			cfgfn: UIConfig(map[string]string{
				"urls": `["https://example.org/doc1.json","https://example.org/doc1.json"],`,
			}),
		},
		{
			desc: "configure BeforeScript",
			exp: &Config{
				BeforeScript: `const SomePlugin = (system) => ({
    // Some plugin
  });`,
			},
			cfgfn: BeforeScript(`const SomePlugin = (system) => ({
    // Some plugin
  });`),
		},
		{
			desc: "configure AfterScript",
			exp: &Config{
				AfterScript: `const SomePlugin = (system) => ({
    // Some plugin
  });`,
			},
			cfgfn: AfterScript(`const SomePlugin = (system) => ({
    // Some plugin
  });`),
		},
	}

	for _, fix := range fixtures {
		t.Run(fix.desc, func(t *testing.T) {
			cfg := &Config{}
			fix.cfgfn(cfg)
			assert.Equal(t, cfg, fix.exp)
		})
	}
}

func TestUIConfigOptions(t *testing.T) {

	type fixture struct {
		desc string
		cfg  *Config
		exp  string
	}

	fixtures := []fixture{
		{
			desc: "default configuration",
			cfg: &Config{
				URL:                      "doc.json",
				DeepLinking:              true,
				DocExpansion:             "list",
				DomID:                    "swagger-ui",
				PersistAuthorization:     false,
				Layout:                   StandaloneLayout,
				DefaultModelsExpandDepth: ShowModel,
			},
			exp: `
window.onload = function() {
  // Build a system
  const ui = SwaggerUIBundle({
    url: "doc.json",
    deepLinking: true,
    docExpansion: "list",
    dom_id: "#swagger-ui",
    persistAuthorization: false,
    validatorUrl: null,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout",
    defaultModelsExpandDepth: 1,
    showExtensions: false
  })

  window.ui = ui
}
`,
		},
		{
			desc: "script configuration",
			cfg: &Config{
				URL:                  "swagger.json",
				DeepLinking:          false,
				PersistAuthorization: true,
				DocExpansion:         "none",
				DomID:                "swagger-ui-id",
				Layout:               StandaloneLayout,
				BeforeScript: `const SomePlugin = (system) => ({
    // Some plugin
  });
`,
				AfterScript: `const someOtherCode = function(){
    // Do something
  };
  someOtherCode();`,
				Plugins: []template.JS{
					"SomePlugin",
					"AnotherPlugin",
				},
				UIConfig: map[template.JS]template.JS{
					"showExtensions":        "true",
					"onComplete":            `() => { window.ui.setBasePath('v3'); }`,
					"defaultModelRendering": `"model"`,
				},
				DefaultModelsExpandDepth: HideModel,
			},
			exp: `
window.onload = function() {
  const SomePlugin = (system) =&gt; ({
    // Some plugin
  });

  // Build a system
  const ui = SwaggerUIBundle({
    url: "swagger.json",
    deepLinking: false,
    docExpansion: "none",
    dom_id: "#swagger-ui-id",
    persistAuthorization: true,
    validatorUrl: null,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl,
      SomePlugin,
      AnotherPlugin
    ],
    defaultModelRendering: &#34;model&#34;,
    onComplete: () =&gt; { window.ui.setBasePath(&#39;v3&#39;); },
    showExtensions: true,
    layout: "StandaloneLayout",
    defaultModelsExpandDepth: -1,
    showExtensions: false
  })

  window.ui = ui
  const someOtherCode = function(){
    // Do something
  };
  someOtherCode();
}
`,
		},
	}

	for _, fix := range fixtures {
		t.Run(fix.desc, func(t *testing.T) {
			tmpl := template.New("swagger_index.js")
			indexJs, err := tmpl.Parse(indexJsTempl)
			if err != nil {
				t.Fatal(err)
			}

			buf := bytes.NewBuffer(nil)
			if err := indexJs.Execute(buf, fix.cfg); err != nil {
				t.Fatal(err)
			}

			exp := fix.exp

			// Compare line by line
			explns := strings.Split(exp, "\n")
			buflns := strings.Split(buf.String(), "\n")

			explen, buflen := len(explns), len(buflns)
			if explen != buflen {
				t.Errorf("expected %d lines, but got %d", explen, buflen)
			}

			printContext := func(idx int) {
				lines := 3

				firstIdx := idx - lines
				if firstIdx < 0 {
					firstIdx = 0
				}
				lastIdx := idx + lines
				if lastIdx > explen {
					lastIdx = explen
				}
				if lastIdx > buflen {
					lastIdx = buflen
				}
				t.Logf("expected:\n")
				for i := firstIdx; i < lastIdx; i++ {
					t.Logf(explns[i])
				}
				t.Logf("got:\n")
				for i := firstIdx; i < lastIdx; i++ {
					t.Logf(buflns[i])
				}
			}

			for i, expln := range explns {
				if i >= buflen {
					printContext(i)
					t.Fatalf(`first unequal line: expected "%s" but got EOF`, expln)
				}
				bufln := buflns[i]
				if bufln != expln {
					printContext(i)
					t.Fatalf(`first unequal line: expected "%s" but got "%s"`, expln, bufln)
				}
			}

			if buflen > explen {
				printContext(explen - 1)
				t.Fatalf(`first unequal line: expected EOF, but got "%s"`, buflns[explen])
			}
		})
	}
}

func TestDefaultModelsExpandDepth(t *testing.T) {
	cfg := newConfig()
	// Default value
	assert.Equal(t, ShowModel, cfg.DefaultModelsExpandDepth)

	// Set hide
	DefaultModelsExpandDepth(HideModel)(cfg)
	assert.Equal(t, HideModel, cfg.DefaultModelsExpandDepth)

	// Set show
	DefaultModelsExpandDepth(ShowModel)(cfg)
	assert.Equal(t, ShowModel, cfg.DefaultModelsExpandDepth)
}

func TestShowExtensions(t *testing.T) {
	var cfg *Config

	cfg = newConfig()
	assert.False(t, cfg.ShowExtensions)

	cfg = newConfig(ShowExtensions(true))
	assert.True(t, cfg.ShowExtensions)

	cfg = newConfig(ShowExtensions(false))
	assert.False(t, cfg.ShowExtensions)
}
