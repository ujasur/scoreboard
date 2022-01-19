package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/casbin/casbin"
)

var (
	testTeam    *team
	testHandler *endpoints
	testClock   *clock
)

func TestMain(m *testing.M) {
	testClock = new(clock)
	testTeam = newDefaultTeam()

	workdir, err := filepath.Abs(".")
	if err != nil {
		log.Fatal(err)
	}
	// Open the dbFile data file in your current directory.
	// It will be created if it doesn't exist.
	dbFile, err := ioutil.TempFile(workdir, "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(dbFile.Name())

	db, err := bolt.Open(dbFile.Name(), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal(err)
	}

	enf := casbin.NewEnforcer(filepath.Join(workdir, authConfPath), filepath.Join(workdir, policyConfPath))

	// init user store and load users
	users, err := newUserStore(db, testTeam.Name, 10)
	if err != nil {
		log.Fatal(err)
	}

	// init user store and load users
	links, err := newLinkStore(db, testTeam.Name, 10)
	if err != nil {
		log.Fatal(err)
	}

	templates := newTemplateMgr(filepath.Join(workdir, templateDir), &page{
		Version: "0.0.0",
		Team:    testTeam.Name,
	})

	testHandler = newEndpoints(&endpointsConfig{
		team:        testTeam,
		enforcer:    enf,
		templateMgr: templates,
		userStore:   users,
		linkStore:   links,
		clock:       testClock,
	})

	code := m.Run()

	os.Remove(dbFile.Name())
	os.Exit(code)
}
