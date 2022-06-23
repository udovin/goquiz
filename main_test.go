package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/udovin/goquiz/config"
	"github.com/udovin/goquiz/core"
	"github.com/udovin/goquiz/migrations"
)

var (
	testConfigFile *os.File
	testConfig     = config.Config{
		DB: config.DB{
			Options: config.SQLiteOptions{Path: ":memory:"},
		},
		SocketFile: fmt.Sprintf("/tmp/test-solve-%d.sock", rand.Int63()),
		Server:     &config.Server{},
		Security: &config.Security{
			PasswordSalt: "qwerty123",
		},
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func testSetup(tb testing.TB) {
	var err error
	func() {
		testConfigFile, err = ioutil.TempFile(os.TempDir(), "test-")
		if err != nil {
			tb.Fatal("Error:", err)
		}
		defer testConfigFile.Close()
		err := json.NewEncoder(testConfigFile).Encode(testConfig)
		if err != nil {
			tb.Fatal("Error:", err)
		}
	}()
	c, err := core.NewCore(testConfig)
	if err != nil {
		tb.Fatal("Error:", err)
	}
	c.SetupAllStores()
	if err := migrations.Apply(c); err != nil {
		tb.Fatal("Error:", err)
	}
}

func testTeardown(tb testing.TB) {
	os.RemoveAll(testConfigFile.Name())
	c, err := core.NewCore(testConfig)
	if err != nil {
		tb.Fatal("Error:", err)
	}
	c.SetupAllStores()
	if err := migrations.Unapply(c, true); err != nil {
		tb.Fatal("Error:", err)
	}
}

func TestServerMain(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)
	cmd := cobra.Command{}
	cmd.Flags().String("config", "", "")
	cmd.Flags().Set("config", testConfigFile.Name())
	go func() {
		shutdown <- os.Interrupt
	}()
	serverMain(&cmd, nil)
}

func TestDBApplyMain(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)
	cmd := cobra.Command{}
	cmd.Flags().String("config", "", "")
	cmd.Flags().Set("config", testConfigFile.Name())
	go func() {
		shutdown <- os.Interrupt
	}()
	dbApplyMain(&cmd, nil)
}

func TestDBUnapplyMain(t *testing.T) {
	testSetup(t)
	defer testTeardown(t)
	cmd := cobra.Command{}
	cmd.Flags().String("config", "", "")
	cmd.Flags().Set("config", testConfigFile.Name())
	go func() {
		shutdown <- os.Interrupt
	}()
	dbUnapplyMain(&cmd, nil)
}

func TestVersionMain(t *testing.T) {
	cmd := cobra.Command{}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Unexpected panic: %v", r)
		}
	}()
	versionMain(&cmd, nil)
}

func TestGetConfigUnknown(t *testing.T) {
	cmd := cobra.Command{}
	if _, err := getConfig(&cmd); err == nil {
		t.Fatal("Expected error")
	}
}

func TestCommand(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic")
		}
	}()
	args := os.Args
	os.Args = []string{"solve", "--config", "not-found", "server"}
	main()
	os.Args = args
}
