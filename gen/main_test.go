package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenMain(t *testing.T) {
	err := run("..")
	require.NoError(t, err)
}
