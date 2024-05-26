// Package cmd provides the root command for the vai CLI.
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"github.com/noxsios/vai"
	"github.com/spf13/cobra"
)

// NewRootCmd creates the root command for the vai CLI.
func NewRootCmd() *cobra.Command {
	var w map[string]string
	var level string
	var ver bool
	var list bool
	var filename string
	var timeout time.Duration

	root := &cobra.Command{
		Use:   "vai",
		Short: "A simple task runner",
		ValidArgsFunction: func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			if filename == "" {
				filename = vai.DefaultFileName
			}
			f, err := os.Open(filename)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			defer f.Close()

			wf, err := vai.ReadAndValidate(f)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			return wf.OrderedTaskNames(), cobra.ShellCompDirectiveNoFileComp
		},
		PreRunE: func(_ *cobra.Command, _ []string) error {
			l, err := log.ParseLevel(level)
			if err != nil {
				return err
			}
			vai.SetLogLevel(l)
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := vai.Logger()

			if ver && len(args) == 0 {
				bi, ok := debug.ReadBuildInfo()
				if !ok {
					return fmt.Errorf("version information not available")
				}
				logger.Printf("%s", bi.Main.Version)
				return nil
			}

			if cmpl := os.Getenv("VAI_COMPLETION"); cmpl == "true" && len(args) == 2 && args[0] == "completion" {
				switch args[1] {
				case "bash":
					return cmd.GenBashCompletion(os.Stdout)
				case "zsh":
					return cmd.GenZshCompletion(os.Stdout)
				case "fish":
					return cmd.GenFishCompletion(os.Stdout, false)
				case "powershell":
					return cmd.GenPowerShellCompletionWithDesc(os.Stdout)
				default:
					return fmt.Errorf("unsupported shell: %s", cmpl)
				}
			}

			if filename == "" {
				filename = vai.DefaultFileName
			}

			f, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer f.Close()

			wf, err := vai.ReadAndValidate(f)
			if err != nil {
				return err
			}

			if list {
				names := wf.OrderedTaskNames()

				if len(names) == 0 {
					return fmt.Errorf("no tasks available")
				}

				logger.Print("Available:\n")
				for _, n := range names {
					logger.Printf("- %s", n)
				}

				return nil
			}

			with := make(vai.With)
			for k, v := range w {
				with[k] = v
			}

			if len(args) == 0 {
				args = append(args, vai.DefaultTaskName)
			}

			ctx := cmd.Context()

			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			for _, call := range args {
				if err := vai.Run(ctx, wf, call, with); err != nil {
					if errors.Is(ctx.Err(), context.DeadlineExceeded) {
						return fmt.Errorf("task %q timed out", call)
					}
					return err
				}
			}
			return nil
		},
	}

	root.Flags().StringToStringVarP(&w, "with", "w", nil, "key=value pairs to pass to the called task(s)")
	root.Flags().StringVarP(&level, "log-level", "l", "info", "log level")
	root.Flags().BoolVarP(&ver, "version", "V", false, "print version")
	root.Flags().BoolVarP(&vai.Force, "force", "F", false, "ignore checksum mismatch for cached remote files")
	root.Flags().BoolVar(&list, "list", false, "list available tasks")
	root.Flags().StringVarP(&filename, "file", "f", "", "read file as workflow definition")
	root.Flags().DurationVarP(&timeout, "timeout", "t", 0, "timeout for task execution")

	return root
}

// Main executes the root command for the vai CLI.
//
// It returns 0 on success, 1 on failure and logs any errors.
func Main() int {
	cli := NewRootCmd()

	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	logger := vai.Logger()

	if err := cli.ExecuteContext(ctx); err != nil {
		logger.Print("")
		logger.Error(err)
		return 1
	}
	return 0
}
