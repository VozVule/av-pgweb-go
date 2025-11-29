package migrate

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	defaultAtlasBinary = "atlas"
	defaultDir         = "migrations"
)

// Run executes atlas migrations using the configured directory and connection URL.
func Run(ctx context.Context, databaseURL string) error {
	if databaseURL == "" {
		return fmt.Errorf("database URL is required for migrations")
	}

	bin := os.Getenv("PGWEB_ATLAS_BIN")
	if bin == "" {
		bin = defaultAtlasBinary
	}

	dir := os.Getenv("PGWEB_MIGRATIONS_DIR")
	if dir == "" {
		dir = defaultDir
	}

	atlasDir, err := normalizeDir(dir)
	if err != nil {
		return err
	}

	args := []string{"migrate", "apply", "--dir", atlasDir, "--url", databaseURL}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("atlas migrate apply failed: %w", err)
	}

	return nil
}

func normalizeDir(dir string) (string, error) {
	if strings.HasPrefix(dir, "file://") {
		return dir, nil
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve migrations dir: %w", err)
	}

	return "file://" + abs, nil
}
