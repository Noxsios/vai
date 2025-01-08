// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

module github.com/noxsios/vai

go 1.23.4

retract (
	v0.10.1 // v0.10.1 was tagged to retract v0.10.0
	v0.10.0 // accidentally tagged incorrect version
)

require (
	github.com/Masterminds/semver/v3 v3.3.1
	github.com/alecthomas/chroma/v2 v2.15.0
	github.com/charmbracelet/lipgloss v1.0.0
	github.com/charmbracelet/log v0.4.0
	github.com/charmbracelet/x/ansi v0.6.0
	github.com/d5/tengo/v2 v2.17.0
	github.com/goccy/go-yaml v1.15.13
	github.com/google/go-github/v62 v62.0.0
	github.com/invopop/jsonschema v0.13.0
	github.com/muesli/termenv v0.15.2
	github.com/package-url/packageurl-go v0.1.3
	github.com/rogpeppe/go-internal v1.13.1
	github.com/spf13/afero v1.11.0
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.10.0
	github.com/xeipuuv/gojsonschema v1.2.0
	gitlab.com/gitlab-org/api/client-go v0.119.0
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dlclark/regexp2 v1.11.4 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/wk8/go-ordered-map/v2 v2.1.8 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/oauth2 v0.24.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/time v0.8.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
