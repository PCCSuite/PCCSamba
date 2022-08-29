package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"

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
	user, err := lib.CheckToken(c)
	if err != nil {
		log.Print("Failed to check token: ", err)
	}
	if user == "" {
		log.Print("Auth failed: ", err)
		return err
	}
	userdata, err := db.GetData(user)
	var password string
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			password = lib.GeneratePassword()
			msg, err := samba.AddUser(user, password)
			if err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					msg = err.Error()
				}
				return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to adduser to samba: ", msg))
			}
			db.AddUser(user)
		} else {
			log.Print("Failed to get userdata: ", err)
			return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to get userdata: ", err))
		}
	}
	switch userdata.Mode {
	case lib.PasswordModeDynamic:
		if password == "" {
			password = lib.GeneratePassword()
			msg, err := samba.SetPassword(user, password)
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

type SetPasswordRequest struct {
	Mode      lib.PasswordMode `json:"mode"`
	Password  string           `json:"password,omitempty"`
	Encrypted string           `json:"encrypted,omitempty"`
}

func SetPassword(c echo.Context) error {
	user, err := lib.CheckToken(c)
	if err != nil {
		log.Print("Failed to check token: ", err)
	}
	if user == "" {
		log.Print("Auth failed: ", err)
		return err
	}

	data := SetPasswordRequest{}
	err = c.Bind(&data)
	if err != nil {
		return lib.ErrorInvalidRequest.Send(c)
	}
	userdata := db.UserData{
		ID:   user,
		Mode: data.Mode,
	}
	switch data.Mode {
	case lib.PasswordModeDynamic:
		break
	case lib.PasswordModeStaticPlain:
		samba.SetPassword(user, data.Password)
		userdata.Data = data.Password
	case lib.PasswordModeStaticEncrypted:
		samba.SetPassword(user, data.Password)
		userdata.Data = data.Encrypted
	case lib.PasswordModeStaticUnstored:
		samba.SetPassword(user, data.Password)
	default:
		return lib.ErrorInvalidPasswordMode.Send(c)
	}
	err = db.SetData(&userdata)
	if err != nil {
		log.Print("Failed to set userdata: ", err)
		return lib.ErrorInternalError.Send(c, fmt.Sprint("Failed to set userdata: ", err))
	}
	return c.NoContent(http.StatusNoContent)
}
