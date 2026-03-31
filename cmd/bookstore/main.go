package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	mmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"goa.design/clue/debug"
	"goa.design/clue/log"

	books "github.com/Velin-Todorov/zetta-task/gen/books"
	"github.com/Velin-Todorov/zetta-task/internal/config"
	"github.com/Velin-Todorov/zetta-task/internal/repository"
	bookstore "github.com/Velin-Todorov/zetta-task/internal/service"
	"github.com/Velin-Todorov/zetta-task/internal/storage"
)

func main() {
	// Define command line flags, add any other flag required to configure the
	// service.
	var (
		hostF     = flag.String("host", "localhost", "Server host (valid values: localhost)")
		domainF   = flag.String("domain", "", "Host domain name (overrides host domain specified in service design)")
		httpPortF = flag.String("http-port", "", "HTTP port (overrides host HTTP port specified in service design)")
		secureF   = flag.Bool("secure", false, "Use secure scheme (https or grpcs)")
		dbgF      = flag.Bool("debug", false, "Log request and response bodies")
	)
	flag.Parse()

	// Setup logger. Replace logger with your own log package of choice.
	format := log.FormatJSON
	if log.IsTerminal() {
		format = log.FormatTerminal
	}
	ctx := log.Context(context.Background(), log.WithFormat(format))
	if *dbgF {
		ctx = log.Context(ctx, log.WithDebug())
		log.Debugf(ctx, "debug logs enabled")
	}
	log.Print(ctx, log.KV{K: "http-port", V: *httpPortF})

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf(ctx, err, "failed to load config")
	}
	
	database, err := sql.Open("mysql", cfg.Database.DSN())
	if err != nil {
		log.Fatalf(ctx, err, "failed to open database")
	}
	defer database.Close()

	if err := database.PingContext(ctx); err != nil {
		log.Fatalf(ctx, err, "failed to connect to database")
	}

	// Run migrations
	driver, err := mmysql.WithInstance(database, &mmysql.Config{})
	if err != nil {
		log.Fatalf(ctx, err, "failed to create migration driver")
	}
	m, err := migrate.NewWithDatabaseInstance("file://sql/schema", cfg.Database.Name, driver)
	if err != nil {
		log.Fatalf(ctx, err, "failed to create migration instance")
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf(ctx, err, "failed to run migrations")
	}

	repo := repository.NewMySQLBookRepository(database)
	store := storage.NewLocalStorage(cfg.Storage.UploadPath)
	
	// Initialize the services.
	var (
		booksSvc books.Service
	)
	{
		booksSvc = bookstore.NewBooks(repo, store)
	}

	// Wrap the services in endpoints that can be invoked from other services
	// potentially running in different processes.
	var (
		booksEndpoints *books.Endpoints
	)
	{
		booksEndpoints = books.NewEndpoints(booksSvc)
		booksEndpoints.Use(debug.LogPayloads())
		booksEndpoints.Use(log.Endpoint)
	}

	// Create channel used by both the signal handler and server goroutines
	// to notify the main goroutine when to stop the server.
	errc := make(chan error)

	// Setup interrupt handler. This optional step configures the process so
	// that SIGINT and SIGTERM signals cause the services to stop gracefully.
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)

	// Start the servers and send errors (if any) to the error channel.
	switch *hostF {
	case "localhost":
		{
			addr := "http://0.0.0.0:80"
			u, err := url.Parse(addr)
			if err != nil {
				log.Fatalf(ctx, err, "invalid URL %#v\n", addr)
			}
			if *secureF {
				u.Scheme = "https"
			}
			if *domainF != "" {
				u.Host = *domainF
			}
			if *httpPortF != "" {
				h, _, err := net.SplitHostPort(u.Host)
				if err != nil {
					log.Fatalf(ctx, err, "invalid URL %#v\n", u.Host)
				}
				u.Host = net.JoinHostPort(h, *httpPortF)
			} else if u.Port() == "" {
				u.Host = net.JoinHostPort(u.Host, "80")
			}
			handleHTTPServer(ctx, u, booksEndpoints, &wg, errc, *dbgF)
		}

	default:
		log.Fatal(ctx, fmt.Errorf("invalid host argument: %q (valid hosts: localhost)", *hostF))
	}

	// Wait for signal.
	log.Printf(ctx, "exiting (%v)", <-errc)

	// Send cancellation signal to the goroutines.
	cancel()

	wg.Wait()
	log.Printf(ctx, "exited")
}
