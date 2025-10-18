package cmd

import (
	"log/slog"

	"github.com/spf13/cobra"
)

func New(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use:           "client",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	c.AddCommand(newProjectCommand(logger))

	return c
}
