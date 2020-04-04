package actions

import (
	"fmt"
	"log"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/nulls"
	"github.com/gobuffalo/pop"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/pkg/errors"
	"github.com/zulhfreelancer/buffalo_blog_with_oauth2/models"
)

func init() {
	gothic.Store = App().SessionStore

	goth.UseProviders(
		facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"), fmt.Sprintf("%s%s", App().Host, "/auth/facebook/callback")),
	)
}

func AuthCallback(c buffalo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		log.Printf("Callback error: %v\n", err)
		c.Flash().Add("danger", "Something went wrong")
		return c.Redirect(302, "/")
	}

	// Do something with the user, maybe register them/sign them in
	App().Logger.Info("User callback success")
	App().Logger.Info("Token expiry: ", user.ExpiresAt.Format("02-Jan-2006 03:04:05 PM"))
	tx := c.Value("tx").(*pop.Connection)
	q := tx.Where("provider = ? AND provider_id = ?", user.Provider, user.UserID)
	exists, err := q.Exists("users")
	if err != nil {
		return errors.WithStack(err)
	}

	// If exists in DB, load it
	u := &models.User{}
	if exists {
		if err = q.First(u); err != nil {
			return errors.WithStack(err)
		}
	}

	// If does not exist, create it
	u.Name = user.Name
	u.Provider = user.Provider
	u.ProviderID = user.UserID
	u.Email = nulls.NewString(user.Email)
	if err = tx.Save(u); err != nil {
		return errors.WithStack(err)
	}

	// Create session
	c.Session().Set("current_user_id", u.ID)
	if err = c.Session().Save(); err != nil {
		return errors.WithStack(err)
	}

	c.Flash().Add("success", "You have been logged in")
	return c.Redirect(302, "/")
}

func AuthDestroy(c buffalo.Context) error {
	c.Session().Clear()
	c.Flash().Add("success", "You have been logged out")
	return c.Redirect(302, "/")
}

func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid != nil {
			u := &models.User{}
			tx := c.Value("tx").(*pop.Connection)
			if err := tx.Find(u, uid); err != nil {
				return errors.WithStack(err)
			}
			c.Set("current_user", u)
		}
		return next(c)
	}
}

func Authorize(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if uid := c.Session().Get("current_user_id"); uid == nil {
			c.Flash().Add("danger", "Please login to proceed")
			return c.Redirect(302, "/")
		}
		return next(c)
	}
}
