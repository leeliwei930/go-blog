package actions

import (
	"blog/models"
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
)

// PostsResponse - Posts collection response body
type PostsResponse struct {
	Code string       `json:"code"`
	Data models.Posts `json:"data"`
}

// PostResponse - Single post object response body
type PostResponse struct {
	Code string       `json:"code"`
	Data *models.Post `json:"data"`
}

// PostDeletedResponse - Response body when post get removed
type PostDeletedResponse struct {
	Code string       `json:"code"`
	Data *models.Post `json:"data"`
}

// ErrorResponse - Validation errors or any others errors response body
type ErrorResponse struct {
	Code   string              `json:"code"`
	Errors map[string][]string `json:"errors,omitempty"`
	Error  map[string]string   `json:"error,omitempty"`
}

// NewErrorResponse - Single based error response body
func NewErrorResponse(statusCode int, field string, message string) ErrorResponse {

	var errorFields = map[string]string{}
	errorFields[field] = message

	return ErrorResponse{
		Code:  fmt.Sprintf("%d", statusCode),
		Error: errorFields,
	}
}

// NewValidationErrorResponse - Validation fields based errors
func NewValidationErrorResponse(statusCode int, verrs map[string][]string) ErrorResponse {
	return ErrorResponse{
		Code:   fmt.Sprintf("%d", statusCode),
		Errors: verrs,
	}
}

// PostList default implementation.
func ListPost(c buffalo.Context) error {

	db := c.Value("tx").(*pop.Connection)

	posts := &models.Posts{}

	query := db.PaginateFromParams(c.Params())

	if err := query.Order("created_at desc").All(posts); err != nil {
		return err
	}

	response := PostsResponse{
		Code: fmt.Sprintf("%d", http.StatusOK),
		Data: *posts,
	}
	c.Logger().Debug(c.Value("email"))
	return c.Render(http.StatusOK, r.JSON(response))
}

// CreatePost - Validate and create a Post
func CreatePost(c buffalo.Context) error {

	post := &models.Post{}
	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}
	db := c.Value("tx").(*pop.Connection)

	validationErrors, err := db.ValidateAndCreate(post)
	if err != nil {
		return errors.WithStack(err)
	}

	if validationErrors.HasAny() {

		errResponse := NewValidationErrorResponse(
			http.StatusUnprocessableEntity, validationErrors.Errors,
		)
		return c.Render(http.StatusUnprocessableEntity, r.JSON(errResponse))
	}

	postResponse := PostResponse{
		Code: fmt.Sprintf("%d", http.StatusCreated),
		Data: post,
	}
	return c.Render(http.StatusCreated, r.JSON(postResponse))
}

// ShowPost - update the post based on the given ID
func ShowPost(c buffalo.Context) error {
	database := c.Value("tx").(*pop.Connection)

	post := &models.Post{}

	if txErr := database.Find(post, c.Param("post_id")); txErr != nil {

		notFoundResponse := NewErrorResponse(
			http.StatusNotFound,
			"post_id",
			fmt.Sprintf("The requested post %s is removed or move to somewhere else.", c.Param("post_id")),
		)
		return c.Render(http.StatusNotFound, r.JSON(notFoundResponse))
	}

	postResponse := PostResponse{
		Code: fmt.Sprintf("%d", http.StatusOK),
		Data: post,
	}
	return c.Render(http.StatusOK, r.JSON(postResponse))
}

// UpdatePost - Update a single post
func UpdatePost(c buffalo.Context) error {

	post := &models.Post{}
	database := c.Value("tx").(*pop.Connection)
	// retrieve the existing record
	if txErr := database.Find(post, c.Param("post_id")); txErr != nil {

		notFoundResponse := NewErrorResponse(
			http.StatusNotFound,
			"post_id",
			fmt.Sprintf("The requested post %s is removed or move to somewhere else.", c.Param("post_id")),
		)
		return c.Render(http.StatusNotFound, r.JSON(notFoundResponse))
	}
	// bind the form input
	if bindErr := c.Bind(post); bindErr != nil {
		emptyBodyResponse := NewErrorResponse(
			http.StatusUnprocessableEntity,
			"body",
			"The request body cannot be empty",
		)
		return c.Render(http.StatusUnprocessableEntity, r.JSON(emptyBodyResponse))
	}
	validationErrors, err := database.ValidateAndUpdate(post)
	if err != nil {
		return errors.WithStack(err)
	}

	if validationErrors.HasAny() {
		errResponse := ErrorResponse{
			Code:   fmt.Sprintf("%d", http.StatusUnprocessableEntity),
			Errors: validationErrors.Errors,
		}
		return c.Render(http.StatusUnprocessableEntity, r.JSON(errResponse))
	}

	response := PostResponse{
		Code: fmt.Sprintf("%d", http.StatusOK),
		Data: post,
	}

	return c.Render(http.StatusOK, r.JSON(response))
}

func DeletePost(c buffalo.Context) error {
	post := &models.Post{}
	database := c.Value("tx").(*pop.Connection)

	txErr := database.Find(post, c.Param("post_id"))
	if txErr != nil {
		notFoundResponse := NewErrorResponse(
			http.StatusNotFound,
			"post_id",
			fmt.Sprintf("The requested post %s is removed or move to somewhere else.", c.Param("post_id")),
		)
		return c.Render(http.StatusNotFound, r.JSON(notFoundResponse))
	}

	if deleteErr := database.Destroy(post); deleteErr != nil {
		deleteErrResponse := NewErrorResponse(
			http.StatusInternalServerError,
			"post",
			fmt.Sprintf("Unable to delete the post with id %s", c.Param("post_id")),
		)

		return c.Render(http.StatusInternalServerError, r.JSON(deleteErrResponse))
	}

	deleteSuccessResponse := PostDeletedResponse{
		Code: fmt.Sprintf("%d", http.StatusOK),
		Data: post,
	}
	return c.Render(http.StatusOK, r.JSON(deleteSuccessResponse))
}
