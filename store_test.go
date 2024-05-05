package vai

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/package-url/packageurl-go"
	"github.com/stretchr/testify/require"
)

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
	"a": "87428fc522803d31065e7bce3cf03fe475096631e5e07bbd7a0fde60c4cf25c7",
	"b": "0263829989b6fd954f72baaf2fc64bc2e2f01d692d4de72986ea808f6e99813f",
	"c": "a3a5e715f0cc574a73c3f9bebb6bc24f32ffd5b67b387244c2c909da779a1478",
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

	// all keys exist (sha checking is done in store)
	for k := range shaMap {
		ok, err := store.Exists(k, bytes.NewReader([]byte(k)))
		require.NoError(t, err)
		require.True(t, ok)
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
	helloWorldWorkflow := Workflow{"default": {Step{CMD: "echo 'Hello World!'"}}}
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
