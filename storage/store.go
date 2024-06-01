// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

// Package storage provides a cache+clients for storing and retrieving remote workflows.
package storage

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"sync"

	"github.com/spf13/afero"
)

// Fetcher fetches a file from a remote location.
type Fetcher interface {
	Fetch(context.Context, string) (io.ReadCloser, error)
}

// Describer describes a file from a remote location.
type Describer interface {
	Describe(context.Context, string) (Descriptor, error)
}

// Cacher is a combination of a fetcher and a describer.
type Cacher interface {
	Fetcher
	Describer
}

// Descriptor describes a file to use for caching.
type Descriptor struct {
	Size int64
	Hex  string
}

// IndexFileName is the name of the index file.
const IndexFileName = "index.json"

// CacheIndex is a list of files and their digests.
type CacheIndex struct {
	Content []Descriptor `json:"content"`
}

// NewCacheIndex creates a new cache index.
func NewCacheIndex() *CacheIndex {
	return &CacheIndex{
		Content: []Descriptor{},
	}
}

// Find returns the descriptor for a given key.
func (c *CacheIndex) Find(desc Descriptor) (Descriptor, bool) {
	i := slices.Index(c.Content, desc)
	if i == -1 {
		return Descriptor{}, false
	}
	return c.Content[i], true
}

// Add adds an entry to the index.
//
// If the key already exists in the index and the value is the same, nothing will happen.
//
// If the key already exists in the index and the value is different, the key will be removed and re-added.
//
// If the key does not exist in the index, it will be added.
func (c *CacheIndex) Add(desc Descriptor) {
	if _, ok := c.Find(desc); ok {
		return
	}
	c.Content = append(c.Content, desc)
}

// Remove removes an entry from the index.
func (c *CacheIndex) Remove(desc Descriptor) {
	for i, d := range c.Content {
		if d == desc {
			c.Content = append(c.Content[:i], c.Content[i+1:]...)
			return
		}
	}
}

// Store is a cache for storing and retrieving remote workflows.
type Store struct {
	index *CacheIndex

	fs afero.Fs

	sync sync.RWMutex
}

// New creates a new store at the given path.
func New(fs afero.Fs) (*Store, error) {
	index := NewCacheIndex()

	if err := fs.MkdirAll("/", 0644); err != nil {
		return nil, err
	}

	if _, err := fs.Stat(IndexFileName); os.IsNotExist(err) {
		f, err := fs.Create(IndexFileName)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		_, err = f.WriteString("{}")
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		b, err := afero.ReadFile(fs, IndexFileName)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, index); err != nil {
			return nil, err
		}
	}

	return &Store{
		fs:    fs,
		index: index,
	}, nil
}

// Fetch retrieves a workflow from the store
func (s *Store) Fetch(desc Descriptor) (io.ReadCloser, error) {
	s.sync.RLock()
	defer s.sync.RUnlock()

	desc, ok := s.index.Find(desc)
	if !ok {
		return nil, fmt.Errorf("descriptor not found")
	}

	f, err := s.fs.Open(desc.Hex)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Store a workflow in the store.
func (s *Store) Store(r io.Reader) error {
	s.sync.Lock()
	defer s.sync.Unlock()

	hasher := sha256.New()

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if _, err := hasher.Write(b); err != nil {
		return err
	}

	hex := fmt.Sprintf("%x", hasher.Sum(nil))

	if err := afero.WriteFile(s.fs, hex, b, 0644); err != nil {
		return err
	}

	s.index.Add(Descriptor{
		Size: int64(len(b)),
		Hex:  hex,
	})

	b, err = json.Marshal(s.index)
	if err != nil {
		return err
	}

	return afero.WriteFile(s.fs, IndexFileName, b, 0644)
}

// Delete a workflow from the store.
func (s *Store) Delete(desc Descriptor) error {
	s.sync.Lock()
	defer s.sync.Unlock()

	desc, ok := s.index.Find(desc)
	if !ok {
		return fmt.Errorf("descriptor not found")
	}

	s.index.Remove(desc)

	b, err := json.Marshal(s.index)
	if err != nil {
		return err
	}

	if err := afero.WriteFile(s.fs, IndexFileName, b, 0644); err != nil {
		return err
	}

	return s.fs.Remove(desc.Hex)
}

// Exists checks if a workflow exists in the store.
func (s *Store) Exists(desc Descriptor) (bool, error) {
	s.sync.RLock()
	defer s.sync.RUnlock()

	desc, ok := s.index.Find(desc)
	if !ok {
		return false, nil
	}

	fi, err := s.fs.Stat(desc.Hex)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("descriptor exists in index, but no corresponding file was found, possible cache corruption: %s", desc.Hex)
		}
		return false, err
	}

	if fi.Size() != desc.Size {
		return false, fmt.Errorf("size mismatch, expected %d, got %d", desc.Size, fi.Size())
	}

	hasher := sha256.New()

	f, err := s.fs.Open(desc.Hex)
	if err != nil {
		return false, err
	}
	defer f.Close()

	if _, err := io.Copy(hasher, f); err != nil {
		return false, err
	}

	if fmt.Sprintf("%x", hasher.Sum(nil)) != desc.Hex {
		return false, errors.New("hash mismatch")
	}

	return true, nil
}
