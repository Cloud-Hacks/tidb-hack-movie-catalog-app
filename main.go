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
	"sync"
	"syscall"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	msql "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	port *int
	// postgresUser     = "app"
	// postgresPassword *string
	// postgresHost     *string
	// postgresPort     = "5432"
	build = "develop"
	once  sync.Once
)

const (
	service     = "trace-sales-api"
	environment = "production"
	id          = 1
)

func main() {
	port = flag.Int("port", 8081, "Port for test HTTP server")
	/* 	postgresPassword = flag.String("postgres-password", "hiclass@12", "Postgres password")
	   	postgresHost = flag.String("postgres-host", "localhost", "Postgres host")
	   	flag.Parse()

	   	// Connect to postgres
	   	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
	   		*postgresHost, postgresUser, *postgresPassword, "movie", postgresPort)
	   	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	   	if err != nil {
	   		panic(err)
	   	} */

	msql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: "gateway01.eu-central-1.prod.aws.tidbcloud.com",
	})
	sqlDB, err := sql.Open("mysql", "hLnbrpfVNLTaw8b.root:nzXjwvPU5x3q4EjH@tcp(gateway01.eu-central-1.prod.aws.tidbcloud.com:4000)/test?tls=tidb")
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

	//
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

	// ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "foundation.web.respond")
	// span.SetAttributes(attribute.Int("statusCode", statusCode))
	// defer span.End()

	// Start Tracing Support

	log.Info("startup", "status", "initializing OT/Jaeger tracing support")

	traceProvider, err := startTracing(
		service,
		id,
	)
	if err != nil {
		log.Fatal("starting tracing: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tr := traceProvider.Tracer("component-main")

	_, span := tr.Start(ctx, "foo")
	defer span.End()
	api.AddSpan(ctx, "business.sys.main.exec", attribute.String("foo", "main exec"))

	defer traceProvider.Shutdown(context.Background())

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

	shutdown := make(chan os.Signal, 0)
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

// startTracing configure open telemetery to be used with zipkin.
func startTracing(serviceName string, probability float64) (*trace.TracerProvider, error) {
	ctx := context.Background()
	once.Do(func() {
		otel.SetTextMapPropagator(
			propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			))
	})

	// Create the Jaeger exporter
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
	)
	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	tp := trace.NewTracerProvider(
		// Always be sure to batch in production.
		trace.WithBatcher(traceExporter),
		// Record information about this application in a Resource.
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)

	// I can only get this working properly using the singleton :(
	otel.SetTracerProvider(tp)
	return tp, nil
}
