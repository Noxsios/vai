// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024-Present Harry Randazzo

// Package main is the entry point for the application
package main

import (
	"os"

	"github.com/noxsios/vai/cmd"
)

func main() {
	code := cmd.Main()
	os.Exit(code)
}
