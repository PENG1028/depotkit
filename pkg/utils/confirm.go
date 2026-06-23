package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm prompts the user for confirmation. Returns true if the user confirms.
func Confirm(prompt string) bool {
	fmt.Printf("\n%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

// ConfirmProjectName prompts the user to type the project name for destructive operations.
// Returns true if the typed name matches.
func ConfirmProjectName(projectName string) bool {
	fmt.Printf("\n⚠  DANGER: This operation will delete data for project '%s'.\n", projectName)
	fmt.Printf("Type the project name to confirm: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	return input == projectName
}
