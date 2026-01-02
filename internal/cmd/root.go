package cmd

import (
	"log/slog"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/tacokumo/admin-api/pkg/config"
	"github.com/tacokumo/admin-api/pkg/server"
)

func New(logger *slog.Logger) *cobra.Command {
	return &cobra.Command{
		Use:           "api",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fp := os.Getenv("ADMIN_API_CONFIG_FILE")
			cfg, err := config.LoadFromYAMLWithEnvOverride(fp)
			if err != nil {
				return errors.Wrapf(err, "failed to load config from file: %s", fp)
			}

			srv, err := server.New(cmd.Context(), cfg, logger)
			if err != nil {
				return errors.Wrapf(err, "failed to create server")
			}

			if err := srv.Start(cmd.Context()); err != nil {
				return errors.Wrapf(err, "failed to start server")
			}

			return nil
		},
	}
}
