package core

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/labstack/gommon/log"

	"github.com/udovin/goquiz/config"
	"github.com/udovin/goquiz/models"
	"github.com/udovin/gosql"
	"github.com/udovin/solve/db"
)

// Core manages all available resources.
type Core struct {
	// Config contains config.
	Config config.Config
	// Settings contains settings store.
	Settings *models.SettingStore
	// Roles contains role store.
	Roles *models.RoleStore
	// RoleEdges contains role edge store.
	RoleEdges *models.RoleEdgeStore
	// Accounts contains account store.
	Accounts *models.AccountStore
	// AccountRoles contains account role store.
	AccountRoles *models.AccountRoleStore
	// Sessions contains session store.
	Sessions *models.SessionStore
	// Users contains user store.
	Users *models.UserStore
	// Visits contains visit store.
	Visits *models.VisitStore
	//
	Quizes *models.QuizStore
	//
	Pools *models.PoolStore
	//
	Problems *models.ProblemStore
	//
	context context.Context
	cancel  context.CancelFunc
	waiter  sync.WaitGroup
	// DB stores database connection.
	DB *gosql.DB
	// logger contains logger.
	logger *log.Logger
}

// NewCore creates core instance from config.
func NewCore(cfg config.Config) (*Core, error) {
	conn, err := cfg.DB.Create()
	if err != nil {
		return nil, err
	}
	logger := log.New("core")
	logger.SetLevel(log.Lvl(cfg.LogLevel))
	logger.EnableColor()
	return &Core{Config: cfg, DB: conn, logger: logger}, nil
}

// Logger returns logger instance.
func (c *Core) Logger() *log.Logger {
	return c.logger
}

// Start starts application and data synchronization.
func (c *Core) Start() error {
	if c.cancel != nil {
		return fmt.Errorf("core already started")
	}
	c.Logger().Debug("Starting core")
	defer c.Logger().Debug("Core started")
	c.context, c.cancel = context.WithCancel(context.Background())
	return c.startStoreLoops()
}

// Stop stops syncing stores.
func (c *Core) Stop() {
	if c.cancel == nil {
		return
	}
	c.Logger().Debug("Stopping core")
	defer c.Logger().Debug("Core stopped")
	c.cancel()
	c.waiter.Wait()
	c.context, c.cancel = nil, nil
}

// WrapTx runs function with transaction.
func (c *Core) WrapTx(
	ctx context.Context, fn func(ctx context.Context) error,
	options ...gosql.BeginTxOption,
) (err error) {
	return gosql.WrapTx(ctx, c.DB, func(tx *sql.Tx) error {
		return fn(db.WithTx(ctx, tx))
	}, options...)
}

// StartTask starts task in new goroutine.
func (c *Core) StartTask(task func(ctx context.Context)) {
	c.Logger().Debug("Start core task")
	c.waiter.Add(1)
	go func() {
		defer c.Logger().Debug("Core task finished")
		defer c.waiter.Done()
		task(c.context)
	}()
}
