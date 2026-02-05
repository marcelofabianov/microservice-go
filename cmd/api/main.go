package main

import (
	"go.uber.org/fx"

	"github.com/marcelofabianov/course/internal/di"
	"github.com/marcelofabianov/course/pkg/web"
)

// @title           course API
// @version         1.0
// @description     A robust Go course API with authentication, strict validation, and comprehensive security features.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	fx.New(
		di.PkgModule,
		di.AppModule,
		fx.Invoke(func(*web.Server) {}),
	).Run()
}
