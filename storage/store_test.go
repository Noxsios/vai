// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCacheIndex(t *testing.T) {
	index := NewCacheIndex()

	foo := Descriptor{
		Hex:  "foo",
		Size: 3,
	}

	bar := Descriptor{
		Hex:  "bar",
		Size: 3,
	}

	// add
	index.Add(foo)

	// found
	val, ok := index.Find(foo)
	require.Equal(t, foo, val)
	require.True(t, ok)

	// not found
	val, ok = index.Find(bar)
	require.Equal(t, Descriptor{}, val)
	require.False(t, ok)

	// remove
	index.Remove(foo)
	val, ok = index.Find(foo)
	require.Equal(t, Descriptor{}, val)
	require.False(t, ok)

	// not found now
	val, ok = index.Find(foo)
	require.Equal(t, Descriptor{}, val)
	require.False(t, ok)
}

var shaMap = map[string]string{
	"a": "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
	"b": "3e23e8160039594a33894f6564e1b1348bbd7a0088d42c4acb73eeaed59c009d",
	"c": "2e7d2c03a9507ae265ecf5b5356885a53393a2029d241394997265a1a25aefc6",
}

func TestStore(t *testing.T) {
	fs := afero.NewMemMapFs()
	store, err := New(fs)
	require.NoError(t, err)

	// store initializes with empty index
	b, err := afero.ReadFile(fs, IndexFileName)
	require.NoError(t, err)
	require.JSONEq(t, "{}", string(b))

	// new additions cause no errors
	for k := range shaMap {
		require.NoError(t, store.Store(bytes.NewReader([]byte(k))))
	}

	// index is updated
	b, err = afero.ReadFile(fs, IndexFileName)
	require.NoError(t, err)
	var index CacheIndex
	err = json.Unmarshal(b, &index)
	require.NoError(t, err)
	require.ElementsMatch(t, index.Content, store.index.Content)

	// all keys exist at the correct sha
	for _, v := range shaMap {
		ok, err := store.Exists(Descriptor{
			Hex:  v,
			Size: 1,
		})
		require.NoError(t, err)
		require.True(t, ok)
	}

	// delete
	require.NoError(t, store.Delete(Descriptor{
		Hex:  shaMap["a"],
		Size: 1,
	}))
	require.EqualError(t, store.Delete(Descriptor{
		Hex:  shaMap["a"],
		Size: 1,
	}), "descriptor not found")
	b, err = afero.ReadFile(fs, IndexFileName)
	require.NoError(t, err)
	err = json.Unmarshal(b, &index)
	require.NoError(t, err)
	require.ElementsMatch(t, index.Content, store.index.Content)
	ok, err := store.Exists(Descriptor{
		Hex:  shaMap["a"],
		Size: 1,
	})
	require.NoError(t, err)
	require.False(t, ok)

	// store and retrieve a workflow
	// b, err = yaml.Marshal(helloWorldWorkflow)
	// require.NoError(t, err)
	// key := packageurl.PackageURL{
	// 	Type:      "vai",
	// 	Namespace: "github.com/noxsios",
	// 	Name:      "vai",
	// 	Version:   "v0.1.5",
	// 	Subpath:   "vai.yaml",
	// }
	// require.NoError(t, store.Store(key.String(), bytes.NewReader(b)))

	// wf, err := store.Fetch(key.String())
	// require.NoError(t, err)
	// require.Equal(t, helloWorldWorkflow, wf)

	// store can be re-initialized just fine
	store, err = New(fs)
	require.NoError(t, err)

	// cause a mismatch between index and fs, causing cache corruption
	err = fs.Remove(shaMap["b"])
	require.NoError(t, err)
	ok, err = store.Exists(Descriptor{
		Hex:  shaMap["b"],
		Size: 1,
	})
	require.False(t, ok)
	require.EqualError(t, err, fmt.Sprintf("descriptor exists in index, but no corresponding file was found, possible cache corruption: %s", shaMap["b"]))
}
