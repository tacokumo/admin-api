package cmd

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
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
	c.AddCommand(newInteractiveCommand(logger))

	return c
}

func newPingCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use: "ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := v1alpha1.NewDefaultClient(logger, http.Client{Transport: transport})

			fmt.Println("Checking liveness...")
			if err := client.LivenessCheck(cmd.Context()); err != nil {
				fmt.Printf("❌ Liveness check failed: %v\n", err)
				return errors.Wrap(err, "liveness check failed")
			}
			fmt.Println("✅ Liveness check succeeded")

			fmt.Println("Checking readiness...")
			if err := client.ReadinessCheck(cmd.Context()); err != nil {
				fmt.Printf("❌ Readiness check failed: %v\n", err)
				return errors.Wrap(err, "readiness check failed")
			}
			fmt.Println("✅ Readiness check succeeded")

			fmt.Println("🎉 All health checks passed!")
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
	c.AddCommand(newAuthLogoutCommand(logger))

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
				fmt.Printf("❌ Authentication failed: %v\n", err)
				return errors.Wrap(err, "authentication failed")
			}

			fmt.Println("✅ Authentication completed successfully")
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

func newAuthLogoutCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use:   "logout",
		Short: "Logout and remove stored authentication token",
		RunE: func(cmd *cobra.Command, args []string) error {
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := v1alpha1.NewDefaultClient(logger, http.Client{Transport: transport})

			if err := client.Logout(cmd.Context()); err != nil {
				fmt.Printf("❌ Logout failed: %v\n", err)
				return errors.Wrap(err, "logout failed")
			}

			fmt.Println("✅ Successfully logged out")
			return nil
		},
	}
	return c
}

func newInteractiveCommand(logger *slog.Logger) *cobra.Command {
	c := &cobra.Command{
		Use:   "interactive",
		Short: "Start interactive mode for continuous API testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			transport := &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			client := v1alpha1.NewDefaultClient(logger, http.Client{Transport: transport})

			fmt.Println("🥙 Tacokumo Admin API - Interactive Mode")
			fmt.Println("Available commands:")
			fmt.Println("  ping           - Health check (liveness & readiness)")
			fmt.Println("  login          - Authenticate with GitHub OAuth")
			fmt.Println("  logout         - Logout and remove stored token")
			fmt.Println("  projects list  - List all projects")
			fmt.Println("  projects create - Create a new project")
			fmt.Println("  help           - Show this help")
			fmt.Println("  exit           - Exit interactive mode")
			fmt.Println()

			scanner := bufio.NewScanner(os.Stdin)
			for {
				fmt.Print("> ")
				if !scanner.Scan() {
					break
				}

				input := strings.TrimSpace(scanner.Text())
				if input == "" {
					continue
				}

				parts := strings.Fields(input)
				command := parts[0]

				switch command {
				case "exit", "quit", "q":
					fmt.Println("Goodbye! 👋")
					return nil
				case "help", "h":
					fmt.Println("Available commands:")
					fmt.Println("  ping           - Health check (liveness & readiness)")
					fmt.Println("  login          - Authenticate with GitHub OAuth")
					fmt.Println("  logout         - Logout and remove stored token")
					fmt.Println("  projects list  - List all projects")
					fmt.Println("  projects create - Create a new project")
					fmt.Println("  help           - Show this help")
					fmt.Println("  exit           - Exit interactive mode")
				case "ping":
					if err := runPingCommand(cmd.Context(), client, logger); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				case "login":
					if err := runLoginCommand(cmd.Context(), client, logger); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				case "logout":
					if err := runLogoutCommand(cmd.Context(), client, logger); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				case "projects":
					if len(parts) < 2 {
						fmt.Println("Usage: projects [list|create]")
						continue
					}
					subcommand := parts[1]
					switch subcommand {
					case "list":
						if err := runProjectsListCommand(cmd.Context(), client, logger); err != nil {
							fmt.Printf("Error: %v\n", err)
						}
					case "create":
						if err := runProjectsCreateCommand(cmd.Context(), client, logger, scanner); err != nil {
							fmt.Printf("Error: %v\n", err)
						}
					default:
						fmt.Printf("Unknown projects subcommand: %s\n", subcommand)
						fmt.Println("Usage: projects [list|create]")
					}
				default:
					fmt.Printf("Unknown command: %s\n", command)
					fmt.Println("Type 'help' for available commands")
				}
			}

			if err := scanner.Err(); err != nil {
				return errors.Wrap(err, "error reading input")
			}

			return nil
		},
	}
	return c
}

