package cmd

import (
	"crypto/tls"
	"log/slog"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/tacokumo/admin-api/pkg/client/v1alpha1"
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
	c.AddCommand(newPingCommand(logger))

	return c
}

func newPingCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use: "ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.InfoContext(cmd.Context(), "liveness check")
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := v1alpha1.NewDefaultClient(logger, http.Client{Transport: transport})

			if err := client.LivenessCheck(cmd.Context()); err != nil {
				return err
			}

			logger.InfoContext(cmd.Context(), "liveness check succeeded")

			logger.InfoContext(cmd.Context(), "readiness check")
			if err := client.ReadinessCheck(cmd.Context()); err != nil {
				return err
			}

			logger.InfoContext(cmd.Context(), "readiness check succeeded")

			return nil
		},
	}
	return c
}
