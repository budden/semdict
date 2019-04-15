package app

import (
	"net/http"

	"github.com/budden/semdict/pkg/user"

	"github.com/gin-gonic/gin"
)

type menuTemplateParams struct {
	IsLoggedIn bool
	Nickname   string
}

func menuPageHandler(c *gin.Context) {
	mtp := menuTemplateParams{
		IsLoggedIn: user.IsLoggedIn(c),
		Nickname:   "MyFriend"}
	c.HTML(http.StatusOK, "menu.t.html", mtp)
}
