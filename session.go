package main

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// StatusNotVoted not voted
	StatusNotVoted     = -1
	VoterStatusSkipped = "skipped"
)

type sessionModel struct {
	Version int64           `json:"version"`
	Chain   *pollChainModel `json:"chain"`
}

type pollChainModel struct {
	Name     string            `json:"name"`
	Leader   string            `json:"leader"`
	Voters   map[string]string `json:"voters"`
	Result   *pollResult       `json:"result"`
	Unmasked bool              `json:"unmasked"`
	Skipped  int               `json:"skipped"`
}

type pollResult struct {
	Average float64 `json:"average"`
	Scores  []int   `json:"scores"`
}

type modelMasker struct {
	sm   *sessionModel
	noop bool
}

func (msk *modelMasker) slurpModel(s *session) {
	sm := new(sessionModel)
	sm.Version = s.version

	// Copy current poll.
	if s.getChain() != nil {
		chain := s.getChain()
		sm.Chain = new(pollChainModel)
		sm.Chain.Name = chain.current().name
		sm.Chain.Leader = chain.leader.name
		sm.Chain.Voters = make(map[string]string)

		poll := chain.current()
		for _, voter := range chain.voters {
			if poll.hasVoter(voter) {
				if poll.isVoted(voter) {
					sm.Chain.Voters[voter] = strconv.Itoa(poll.getScore(voter))
				} else {
					sm.Chain.Voters[voter] = ""
				}
			} else {
				sm.Chain.Skipped += 1
				sm.Chain.Voters[voter] = VoterStatusSkipped
			}
		}

		if chain.current().isReady() {
			sm.Chain.Result = chain.current().compute()
		}
	}
	msk.sm = sm
}

func (msk *modelMasker) write(c *websocket.Conn, p *principal) error {
	return c.WriteJSON(msk.get(p))
}

func (msk *modelMasker) get(p *principal) interface{} {
	src := msk.sm
	dist := new(sessionModel)
	dist.Version = src.Version
	if src.Chain != nil {
		dist.Chain = new(pollChainModel)
		dist.Chain.Name = src.Chain.Name
		dist.Chain.Leader = src.Chain.Leader
		dist.Chain.Result = src.Chain.Result
		dist.Chain.Voters = make(map[string]string)
		dist.Chain.Unmasked = msk.noop
		dist.Chain.Skipped = src.Chain.Skipped
		// We only allow to an user with "view_all_others" permissions to see others scores.
		viewAll := msk.noop || p.hasPermission("session", "view_all_others")
		for voter, status := range src.Chain.Voters {
			show := len(status) == 0 || status == VoterStatusSkipped || viewAll || p.user.Name == voter
			if show {
				dist.Chain.Voters[voter] = status
			} else {
				dist.Chain.Voters[voter] = "***"
			}
		}
	}
	return dist
}

type leader struct {
	name        string
	clock       *clock
	lastTouched time.Time
	maxLife     time.Duration
}

func (l *leader) is(name string) bool {
	return l.name == name
}

func (l *leader) alive() {
	l.lastTouched = l.clock.Now()
}

func (l *leader) isDead() bool {
	return l.maxLife < time.Since(l.lastTouched)
}

type pollChain struct {
	owner   *session
	leader  *leader
	poll    *poll
	voters  []string
	counter int
}

func newPollChain(l *leader, voters []string) *pollChain {
	c := new(pollChain)
	c.leader = l
	c.voters = voters
	c.leader.alive()
	c.next()
	return c
}

func (c *pollChain) setOwner(s *session) {
	c.owner = s
}

func (c *pollChain) getVoters() []string {
	if c.voters == nil {
		return nil
	}
	return c.voters[:]
}

func (c *pollChain) hasVoter(voter string) bool {
	if len(c.voters) > 0 {
		for _, v := range c.voters {
			if v == voter {
				return true
			}
		}
	}
	return false
}

func (c *pollChain) current() *poll {
	return c.poll
}

func (c *pollChain) next() {
	c.counter++

	voters := make(map[string]int)
	for _, v := range c.voters {
		voters[v] = StatusNotVoted
	}
	c.poll = new(poll)
	c.poll.name = fmt.Sprintf("%d. %s|%s", c.counter, randomName(0), nextColor())
	c.poll.owner = c
	c.poll.voters = voters
	c.touch()
}

func (c *pollChain) touch() {
	if c.owner != nil {
		c.owner.touch()
	}
}

type poll struct {
	owner  *pollChain
	name   string
	voters map[string]int
}

