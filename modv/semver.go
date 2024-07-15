package modv

import (
	"github.com/Masterminds/semver/v3"
	"github.com/d5/tengo/v2"
)

var SemverModule = map[string]tengo.Object{
	"new_version": &tengo.UserFunction{
		Name:  "new_version",
		Value: semverNewVersion,
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
		}}, nil
}
