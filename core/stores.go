package core

import (
	"log"
	"reflect"
	"time"

	"github.com/udovin/goquiz/models"
)

// SetupAllStores prepares all stores.
func (c *Core) SetupAllStores() {
	c.Settings = models.NewSettingStore(
		c.DB, "goquiz_setting", "goquiz_setting_event",
	)
	c.Roles = models.NewRoleStore(
		c.DB, "goquiz_role", "goquiz_role_event",
	)
	c.RoleEdges = models.NewRoleEdgeStore(
		c.DB, "goquiz_role_edge", "goquiz_role_edge_event",
	)
	c.Accounts = models.NewAccountStore(
		c.DB, "goquiz_account", "goquiz_account_event",
	)
	c.AccountRoles = models.NewAccountRoleStore(
		c.DB, "goquiz_account_role", "goquiz_account_role_event",
	)
	c.Sessions = models.NewSessionStore(
		c.DB, "goquiz_session", "goquiz_session_event",
	)
	if c.Config.Security != nil {
		c.Users = models.NewUserStore(
			c.DB, "goquiz_user", "goquiz_user_event",
			c.Config.Security.PasswordSalt,
		)
	}
	c.Visits = models.NewVisitStore(c.DB, "goquiz_visit")
}

func (c *Core) startStores(start func(models.Store, time.Duration)) {
	start(c.Settings, time.Second*5)
	start(c.Roles, time.Second*5)
	start(c.RoleEdges, time.Second*5)
	start(c.Accounts, time.Second)
	start(c.AccountRoles, time.Second)
	start(c.Sessions, time.Second)
	start(c.Users, time.Second)
}

func (c *Core) startStoreLoops() error {
	errs := make(chan error)
	count := 0
	c.startStores(func(s models.Store, d time.Duration) {
		v := reflect.ValueOf(s)
		if s == nil || (v.Kind() == reflect.Ptr && v.IsNil()) {
			return
		}
		count++
		c.waiter.Add(1)
		go c.runStoreLoop(s, d, errs)
	})
	var err error
	for i := 0; i < count; i++ {
		lastErr := <-errs
		if lastErr != nil {
			log.Println("Error:", lastErr)
			err = lastErr
		}
	}
	return err
}

func (c *Core) runStoreLoop(
	s models.Store, d time.Duration, errs chan<- error,
) {
	defer c.waiter.Done()
	err := s.Init(c.context)
	errs <- err
	if err != nil {
		return
	}
	ticker := time.NewTicker(d)
	defer ticker.Stop()
	for {
		select {
		case <-c.context.Done():
			return
		case <-ticker.C:
			if err := s.Sync(c.context); err != nil {
				log.Println("Error:", err)
			}
		}
	}
}
