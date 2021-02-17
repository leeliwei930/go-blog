package actions

import (
	"blog/models"
	"blog/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gobuffalo/validate/v3/validators"
	"golang.org/x/crypto/bcrypt"
)

type LogInPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LogInResponse struct {
	AccessToken string      `json:"access_token"`
	User        models.User `json:"user"`
}

type RegisterResponse struct {
	Code string      `json:"code"`
	Data models.User `json:"data"`
}

type UserAuthResponse struct {
	Code string      `json:"code"`
	Data models.User `json:"data"`
}

type Claims struct {
	Email          string `json:"email"`
	StandardClaims jwt.StandardClaims
}

func (payload *LogInPayload) Validate() (*validate.Errors, error) {
	return validate.Validate(
		&validators.EmailIsPresent{Name: "Email", Field: payload.Email},
		&validators.StringIsPresent{Name: "Password", Field: payload.Password},
	), nil
}

// AttemptAuth - Attempt on authentication
func AttemptAuth(payload LogInPayload, db *pop.Connection) (bool, *models.User) {
	user := &models.User{}

	userNotFound := db.Where("email = ?", payload.Email).First(user)

	if userNotFound != nil {
		return false, user
	}

	passwordNotMatched := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))

	if passwordNotMatched != nil {
		return false, user
	}
	return true, user
}

// JwtAuthLogIn default implementation.
func JwtAuthLogIn(c buffalo.Context) error {
	request := &LogInPayload{}
	c.Bind(request)

	verrs, err := request.Validate()
	if err != nil {
		return c.Error(http.StatusInternalServerError, err)
	}
	db := c.Value("tx").(*pop.Connection)
	exist, user := AttemptAuth(*request, db)

	if !exist {
		verrs.Add("email", "The account credentials doesn't match with our database records")
		errorResponse := utils.NewValidationErrorResponse(http.StatusUnprocessableEntity, verrs.Errors)
		return c.Render(http.StatusUnprocessableEntity, r.JSON(errorResponse))
	}

	if verrs.HasAny() {
		errorResponse := utils.NewValidationErrorResponse(http.StatusUnprocessableEntity, verrs.Errors)
		return c.Render(http.StatusUnprocessableEntity, r.JSON(errorResponse))
	}

	tokenExpiration := &jwt.Time{
		Time: time.Now().Add(10080 * time.Minute),
	}
	claims := &jwt.StandardClaims{
		ExpiresAt: tokenExpiration,
		Issuer:    "buffalo-cms.api.dev",
		ID:        user.ID.String(),
	}

	tokenAlgo := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	key, keyErr := utils.ReadJWTKey()
	if keyErr != nil {
		c.Error(http.StatusInternalServerError, keyErr)
	}

	token, tokenSignedErr := tokenAlgo.SignedString(key)

	if tokenSignedErr != nil {
		c.Error(http.StatusInternalServerError, tokenSignedErr)
	}

	return c.Render(http.StatusOK, r.JSON(LogInResponse{
		AccessToken: token,
		User:        *user,
	}))
}

// RegisterUser - Create a user
func RegisterUser(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	request := &RegisterPayload{}

	c.Bind(request)

	verrs := validate.Validate(
		&validators.EmailIsPresent{Field: request.Email, Name: "email"},
		&validators.StringLengthInRange{Field: request.Name, Name: "name", Min: 3, Max: 255},
		&validators.StringLengthInRange{Field: request.Password, Name: "password", Min: 5, Max: 32},
	)

	existUser := &models.User{}
	dbError := tx.Where("email = ? ", request.Email).First(existUser)
	// if the db find a user
	if dbError == nil {
		verrs.Add("email", "The email has been taken.")
	}

	if verrs.HasAny() {
		errorResponse := utils.NewValidationErrorResponse(http.StatusUnprocessableEntity, verrs.Errors)
		return c.Render(http.StatusUnprocessableEntity, r.JSON(errorResponse))
	}
	user := &models.User{
		Email:    request.Email,
		Password: request.Password,
		Name:     request.Name,
	}
	_, createUserErr := user.Create(tx)

	if createUserErr != nil {
		errorResponse := utils.NewErrorResponse(http.StatusInternalServerError, "user", "There is a problem while creating a user please try again later")
		return c.Render(http.StatusInternalServerError, r.JSON(errorResponse))
	}

	response := RegisterResponse{
		Code: fmt.Sprintf("%d", http.StatusCreated),
		Data: *user,
	}
	return c.Render(http.StatusCreated, r.JSON(response))
}

func GetUser(c buffalo.Context) error {

	user := c.Value("authUser").(models.User)
	response := UserAuthResponse{
		Code: fmt.Sprintf("%d", http.StatusOK),
		Data: user,
	}
	return c.Render(http.StatusOK, r.JSON(response))
}
