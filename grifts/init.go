package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/zulhfreelancer/blog_with_oauth2/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
