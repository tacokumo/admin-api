package v1alpha1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
)

func readResponseError(resp *http.Response) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("unexpected status code: %d (failed to read response body: %v)", resp.StatusCode, err)
	}

	var errorBody map[string]any
	if json.Unmarshal(bodyBytes, &errorBody) == nil {
		return errors.Errorf("unexpected status code: %d, response: %v", resp.StatusCode, errorBody)
	}

	return errors.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(bodyBytes))
}

func (c *DefaultClient) CreateProject(ctx context.Context, req *generated.CreateProjectRequest) (err error) {
	resp, err := c.post(ctx, "/v1alpha1/projects", req)
	if err != nil {
		return errors.Wrapf(err, "failed to create project")
	}
	defer func() {
		if err == nil {
			err = resp.Body.Close()
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		return readResponseError(resp)
	}

	return nil
}

func (c *DefaultClient) ListProjects(
	ctx context.Context,
) (projects []generated.Project, err error) {
	resp, err := c.get(ctx, "/v1alpha1/projects", map[string]string{
		"limit":  "100",
		"offset": "0",
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list projects")
	}
	defer func() {
		if err == nil {
			err = resp.Body.Close()
			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, readResponseError(resp)
	}

	var listResp generated.ListProjectsOKApplicationJSON
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, errors.Wrapf(err, "failed to decode list projects response")
	}
	return listResp, nil
}
