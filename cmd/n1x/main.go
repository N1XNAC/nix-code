package main

import (
	"fmt"
	"os"

	"github.com/n1xcode/n1x/internal/app"
	"github.com/n1xcode/n1x/internal/config"
	"github.com/n1xcode/n1x/internal/tui"
	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	debug   bool
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "n1x",
		Short: "N1X Code - Terminal AI Coding Agent",
		Long: tui.Banner + "\n\nN1X Code is a terminal-based AI coding agent.\n" +
			"Connect your own API keys and code with AI assistance.\n" +
			"More info: https://github.com/n1xcode/n1x",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "version" || cmd.Name() == "help" {
				return nil
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("", debug)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			a, err := app.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create app: %w", err)
			}
			defer a.Shutdown()
			return a.RunTUI(cmd.Context())
		},
	}

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.Flags().Bool("version", false, "Print version")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(tui.Banner)
			fmt.Printf("Version: %s\n", version)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "config",
		Short: "Open web configuration UI in browser",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("", debug)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			a, err := app.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create app: %w", err)
			}
			defer a.Shutdown()
			return a.RunConfigServer(cmd.Context())
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "run [prompt]",
		Short: "Run in non-interactive mode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load("", debug)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			a, err := app.New(cfg)
			if err != nil {
				return fmt.Errorf("failed to create app: %w", err)
			}
			defer a.Shutdown()
			return a.RunNonInteractive(cmd.Context(), args[0])
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
