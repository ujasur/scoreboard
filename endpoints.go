package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/casbin/casbin"
	"github.com/gorilla/websocket"
)

var (
	validUserName     = regexp.MustCompile(`^[a-zA-Z]+[a-zA-Z0-9]*$`)
	reservedUserNames = regexp.MustCompile(`^master$`)
	maxUsernameLength = 20
	voterSkipScore    = -2
)

type endpointsConfig struct {
	templateMgr   *templateMgr
	userStore     *userStore
	linkStore     *linksStore
	enforcer      *casbin.Enforcer
	team          *team
	clock         *clock
	chromeExtFile string
}

type endpoints struct {
	config       *endpointsConfig
	auth         *auth
	sessionTopic *sessionTopic
	templateMgr  *templateMgr
	userStore    *userStore
	linkStore    *linksStore
	leader       *leader
	online       *online
	quota        map[string]int
}

func newEndpoints(config *endpointsConfig) *endpoints {
	h := new(endpoints)
	h.config = config

	if onlineEnabledFlag {
		// h.online = newOnline()
	}

	h.templateMgr = config.templateMgr
	h.sessionTopic = newSessionTopic(newSession(config.clock), notificationBufferSize)
	h.linkStore = config.linkStore
	h.userStore = config.userStore
	h.leader = &leader{
		clock:   config.clock,
		maxLife: config.team.getLeaderDuration(),
	}
	users, err := h.userStore.list()
	if err != nil {
		log.Fatal(err)
	}
	// set master from settings
	for _, u := range users {
		if u.Role == roleMaster {
			if err = h.userStore.delete(u.Name); err != nil {
				log.Fatal(err)
			}
		}
	}
	master := newUser(h.config.team.Master, roleMaster)
	if err = h.userStore.create(master); err != nil {
		log.Fatal(err)
	}

	// init authorization
	h.auth = &auth{store: h.userStore, enforcer: h.config.enforcer}

	h.quota = make(map[string]int)
	h.quota["links"] = h.linkStore.getMaxLinks()
	h.quota["users"] = h.userStore.getMaxUsers()
	h.quota["leaderLife"] = int(config.team.getLeaderDuration() / time.Hour)
	return h
}

func (h *endpoints) sessionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}
	model, _ := h.sessionTopic.read(nil)
	json.NewEncoder(w).Encode(model.get(p))
}

func (h *endpoints) sessionOpenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	if !p.hasPermission("session", "open") {
		writeAPIError(w, errUnauthorized)
		return
	}

	voters := queryKey(r, "name")
	if len(voters) == 0 {
		writeAPIError(w, newClientError("at least 1 voter must be provided"))
		return
	}

	// Do not allow openning session if p.user is not in
	// the voters list unless it is master
	var accept bool
	for _, voter := range voters {
		if voter == p.user.Name {
			accept = true
			break
		}
	}
	if !accept {
		accept = p.user.Role == roleMaster
	}
	if !accept {
		writeAPIError(w, newClientError("a session can not be started without you :)"))
		return
	}

	model, err := h.sessionTopic.write(func(s *session, m *modelMasker) error {
		c := s.getChain()
		if c != nil {
			return errSessionOpen
		}
		h.leader.name = p.user.Name
		s.setChain(newPollChain(h.leader, voters))
		return nil
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}
	json.NewEncoder(w).Encode(model.get(p))
}

func (h *endpoints) sessionCloseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	hasPrem := p.hasPermission("session", "close@other")
	model, err := h.sessionTopic.write(func(s *session, m *modelMasker) error {
		c := s.getChain()
		if c == nil {
			return errSessionClosed
		}

		close := c.leader.isDead() || c.leader.is(p.user.Name) || hasPrem
		if !close {
			return newClientError("you are not leader or master")
		}
		s.setChain(nil)

		return nil
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}
	json.NewEncoder(w).Encode(model.get(p))
}

