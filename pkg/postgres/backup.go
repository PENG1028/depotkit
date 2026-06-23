package postgres

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Backup creates a pg_dump -Fc backup file.
func Backup(user, password, host string, port int, database, backupDir string) (string, error) {
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("creating backup directory: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := timestamp + ".dump"
	filePath := filepath.Join(backupDir, filename)

	cmd := exec.Command("pg_dump", "-Fc",
		"-h", host,
		"-p", fmt.Sprintf("%d", port),
		"-U", user,
		"-d", database,
		"-f", filePath,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pg_dump failed: %w\n%s", err, string(output))
	}

	return filePath, nil
}

// Restore restores a backup file using pg_restore.
func Restore(user, password, host string, port int, database, backupFile string, dropExisting bool) error {
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	args := []string{
		"-h", host,
		"-p", fmt.Sprintf("%d", port),
		"-U", user,
		"-d", database,
	}

	if dropExisting {
		args = append(args, "--clean", "--if-exists")
	}

	args = append(args, backupFile)

	cmd := exec.Command("pg_restore", args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("pg_restore failed: %w\n%s", err, string(output))
	}

	return nil
}
