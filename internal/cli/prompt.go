package cli

import (
	"fmt"
	"strings"
)

// Prompts the user with "[y/N]" and returns true only on an explicit yes.
func promptConfirm(question string) bool {
	fmt.Printf("%s [y/N] ", question)
	var answer string
	fmt.Scanln(&answer) //nolint:errcheck
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}