func (p *poll) cancel(voter string) bool {
	return p.accept(voter, StatusNotVoted)
}

func (p *poll) getScore(voter string) int {
	if !p.hasVoter(voter) {
		return -2
	}
	return p.voters[voter]
}

func (p *poll) removeVoter(voter string) bool {
	_, ok := p.voters[voter]
	if ok {
		delete(p.voters, voter)
		p.owner.touch()
	}
	return ok
}

func (p *poll) addVoter(voter string) bool {
	if !p.hasVoter(voter) {
		p.voters[voter] = StatusNotVoted
		p.owner.touch()
		return true
	}
	return false
}

func (p *poll) accept(voter string, score int) bool {
	if score < StatusNotVoted {
		return false
	}
	cv, ok := p.voters[voter]
	if ok {
		if cv != score {
			p.voters[voter] = score
			p.owner.touch()
		}
	}
	return ok
}

func (p *poll) hasVoter(voter string) bool {
	_, has := p.voters[voter]
	return has
}

func (p *poll) isVoted(voter string) bool {
	return p.hasVoter(voter) && p.voters[voter] != StatusNotVoted
}

func (p *poll) isReady() bool {
	for _, score := range p.voters {
		if score == StatusNotVoted {
			return false
		}
	}
	return len(p.voters) > 0
}

func (p *poll) compute() *pollResult {
	r := new(pollResult)
	r.Scores = make([]int, 0)
	if !p.isReady() {
		return r
	}
	var voted int
	var sum float64
	for _, score := range p.voters {
		if score != StatusNotVoted {
			voted++
			sum += float64(score)
			r.Scores = append(r.Scores, score)
		}
	}
	r.Average = math.Round(float64(sum)/float64(voted)*100) / 100
	return r
}

type session struct {
	version int64
	chain   *pollChain
}

func newSession(c *clock) *session {
	s := new(session)
	s.version = c.Now().Unix()
	return s
}

func (s *session) getVersion() int64 {
	return s.version
}

func (s *session) getChain() *pollChain {
	return s.chain
}

func (s *session) setChain(ch *pollChain) {
	if s.chain != ch {
		if s.chain != nil {
			s.chain.setOwner(nil)
		}
		s.chain = ch
		if s.chain != nil {
			s.chain.setOwner(s)
		}
		s.touch()
	}
}

func (s *session) touch() {
	s.version = s.version + 1
}

type sessionTopic struct {
	clients  map[*client]bool
	entering chan *client
	leaving  chan *client
	changes  chan *modelMasker

	mux     sync.RWMutex
	session *session
}

func newSessionTopic(s *session, size int) *sessionTopic {
	t := new(sessionTopic)
	t.session = s
	t.clients = make(map[*client]bool)
	t.entering = make(chan *client)
	t.leaving = make(chan *client)
	t.changes = make(chan *modelMasker, size)
	go t.broadcaster()
	return t
}

func (t *sessionTopic) sync() msgWriter {
	m, _ := t.read(nil)
	return m
}

func (t *sessionTopic) enter(c *client) {
	t.entering <- c
}

func (t *sessionTopic) leave(c *client) {
	t.leaving <- c
}

func (t *sessionTopic) notify(m *modelMasker) {
	t.changes <- m
}

func (t *sessionTopic) broadcaster() {
	for {
		select {
		case m := <-t.changes:
			for c := range t.clients {
				c.msg <- m
			}
		case c := <-t.entering:
			t.clients[c] = true
		case c := <-t.leaving:
			delete(t.clients, c)
			close(c.msg)
		}
	}
}

func (t *sessionTopic) read(reader func(s *session) error) (*modelMasker, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if reader != nil {
		if err := reader(t.session); err != nil {
			return nil, err
		}
	}
	msk := new(modelMasker)
	msk.slurpModel(t.session)
	return msk, nil
}

func (t *sessionTopic) readPartial(reader func(s *session) error) error {
	t.mux.RLock()
	defer t.mux.RUnlock()
	if reader != nil {
		if err := reader(t.session); err != nil {
			return err
		}
	}
	return nil
}

func (t *sessionTopic) write(writer func(s *session, m *modelMasker) error) (*modelMasker, error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	old := t.session.getVersion()
	msk := &modelMasker{}
	if err := writer(t.session, msk); err != nil {
		return nil, err
	}
	msk.slurpModel(t.session)
	if old != t.session.getVersion() {
		t.notify(msk)
	}
	return msk, nil
}
