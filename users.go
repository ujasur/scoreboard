package main

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/casbin/casbin"
)

const defaultMaxUsers = 20
const usersBucketName = "users"

type role string

const (
	roleVoter  role = "voter"
	roleMaster role = "scrum_master"
)

type user struct {
	Name     string `json:"name"`
	Role     role   `json:"role"`
	Passcode string `json:"passcode,omitempty"`
}

func newUser(username string, userRole role) *user {
	newUser := new(user)
	newUser.Name = username
	newUser.Passcode = username // being lazy
	newUser.Role = role(userRole)
	return newUser
}

type userStore struct {
	db       *bolt.DB
	bucket   []byte
	maxUsers int
}

func newUserStore(db *bolt.DB, shard string, maxUsers int) (*userStore, error) {
	s := new(userStore)
	s.db = db
	s.bucket = []byte(fmt.Sprintf("%s_%s", shard, usersBucketName))

	if maxUsers <= 0 {
		s.maxUsers = defaultMaxUsers
	} else {
		s.maxUsers = maxUsers
	}

	tx, err := s.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.CreateBucketIfNotExists(s.bucket)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *userStore) getMaxUsers() int {
	return s.maxUsers
}

func (s *userStore) create(u *user) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)

		var n int
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			n++
		}

		if s.maxUsers < n {
			return newClientError(fmt.Sprintf("maximum %d allowed users is reached", s.maxUsers))
		}

		buf, err := json.Marshal(u)
		if err != nil {
			return err
		}
		return b.Put([]byte(u.Name), buf)
	})
}

func (s *userStore) get(username string) (*user, error) {
	u := &user{}
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		data := b.Get([]byte(username))
		json.Unmarshal(data, &u)
		return nil
	})
	return u, err
}

func (s *userStore) delete(username string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		return b.Delete([]byte(username))
	})
}

func (s *userStore) list() ([]*user, error) {
	var users []*user
	err := s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(s.bucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			u := new(user)
			err := json.Unmarshal(v, &u)
			if err != nil {
				return err
			}
			users = append(users, u)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		users = make([]*user, 0)
	}
	return users, err
}

type auth struct {
	store    *userStore
	enforcer *casbin.Enforcer
}

type principal struct {
	user          *user
	enforcer      *casbin.Enforcer
	authenticated bool
}

func (p *principal) isAnonymus() bool {
	return !p.authenticated || p.user == nil
}

func (p *principal) hasPermission(obj string, act string) bool {
	if p.isAnonymus() {
		return false
	}
	return p.enforcer.Enforce(string(p.user.Role), obj, act)
}

func (a *auth) authenticate(tokenRaw string) (*principal, error) {
	p := new(principal)
	p.enforcer = a.enforcer

	if len(tokenRaw) == 0 {
		return p, errAuthMissing
	}

	token, err := b64.URLEncoding.DecodeString(tokenRaw)
	if err != nil {
		return p, errAuthTokenType
	}
	parts := strings.Split(string(token), ",")
	if len(parts) < 2 {
		return p, errAuthTokenType
	}
	username, passcode := parts[0], parts[1]

	user, err := a.store.get(username)
	if err != nil {
		return nil, &systemError{err: err, msg: fmt.Sprintf("auth: failed to get user %s from store", username)}
	}
	if user == nil || user.Passcode != passcode {
		return p, errAuthInvalid
	}
	p.user = user
	p.authenticated = true

	return p, nil
}

func (a *auth) login(username string, passcode string) (string, error) {
	if len(username) == 0 {
		return "", errAuthMissing
	}
	if len(passcode) == 0 {
		return "", errAuthMissing
	}

	u, err := a.store.get(username)
	if err != nil {
		return "", &systemError{err: err, msg: fmt.Sprintf("auth: failed to get user %s from store", username)}
	}
	if u == nil {
		return "", errAuthInvalid
	}
	if u.Passcode != passcode {
		return "", errAuthInvalid
	}

	token := fmt.Sprintf("%s,%s,%s", u.Name, u.Passcode, string(u.Role))
	return b64.URLEncoding.EncodeToString([]byte(token)), nil
}
