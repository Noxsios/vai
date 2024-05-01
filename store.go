package vai

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/goccy/go-yaml"
)

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

// Find returns the digest of a file by name and a boolean indicating if the file was found.
func (c CacheIndex) Find(key string) (string, bool) {
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
	root  string
	index CacheIndex

	sync      sync.RWMutex
	indexLock sync.Mutex
}

// NewStore creates a new store at the given path.
func NewStore(path string) (*Store, error) {
	index := CacheIndex{}
	indexPath := filepath.Join(path, "index.yaml")

	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return nil, err
		}
		_, err = os.Create(indexPath)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(b, &index); err != nil {
		return nil, err
	}

	return &Store{
		root:  path,
		index: index,
	}, nil
}

// DefaultStore creates a new store in the default location:
//
// $VAI_CACHE || $HOME/.vai/cache
func DefaultStore() (*Store, error) {
	if cache := os.Getenv(CacheEnvVar); cache != "" {
		return NewStore(cache)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cache := filepath.Join(home, ".vai", "cache")

	return NewStore(cache)
}

// Fetch retrieves a workflow from the store by key (SHA256).
func (s *Store) Fetch(key string) (Workflow, error) {
	s.sync.RLock()
	defer s.sync.RUnlock()

	s.indexLock.Lock()
	defer s.indexLock.Unlock()

	sha, ok := s.index.Find(key)
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	path := filepath.Join(s.root, sha)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	hasher := sha256.New()

	if _, err := io.Copy(hasher, f); err != nil {
		return nil, err
	}

	if fmt.Sprintf("%x", hasher.Sum(nil)) != sha {
		return nil, ErrHashMismatch
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	return wf, yaml.Unmarshal(b, &wf)
}

func (s *Store) Store(key string, r io.Reader) error {
	s.sync.Lock()
	defer s.sync.Unlock()

	s.indexLock.Lock()
	defer s.indexLock.Unlock()

	sha := sha256.New()

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if _, err := sha.Write(b); err != nil {
		return err
	}

	hash := fmt.Sprintf("%x", sha.Sum(nil))
	path := filepath.Join(s.root, hash)

	if err := os.WriteFile(path, b, 0644); err != nil {
		return err
	}

	s.index.Add(key, hash)

	indexPath := filepath.Join(s.root, "index.yaml")
	b, err = yaml.Marshal(s.index)
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, b, 0644)
}

func (s *Store) Delete(key string) error {
	s.sync.Lock()
	defer s.sync.Unlock()

	s.indexLock.Lock()
	defer s.indexLock.Unlock()

	sha, ok := s.index.Find(key)
	if !ok {
		return fmt.Errorf("key not found")
	}

	s.index.Remove(key)

	indexPath := filepath.Join(s.root, "index.yaml")
	b, err := yaml.Marshal(s.index)
	if err != nil {
		return err
	}

	if err := os.WriteFile(indexPath, b, 0644); err != nil {
		return err
	}

	path := filepath.Join(s.root, sha)
	return os.Remove(path)
}

func (s *Store) Exists(key string, r io.Reader) (bool, error) {
	s.sync.RLock()
	defer s.sync.RUnlock()

	s.indexLock.Lock()
	defer s.indexLock.Unlock()

	sha, ok := s.index.Find(key)
	if !ok {
		return false, nil
	}

	path := filepath.Join(s.root, sha)
	_, err := os.Stat(path)
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

	new := fmt.Sprintf("%x", hasher.Sum(nil))

	logger.Debug("hashes", "new", new, "old", sha)

	if new != sha {
		return false, ErrHashMismatch
	}

	return true, nil
}

func (s *Store) Index() CacheIndex {
	s.indexLock.Lock()
	defer s.indexLock.Unlock()

	return s.index
}
