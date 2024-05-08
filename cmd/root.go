// Package cmd provides the root command for the vai CLI.
package cmd

import (
	"cmp"
	"runtime/debug"
	"slices"

	"github.com/charmbracelet/log"
	"github.com/noxsios/vai"
	"github.com/spf13/cobra"
)

var w map[string]string
var level string
var ver bool
var list bool

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vai",
		Short: "A simple task runner",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			l, err := log.ParseLevel(level)
			if err != nil {
				return err
			}
			vai.SetLogLevel(l)
			return nil
		},
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, args []string) error {
			logger := vai.Logger()

			if ver {
				bi, ok := debug.ReadBuildInfo()
				if ok {
					logger.Printf("%s", bi.Main.Version)
				}
				return nil
			}

			wf, err := vai.ReadAndValidate(vai.DefaultFileName)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				args = append(args, vai.DefaultTaskName)
			}

			with := make(vai.With)
			for k, v := range w {
				with[k] = v
			}

			if list {
				logger.Print("Available:\n")
				names := []string{}
				for k := range wf {
					names = append(names, k)
				}
				slices.SortStableFunc(names, func(a, b string) int {
					if a == vai.DefaultTaskName {
						return -1
					}
					return cmp.Compare(a, b)
				})

				for _, n := range names {
					logger.Printf("- %s", n)
				}

				return nil
			}
			for _, call := range args {
				if err := vai.Run(wf, call, with); err != nil {
					return err
				}
			}
			return nil
		},
	}

	cmd.Flags().StringToStringVarP(&w, "with", "w", nil, "variables to pass to the called task(s)")
	cmd.Flags().StringVarP(&level, "log-level", "l", "info", "log level")
	cmd.Flags().BoolVarP(&ver, "version", "V", false, "print version")
	cmd.Flags().BoolVarP(&vai.Force, "force", "F", false, "ignore checksum mismatch for cached remote files")
	cmd.Flags().BoolVar(&list, "list", false, "list available tasks")

	return cmd
}
