package middleware

import (
	"blog/models"
	"blog/utils"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v5"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func JWTMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {

		// get the bearer token from authorization header
		authorizationHeader := c.Request().Header.Get("Authorization")

		splitToken := strings.SplitAfter(authorizationHeader, "Bearer")
		var jwtToken string
		if len(authorizationHeader) >= 2 {
			jwtToken = strings.TrimSpace(splitToken[1])
		} else {
			unauthResponse := ErrorResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid JWT token",
			}

			return c.Render(http.StatusUnauthorized, render.JSON(unauthResponse))
		}
		claims := &jwt.StandardClaims{}
		jwtKey, readErr := utils.ReadJWTKey()

		if readErr != nil {
			return c.Error(http.StatusInternalServerError, readErr)
		}
		token, err := jwt.ParseWithClaims(jwtToken, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if !token.Valid || err != nil {
			unauthResponse := ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid JWT token"}

			return c.Render(http.StatusUnauthorized, render.JSON(unauthResponse))
		}
		// verify the email is available in DB
		database := c.Value("tx").(*pop.Connection)
		tokenUser := &models.User{}
		dbErr := database.Find(tokenUser, claims.ID)
		if dbErr != nil {
			unauthResponse := ErrorResponse{Code: http.StatusUnauthorized, Message: "Invalid User ID"}
			return c.Render(http.StatusUnauthorized, render.JSON(unauthResponse))
		}

		c.Set("authUser", *tokenUser)

		middlewareErr := next(c)
		// do some work after calling the next handler

		return middlewareErr
	}
}
