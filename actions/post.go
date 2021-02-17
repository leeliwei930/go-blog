package actions

import (
	"blog/models"
	"blog/utils"
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"
)

// PostsResponse - Posts collection response body
type PostsResponse struct {
	Code string        `json:"code"`
	Data models.Posts  `json:"data"`
	Meta pop.Paginator `json:"meta"`
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

// ListPost - list a collection of post with user
func ListPost(c buffalo.Context) error {

	db := c.Value("tx").(*pop.Connection)

	posts := &models.Posts{}

	query := db.PaginateFromParams(c.Params())

	if err := query.Order("created_at desc").Eager().All(posts); err != nil {

		errorResponse := utils.NewErrorResponse(http.StatusInternalServerError, "user", "There is a problem while loading the relationship user")
		return c.Render(http.StatusInternalServerError, r.JSON(errorResponse))
	}

	response := PostsResponse{
		Code: fmt.Sprintf("%d", http.StatusOK),
		Data: *posts,
		Meta: *query.Paginator,
	}
	c.Logger().Debug(c.Value("email"))
	return c.Render(http.StatusOK, r.JSON(response))
}

// CreatePost - Validate and create a Post
func CreatePost(c buffalo.Context) error {
	authUser := c.Value("authUser").(models.User)
	post := &models.Post{}
	if err := c.Bind(post); err != nil {
		return errors.WithStack(err)
	}
	db := c.Value("tx").(*pop.Connection)
	post.UserID = authUser.ID
	post.User = &authUser
	validationErrors, err := db.Eager().ValidateAndCreate(post)
	if err != nil {
		return errors.WithStack(err)
	}

	if validationErrors.HasAny() {

		errResponse := utils.NewValidationErrorResponse(
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

	if txErr := database.Eager().Find(post, c.Param("post_id")); txErr != nil {

		notFoundResponse := utils.NewErrorResponse(
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
	authUser := c.Value("authUser").(models.User)
	post := &models.Post{}
	database := c.Value("tx").(*pop.Connection)
	// retrieve the existing record
	if txErr := database.Find(post, c.Param("post_id")); txErr != nil {

		notFoundResponse := utils.NewErrorResponse(
			http.StatusNotFound,
			"post_id",
			fmt.Sprintf("The requested post %s is removed or move to somewhere else.", c.Param("post_id")),
		)
		return c.Render(http.StatusNotFound, r.JSON(notFoundResponse))
	}
	// bind the form input
	if bindErr := c.Bind(post); bindErr != nil {
		emptyBodyResponse := utils.NewErrorResponse(
			http.StatusUnprocessableEntity,
			"body",
			"The request body cannot be empty",
		)
		return c.Render(http.StatusUnprocessableEntity, r.JSON(emptyBodyResponse))
	}
	post.UserID = authUser.ID
	validationErrors, err := database.ValidateAndUpdate(post)
	if err != nil {
		return errors.WithStack(err)
	}

	if validationErrors.HasAny() {
		errResponse := utils.NewValidationErrorResponse(
			http.StatusUnprocessableEntity,
			validationErrors.Errors,
		)
		return c.Render(http.StatusUnprocessableEntity, r.JSON(errResponse))
	}

	response := PostResponse{
		Code: fmt.Sprintf("%d", http.StatusOK),
		Data: post,
	}

	return c.Render(http.StatusOK, r.JSON(response))
}

// DeletePost - Remove a post based on ID
func DeletePost(c buffalo.Context) error {
	post := &models.Post{}
	database := c.Value("tx").(*pop.Connection)

	txErr := database.Find(post, c.Param("post_id"))
	if txErr != nil {
		notFoundResponse := utils.NewErrorResponse(
			http.StatusNotFound,
			"post_id",
			fmt.Sprintf("The requested post %s is removed or move to somewhere else.", c.Param("post_id")),
		)
		return c.Render(http.StatusNotFound, r.JSON(notFoundResponse))
	}

	if deleteErr := database.Destroy(post); deleteErr != nil {
		deleteErrResponse := utils.NewErrorResponse(
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
