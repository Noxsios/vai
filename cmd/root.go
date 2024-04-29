package cmd

import (
	"os"
	"runtime/debug"

	"github.com/Noxsios/vai"
	"github.com/charmbracelet/log"
	"github.com/goccy/go-yaml"
	"github.com/spf13/cobra"
)

var w map[string]string
var ll string
var ver bool

var rootCmd = &cobra.Command{
	Use:   "vai",
	Short: "A simple task runner",
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		l, err := log.ParseLevel(ll)
		if err != nil {
			return err
		}
		vai.SetLogLevel(l)
		return nil
	},
	RunE: func(_ *cobra.Command, args []string) error {
		logger := vai.Logger()

		if ver {
			bi, ok := debug.ReadBuildInfo()
			if ok {
				logger.Printf("%s", bi.Main.Version)
			}
			return nil
		}

		var wf vai.Workflow

		b, err := os.ReadFile("vai.yaml")
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(b, &wf); err != nil {
			return err
		}

		if len(args) == 0 {
			logger.Print("Available:\n")
			for k := range wf {
				logger.Printf("- %s", k)
			}
			return nil
		}

		with := make(vai.With)
		for k, v := range w {
			with[k] = v
		}

		for _, call := range args {
			tasks, err := wf.Find(call)
			if err != nil {
				return err
			}

			if err := vai.Run(tasks, with); err != nil {
				return err
			}
		}
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.SilenceUsage = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringToStringVarP(&w, "with", "w", nil, "variables to pass to tasks")
	rootCmd.Flags().StringVarP(&ll, "log-level", "l", "info", "log level")
	rootCmd.Flags().BoolVar(&ver, "version", false, "print version")
}
