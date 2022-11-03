package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PCCSuite/PCCSamba/SambaAPI/lib"
	"github.com/PCCSuite/PCCSamba/SambaAPI/lib/db"
	"github.com/PCCSuite/PCCSamba/SambaAPI/lib/samba"
	"github.com/labstack/echo/v4"
)

type GetPasswordResponce struct {
	Mode lib.PasswordMode `json:"mode"`
	Data string           `json:"data,omitempty"`
}

func GetPassword(c echo.Context) error {
	auth, err := lib.CheckToken(c)
	if err != nil {
		log.Print("Failed to send error: ", err)
	}
	if auth == nil {
		log.Print("Auth failed: ", err)
		return err
	}
	userdata, err := db.GetData(auth.Username)
	var password string
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			userdata, password, err = initUser(auth.Username)
			if err != nil {
				return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to init user: ", err))
			}
		} else {
			return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to get userdata: ", err))
		}
	}
	err = checkGroup(auth.Username, auth.ResourceAccess[lib.TokenInfo.Client].Roles)
	if err != nil {
		log.Print("Failed to check groups: ", err)
		return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to check groups: ", err))
	}
	switch userdata.Mode {
	case lib.PasswordModeDynamic:
		if password == "" {
			password = lib.GeneratePassword()
			msg, err := samba.SetPassword(auth.Username, password)
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					msg = err.Error()
				}
				return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to set password in samba: ", msg))
			}
		}
		return c.JSON(http.StatusOK, GetPasswordResponce{
			Mode: userdata.Mode,
			Data: password,
		})
	case lib.PasswordModeStaticUnstored, lib.PasswordModeStaticEncrypted, lib.PasswordModeStaticPlain:
		return c.JSON(http.StatusOK, GetPasswordResponce{
			Mode: userdata.Mode,
			Data: userdata.Data,
		})
	default:
		return lib.ErrorInternalError.Send(c, fmt.Sprintf("Unknown password mode: %d", userdata.Mode))
	}
}

var homesPath = os.Getenv("PCC_SAMBAAPI_HOMES_FILEPATH")

// return user data and password
func initUser(user string) (*db.UserData, string, error) {
	password := lib.GeneratePassword()
	_, err := samba.AddUser(user, password)
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// if user exists, skip add
			password = ""
		} else {
			return nil, "", fmt.Errorf("failed to add user to samba: %w", err)
		}
	}
	uid, err := samba.GetUID(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get uid: %w", err)
	}
	homes := filepath.Join(homesPath, user)
	err = os.Mkdir(homes, 0700)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return nil, "", fmt.Errorf("failed to make homes: %w", err)
	}
	err = filepath.WalkDir(homes, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		err = os.Chown(path, uid, uid)
		if err != nil {
			return err
		}
		return os.Chmod(path, 0700)
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to set permission: %w", err)
	}
	data := &db.UserData{
		ID:   user,
		Mode: lib.PasswordModeDynamic,
		Data: "",
	}
	err = db.AddUser(data)
	return data, password, err
}

var roleGroups = strings.Split(os.Getenv("PCC_SAMBA_ROLE_GROUPS"), " ")

func checkGroup(user string, roles []string) error {
	sambaGroups, err := samba.GetUserGroups(user)
	if err != nil {
		return fmt.Errorf("failed to get samba group: %w", err)
	}
	for _, group := range roleGroups {
		inSamba := false
		for _, v := range sambaGroups {
			if v == group {
				inSamba = true
				break
			}
		}
		inRole := false
		for _, v := range roles {
			if v == group {
				inRole = true
				break
			}
		}
		if inRole {
			if !inSamba {
				msg, err := samba.AddUserToGroup(user, group)
				if err != nil {
					if _, ok := err.(*exec.ExitError); ok {
						err = fmt.Errorf("%s: %w", msg, err)
					}
					return fmt.Errorf("failed to add user to group %v: %w", group, err)
				}
			}
		} else {
			if inSamba {
				msg, err := samba.RemoveUserFromGroup(user, group)
				if err != nil {
					if _, ok := err.(*exec.ExitError); ok {
						err = fmt.Errorf("%s: %w", msg, err)
					}
					return fmt.Errorf("failed to remove user from group %v: %w", group, err)
				}
			}
		}
	}
	return nil
}

type SetPasswordRequest struct {
	Mode      lib.PasswordMode `json:"mode"`
	Password  string           `json:"password,omitempty"`
	Encrypted string           `json:"encrypted,omitempty"`
}

func SetPassword(c echo.Context) error {
	auth, err := lib.CheckToken(c)
	if err != nil {
		log.Print("Failed to check token: ", err)
	}
	if auth == nil {
		log.Print("Auth failed: ", err)
		return err
	}

	data := SetPasswordRequest{}
	err = c.Bind(&data)
	if err != nil {
		return lib.ErrorInvalidRequest.Send(c)
	}
	userdata := db.UserData{
		ID:   auth.Username,
		Mode: data.Mode,
	}
	switch data.Mode {
	case lib.PasswordModeDynamic:
		break
	case lib.PasswordModeStaticPlain, lib.PasswordModeStaticEncrypted, lib.PasswordModeStaticUnstored:
		msg, err := samba.SetPassword(auth.Username, data.Password)
		if err != nil {
			if _, ok := err.(*exec.ExitError); !ok {
				msg = err.Error()
			}
			return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to set password in samba: ", msg))
		}
		userdata.Data = data.Password
	default:
		return lib.ErrorInvalidPasswordMode.Send(c)
	}
	switch data.Mode {
	case lib.PasswordModeStaticPlain:
		userdata.Data = data.Password
	case lib.PasswordModeStaticEncrypted:
		userdata.Data = data.Encrypted
	}
	err = db.SetData(&userdata)
	if err != nil {
		log.Print("Failed to set userdata: ", err)
		return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to set userdata: ", err))
	}
	return c.NoContent(http.StatusNoContent)
}
