package cmd

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samber/lo"
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
	c.AddCommand(newProjectListCommand(logger))
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
			client := v1alpha1.NewDefaultClient(logger, http.Client{Transport: transport})

			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return errors.Wrapf(err, "failed to get name flag")
			}
			if name == "" {
				return errors.New("name is required")
			}
			description, err := cmd.Flags().GetString("description")
			if err != nil {
				return errors.Wrapf(err, "failed to get description flag")
			}
			kind, err := cmd.Flags().GetString("kind")
			if err != nil {
				return errors.Wrapf(err, "failed to get kind flag")
			}
			ownerIds, err := cmd.Flags().GetStringSlice("owner-ids")
			if err != nil {
				return errors.Wrapf(err, "failed to get owner-ids flag")
			}
			ownerGroupIds, err := cmd.Flags().GetStringSlice("owner-group-ids")
			if err != nil {
				return errors.Wrapf(err, "failed to get owner-group-ids flag")
			}
			reqBody := generated.CreateProjectRequest{
				Name:          name,
				Description:   description,
				Kind:          generated.CreateProjectRequestKind(kind),
				OwnerIds:      ownerIds,
				OwnerGroupIds: ownerGroupIds,
			}

			if err := client.CreateProject(cmd.Context(), &reqBody); err != nil {
				return errors.Wrapf(err, "failed to create project")
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

func newProjectListCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := v1alpha1.NewDefaultClient(logger, http.Client{Transport: transport})

			projects, err := client.ListProjects(cmd.Context())
			if err != nil {
				return errors.Wrapf(err, "failed to list projects")
			}

			logger.DebugContext(cmd.Context(), "listed projects successfully", slog.Int("count", len(projects)))
			t := table.NewWriter()
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"ID", "Name", "Description", "Kind"})
			t.AppendRows(lo.Map(projects, func(p generated.Project, index int) table.Row {
				return table.Row{p.ID, p.Name, p.Description, p.Kind}
			}))
			t.Render()
			return nil
		},
	}
	return c
}
