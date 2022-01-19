package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/boltdb/bolt"
)

const linksBucketName = "links"
const defaultMaxLinks = 10

type link struct {
	ID          int    `json:"id"`
	URI         string `json:"uri"`
	DisplayName string `json:"display_name"`
}

func (l *link) validate() error {
	if len(l.DisplayName) == 0 {
		return newClientError("name is required")
	}
	if len(l.URI) == 0 {
		return newClientError("uri is required")
	}
	if len(l.DisplayName) > 54 {
		return newClientError("name is too long, limit is 54 chars")
	}
	if len(l.URI) > 2083 {
		return newClientError("uri is too long, limit is 2083 chars")
	}
	if _, err := url.Parse(l.URI); err != nil {
		return newClientError("url format is invalid")
	}
	return nil
}

type linksStore struct {
	db       *bolt.DB
	bucket   []byte
	maxLinks int
}

func newLinkStore(db *bolt.DB, shard string, maxLinks int) (*linksStore, error) {
	s := new(linksStore)
	s.db = db
	s.bucket = []byte(fmt.Sprintf("%s_%s", shard, linksBucketName))
	if maxLinks <= 0 {
		s.maxLinks = defaultMaxLinks
	} else {
		s.maxLinks = maxLinks
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

func (s *linksStore) getMaxLinks() int {
	return s.maxLinks
}

func (s *linksStore) create(l *link) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)

		var n int
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			n++
		}
		if s.maxLinks <= n {
			return newClientError(fmt.Sprintf("maximum %d allowed links is reached", s.maxLinks))
		}

		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		l.ID = int(id)

		buf, err := json.Marshal(l)
		if err != nil {
			return err
		}
		return b.Put(itob(l.ID), buf)
	})
}

func (s *linksStore) deleteByID(id int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		return b.Delete(itob(id))
	})
}

func (s *linksStore) list() ([]*link, error) {
	var links []*link
	err := s.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket(s.bucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			l := new(link)
			err := json.Unmarshal(v, &l)
			if err != nil {
				return err
			}
			links = append(links, l)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(links) == 0 {
		links = make([]*link, 0)
	}
	return links, err
}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
