package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type testerModel struct {
	Name     string `json:"name"`
	Passcode string `json:"passcode"`
	Role     string `json:"role"`
}

type clientModel struct {
	Version int64             `json:"version"`
	Chain   *clientChainModel `json:"chain"`
}

type clientChainModel struct {
	Name     string            `json:"name"`
	Leader   string            `json:"leader"`
	Voters   map[string]string `json:"voters"`
	Result   *pollResult       `json:"result"`
	Unmasked bool              `json:"unmasked"`
}

var (
	master = &testerModel{"master", "master", "scrum_master"}
	voter1 = &testerModel{"voter1", "voter1", "voter"}
	voter2 = &testerModel{"voter2", "voter2", "voter"}
)

func TestInitialUsersList(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	r, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	http.HandlerFunc(testHandler.usersHandler).ServeHTTP(w, r)
	assertStatus(t, w, http.StatusOK)

	expected := `[{"name":"master","role":"scrum_master"}]`
	if strings.TrimSuffix(w.Body.String(), "\n") != expected {
		t.Fatalf("exptected to have a master initially: got %v want %v", w.Body.String(), expected)
	}
}

func TestMasterAuthentication(t *testing.T) {
	signinUser(t, master)
}

func TesAddRemoveUserByMaster(t *testing.T) {
	// Test user list
	assertUsers := func(wantedUsers []*testerModel) {
		r, err := http.NewRequest("GET", "/users", nil)
		if err != nil {
			t.Fatal(err)
		}
		w := httptest.NewRecorder()
		http.HandlerFunc(testHandler.usersHandler).ServeHTTP(w, r)
		assertStatus(t, w, http.StatusOK)

		users := make([]*testerModel, 0)
		if err := json.Unmarshal(w.Body.Bytes(), &users); err != nil {
			t.Fatalf("users list must be fetched: got %v", w.Body.String())
			return
		}
		areUsersEq(t, wantedUsers, users)
	}

	// Test user addition
	voter := &testerModel{"voter3", "voter3", "voter"}
	addVoter(t, voter)
	signinUser(t, voter)
	assertUsers([]*testerModel{master, voter})

	// Test user removal
	r, err := http.NewRequest("POST", fmt.Sprintf("/users/remove?name=%s", voter.Name), nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("authorization", signinUser(t, master))
	w := httptest.NewRecorder()
	http.HandlerFunc(testHandler.usersRemoveHandler).ServeHTTP(w, r)
	assertStatus(t, w, http.StatusOK)
	assertUsers([]*testerModel{master})
}

func TestVoting(t *testing.T) {
	voters := []*testerModel{voter1, voter2}
	scores := []int{22, 2}

	for lid, leader := range voters {
		// Verify closed state
		m := fetchSession(t, master)
		if m.Version == 0 {
			t.Fatalf("version is zero")
		}
		if m.Chain != nil {
			t.Fatalf("chain must be nil for closed session, got %v", m.Chain)
		}

		var nonLeader *testerModel
		if 0 < lid {
			nonLeader = voters[lid-1]
		} else {
			nonLeader = voters[lid+1]
		}

		// Verify opening of a session.
		r, err := http.NewRequest("POST", "/session/open?"+addVoters(t, voters), nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("authorization", signinUser(t, leader))
		w := httptest.NewRecorder()
		http.HandlerFunc(testHandler.sessionOpenHandler).ServeHTTP(w, r)
		assertStatus(t, w, http.StatusOK)
		assertOpenSession(t, master, lid, voters, []string{"", ""}, false)

		// Verify voting.
		var rounds int
		for rounds < 2 {
			rounds++

			submitted := make([]string, len(scores))
			strscores := make([]string, len(scores))

			for j, v := range voters {
				r, err := http.NewRequest("POST", fmt.Sprintf("/session/vote?score=%d", scores[j]), nil)
				if err != nil {
					t.Fatal(err)
				}
				r.Header.Set("authorization", signinUser(t, v))
				w := httptest.NewRecorder()
				http.HandlerFunc(testHandler.sessionVoteHandler).ServeHTTP(w, r)
				assertStatus(t, w, http.StatusAccepted)

				// Others score must be masked.
				if 0 < j {
					submitted[j-1] = "***"
				}
				submitted[j] = strconv.Itoa(scores[j])
				strscores[j] = submitted[j]
				assertOpenSession(t, v, lid, voters, submitted, false)
			}

			// Verify result.
			m = fetchSession(t, master)
			var sum int
			for s := range scores {
				sum += scores[s]
			}
			result := m.Chain.Result
			if result == nil {
				t.Fatalf("all voters have voted but result is not ready")
			}
			avg := math.Round(float64(sum)/float64(len(voters))*100) / 100
			if result.Average != avg {
				t.Fatalf("expected average to be %f, but got %f", avg, result.Average)
			}
			sort.Ints(scores)
			sort.Ints(result.Scores)
			if len(result.Scores) != len(scores) {
				t.Fatalf("result does not reflect all scores, wanted %d but got %d", len(scores), len(result.Scores))
			}
			for j, s := range scores {
				if s != result.Scores[j] {
					t.Fatalf("expected score %d but got %d", s, result.Scores[j])
				}
			}

			// Verify unmasking by leader
			r, err = http.NewRequest("POST", "/session/unmask", nil)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("authorization", signinUser(t, leader))
			w = httptest.NewRecorder()
			http.HandlerFunc(testHandler.sessionUmaskHandler).ServeHTTP(w, r)
			assertStatus(t, w, http.StatusOK)

			var unmasked *clientModel
			err = json.Unmarshal(w.Body.Bytes(), &unmasked)
			if err != nil {
				t.Fatalf("handler returned unexpected body: got %v want %v", w.Body.String(), m)
			}
			assertOpenSession2(t, unmasked, lid, voters, strscores, true)

			// Verify reset by non leader
			r, err = http.NewRequest("POST", "/session/reset", nil)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("authorization", signinUser(t, nonLeader))
			w = httptest.NewRecorder()
			http.HandlerFunc(testHandler.sessionResetHandler).ServeHTTP(w, r)
			assertStatus(t, w, http.StatusBadRequest)
			assertOpenSession(t, master, lid, voters, strscores, false)

			// Verify reset by leader
			r, err = http.NewRequest("POST", "/session/reset", nil)
			if err != nil {
				t.Fatal(err)
			}
			r.Header.Set("authorization", signinUser(t, leader))
			w = httptest.NewRecorder()
			http.HandlerFunc(testHandler.sessionResetHandler).ServeHTTP(w, r)
			assertStatus(t, w, http.StatusOK)

			oldV := unmasked.Version
			m = assertOpenSession(t, master, lid, voters, []string{"", ""}, false)
			if m.Version == oldV {
				t.Fatalf("Reset did not change version")
			}
		}

		// Verify close by non-leader
		r, err = http.NewRequest("POST", "/session/close", nil)
		if err != nil {
			t.Fatal(err)
		}

		r.Header.Set("authorization", signinUser(t, nonLeader))
		w = httptest.NewRecorder()
		http.HandlerFunc(testHandler.sessionCloseHandler).ServeHTTP(w, r)
		assertStatus(t, w, http.StatusBadRequest)
		assertOpenSession(t, master, lid, voters, []string{"", ""}, false)

		// Verify close by leader
		r, err = http.NewRequest("POST", "/session/close", nil)
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("authorization", signinUser(t, leader))
		w = httptest.NewRecorder()
		http.HandlerFunc(testHandler.sessionCloseHandler).ServeHTTP(w, r)
		assertStatus(t, w, http.StatusOK)
	}

}

func fetchSession(t *testing.T, user *testerModel) *clientModel {
	r, err := http.NewRequest("GET", "/session", nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("authorization", signinUser(t, user))
	w := httptest.NewRecorder()
	http.HandlerFunc(testHandler.sessionHandler).ServeHTTP(w, r)
	assertStatus(t, w, http.StatusOK)

	var m *clientModel
	err = json.Unmarshal(w.Body.Bytes(), &m)
	if err != nil {
		t.Fatalf("handler returned unexpected body: got %v want %v", w.Body.String(), m)
	}
	return m
}

func assertOpenSession(t *testing.T, user *testerModel, lid int, voters []*testerModel, scores []string, unmasked bool) *clientModel {
	m := fetchSession(t, user)
	assertOpenSession2(t, m, lid, voters, scores, unmasked)
	return m
}

func assertOpenSession2(t *testing.T, m *clientModel, lid int, voters []*testerModel, scores []string, unmasked bool) {
	if m.Chain == nil {
		t.Fatalf("chain must not be nil for open session")
	}
	if len(m.Chain.Name) == 0 {
		t.Fatalf("open session must have name")
	}
	if m.Chain.Leader != voters[lid].Name {
		t.Fatalf("open session has unexpected leader %s", m.Chain.Leader)
	}
	if m.Chain.Unmasked != unmasked {
		t.Fatalf("open session is unmasked must be %v but got %v", unmasked, m.Chain.Unmasked)
	}

	var scored int
	for i, v := range voters {
		score, ok := m.Chain.Voters[v.Name]
		if !ok {
			t.Fatalf("missing voter %s", v)
		}
		if score != scores[i] {
			t.Fatalf("expected score to be %s but got %v", scores[i], score)
		}
		if scores[i] != "" {
			scored++
		}
	}
	if scored == len(scores) {
		if m.Chain.Result == nil {
			t.Fatalf("result must be computed")
		}
	} else {
		if m.Chain.Result != nil {
			t.Fatalf("result must not exit yet, but got %v", m.Chain.Result)
		}
	}
}

var authTokens sync.Map

func signinUser(t *testing.T, c *testerModel) string {
	if t, ok := authTokens.Load(c.Name); ok {
		return fmt.Sprintf("%v", t)
	}

	w := httptest.NewRecorder()
	r, err := http.NewRequest("POST", fmt.Sprintf("/users/auth?name=%s&passcode=%s", c.Name, c.Passcode), nil)
	if err != nil {
		t.Fatal(err)
	}
	http.HandlerFunc(testHandler.usersAuthHandler).ServeHTTP(w, r)
	if status := w.Code; status != http.StatusOK {
		t.Fatalf("auth failed with status code: got %v want %v", status, http.StatusOK)
	}

	token := w.Header().Get("authorization")
	if len(token) == 0 {
		t.Fatalf("auth token is missing")
	}
	// token := fmt.Sprintf("%s,%s,%s", u.Name, u.Passcode, string(u.Role))
	if _, err = b64.URLEncoding.DecodeString(token); err != nil {
		t.Fatalf("failed to decode auth token: got %v", err)
	}

	authTokens.Store(c.Name, token)
	return token
}

func addVoter(t *testing.T, c *testerModel) {
	r, err := http.NewRequest("POST", fmt.Sprintf("/users/add?name=%s&role=%s", c.Name, c.Role), nil)
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("authorization", signinUser(t, master))

	w := httptest.NewRecorder()
	http.HandlerFunc(testHandler.usersAddHandler).ServeHTTP(w, r)
	if status := w.Code; status != http.StatusCreated {
		t.Fatalf("handler failed to add user with status code: got %v want %v", status, http.StatusCreated)
	}
}

func addVoters(t *testing.T, voters []*testerModel) string {
	var query string
	for i, v := range voters {
		addVoter(t, v)
		if i != 0 {
			query += "&"
		}
		query += "name=" + v.Name
	}
	return query
}

func areUsersEq(t *testing.T, expected []*testerModel, actual []*testerModel) {
	if len(expected) != len(actual) {
		t.Fatalf("users list size is invalid: got %d expected %d", len(actual), len(expected))
	}
	sort.Slice(expected, func(i, j int) bool { return expected[i].Name < expected[j].Name })
	sort.Slice(actual, func(i, j int) bool { return actual[i].Name < actual[j].Name })
	for i := 0; i < len(expected); i++ {
		if expected[i].Name != actual[i].Name || expected[i].Role != actual[i].Role {
			t.Fatalf("invalid list of users")
		}
	}
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, wantedStatus int) {
	if status := w.Code; status != wantedStatus {
		t.Fatalf("handler returned wrong status code: got %v want %v", status, wantedStatus)
	}
}
