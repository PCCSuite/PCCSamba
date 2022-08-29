package lib

import (
	"context"
	"errors"
	"log"
	"os"
	"strings"

	"github.com/Nerzal/gocloak/v11"
	"github.com/labstack/echo/v4"
)

var ErrNoUserID = errors.New("token isn't contain username")

type KeyCloak struct {
	client       gocloak.GoCloak
	realm        string
	clientId     string
	clientSecret string
}

var keycloak = KeyCloak{
	client:       gocloak.NewClient(os.Getenv("PCC_SAMBAAPI_KEYCLOAK_HOST"), gocloak.SetAuthAdminRealms("admin/realms"), gocloak.SetAuthRealms("realms")),
	realm:        os.Getenv("PCC_SAMBAAPI_KEYCLOAK_REALM"),
	clientId:     os.Getenv("PCC_SAMBAAPI_KEYCLOAK_CLIENT_ID"),
	clientSecret: os.Getenv("PCC_SAMBAAPI_KEYCLOAK_CLIENT_SECRET"),
}

func CheckToken(c echo.Context) (id string, err error) {

	authorization := c.Request().Header.Get("Authorization")
	if authorization == "" {
		c.Response().Header().Add("WWW-Authenticate", "Bearer realm=\""+keycloak.realm+"\"")
		return "", ErrorTokenRequired.Send(c)
	}
	if strings.HasPrefix(authorization, "Bearer ") {
		authorization = strings.TrimPrefix(authorization, "Bearer ")
	} else {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_request\"")
		return "", ErrorInvalidAuthorization.Send(c)
	}

	jwt, claims, err := keycloak.client.DecodeAccessToken(context.Background(), authorization, keycloak.realm)
	if err != nil {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		log.Printf("Failed to decode token: %s error: %v", authorization, err)
		return "", ErrorInvalidToken3.Send(c)
	}

	if !jwt.Valid {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		return "", ErrorInvalidToken4.Send(c)
	}

	if claims.Valid() != nil {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"invalid_token\"")
		return "", ErrorInvalidToken5.Send(c)
	}
	scope, ok := (*claims)["scope"].(string)
	if !ok {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"insufficient_scope\"")
		return "", ErrorNoScope.Send(c)
	}
	scope_ok := false
	for _, v := range strings.Split(scope, " ") {
		if v == "samba" {
			scope_ok = true
			break
		}
	}
	if !scope_ok {
		c.Response().Header().Add("WWW-Authenticate", "Bearer error=\"insufficient_scope\"")
		return "", ErrorInsufficientScope.Send(c)
	}
	username, ok := (*claims)["preferred_username"].(string)
	if !ok {
		return "", ErrNoUserID
	}
	return username, nil
}
