// Command atlas runs the Atlas HTTP service and its operator commands.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/prabhavalabs/atlas/internal/identity"
	"github.com/prabhavalabs/atlas/internal/platform/buildinfo"
	"github.com/prabhavalabs/atlas/internal/platform/config"
	appRuntime "github.com/prabhavalabs/atlas/internal/platform/runtime"
)

func main() {
	os.Exit(execute())
}

func execute() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, logger); err != nil {
		logger.Error("atlas command failed", "error", err)
		return 1
	}
	return 0
}

func run(ctx context.Context, arguments []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, logger *slog.Logger) error {
	if len(arguments) == 0 {
		return errors.New("a command is required: serve, migrate, import-fixture, create-admin, or version")
	}
	if arguments[0] == "version" {
		info := buildinfo.Current()
		_, err := fmt.Fprintf(stdout, "atlas %s (%s, %s)\n", info.Version, info.Commit, info.BuildTime)
		return err
	}
	if arguments[0] == "healthcheck" {
		return runHealthcheck(ctx)
	}

	configuration, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}
	switch arguments[0] {
	case "serve":
		if len(arguments) != 1 {
			return errors.New("serve does not accept arguments")
		}
		return appRuntime.Serve(ctx, configuration, logger)
	case "migrate":
		if len(arguments) != 1 {
			return errors.New("migrate does not accept arguments")
		}
		return appRuntime.Migrate(ctx, configuration)
	case "import-fixture":
		return runImportFixture(ctx, arguments[1:], configuration, stdout, stderr)
	case "create-admin":
		return runCreateAdmin(ctx, arguments[1:], configuration, stdin, stdout, stderr)
	default:
		return fmt.Errorf("unknown command %q", arguments[0])
	}
}

func runHealthcheck(ctx context.Context) error {
	healthURL := strings.TrimSpace(os.Getenv("ATLAS_HEALTH_URL"))
	if healthURL == "" {
		healthURL = "http://127.0.0.1:8080/health/live"
	}
	requestContext, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(requestContext, http.MethodGet, healthURL, nil)
	if err != nil {
		return errors.New("create health request")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.New("health endpoint unavailable")
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned status %d", response.StatusCode)
	}
	return nil
}

func runImportFixture(ctx context.Context, arguments []string, configuration config.Config, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("import-fixture", flag.ContinueOnError)
	flags.SetOutput(stderr)
	path := flags.String("file", "", "fixture JSON file")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *path == "" || flags.NArg() != 0 {
		return errors.New("import-fixture requires --file and no positional arguments")
	}
	file, err := os.Open(*path)
	if err != nil {
		return fmt.Errorf("open fixture: %w", err)
	}
	defer func() { _ = file.Close() }()
	result, err := appRuntime.ImportFixture(ctx, configuration, file)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "event=%s changed=%t\n", result.EventID, result.Changed)
	return err
}

func runCreateAdmin(ctx context.Context, arguments []string, configuration config.Config, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	flags := flag.NewFlagSet("create-admin", flag.ContinueOnError)
	flags.SetOutput(stderr)
	email := flags.String("email", "", "administrator email")
	displayName := flags.String("name", "", "administrator display name")
	role := flags.String("role", string(identity.RoleAdministrator), "administrator, editor, or viewer")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	if *email == "" || *displayName == "" || flags.NArg() != 0 {
		return errors.New("create-admin requires --email and --name")
	}
	passwordBytes, err := io.ReadAll(io.LimitReader(stdin, 2049))
	if err != nil {
		return fmt.Errorf("read password from stdin: %w", err)
	}
	if len(passwordBytes) > 2048 {
		return errors.New("password input exceeds 2048 bytes")
	}
	password := strings.TrimRight(string(passwordBytes), "\r\n")
	user, err := appRuntime.CreateAdmin(ctx, configuration, *email, *displayName, password, identity.Role(*role))
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(stdout, "created administrator %s with role %s\n", user.Email, user.Role)
	return err
}
