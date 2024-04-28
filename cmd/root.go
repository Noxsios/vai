package cmd

import (
	"fmt"
	"os"

	"github.com/Noxsios/vai"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var w map[string]string

var rootCmd = &cobra.Command{
	Use:   "vai",
	Short: "A simple task runner",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		call := ""
		if len(args) > 0 {
			call = args[0]
		}

		var wf vai.Workflow

		b, err := os.ReadFile("vai.yaml")
		if err != nil {
			return err
		}

		if err := yaml.Unmarshal(b, &wf); err != nil {
			return err
		}

		if call == "" {
			fmt.Println("Available:")
			fmt.Println()
			for k := range wf {
				fmt.Println("-", k)
			}
			return nil
		}

		tg, err := wf.Find(call)
		if err != nil {
			return err
		}

		with := make(vai.With)
		with.FromStringMap(w)

		return vai.Run(tg, with)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringToStringVarP(&w, "with", "w", nil, "variables to pass to tasks")
}
