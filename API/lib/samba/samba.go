package samba

import (
	"os/exec"
)

func execSambaTool(args []string) (string, error) {
	cmd := exec.Command("samba-tool", args...)
	out, err := cmd.CombinedOutput()

	return string(out), err
}

func SetPassword(username, password string) (string, error) {
	return execSambaTool([]string{"user", "setpassword", "--newpassword", password, username})
}

func AddUser(username, password string) (string, error) {
	return execSambaTool([]string{"user", "add", username, password})
}

func AddUserToGroup(username, group string) (string, error) {
	return execSambaTool([]string{"group", "addmembers", group, username})
}
