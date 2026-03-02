package handler

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	//go:embed docs/openapi.yaml
	openAPISpec []byte
)

const redocHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>MediaHub API Docs</title>
    <style>
      body { margin: 0; padding: 0; }
    </style>
  </head>
  <body>
    <redoc spec-url="/docs/openapi.yaml"></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </body>
</html>`

const swaggerHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
    <title>MediaHub API Docs (Swagger UI)</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
    <style>
      body { margin: 0; padding: 0; }
      #swagger-ui { max-width: 1200px; margin: 0 auto; }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
      window.onload = function () {
        window.ui = SwaggerUIBundle({
          url: "/docs/openapi.yaml",
          dom_id: "#swagger-ui",
          deepLinking: true,
          presets: [SwaggerUIBundle.presets.apis],
          layout: "BaseLayout"
        });
      };
    </script>
  </body>
</html>`

// ServeOpenAPISpec serves the embedded OpenAPI YAML.
func ServeOpenAPISpec(c *gin.Context) {
	c.Data(http.StatusOK, "application/yaml; charset=utf-8", openAPISpec)
}

// ServeReDoc renders a ReDoc page bound to /docs/openapi.yaml.
func ServeReDoc(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(redocHTML))
}

// ServeSwaggerUI renders a Swagger UI page bound to /docs/openapi.yaml.
func ServeSwaggerUI(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerHTML))
}
