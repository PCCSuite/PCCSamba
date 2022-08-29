package lib

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResp struct {
	Status int
	Data   ErrorResponceData
}

func (e ErrorResp) Send(c echo.Context, args ...any) error {
	e.Data.Error = fmt.Sprintf(e.Data.Error, args...)
	return c.JSON(e.Status, e.Data)
}

type ErrorResponceData struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

var ErrorTokenRequired = ErrorResp{
	Status: http.StatusUnauthorized,
	Data: ErrorResponceData{
		Error: "Token required",
		Code:  1001,
	},
}

var ErrorInvalidAuthorization = ErrorResp{
	Status: http.StatusUnauthorized,
	Data: ErrorResponceData{
		Error: "Invalid Authorization",
		Code:  1002,
	},
}

var ErrorInvalidToken3 = ErrorResp{
	Status: http.StatusUnauthorized,
	Data: ErrorResponceData{
		Error: "Invalid Token",
		Code:  1003,
	},
}

var ErrorInvalidToken4 = ErrorResp{
	Status: http.StatusUnauthorized,
	Data: ErrorResponceData{
		Error: "Invalid Token",
		Code:  1004,
	},
}

var ErrorInvalidToken5 = ErrorResp{
	Status: http.StatusUnauthorized,
	Data: ErrorResponceData{
		Error: "Invalid Token",
		Code:  1005,
	},
}

var ErrorNoScope = ErrorResp{
	Status: http.StatusUnauthorized,
	Data: ErrorResponceData{
		Error: "Token isn't have scope",
		Code:  1006,
	},
}

var ErrorInsufficientScope = ErrorResp{
	Status: http.StatusUnauthorized,
	Data: ErrorResponceData{
		Error: "Insufficient Scope",
		Code:  1007,
	},
}

var ErrorInvalidRequest = ErrorResp{
	Status: http.StatusBadRequest,
	Data: ErrorResponceData{
		Error: "Invalid Request: Failed to parse",
		Code:  1101,
	},
}

var ErrorInvalidPasswordMode = ErrorResp{
	Status: http.StatusBadRequest,
	Data: ErrorResponceData{
		Error: "Invalid Request: Unknown password mode",
		Code:  1102,
	},
}

var ErrorInternalError = ErrorResp{
	Status: http.StatusInternalServerError,
	Data: ErrorResponceData{
		Error: "Internal error: %v",
		Code:  2001,
	},
}
