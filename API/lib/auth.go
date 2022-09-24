package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	req, err := http.NewRequest("POST", TokenInfo.IntrospectURL, strings.NewReader(body.Encode()))
	if err != nil {
		err = fmt.Errorf("failed to create introspection request: %w", err)
		log.Print(err)
		return nil, ErrorInternalError.Send(c, err)
	}

	req.Header.Set("Authorization", TokenInfo.IntrospectAuth)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to request introspection: %w", err)
		log.Print(err)
		return nil, ErrorInternalError.Send(c, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = errors.New("not success code while introspection: " + resp.Status)
		log.Print(err)
		return nil, ErrorInternalError.Send(c, err)
	}

	respRaw, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read introspection result: %w", err)
		log.Print(err)
		return nil, ErrorInternalError.Send(c, err)
	}

	result := IntrospectionResult{}
	err = json.Unmarshal(respRaw, &result)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal introspection result: %w", err)
		log.Print(err)
		return nil, ErrorInternalError.Send(c, err)
	}

	if !result.Active {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		return nil, ErrorInvalidToken3.Send(c)
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
