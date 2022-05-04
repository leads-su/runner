package runner

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetMyUsername returns username for current user
func GetMyUsername() (string, error) {
	command := exec.Command("whoami")
	stdOut, stdErr := command.Output()
	if stdErr != nil {
		return "", stdErr
	}
	return strings.TrimSpace(string(stdOut)), nil
}

// HasSudoPrivileges checks whether specified user has sudo privileges
func HasSudoPrivileges(username string) (bool, error) {
	if len(strings.TrimSpace(username)) == 0 {
		myUsername, err := GetMyUsername()
		if err != nil {
			return false, err
		}
		username = myUsername
	}

	command := exec.Command("groups", username)
	stdOut, stdErr := command.Output()

	if stdErr != nil {
		return false, fmt.Errorf("user `%s` could not be found", username)
	}
	output := strings.TrimSpace(string(stdOut))

	if strings.Contains(output, "sudo") {
		return true, nil
	}

	return false, fmt.Errorf("user `%s` has no sudo privileges", username)
}
