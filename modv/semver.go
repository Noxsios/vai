// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

// Package modv provides extra tengo modules for vai.
package modv

import (
	"github.com/Masterminds/semver/v3"
	"github.com/d5/tengo/v2"
)

// SemverModule is a map of semver-related functions.
var SemverModule = map[string]tengo.Object{
	"new_version": &tengo.UserFunction{
		Name:  "new_version",
		Value: semverNewVersion,
	},
	"to_string": &tengo.UserFunction{
		Name:  "to_string",
		Value: semverToString,
	},
}

func semverNewVersion(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	s, ok := tengo.ToString(args[0])
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "string",
			Found:    args[0].TypeName(),
		}
	}

	v, err := semver.NewVersion(s)
	if err != nil {
		return nil, err
	}

	var prefix string
	// logic copied from semver's originalVPrefix
	if v.Original() != "" && v.Original()[0:1] == "v" {
		prefix = "v"
	}

	return &tengo.Map{
		Value: map[string]tengo.Object{
			"major": &tengo.Int{
				Value: int64(v.Major()),
			},
			"minor": &tengo.Int{
				Value: int64(v.Minor()),
			},
			"patch": &tengo.Int{
				Value: int64(v.Patch()),
			},
			"prerelease": &tengo.String{
				Value: v.Prerelease(),
			},
			"metadata": &tengo.String{
				Value: v.Metadata(),
			},
			"prefix": &tengo.String{
				Value: prefix,
			},
		},
	}, nil
}

func semverToString(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	m, ok := args[0].(*tengo.Map)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "first",
			Expected: "map",
			Found:    args[0].TypeName(),
		}
	}

	major, ok := (m.Value["major"].(*tengo.Int))
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "major",
			Expected: "int",
			Found:    m.Value["major"].TypeName(),
		}
	}

	minor, ok := m.Value["minor"].(*tengo.Int)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "minor",
			Expected: "int",
			Found:    m.Value["minor"].TypeName(),
		}
	}

	patch, ok := m.Value["patch"].(*tengo.Int)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "patch",
			Expected: "int",
			Found:    m.Value["patch"].TypeName(),
		}
	}

	prerelease, ok := m.Value["prerelease"].(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "prerelease",
			Expected: "string",
			Found:    m.Value["prerelease"].TypeName(),
		}
	}

	metadata, ok := m.Value["metadata"].(*tengo.String)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "metadata",
			Expected: "string",
			Found:    m.Value["metadata"].TypeName(),
		}
	}

	prefix, ok := (m.Value["prefix"].(*tengo.String))
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "prefix",
			Expected: "string",
			Found:    m.Value["prefix"].TypeName(),
		}
	}

	if prefix.Value != "" && prefix.Value != "v" {
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "prefix",
			Expected: "literal 'v' or empty string",
			Found:    prefix.Value,
		}
	}

	v := semver.New(uint64(major.Value), uint64(minor.Value), uint64(patch.Value), prerelease.Value, metadata.Value)
	return &tengo.String{
		Value: prefix.Value + v.String(),
	}, nil
}
