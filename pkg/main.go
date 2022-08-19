package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"movie-data-api/provider"
	"movie-data-api/server"

	"github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var port = flag.Int("port", 8080, "Port for test HTTP server")
	var postgresUser = flag.String("postgres-user", "app", "Postgres user")
	var postgresPassword = flag.String("postgres-password", "secret", "Postgres password")
	var postgresHost = flag.String("postgres-url", "localhost", "Postgres URL")
	var postgresPort = flag.String("postgres-port", "5432", "Postgres port")

	flag.Parse()
	// Init global error group
	g := new(errgroup.Group)
	// Init Postgres database connection ---------------------------
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=movie port=%s sslmode=disable",
		*postgresHost, *postgresUser, *postgresPassword, *postgresPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db.AutoMigrate(&server.Movie{})
	db.AutoMigrate(&provider.S3ItemMapping{})

	if err != nil {
		panic("failed to connect database")
	}
	// Init provider service ---------------------------------------
	g.Go(func() error {

		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		if accessKey == "" {
			return fmt.Errorf("AWS_ACCESS_KEY_ID is not set")
		}
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
		if secretKey == "" {
			return fmt.Errorf("AWS_SECRET_ACCESS_KEY is not set")
		}
		bucketName := os.Getenv("AWS_BUCKET_NAME")
		if bucketName == "" {
			return fmt.Errorf("AWS_BUCKET_NAME is not set")
		}
		region := os.Getenv("AWS_REGION")
		if region == "" {
			return fmt.Errorf("AWS_REGION is not set")
		}

		awsProvider, err := provider.NewAWSProvider(accessKey, secretKey, bucketName, region, db)
		if err != nil {
			return err
		}
		awsProvider.Run()

		return nil
	})
	// Init echo server --------------------------------------------
	g.Go(func() error {
		swagger, err := server.GetSwagger()
		if err != nil {
			return err
		}

		swagger.Servers = nil

		// Create an instance of our handler which satisfies the generated interface
		movieServer := server.NewMovieServer(db)

		e := echo.New()

		// e.Use(echomiddleware.TimeoutWithConfig(echomiddleware.TimeoutConfig{
		// 	Timeout: time.Millisecond * 5,
		// }))

		// Log all requests
		e.Use(echomiddleware.Logger())
		// Use our validation middleware to check all requests against the
		// OpenAPI schema.
		e.Use(middleware.OapiRequestValidator(swagger))

		// We now register our petStore above as the handler for the interface
		server.RegisterHandlers(e, movieServer)

		// And we serve HTTP until the world ends.
		return e.Start(fmt.Sprintf("0.0.0.0:%d", *port))
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
