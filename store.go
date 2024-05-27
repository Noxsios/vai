package vai

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/spf13/afero"
)

// IndexFileName is the name of the index file.
const IndexFileName = "index.json"

// ErrHashMismatch is returned when the hash of the stored file does not match the hash in the index.
var ErrHashMismatch = fmt.Errorf("hash mismatch")

// IsHashMismatch returns true if the error is a hash mismatch error.
func IsHashMismatch(err error) bool {
	return errors.Is(err, ErrHashMismatch)
}

// CacheIndex is a list of files and their digests.
type CacheIndex struct {
	Files []struct {
		Name   string `json:"name"`
		Digest string `json:"digest"`
	} `json:"files"`
}

// NewCacheIndex creates a new cache index.
func NewCacheIndex() *CacheIndex {
	return &CacheIndex{
		Files: []struct {
			Name   string `json:"name"`
			Digest string `json:"digest"`
		}{},
	}
}

// Find returns the digest of a file by name and a boolean indicating if the file was found.
func (c *CacheIndex) Find(key string) (string, bool) {
	for _, f := range c.Files {
		if f.Name == key {
			return f.Digest, true
		}
	}
	return "", false
}

// Add adds an entry to the index.
//
// If the key already exists in the index and the value is the same, nothing will happen.
//
// If the key already exists in the index and the value is different, the key will be removed and re-added.
//
// If the key does not exist in the index, it will be added.
func (c *CacheIndex) Add(key, value string) {
	if d, ok := c.Find(key); ok && d == value {
		return
	} else if ok {
		c.Remove(key)
	}
	c.Files = append(c.Files, struct {
		Name   string `json:"name"`
		Digest string `json:"digest"`
	}{
		Name:   key,
		Digest: value,
	})
}

// Remove removes an entry from the index.
func (c *CacheIndex) Remove(key string) {
	for i, f := range c.Files {
		if f.Name == key {
			c.Files = append(c.Files[:i], c.Files[i+1:]...)
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

// NewStore creates a new store at the given path.
func NewStore(fs afero.Fs) (*Store, error) {
	index := NewCacheIndex()

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

// DefaultStore creates a new store in the default location:
//
// $VAI_CACHE || $HOME/.vai/cache
func DefaultStore() (*Store, error) {
	if cache := os.Getenv(CacheEnvVar); cache != "" {
		return NewStore(afero.NewBasePathFs(afero.NewOsFs(), cache))
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cache := filepath.Join(home, ".vai", "cache")

	return NewStore(afero.NewBasePathFs(afero.NewOsFs(), cache))
}

// Fetch retrieves a workflow from the store
func (s *Store) Fetch(key string) (Workflow, error) {
	s.sync.RLock()
	defer s.sync.RUnlock()

	hex, ok := s.index.Find(key)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	f, err := s.fs.Open(hex)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hasher := sha256.New()

	if _, err := io.Copy(hasher, f); err != nil {
		return nil, err
	}

	if fmt.Sprintf("%x", hasher.Sum(nil)) != hex {
		return nil, ErrHashMismatch
	}

	return ReadAndValidate(f)
}

// Store a workflow in the store.
func (s *Store) Store(key string, r io.Reader) error {
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

	s.index.Add(key, hex)

	b, err = json.Marshal(s.index)
	if err != nil {
		return err
	}

	return afero.WriteFile(s.fs, IndexFileName, b, 0644)
}

// Delete a workflow from the store.
func (s *Store) Delete(key string) error {
	s.sync.Lock()
	defer s.sync.Unlock()

	hex, ok := s.index.Find(key)
	if !ok {
		return fmt.Errorf("key not found")
	}

	s.index.Remove(key)

	b, err := json.Marshal(s.index)
	if err != nil {
		return err
	}

	if err := afero.WriteFile(s.fs, IndexFileName, b, 0644); err != nil {
		return err
	}

	return s.fs.Remove(hex)
}

// Exists checks if a workflow exists in the store.
func (s *Store) Exists(key string, r io.Reader) (bool, error) {
	s.sync.RLock()
	defer s.sync.RUnlock()

	hex, ok := s.index.Find(key)
	if !ok {
		return false, nil
	}

	_, err := s.fs.Stat(hex)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	hasher := sha256.New()

	if _, err := io.Copy(hasher, r); err != nil {
		return false, err
	}

	nhex := fmt.Sprintf("%x", hasher.Sum(nil))

	if nhex != hex {
		logger.Debug("hashes", "new", nhex, "old", hex)
		return false, ErrHashMismatch
	}

	return true, nil
}
