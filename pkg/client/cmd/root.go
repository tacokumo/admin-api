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
	c.AddCommand(newAuthCommand(logger))

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

func newAuthCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	c.AddCommand(newAuthLoginCommand(logger))
	c.AddCommand(newAuthMeCommand(logger))

	return c
}

func newAuthLoginCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with GitHub OAuth",
		RunE: func(cmd *cobra.Command, args []string) error {
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := v1alpha1.NewDefaultClient(logger, http.Client{Transport: transport})

			if err := client.Authenticate(cmd.Context()); err != nil {
				return err
			}

			logger.InfoContext(cmd.Context(), "Authentication completed successfully")
			return nil
		},
	}
	return c
}

func newAuthMeCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use:   "me",
		Short: "Show current user information",
		RunE: func(cmd *cobra.Command, args []string) error {
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := v1alpha1.NewDefaultClient(logger, http.Client{Transport: transport})

			// This will trigger authentication if needed
			_, err := client.ListProjects(cmd.Context())
			if err != nil {
				return err
			}

			logger.InfoContext(cmd.Context(), "Successfully authenticated and able to access API")
			return nil
		},
	}
	return c
}
