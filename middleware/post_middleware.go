package middleware

import (
	"blog/models"
	"blog/utils"
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/pop/v5"
)

// PostGuardMiddleware - check for the ownership of the post
func PostGuardMiddleware(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		authUser := c.Value("authUser").(models.User)

		post := &models.Post{}

		db := c.Value("tx").(*pop.Connection)

		queryError := db.Eager().Find(post, c.Param("post_id"))
		errorResponse := utils.NewErrorResponse(http.StatusUnauthorized, "post", "Unauthorized access")

		if queryError != nil {
			return c.Render(http.StatusUnauthorized, render.JSON(errorResponse))
		}
		if strings.Compare(authUser.ID.String(), post.User.ID.String()) != 0 {
			return c.Render(http.StatusUnauthorized, render.JSON(errorResponse))
		}
		return next(c)
	}
}
