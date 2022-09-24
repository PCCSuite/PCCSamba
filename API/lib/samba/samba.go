package samba

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
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

func RemoveUserFromGroup(username, group string) (string, error) {
	return execSambaTool([]string{"group", "removemembers", group, username})
}

func GetUserGroups(username string) (groups []string, err error) {
	result, err := execSambaTool([]string{"user", "getgroups", username})
	if err != nil {
		return
	}
	for _, v := range strings.Split(result, "\n") {
		v = strings.TrimSpace(v)
		if len(v) != 0 {
			groups = append(groups, v)
		}
	}
	return
}

var ErrInvalidResult = errors.New("process result invalid")

func GetUID(username string) (int, error) {
	cmd := exec.Command("wbinfo", "-i", username)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return -1, err
	}
	dataRaw := string(out)
	dataSplit := strings.Split(dataRaw, ":")
	if len(dataSplit) != 7 {
		return -1, ErrInvalidResult
	}
	return strconv.Atoi(dataSplit[2])
}
