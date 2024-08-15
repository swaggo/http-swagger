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
	"github.com/swaggo/swag/v2"
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

	hdr := `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Swagger UI</title>
  <link rel="stylesheet" type="text/css" href="./swagger-ui.css" >
  <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
  <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
  <style>
    html
    {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
    }
    *,
    *:before,
    *:after
    {
        box-sizing: inherit;
    }

    body {
      margin:0;
      background: #fafafa;
    }
  </style>
</head>

<body>

<svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0">
  <defs>
    <symbol viewBox="0 0 20 20" id="unlocked">
      <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V6h2v-.801C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8z"></path>
    </symbol>

    <symbol viewBox="0 0 20 20" id="locked">
      <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8zM12 8H8V5.199C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="close">
      <path d="M14.348 14.849c-.469.469-1.229.469-1.697 0L10 11.819l-2.651 3.029c-.469.469-1.229.469-1.697 0-.469-.469-.469-1.229 0-1.697l2.758-3.15-2.759-3.152c-.469-.469-.469-1.228 0-1.697.469-.469 1.228-.469 1.697 0L10 8.183l2.651-3.031c.469-.469 1.228-.469 1.697 0 .469.469.469 1.229 0 1.697l-2.758 3.152 2.758 3.15c.469.469.469 1.229 0 1.698z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow">
      <path d="M13.25 10L6.109 2.58c-.268-.27-.268-.707 0-.979.268-.27.701-.27.969 0l7.83 7.908c.268.271.268.709 0 .979l-7.83 7.908c-.268.271-.701.27-.969 0-.268-.269-.268-.707 0-.979L13.25 10z"/>
    </symbol>

    <symbol viewBox="0 0 20 20" id="large-arrow-down">
      <path d="M17.418 6.109c.272-.268.709-.268.979 0s.271.701 0 .969l-7.908 7.83c-.27.268-.707.268-.979 0l-7.908-7.83c-.27-.268-.27-.701 0-.969.271-.268.709-.268.979 0L10 13.25l7.418-7.141z"/>
    </symbol>

    <symbol viewBox="0 0 24 24" id="jump-to">
      <path d="M19 7v4H5.83l3.58-3.59L8 6l-6 6 6 6 1.41-1.41L5.83 13H21V7z"/>
    </symbol>

    <symbol viewBox="0 0 24 24" id="expand">
      <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/>
    </symbol>
  </defs>
</svg>

<div id="swagger-ui"></div>

<script src="./swagger-ui-bundle.js"> </script>
<script src="./swagger-ui-standalone-preset.js"> </script>
<script>
`
	ftr := `
</script>
</body>

</html>
`

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
			exp: `window.onload = function() {
  
  const ui = SwaggerUIBundle({
    url: "doc.json",
    deepLinking:  true ,
    docExpansion: "list",
    dom_id: "#swagger-ui",
    persistAuthorization:  false ,
    validatorUrl: null,
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    plugins: [
      SwaggerUIBundle.plugins.DownloadUrl
    ],
    layout: "StandaloneLayout",
    defaultModelsExpandDepth:  1 
  })

  window.ui = ui
}`,
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
			exp: `window.onload = function() {
  const SomePlugin = (system) => ({
    // Some plugin
  });

  
  const ui = SwaggerUIBundle({
    url: "swagger.json",
    deepLinking:  false ,
    docExpansion: "none",
    dom_id: "#swagger-ui-id",
    persistAuthorization:  true ,
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
    defaultModelRendering: "model",
    onComplete: () => { window.ui.setBasePath('v3'); },
    showExtensions: true,
    layout: "StandaloneLayout",
    defaultModelsExpandDepth:  -1 
  })

  window.ui = ui
  const someOtherCode = function(){
    // Do something
  };
  someOtherCode();
}`,
		},
	}

	for _, fix := range fixtures {
		t.Run(fix.desc, func(t *testing.T) {
			tmpl := template.New("swagger_index.html")
			index, err := tmpl.Parse(indexTempl)
			if err != nil {
				t.Fatal(err)
			}

			buf := bytes.NewBuffer(nil)
			if err := index.Execute(buf, fix.cfg); err != nil {
				t.Fatal(err)
			}

			exp := hdr + fix.exp + ftr

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
