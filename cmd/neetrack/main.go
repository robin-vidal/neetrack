// Command neetrack runs the Neetrack HTTP server and its maintenance CLI.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"

	"github.com/robin-vidal/neetrack/internal/auth"
	"github.com/robin-vidal/neetrack/internal/catalog"
	"github.com/robin-vidal/neetrack/internal/config"
	"github.com/robin-vidal/neetrack/internal/httpapi"
	"github.com/robin-vidal/neetrack/internal/scheduler"
	"github.com/robin-vidal/neetrack/internal/seed"
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
		case "replay":
			return replay(cfg, args[1:])
		case "seed":
			return seedHistory(cfg, args[1:])
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

func replay(cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	name := fs.String("user", "", "name of the user whose pattern cards to rebuild")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("usage: neetrack replay --user=<name>")
	}

	ctx := context.Background()
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer st.Close()

	creds, err := st.UserByName(ctx, *name)
	if err != nil {
		return fmt.Errorf("looking up user %q: %w", *name, err)
	}

	userSettings, err := st.Settings(ctx, creds.UserID)
	if err != nil {
		return fmt.Errorf("loading settings for %q: %w", *name, err)
	}

	settings := scheduler.Settings{
		TargetRetention: userSettings.TargetRetention,
		TargetDate:      userSettings.TargetDate,
	}
	if err := st.ReplayCards(ctx, creds.UserID, settings); err != nil {
		return fmt.Errorf("replaying attempts: %w", err)
	}

	fmt.Printf("Cartes reconstruites pour %q.\n", *name)
	return nil
}

func seedHistory(cfg config.Config, args []string) error {
	fs := flag.NewFlagSet("seed", flag.ContinueOnError)
	name := fs.String("user", "", "name of the user to fill")
	days := fs.Int("days", 75, "number of days of history to simulate")
	randSeed := fs.Uint64("seed", 1, "random seed, the same seed gives the same history")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" || *days < 1 {
		return fmt.Errorf("usage: neetrack seed --user=<name> [--days=75] [--seed=1]")
	}

	cat, err := catalog.Load()
	if err != nil {
		return fmt.Errorf("loading catalog: %w", err)
	}

	ctx := context.Background()
	st, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	defer st.Close()

	creds, err := st.UserByName(ctx, *name)
	if err != nil {
		return fmt.Errorf("looking up user %q: %w", *name, err)
	}

	existing, err := st.Attempts(ctx, creds.UserID)
	if err != nil {
		return fmt.Errorf("checking existing attempts: %w", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("user %q already has %d attempts, seed only fills an empty history", *name, len(existing))
	}

	userSettings, err := st.Settings(ctx, creds.UserID)
	if err != nil {
		return fmt.Errorf("loading settings for %q: %w", *name, err)
	}

	generated := seed.Generate(cat, seed.Options{
		Days:            *days,
		Now:             time.Now(),
		TargetRetention: userSettings.TargetRetention,
		Rand:            rand.New(rand.NewPCG(*randSeed, *randSeed)),
	})

	attempts := make([]store.LoggedAttempt, len(generated))
	for i, a := range generated {
		attempts[i] = store.LoggedAttempt(a)
	}
	if err := st.AppendAttempts(ctx, creds.UserID, attempts); err != nil {
		return fmt.Errorf("writing attempts: %w", err)
	}

	settings := scheduler.Settings{
		TargetRetention: userSettings.TargetRetention,
		TargetDate:      userSettings.TargetDate,
	}
	if err := st.ReplayCards(ctx, creds.UserID, settings); err != nil {
		return fmt.Errorf("replaying attempts: %w", err)
	}

	fmt.Printf("%d tentatives simulées sur %d jours pour %q.\n", len(attempts), *days, *name)
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

	problemCatalog, err := catalog.Load()
	if err != nil {
		return fmt.Errorf("loading catalog: %w", err)
	}

	srv := httpapi.New(logger, st, authService, problemCatalog)
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
