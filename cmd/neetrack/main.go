// Command neetrack runs the Neetrack HTTP server and its maintenance CLI.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/config"
	"github.com/robin-vidal/neetrack/internal/httpapi"
	"github.com/robin-vidal/neetrack/internal/store"
)

const shutdownGrace = 10 * time.Second

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	if err := store.Migrate(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	if len(args) > 0 {
		switch args[0] {
		case "migrate":
			logger.Info("migrations applied")
			return nil
		case "adduser":
			if len(args) != 2 {
				return fmt.Errorf("usage: neetrack adduser <name>")
			}
			return addUser(cfg, args[1])
		}
	}

	return serve(cfg, logger)
}

func addUser(cfg config.Config, name string) error {
	fmt.Fprint(os.Stderr, "Mot de passe : ")
	passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return fmt.Errorf("reading password: %w", err)
	}

	hash, err := auth.HashPassword(string(passwordBytes))
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	ctx := context.Background()
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer st.Close()

	if _, err := st.CreateUser(ctx, name, hash); err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	fmt.Printf("Utilisateur %q créé.\n", name)
	return nil
}

func serve(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer st.Close()

	authService := auth.NewService(st, st, cfg.CookieSecure)

	srv := httpapi.New(logger, st, authService)
	httpServer := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: srv.Routes(),
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("listening", "addr", cfg.ListenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownGrace)
	defer cancel()

	logger.Info("shutting down")
	return httpServer.Shutdown(shutdownCtx)
}
