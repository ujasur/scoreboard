package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/casbin/casbin"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type team struct {
	Name                  string        `json:"name"`
	Master                string        `json:"master"`
	Port                  int           `json:"port"`
	Preference            *preference   `json:"preference"`
	LeaderMaxIdlePeriod   string        `json:"leader_max_idle_period"`
	LeaderMaxIdleDuration time.Duration `json:"-"`
}

type preference struct {
	MaxFib           int    `json:"max_fib"`
	OutOfBucketLimit int    `json:"out_of_bucket_limit"`
	PrimaryAggrFunc  string `json:"primary_aggr_func"`
}

func newDefaultTeam() *team {
	t := new(team)
	t.Name = "team"
	t.Master = "master"
	t.LeaderMaxIdlePeriod = defaultMaxLeaderIdlePeriod
	t.Preference = &preference{
		MaxFib:           defaultMaxFib,
		OutOfBucketLimit: defaultOutBucket,
		PrimaryAggrFunc:  "closestFib",
	}
	return t
}

func (t *team) validate() error {
	if t.Port <= 3000 {
		return fmt.Errorf("wanted port be greater of 3000, but got %d", t.Port)
	}
	if t.Preference == nil {
		return fmt.Errorf("Preference is nil")
	}
	_, err := time.ParseDuration(t.LeaderMaxIdlePeriod)
	if err != nil {
		return err
	}
	return nil
}

func (t *team) getLeaderDuration() time.Duration {
	period, err := time.ParseDuration(t.LeaderMaxIdlePeriod)
	if err != nil {
		panic(err)
	}
	return period
}

func (t *team) extend(src *team) {
	if len(t.Name) == 0 {
		t.Name = src.Name
	}
	if len(t.Master) == 0 {
		t.Master = src.Master
	}
	if len(t.LeaderMaxIdlePeriod) == 0 {
		t.LeaderMaxIdlePeriod = src.LeaderMaxIdlePeriod
	}
	if t.Preference == nil {
		t.Preference = &preference{}
	}
	if src.Preference != nil {
		t.Preference.extend(src.Preference)
	}
}

func (p *preference) extend(src *preference) {
	if p.MaxFib <= 0 {
		p.MaxFib = src.MaxFib
	}
	if p.OutOfBucketLimit <= 0 {
		p.OutOfBucketLimit = src.OutOfBucketLimit
	}
	if len(p.PrimaryAggrFunc) == 0 {
		p.PrimaryAggrFunc = src.PrimaryAggrFunc
	}
}

type teamServerOpts struct {
	team        *team
	db          *bolt.DB
	enf         *casbin.Enforcer
	addr        string
	connlimit   int
	templates   string
	staticDir   string
	sigshutdown chan bool
	sigstop     <-chan bool
}

func startTeamServer(opts *teamServerOpts) {
	users, err := newUserStore(opts.db, opts.team.Name, usersLimitPerTeam)
	if err != nil {
		log.Fatal(err)
	}

	links, err := newLinkStore(opts.db, opts.team.Name, linksLimitPerTeam)
	if err != nil {
		log.Fatal(err)
	}

	templates := newTemplateMgr(opts.templates, &page{
		Version: version, // Referencing global variable :(
		Team:    opts.team.Name,
	})

	h := newEndpoints(&endpointsConfig{
		team:        opts.team,
		enforcer:    opts.enf,
		clock:       new(clock),
		templateMgr: templates,
		userStore:   users,
		linkStore:   links,
	})

	r := mux.NewRouter()

	r.Use(metricMiddleware([]string{"/metrics"}))
	r.Use(connLimitMiddleware(opts.connlimit))

	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", newFsWrapper(opts.staticDir, 1*time.Hour)))

	r.HandleFunc("/", h.pageIndexHandler)
	r.HandleFunc("/ui/users", h.pageUsersHandler)
	r.HandleFunc("/ui/login", h.pageLoginHandler)
	r.HandleFunc("/ui/links", h.pageLinksHandler)
	r.HandleFunc("/ui/docs", h.pageDocHandler)

	r.HandleFunc("/session", h.sessionHandler)
	r.HandleFunc("/session/open", h.sessionOpenHandler)
	r.HandleFunc("/session/close", h.sessionCloseHandler)
	r.HandleFunc("/session/vote", h.sessionVoteHandler)
	r.HandleFunc("/session/reset", h.sessionResetHandler)
	r.HandleFunc("/session/unmask", h.sessionUmaskHandler)
	r.HandleFunc("/session/changes", h.acceptChangeLogListener)
	if onlineEnabledFlag {
		r.HandleFunc("/session/live_users", h.acceptOnlineListener)
	}

	r.HandleFunc("/users/auth", h.usersAuthHandler)
	r.HandleFunc("/users", h.usersHandler)
	r.HandleFunc("/users/add", h.usersAddHandler)
	r.HandleFunc("/users/remove", h.usersRemoveHandler)

	r.HandleFunc("/links", h.linksListHandler)
	r.HandleFunc("/links/add", h.linksAddHandler)
	r.HandleFunc("/links/remove", h.linksRemoveHandler)

	r.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:           opts.addr,
		Handler:        r,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		<-opts.sigstop
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("server team - %s shutdown: %v", opts.team.Name, err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil {
		log.Printf("server team - %s listen: %v", opts.team.Name, err)
	}

	opts.sigshutdown <- true
}

type fsWrapper struct {
	maxage  time.Duration
	handler http.Handler
}

func newFsWrapper(path string, maxage time.Duration) http.Handler {
	f := new(fsWrapper)
	f.handler = http.FileServer(http.Dir(path))
	f.maxage = maxage
	return f
}

func (f *fsWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Expires", time.Now().Add(f.maxage).Format(http.TimeFormat))
	f.handler.ServeHTTP(w, r)
}
