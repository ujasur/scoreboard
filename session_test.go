package main

import (
	"math"
	"testing"
)

var (
	voterA = "va"
	voterB = "vb"
	voterC = "vc"
)

func TestSessionInit(t *testing.T) {
	s := newSession(testClock)
	if s.getChain() != nil {
		t.Fatal("session must not have chain by default")
	}
}

func TestSessionSetChain(t *testing.T) {
	s := newSession(testClock)

	c := newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB})

	v := s.getVersion()
	s.setChain(c)
	if s.getVersion() <= v {
		t.Fatalf("setting a new chain must increment the version, before=%d after=%d", v, s.getVersion())
	}
}

func TestSessionSetChainAgain(t *testing.T) {
	s := newSession(testClock)

	c := newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB})

	s.setChain(c)
	v := s.getVersion()
	s.setChain(c)
	if s.getVersion() != v {
		t.Fatalf("setting the same chain must not increment the version wanted %d, got %d", v, s.getVersion())
	}
}

func TestSessionSetNewChain(t *testing.T) {
	s := newSession(testClock)

	a := newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB})
	s.setChain(a)

	v := s.getVersion()
	b := newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB})
	s.setChain(b)
	if s.getVersion() <= v {
		t.Fatalf("setting a new chain must increment the version, before=%d after=%d", v, s.getVersion())
	}
}

func TestSessionSetChainNil(t *testing.T) {
	s := newSession(testClock)

	c := newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB})
	s.setChain(c)

	v := s.getVersion()
	s.setChain(nil)
	if s.getVersion() <= v {
		t.Fatalf("clearing a chain must increment version, before=%d after=%d", v, s.getVersion())
	}
}

func TestVotingMultiple(t *testing.T) {
	c := newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB})

	p := c.current()
	if p.isReady() {
		t.Fatal("Expected poll result being not ready")
	}
	checkUnreadyResult(t, p)

	voters := c.getVoters()
	for rounds := 0; rounds < 4; rounds++ {
		var sum int

		// check voting all
		for i := 0; i < len(voters); i++ {
			score := (rounds + i + 1)
			sum += score
			if accepted := p.accept(voters[i], score); !accepted {
				t.Fatal("Expected a valid vote to be accepted")
			}
			if i == len(voters)-1 {
				if !p.isReady() {
					t.Fatal("Result must be ready after every voters has voted")
				}
				avg := math.Round(float64(sum)/float64(len(voters))*100) / 100
				res := p.compute()
				if res.Average != avg {
					t.Fatalf("Poll result invalid, wanted %f got %f", avg, res.Average)
				}
			} else {
				checkUnreadyResult(t, p)
			}
		}

		// checking cancelling votes
		for i := 0; i < len(voters); i++ {
			if accepted := p.cancel(voters[i]); !accepted {
				t.Fatal("Expected a no vote to be accepted")
			}
			checkUnreadyResult(t, p)
		}
	}
}

func TestVotingSingle(t *testing.T) {
	c := newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB})
	p := c.current()

	p.accept(voterA, 2)
	checkUnreadyResult(t, p)

	p.accept(voterA, 2)
	checkUnreadyResult(t, p)

	rounds := 3
	for rounds >= 0 {
		p.accept(voterA, rounds)
		checkUnreadyResult(t, p)
		rounds--
	}

	p.accept(voterA, StatusNotVoted)
	p.cancel(voterA)
	p.cancel(voterA)
	checkUnreadyResult(t, p)

	p.accept(voterA, 4)
	p.accept(voterB, 8)
	checkReadyResult(t, p, 6, 2)

	p.accept(voterB, StatusNotVoted)
	checkUnreadyResult(t, p)
}

func TestNextPoll(t *testing.T) {
	c := newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB})

	p1 := c.current()
	p1.accept(voterA, 1)
	p1.accept(voterB, 2)

	c.next()

	p2 := c.current()
	if p1 == p2 {
		t.Fatal("next poll failed to generate a new poll")
	}
	checkUnreadyResult(t, p2)

	if accepted := p2.accept(voterA, 3); !accepted {
		t.Fatal("Expected acceptance of vote")
	}
	checkUnreadyResult(t, p2)
	if accepted := p2.accept(voterB, 5); !accepted {
		t.Fatal("Expected acceptance of vote")
	}

	checkReadyResult(t, p2, 4, 2)

	p1.cancel(voterA)
	checkUnreadyResult(t, p1)
	checkReadyResult(t, p2, 4, 2)
}

func TestVersionChange(t *testing.T) {
	s := newSession(testClock)
	s.setChain(newPollChain(&leader{name: "leader", clock: testClock}, []string{voterA, voterB}))
	c := s.getChain()
	actions := [...]func() string{
		func() string {
			c.current().accept(voterA, 1)
			return "accept()"
		},
		func() string {
			c.current().accept(voterB, 2)
			return "accept()"
		},
		func() string {
			c.next()
			return "next()"
		},
		func() string {
			c.current().accept(voterA, 3)
			return "accept()"
		},
		func() string {
			c.current().cancel(voterA)
			return "cancel()"
		},
		func() string {
			s.setChain(nil)
			return "setChain() - nil"
		},
		func() string {
			s.setChain(newPollChain(&leader{name: "leader", clock: testClock}, []string{voterB}))
			return "setChain() - new poll"
		},
	}

	for _, act := range actions {
		old := s.getVersion()
		op := act()
		if s.getVersion() <= old {
			t.Fatalf("%s must increment the version, before=%d after=%d", op, old, s.getVersion())
		}
	}
}

func checkUnreadyResult(t *testing.T, p *poll) {
	if p.isReady() {
		t.Fatal("Result must not be ready until every voter has voted")
	}
	r := p.compute()
	if r.Average != 0 {
		t.Fatal("Expected poll result to be 0, got ", r.Average)
	}
	if len(r.Scores) != 0 {
		t.Fatal("Unready poll result must have zero scores, got", len(r.Scores))
	}
}

func checkReadyResult(t *testing.T, p *poll, average float64, scores int) {
	if !p.isReady() {
		t.Fatal("Poll result must be ready")
	}
	r := p.compute()
	if r.Average != average {
		t.Fatalf("Invalid average, wanted %f, got %f", average, r.Average)
	}
	if len(r.Scores) != scores {
		t.Fatal("Ready result must have all scores")
	}
}
