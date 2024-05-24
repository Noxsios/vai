package vai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/require"
)

func TestIsHashMismatch(t *testing.T) {
	require.True(t, IsHashMismatch(fmt.Errorf("additional context: %w", ErrHashMismatch)))
	require.False(t, IsHashMismatch(errors.New("some other error")))
}

func TestCacheIndex(t *testing.T) {
	index := NewCacheIndex()

	// add
	index.Add("foo", "bar")

	// found
	val, ok := index.Find("foo")
	require.Equal(t, "bar", val)
	require.True(t, ok)

	// not found
	val, ok = index.Find("baz")
	require.Equal(t, "", val)
	require.False(t, ok)

	// key:value overrwritten if different value
	index.Add("foo", "baz")
	val, ok = index.Find("foo")
	require.True(t, ok)
	require.Equal(t, "baz", val)

	// no op if key:value same key
	index.Add("foo", "baz")
	val, ok = index.Find("foo")
	require.True(t, ok)
	require.Equal(t, "baz", val)

	// remove
	index.Remove("foo")
	val, ok = index.Find("foo")
	require.Equal(t, "", val)
	require.False(t, ok)
}

var shaMap = map[string]string{
	"a": "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
	"b": "3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
	"c": "2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
}

func TestStore(t *testing.T) {
	tmp := t.TempDir()

	store, err := NewStore(tmp)
	require.NoError(t, err)

	// store initializes with empty index
	b, err := os.ReadFile(filepath.Join(store.root, "index.json"))
	require.NoError(t, err)
	require.JSONEq(t, "{}", string(b))

	// new additions cause no errors
	for k := range shaMap {
		require.NoError(t, store.Store(k, bytes.NewReader([]byte(k))))
	}

	// index is updated
	b, err = os.ReadFile(filepath.Join(store.root, "index.json"))
	require.NoError(t, err)
	var index CacheIndex
	err = json.Unmarshal(b, &index)
	require.NoError(t, err)
	require.ElementsMatch(t, index.Files, store.index.Files)

	// all keys exist at the correct sha
	for k, v := range shaMap {
		ok, err := store.Exists(k, bytes.NewReader([]byte(k)))
		require.NoError(t, err)
		require.True(t, ok)

		for _, f := range store.index.Files {
			if f.Name == k {
				require.Equal(t, v, f.Digest)
			}
		}
	}

	// delete
	require.NoError(t, store.Delete("a"))
	b, err = os.ReadFile(filepath.Join(store.root, "index.json"))
	require.NoError(t, err)
	err = json.Unmarshal(b, &index)
	require.NoError(t, err)
	require.ElementsMatch(t, index.Files, store.index.Files)
	ok, err := store.Exists("a", bytes.NewReader([]byte("a")))
	require.NoError(t, err)
	require.False(t, ok)

	// store and retrieve a workflow
	b, err = yaml.Marshal(helloWorldWorkflow)
	require.NoError(t, err)
	key := packageurl.PackageURL{
		Type:      "vai",
		Namespace: "github.com/noxsios",
		Name:      "vai",
		Version:   "v0.1.5",
		Subpath:   "vai.yaml",
	}
	require.NoError(t, store.Store(key.String(), bytes.NewReader(b)))

	wf, err := store.Fetch(key.String())
	require.NoError(t, err)
	require.Equal(t, helloWorldWorkflow, wf)
}

func TestDefaultStore(t *testing.T) {
	defer os.Unsetenv(CacheEnvVar)
	tmp := t.TempDir()
	os.Setenv(CacheEnvVar, tmp)
	store, err := DefaultStore()
	require.NoError(t, err)
	require.Equal(t, store.root, tmp)
}
