package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/client/auth"
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
			logger.DebugContext(cmd.Context(), "trying to retrieve auth0 token")
			// Implementation for project creation goes here
			v, err := auth.RetrieveToken(os.Getenv("AUTH0_DOMAIN"), os.Getenv("AUTH0_CLIENT_ID"), os.Getenv("AUTH0_CLIENT_SECRET"), os.Getenv("AUTH0_AUDIENCE"))
			if err != nil {
				return errors.WithStack(err)
			}

			serverHost := os.Getenv("SERVER_HOST")
			if serverHost == "" {
				serverHost = "localhost"
			}
			serverURL := fmt.Sprintf("https://%s:8444", serverHost)
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := &http.Client{
				Transport: transport,
			}
			reqBody := generated.CreateProjectRequest{
				Name:          "example-project",
				Description:   "the example project",
				OwnerIds:      []string{},
				OwnerGroupIds: []string{},
			}
			reqBodyBytes, err := json.Marshal(reqBody)
			if err != nil {
				return errors.WithStack(err)
			}
			endpoint := fmt.Sprintf("%s/v1alpha1/projects", serverURL)
			req, err := http.NewRequestWithContext(cmd.Context(), http.MethodPost, endpoint, bytes.NewReader(reqBodyBytes))
			if err != nil {
				return errors.WithStack(err)
			}
			fmt.Println(v)
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", v.AccessToken))
			req.Header.Add("Content-Type", "application/json")

			logger.DebugContext(cmd.Context(), "sending request to create project")
			resp, err := client.Do(req)
			if err != nil {
				return errors.WithStack(err)
			}
			defer func() {
				if e := resp.Body.Close(); e != nil && err == nil {
					err = errors.WithStack(e)
				}
			}()

			logger.DebugContext(cmd.Context(), "trying to read body from response")
			respBody := &bytes.Buffer{}
			_, err = respBody.ReadFrom(resp.Body)
			if err != nil {
				return errors.WithStack(err)
			}
			if resp.StatusCode != http.StatusOK {
				return errors.Newf("unexpected status code: %d, body: %s", resp.StatusCode, respBody.String())
			}
			logger.InfoContext(cmd.Context(), "project created successfully", "response", respBody.String())
			return nil
		},
	}
	return c
}
