package lib

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

var ErrNoUserID = errors.New("token isn't contain username")

var TokenInfo = struct {
	IntrospectURL  string
	IntrospectAuth string
	Realm          string
	Client         string
}{
	IntrospectURL:  os.Getenv("PCC_SAMBAAPI_TOKEN_INTROSPECT_URL"),
	IntrospectAuth: os.Getenv("PCC_SAMBAAPI_TOKEN_INTROSPECT_AUTH"),
	Realm:          os.Getenv("PCC_SAMBAAPI_TOKEN_REALM"),
	Client:         os.Getenv("PCC_SAMBAAPI_TOKEN_CLIENT"),
}

var person = struct {
	Name string
	Age  int
}{
	Name: "tenntenn",
	Age:  30,
}

var TokenClientId = os.Getenv("PCC_SAMBAAPI_TOKEN_CLIENT")

type IntrospectionResult struct {
	Active      bool     `json:"active"`
	Audience    []string `json:"aud"`
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	Scope    string `json:"scope"`
	Username string `json:"username"`
}

func CheckToken(c echo.Context) (*IntrospectionResult, error) {

	authorization := c.Request().Header.Get("Authorization")
	if authorization == "" {
		c.Response().Header().Add("WWW-Authenticate", "Bearer realm=\""+TokenInfo.Realm+"\"")
		return nil, ErrorTokenRequired.Send(c)
	}
	if strings.HasPrefix(authorization, "Bearer ") {
		authorization = strings.TrimPrefix(authorization, "Bearer ")
	} else {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_request\"")
		return nil, ErrorInvalidAuthorization.Send(c)
	}

	body := url.Values{}
	body.Add("token", authorization)
	resp, err := http.NewRequest("POST", TokenInfo.IntrospectURL, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, ErrorInternalError.Send(c, err)
	}

	respRaw := make([]byte, 8192)
	i, err := resp.Body.Read(respRaw)
	if err != nil {
		log.Printf("Failed to read introspection result: %v", err)
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		return nil, ErrorInvalidToken3.Send(c)
	}

	result := IntrospectionResult{}
	err = json.Unmarshal(respRaw[:i], &result)
	if err != nil {
		log.Printf("Failed to unmarshal introspection result: %v", err)
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		return nil, ErrorInvalidToken4.Send(c)
	}

	if !result.Active {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		return nil, ErrorInvalidToken5.Send(c)
	}

	scope_ok := false
	for _, v := range strings.Split(result.Scope, " ") {
		if v == "samba" {
			scope_ok = true
			break
		}
	}
	if !scope_ok {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"insufficient_scope\"")
		return nil, ErrorInsufficientScope.Send(c)
	}
	return &result, nil
}
