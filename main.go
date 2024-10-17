package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"ohohestudio/sogorro/metadata"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/logging"
	"github.com/gorilla/mux"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	*http.Server
	projectId string
	log       *logging.Logger
}

func main() {
	ctx := context.Background()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	projectId := os.Getenv("GOOGLE_CLOUD_PROJECT")

	app, err := newApp(ctx, port, projectId)
	if err != nil {
		log.Fatalf("unable to initialize sogorro API server: %v", err)
	}

	log.Printf("starting sogorro API server, running on port: %s\n", port)

	go func() {
		if err := app.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("sogorro API server closed unexpectedly: %v", err)
		}
	}()

	nofityCtx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()
	<-nofityCtx.Done()
	log.Println("manually shutdown sogorro API server")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	app.Shutdown(ctx)
	log.Println("sogorro API server has been shutdown")
}

func newApp(ctx context.Context, port, projectId string) (*App, error) {
	app := &App{
		Server: &http.Server{
			Addr:           fmt.Sprintf(":%s", port),
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	if projectId == "" {
		projId, err := metadata.ProjectId(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to detect Project ID from GOOGLE_CLOUD_PROJECT or Google metadata server: %v", err)
		}
		projectId = projId
	}

	app.projectId = projectId

	client, err := logging.NewClient(
		ctx,
		fmt.Sprintf("projects/%s", app.projectId),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize logging client: %v", err)
	}

	app.log = client.Logger("test-log", logging.RedirectAsJSON(os.Stderr))

	r := mux.NewRouter()
	r.HandleFunc("/", app.Home).Methods("GET")
	app.Handler = r

	return app, nil
}

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	a.log.Log(logging.Entry{
		Severity: logging.Info,
		HTTPRequest: &logging.HTTPRequest{
			Request: r,
		},
		Labels:  map[string]string{"arbitraryField": "custom entry"},
		Payload: "Structured logging example",
	})

	fmt.Fprintf(w, "Hello World!\n")
}
