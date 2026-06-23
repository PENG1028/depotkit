package postgres

import (
	"fmt"
	"os"
	"os/exec"
)

// SchemaDump runs pg_dump --schema-only and writes to the specified file.
func SchemaDump(user, password, host string, port int, database, outputFile string) error {
	dir := outputFile
	if idx := len(outputFile) - 1; idx >= 0 {
		// Extract directory from path
		for i := idx; i >= 0; i-- {
			if outputFile[i] == '/' || outputFile[i] == '\\' {
				dir = outputFile[:i]
				break
			}
		}
	}
	if dir != outputFile {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating schema directory: %w", err)
		}
	}

	cmd := exec.Command("pg_dump", "--schema-only",
		"-h", host,
		"-p", fmt.Sprintf("%d", port),
		"-U", user,
		"-d", database,
		"-f", outputFile,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_dump --schema-only failed: %w\n%s", err, string(output))
	}

	return nil
}
