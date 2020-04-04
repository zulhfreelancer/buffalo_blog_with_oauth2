package actions

import "github.com/gobuffalo/buffalo"

// HomeHandler is a default handler to serve
func HomeHandler(c buffalo.Context) error {
	// return c.Render(200, r.HTML("index.html"))

	// When user goes to "/", he will see the same content as at "/posts".
	// This is similar to `root` behaviour in Rails routes.
	pr := PostsResource{&buffalo.BaseResource{}}
	return pr.List(c)
}
