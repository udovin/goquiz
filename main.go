package main

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/spf13/cobra"

	"github.com/udovin/goquiz/api"
	"github.com/udovin/goquiz/config"
	"github.com/udovin/goquiz/core"
	"github.com/udovin/goquiz/migrations"
)

var shutdown = make(chan os.Signal, 1)

// getConfig reads config with filename from '--config' flag.
func getConfig(cmd *cobra.Command) (config.Config, error) {
	filename, err := cmd.Flags().GetString("config")
	if err != nil {
		return config.Config{}, err
	}
	return config.LoadFromFile(filename)
}

func isServerError(err error) bool {
	return err != nil && err != http.ErrServerClosed
}

func newServer(logger *log.Logger) *echo.Echo {
	srv := echo.New()
	srv.Logger = logger
	srv.HideBanner, srv.HidePort = true, true
	srv.Pre(middleware.RemoveTrailingSlash())
	srv.Use(middleware.Recover(), middleware.Gzip(), middleware.Logger())
	return srv
}

func registerStatic(srv *echo.Echo, static string) {
	srv.Any("/*", func(c echo.Context) error {
		p, err := url.PathUnescape(c.Param("*"))
		if err != nil {
			return err
		}
		name := filepath.Join(static, path.Clean("/"+p))
		if _, err := os.Stat(name); os.IsNotExist(err) {
			name = filepath.Join(static, "index.html")
		}
		return c.File(name)
	})
}

// serverMain starts GoQuiz server.
//
// Simply speaking this function does following things:
//  1. Setup Core instance (with all managers).
//  2. Setup Echo server instance (HTTP + unix socket).
//  3. Register API View to Echo server.
func serverMain(cmd *cobra.Command, _ []string) {
	cfg, err := getConfig(cmd)
	if err != nil {
		panic(err)
	}
	if cfg.Server == nil {
		panic("section 'server' should be configured")
	}
	c, err := core.NewCore(cfg)
	if err != nil {
		panic(err)
	}
	c.SetupAllStores()
	if err := c.Start(); err != nil {
		panic(err)
	}
	defer c.Stop()
	v := api.NewView(c)
	var waiter sync.WaitGroup
	defer waiter.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	waiter.Add(1)
	go func() {
		defer waiter.Done()
		select {
		case <-ctx.Done():
		case <-shutdown:
			cancel()
		}
	}()
	if file := cfg.SocketFile; file != "" {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			panic(err)
		}
		srv := newServer(c.Logger())
		if srv.Listener, err = net.Listen("unix", file); err != nil {
			panic(err)
		}
		v.RegisterSocket(srv.Group("/socket"))
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			defer cancel()
			if err := srv.Start(""); isServerError(err) {
				c.Logger().Error(err)
			}
		}()
		defer func() {
			if err := srv.Shutdown(context.Background()); err != nil {
				c.Logger().Error(err)
			}
		}()
	}
	if config := cfg.Server; config != nil {
		srv := newServer(c.Logger())
		v.Register(srv.Group("/api"))
		if len(config.Static) > 0 {
			registerStatic(srv, config.Static)
		}
		waiter.Add(1)
		go func() {
			defer waiter.Done()
			defer cancel()
			if err := srv.Start(config.Address()); isServerError(err) {
				c.Logger().Error(err)
			}
		}()
		defer func() {
			if err := srv.Shutdown(context.Background()); err != nil {
				c.Logger().Error(err)
			}
		}()
	}
	<-ctx.Done()
}

func dbApplyMain(cmd *cobra.Command, _ []string) {
	cfg, err := getConfig(cmd)
	if err != nil {
		panic(err)
	}
	c, err := core.NewCore(cfg)
	if err != nil {
		panic(err)
	}
	c.SetupAllStores()
	if err := migrations.Apply(c); err != nil {
		panic(err)
	}
}

func dbUnapplyMain(cmd *cobra.Command, _ []string) {
	cfg, err := getConfig(cmd)
	if err != nil {
		panic(err)
	}
	c, err := core.NewCore(cfg)
	if err != nil {
		panic(err)
	}
	c.SetupAllStores()
	if err := migrations.Unapply(c, false); err != nil {
		panic(err)
	}
}

func versionMain(cmd *cobra.Command, _ []string) {
	println("GoQuiz version:", core.Version)
}

// main is a main entry point.
func main() {
	rootCmd := cobra.Command{Use: os.Args[0]}
	rootCmd.PersistentFlags().String("config", "config.json", "")
	rootCmd.AddCommand(&cobra.Command{
		Use:   "server",
		Run:   serverMain,
		Short: "Starts API server",
	})
	dbCmd := cobra.Command{
		Use:   "db",
		Short: "Commands for managing database",
	}
	dbCmd.AddCommand(&cobra.Command{
		Use:   "apply",
		Run:   dbApplyMain,
		Short: "Applies all new migrations to database",
	})
	dbCmd.AddCommand(&cobra.Command{
		Use:   "unapply",
		Run:   dbUnapplyMain,
		Short: "Rolls back all applied migrations",
	})
	rootCmd.AddCommand(&dbCmd)
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Run:   versionMain,
		Short: "Prints information about version",
	})
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
