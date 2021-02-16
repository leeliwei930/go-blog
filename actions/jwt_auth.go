package actions

import (
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
)

type LogInPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (payload *LogInPayload) Validate() (*validate.Errors, error) {
	return validate.Validate(
		&validators.EmailIsPresent{Name: "Email", Field: payload.Email},
		&validators.StringIsPresent{Name: "Password", Field: payload.Password},
	), nil
}

type LogInResponse struct {
	AccessToken string `json:"access_token"`
}

type Claims struct {
	Email          string `json:"email"`
	StandardClaims jwt.StandardClaims
}

func readJWTKey() ([]byte, error) {
	keyPath := envy.Get("JWT_KEY_PATH", "")

	content, error := ioutil.ReadFile(keyPath)

	return content, error
}

// JwtAuthLogIn default implementation.
func JwtAuthLogIn(c buffalo.Context) error {
	request := &LogInPayload{}
	c.Bind(request)

	verrs, err := request.Validate()
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}

	if verrs.HasAny() {
		errorResponse := NewValidationErrorResponse(http.StatusUnprocessableEntity, verrs.Errors)
		return c.Render(http.StatusUnprocessableEntity, r.JSON(errorResponse))
	}
	tokenExpiration := &jwt.Time{
		Time: time.Now().Add(10080 * time.Minute),
	}
	claims := &jwt.StandardClaims{
		ExpiresAt: tokenExpiration,
		Issuer:    "buffalo-cms.api.dev",
		ID:        request.Email,
	}

	tokenAlgo := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	key, keyErr := readJWTKey()
	if keyErr != nil {
		c.Error(http.StatusInternalServerError, keyErr)
	}

	token, tokenSignedErr := tokenAlgo.SignedString(key)

	if tokenSignedErr != nil {
		c.Error(http.StatusInternalServerError, tokenSignedErr)
	}

	return c.Render(http.StatusOK, r.JSON(LogInResponse{
		AccessToken: token,
	}))
}

// func JwtAuthVerify(c buffalo.Context) error {

// }
