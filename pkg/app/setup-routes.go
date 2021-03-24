package app

import (
	"github.com/budden/semdict/pkg/query"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

func setupRoutes(engine *gin.Engine) {
	engine.GET("/", homePageHandler)
	engine.GET("/menu", menuPageHandler)
	engine.GET("/wordsearchform", query.WordSearchFormRouteHandler)
	engine.GET("/wordsearchresultform", query.WordSearchResultRouteHandler)
	engine.GET("/wordsearchquery", query.WordSearchQueryRouteHandler)
	// FIXME add reference from proposal to the origin
	engine.GET("/sensebyidview/:senseid", query.SenseByIdViewDirHandler)

	engine.GET("/senseedit/:senseid", query.SenseEditDirHandler)
	// sensedeleteconfirm returns a form which asks if you really want to delete the item,
	// and, if yes, posts /sensedelete request
	engine.GET("/sensedeleteconfirm/:senseid", query.SenseDeleteConfirmRequestHandler)
	// sensedelete deletes sense w/o confirmation
	engine.POST("/sensedelete/:senseid", query.SenseDeleteRequestHandler)

	engine.POST("/sensenewedit", query.SenseNewEditRequestHandler)
	engine.POST("/sensenewsubmit", query.SenseNewSubmitPostHandler)

	engine.GET("/registrationform", user.RegistrationFormPageHandler)
	engine.POST("/registrationsubmit", user.RegistrationSubmitPostHandler)
	engine.GET("/registrationconfirmation", user.RegistrationConfirmationPageHandler)

	engine.GET("/loginform", user.LoginFormPageHandler)
	engine.POST("/loginsubmit", user.LoginSubmitPostHandler) // FIXME rename handler
	engine.GET("/logout", user.Logout)
	engine.Static("/static", *TemplateBaseDir+"static")

	engine.POST("/senseeditsubmit", query.SenseEditSubmitPostHandler)
}