func runPingCommand(ctx context.Context, client v1alpha1.Client, logger *slog.Logger) error {
	fmt.Println("Checking liveness...")
	if err := client.LivenessCheck(ctx); err != nil {
		return errors.Wrap(err, "liveness check failed")
	}
	fmt.Println("✅ Liveness check succeeded")

	fmt.Println("Checking readiness...")
	if err := client.ReadinessCheck(ctx); err != nil {
		return errors.Wrap(err, "readiness check failed")
	}
	fmt.Println("✅ Readiness check succeeded")

	return nil
}

func runLoginCommand(ctx context.Context, client v1alpha1.Client, logger *slog.Logger) error {
	if err := client.Authenticate(ctx); err != nil {
		return errors.Wrap(err, "authentication failed")
	}
	fmt.Println("✅ Authentication completed successfully")
	return nil
}

func runLogoutCommand(ctx context.Context, client v1alpha1.Client, logger *slog.Logger) error {
	if err := client.Logout(ctx); err != nil {
		return errors.Wrap(err, "logout failed")
	}
	fmt.Println("✅ Successfully logged out")
	return nil
}

func runProjectsListCommand(ctx context.Context, client v1alpha1.Client, logger *slog.Logger) error {
	projects, err := client.ListProjects(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to list projects")
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	fmt.Printf("Found %d project(s):\n", len(projects))
	for i, project := range projects {
		fmt.Printf("%d. ID: %s, Name: %s, Description: %s, Kind: %s\n",
			i+1, project.ID, project.Name, project.Description, project.Kind)
	}
	return nil
}

func runProjectsCreateCommand(ctx context.Context, client v1alpha1.Client, logger *slog.Logger, scanner *bufio.Scanner) error {
	fmt.Print("Project name: ")
	if !scanner.Scan() {
		return errors.New("failed to read project name")
	}
	name := strings.TrimSpace(scanner.Text())
	if name == "" {
		return errors.New("project name cannot be empty")
	}

	fmt.Print("Project description: ")
	if !scanner.Scan() {
		return errors.New("failed to read project description")
	}
	description := strings.TrimSpace(scanner.Text())

	fmt.Print("Project kind (personal/shared) [personal]: ")
	if !scanner.Scan() {
		return errors.New("failed to read project kind")
	}
	kind := strings.TrimSpace(scanner.Text())
	if kind == "" {
		kind = "personal"
	}

	fmt.Print("Owner IDs (comma-separated, optional): ")
	if !scanner.Scan() {
		return errors.New("failed to read owner IDs")
	}
	ownerIdsInput := strings.TrimSpace(scanner.Text())
	var ownerIds []string
	if ownerIdsInput != "" {
		ownerIds = strings.Split(ownerIdsInput, ",")
		for i := range ownerIds {
			ownerIds[i] = strings.TrimSpace(ownerIds[i])
		}
	}

	fmt.Print("Owner group IDs (comma-separated, optional): ")
	if !scanner.Scan() {
		return errors.New("failed to read owner group IDs")
	}
	ownerGroupIdsInput := strings.TrimSpace(scanner.Text())
	var ownerGroupIds []string
	if ownerGroupIdsInput != "" {
		ownerGroupIds = strings.Split(ownerGroupIdsInput, ",")
		for i := range ownerGroupIds {
			ownerGroupIds[i] = strings.TrimSpace(ownerGroupIds[i])
		}
	}

	reqBody := generated.CreateProjectRequest{
		Name:          name,
		Description:   description,
		Kind:          generated.CreateProjectRequestKind(kind),
		OwnerIds:      ownerIds,
		OwnerGroupIds: ownerGroupIds,
	}

	if err := client.CreateProject(ctx, &reqBody); err != nil {
		return errors.Wrap(err, "failed to create project")
	}

	fmt.Println("✅ Project created successfully")
	return nil
}
