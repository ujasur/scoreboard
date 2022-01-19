package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/casbin/casbin"
)

var version string

// Command line arguments
var (
	databaseDir     = flag.String("db_dir", "", "Database directory path")
	databasePerTeam = flag.Bool("db_per_team", false, "Must each team has own database")
)

const (
	dbPath                     = "scoreboard.db"
	authConfPath               = "config/auth.conf"
	policyConfPath             = "config/policy.csv"
	teamsPath                  = "config/teams.json"
	templateDir                = "templates/"
	staticDir                  = "static/"
	usersLimitPerTeam          = 25
	linksLimitPerTeam          = 10
	connLimitPerTeamServer     = 25
	webSocketPingPeriod        = 5 * time.Second
	defaultMaxFib              = 14
	defaultOutBucket           = 3
	defaultMaxLeaderIdlePeriod = "1h"
	notificationBufferSize     = 10
	onlineEnabledFlag          = false
)

func main() {
	flag.Parse()

	appdir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	dbdir := *databaseDir
	if len(*databaseDir) == 0 {
		dbdir = appdir
	}

	teams, err := readTeams(filepath.Join(appdir, teamsPath))
	if err != nil {
		log.Fatalf("failed to read teams %v", err)
	}
	start(appdir, dbdir, teams)
}

func start(appdir string, dbdir string, teams map[string]*team) {
	done, broadcast := make(chan bool, len(teams)), make(chan bool)
	enforcer := casbin.NewEnforcer(filepath.Join(appdir, authConfPath), filepath.Join(appdir, policyConfPath))

	var db *bolt.DB
	for _, team := range teams {
		if *databasePerTeam || db == nil {
			// It will be created if it doesn't exist.
			var dbfile = dbPath
			if *databasePerTeam {
				dbfile = fmt.Sprintf("%s.%s", strings.ToLower(team.Name), dbfile)
			}
			teamdb, err := bolt.Open(
				filepath.Join(dbdir, dbfile), 0600, &bolt.Options{Timeout: 1 * time.Second})
			if err != nil {
				log.Fatal(err)
			}
			db = teamdb
			defer func(b *bolt.DB) { b.Close() }(teamdb)
		}

		if db == nil {
			log.Fatal("database was not found")
		}

		go startTeamServer(&teamServerOpts{
			db:          db,
			enf:         enforcer,
			team:        team,
			addr:        fmt.Sprintf(":%d", team.Port),
			templates:   filepath.Join(appdir, templateDir),
			staticDir:   filepath.Join(appdir, staticDir),
			connlimit:   connLimitPerTeamServer,
			sigstop:     broadcast,
			sigshutdown: done,
		})
		log.Printf("server team - %s has started at port %d", team.Name, team.Port)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	log.Printf("shutting down %d servers", len(teams))

	close(broadcast)
	for range teams {
		<-done
	}
}

func readTeams(path string) (map[string]*team, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	teams := make(map[string]*team)
	if err := json.Unmarshal(data, &teams); err != nil {
		return nil, err
	}
	if len(teams) == 0 {
		return nil, fmt.Errorf("at least one team is required")
	}

	opts, listenPorts := newDefaultTeam(), make(map[int]bool)
	for name, team := range teams {
		team.Name = name
		if _, ok := listenPorts[team.Port]; ok {
			return nil, fmt.Errorf("duplicate port %d", team.Port)
		}
		listenPorts[team.Port] = true
		team.extend(opts)
		if err := team.validate(); err != nil {
			return nil, err
		}
	}

	return teams, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
