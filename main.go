//go:generate scripts/generate-openapi.sh
package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"flag"
	"fmt"
	"movie-catalogue/pkg/api"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	msql "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	port  *int
	build = "develop"
	tidb_db_nm = "movie_catalogue"
)

func main() {
	port = flag.Int("port", 8081, "Port for test HTTP server")

	// Retrieve MySQL user and password from environment variables
	mysqlUser := os.Getenv("TIDB_SQL_USER")
	mysqlPassword := os.Getenv("TIDB_SQL_PASSWORD")

	msql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: "gateway01.eu-central-1.prod.aws.tidbcloud.com",
	})
	sqlDB, err := sql.Open("mysql", mysqlUser+".root:"+mysqlPassword+"@tcp(gateway01.eu-central-1.prod.aws.tidbcloud.com:4000)/" + tidb_db_nm + "?tls=tidb")
	if err != nil {
		panic("failed to open database")
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&api.Movie{})

	//get swagger api
	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	serverImpl := &api.ServerImplementation{
		DB: db,
	}

	e := echo.New()
	e.Use(echomiddleware.Logger())

	e.Use(middleware.OapiRequestValidator(swagger))

	// We now register our petStore above as the handler for the interface
	api.RegisterHandlers(e, serverImpl)

	if err != nil {
		log.Fatal("starting tracing: %w", err)
	}

	// Construct the mux for the debug calls.
	debugMux := api.DebugStandardLibraryMux()

	// Start the service listening for debug requests.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe("0.0.0.0:4000", debugMux); err != nil {
			log.Error("shutdown", "status", "debug router closed", "host", "0.0.0.0:4000", "ERROR", err)
		}
	}()

	// App Starting

	ctxi := make(chan error, 1)

	log.Info("starting service", "version", build)
	defer log.Info("shutdown complete")

	// And we serve HTTP until the world ends.
	go func() {
		// ctx <- context.Background()
		log.Info("api", "status", "api started", "handlers", context.Background())
		ctxi <- (e.Start(fmt.Sprintf("0.0.0.0:%d", *port)))
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-ctxi:
		log.Fatal("server error: %w", err)
	case sig := <-shutdown:
		log.Info("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		_, cancel := context.WithTimeout(context.Background(), e.Server.ReadTimeout)
		defer cancel()
	}

}
