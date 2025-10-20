package cmd

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/client/v1alpha1"
)

func newProjectCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use: "project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	c.AddCommand(newProjectCreateCommand(logger))
	return c
}

func newProjectCreateCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := v1alpha1.NewDefaultClient(http.Client{Transport: transport})

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return errors.WithStack(err)
			}
			if name == "" {
				return errors.New("name is required")
			}
			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return errors.WithStack(err)
			}
			kind, err := cmd.Flags().GetString("kind")
			if err != nil {
				return errors.WithStack(err)
			}
			ownerIds, err := cmd.Flags().GetStringSlice("owner-ids")
			if err != nil {
				return errors.WithStack(err)
			}
			ownerGroupIds, err := cmd.Flags().GetStringSlice("owner-group-ids")
			if err != nil {
				return errors.WithStack(err)
			}
			reqBody := generated.CreateProjectRequest{
				Name:          name,
				Description:   description,
				Kind:          generated.CreateProjectRequestKind(kind),
				OwnerIds:      ownerIds,
				OwnerGroupIds: ownerGroupIds,
			}

			if err := client.CreateProject(cmd.Context(), &reqBody); err != nil {
				return errors.WithStack(err)
			}

			fmt.Println("project created successfullly")
			return nil
		},
	}

	c.Flags().String("name", "project", "プロジェクト名")
	c.Flags().String("description", "the sample project", "プロジェクトの説明")
	c.Flags().StringSlice("owner-ids", []string{}, "プロジェクトのオーナーID")
	c.Flags().StringSlice("owner-group-ids", []string{}, "プロジェクトのオーナーグループID")
	c.Flags().String("kind", "personal", "プロジェクトの種類 (personal | shared)")
	return c
}