func (h *endpoints) sessionVoteHandler(w http.ResponseWriter, r *http.Request) {
	score, err := strconv.Atoi(queryKeySingular(r, "score"))
	if err != nil || score < voterSkipScore {
		writeAPIError(w, newClientError("score is missing or invalid"))
		return
	}

	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	if !p.hasPermission("session", "vote") {
		writeAPIError(w, errUnauthorized)
		return
	}

	_, err = h.sessionTopic.write(func(s *session, m *modelMasker) error {
		c := s.getChain()
		if c == nil {
			return errSessionClosed
		}

		poll := c.current()
		if score == voterSkipScore {
			poll.removeVoter(p.user.Name)
		} else {
			if !poll.hasVoter(p.user.Name) && c.hasVoter(p.user.Name) {
				poll.addVoter(p.user.Name)
			}
			if accepted := poll.accept(p.user.Name, score); !accepted {
				return errVoteRejected
			}
		}

		if c.leader.is(p.user.Name) {
			c.leader.alive()
		}
		return nil
	})

	if err != nil {
		writeAPIError(w, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *endpoints) sessionResetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	if !p.hasPermission("session", "reset") {
		writeAPIError(w, errUnauthorized)
		return
	}

	model, err := h.sessionTopic.write(func(s *session, m *modelMasker) error {
		c := s.getChain()
		if c == nil {
			return errSessionClosed
		}
		if !c.leader.is(p.user.Name) {
			return newClientError("You are not leader")
		}
		c.leader.alive()
		c.next()
		return nil
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}

	json.NewEncoder(w).Encode(model.get(p))
}

func (h *endpoints) sessionUmaskHandler(w http.ResponseWriter, r *http.Request) {
	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	model, err := h.sessionTopic.write(func(s *session, m *modelMasker) error {
		c := s.getChain()
		if c == nil {
			return errSessionClosed
		}
		if !c.leader.is(p.user.Name) {
			return errUnauthorized
		}
		c.leader.alive()
		c.touch()
		m.noop = true
		return nil
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}
	json.NewEncoder(w).Encode(model.get(p))
}

func (h *endpoints) usersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	users, err := h.userStore.list()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Hide passcodes
	for _, u := range users {
		u.Passcode = ""
	}
	json.NewEncoder(w).Encode(users)
}

func (h *endpoints) usersAddHandler(w http.ResponseWriter, r *http.Request) {
	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	n := queryKeySingular(r, "name")
	if len(n) == 0 || strings.Contains(n, " ") {
		writeAPIError(w,
			newClientError("username is invalid"))
		return
	}

	if len(n) > maxUsernameLength {
		writeAPIError(w,
			newClientError(fmt.Sprintf("username is too long than %d chars",
				maxUsernameLength)))
		return
	}

	if !validUserName.MatchString(n) {
		writeAPIError(w,
			newClientError("username must not be number and must contain only letters"))
		return
	}

	if reservedUserNames.MatchString(n) {
		writeAPIError(w,
			newClientError("username is reserved"))
		return
	}

	ur := queryKeySingular(r, "role")
	if len(ur) == 0 {
		writeAPIError(w,
			newClientError("user role is invalid"))
		return
	}

	if !p.hasPermission("users", "add@"+ur) {
		writeAPIError(w, errUnauthorized)
		return
	}

	err = h.userStore.create(newUser(n, role(ur)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *endpoints) usersRemoveHandler(w http.ResponseWriter, r *http.Request) {
	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	n := queryKeySingular(r, "name")
	if len(n) == 0 {
		writeAPIError(w, newClientError("username is required"))
		return
	}

	if reservedUserNames.MatchString(n) {
		writeAPIError(w, newClientError("username is reserved"))
		return
	}

	u, err := h.userStore.get(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if u == nil {
		http.NotFound(w, r)
		return
	}

	if !p.hasPermission("users", "remove@"+string(u.Role)) {
		writeAPIError(w, errUnauthorized)
		return
	}

	err = h.userStore.delete(u.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *endpoints) usersAuthHandler(w http.ResponseWriter, r *http.Request) {
	n := queryKeySingular(r, "name")
	if len(n) == 0 {
		writeAPIError(w, newClientError("username is required"))
		return
	}

	c := queryKeySingular(r, "passcode")
	if len(c) == 0 {
		writeAPIError(w, newClientError("passcode is required"))
		return
	}

	t, err := h.auth.login(n, c)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	w.Header().Set("authorization", t)
}

func (h *endpoints) linksRemoveHandler(w http.ResponseWriter, r *http.Request) {
	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	id, err := strconv.Atoi(queryKeySingular(r, "id"))
	if err != nil || id < -1 {
		writeAPIError(w, newClientError("link id is required"))
		return
	}

	if !p.hasPermission("links", "remove") {
		writeAPIError(w, errUnauthorized)
		return
	}

	if err := h.linkStore.deleteByID(id); err != nil {
		writeAPIError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *endpoints) linksAddHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	p, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	if !p.hasPermission("links", "add") {
		writeAPIError(w, errUnauthorized)
		return
	}

	var l link
	if json.NewDecoder(r.Body).Decode(&l); err != nil {
		writeAPIError(w, newClientError("malformed body"))
		return
	}

	if err := l.validate(); err != nil {
		writeAPIError(w, err)
		return
	}
	if err := h.linkStore.create(&l); err != nil {
		writeAPIError(w, err)
		return
	}
	json.NewEncoder(w).Encode(l)
}

func (h *endpoints) linksListHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, err := h.auth.authenticate(r.Header.Get("authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	links, err := h.linkStore.list()
	if err != nil {
		writeAPIError(w, err)
		return
	}

	json.NewEncoder(w).Encode(links)
}

func (h *endpoints) pageIndexHandler(w http.ResponseWriter, r *http.Request) {
	h.templateMgr.render(w, &page{
		Name: "session.html",
		Data: h.config.team,
	}, r.Header.Get("If-None-Match"))
}

func (h *endpoints) pageUsersHandler(w http.ResponseWriter, r *http.Request) {
	h.templateMgr.render(w, &page{
		Name: "users.html",
		Data: h.quota,
	}, r.Header.Get("If-None-Match"))
}

func (h *endpoints) pageLoginHandler(w http.ResponseWriter, r *http.Request) {
	h.templateMgr.render(w, &page{
		Name: "login.html",
	}, r.Header.Get("If-None-Match"))
}

func (h *endpoints) pageLinksHandler(w http.ResponseWriter, r *http.Request) {
	h.templateMgr.render(w, &page{
		Name: "links.html",
		Data: h.quota,
	}, r.Header.Get("If-None-Match"))
}

func (h *endpoints) pageDocHandler(w http.ResponseWriter, r *http.Request) {
	h.templateMgr.render(w, &page{
		Name: "docs.html",
		Data: h.quota,
	}, r.Header.Get("If-None-Match"))
}

func (h *endpoints) acceptChangeLogListener(w http.ResponseWriter, r *http.Request) {
	h.socketLoop(w, r, h.sessionTopic, 5*time.Second, true)
}

func (h *endpoints) acceptOnlineListener(w http.ResponseWriter, r *http.Request) {
	if onlineEnabledFlag {
		h.socketLoop(w, r, h.online, time.Second, false)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

var anonymID = fmt.Sprintf("anonym45%d", time.Now().Unix())
var upgrader = websocket.Upgrader{} // use default options
func (h *endpoints) socketLoop(w http.ResponseWriter, r *http.Request, socketTopic topic, pingPeriod time.Duration, allowAnonym bool) {
	p, err := h.auth.authenticate(queryKeySingular(r, "authorization"))
	if err != nil {
		writeAPIError(w, err)
		return
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	log.Printf("connection accepted %s \n", r.RemoteAddr)

	c := &client{
		msg: make(chan msgWriter),
		id:  p.user.Name,
	}

	wsStat.Inc()

	socketTopic.enter(c)
	ticker := time.NewTicker(webSocketPingPeriod)

	defer func() {
		ticker.Stop()
		conn.Close()

		// while we are waiting to ack leave, we have to drain
		// all incoming messages to prevent deadlock
		done := make(chan int)
		go func() {
			for {
				select {
				case <-c.msg:
					// drain messages
				case <-done:
					return
				}
			}
		}()
		socketTopic.leave(c)
		close(done)

		wsStat.Dec()
		log.Printf("connection %s disconnected \n", r.RemoteAddr)
	}()

	if _, ok := r.URL.Query()["sync"]; ok {
		msg := socketTopic.sync()
		if msg.write(conn, p) != nil {
			return
		}
	}

	for {
		select {
		case msg := <-c.msg:
			if msg.write(conn, p) != nil {
				return
			}
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
