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
	engine.GET("/sensebycommonidview/:commonid", query.SenseByCommonidViewDirHandler)
	engine.GET("/senseedit/:commonid/:proposalid", query.SenseEditDirHandler)
	engine.POST("/senseproposaldelete/:proposalid", query.SenseProposalDeleteRequestHandler)
	engine.POST("/senseproposaladdform", query.SenseProposalAddFormPageHandler)

	engine.GET("/registrationform", user.RegistrationFormPageHandler)
	engine.POST("/registrationformsubmit", user.RegistrationFormSubmitPostHandler)
	engine.GET("/registrationconfirmation", user.RegistrationConfirmationPageHandler)

	engine.GET("/loginform", user.LoginFormPageHandler)
	engine.POST("/loginformsubmit", user.LoginFormSubmitPostHandler) // FIXME rename handler
	engine.GET("/logout", user.Logout)
	engine.Static("/static", *TemplateBaseDir+"static")

	engine.POST("/senseeditformsubmit", query.SenseEditFormSubmitPostHandler)
	engine.GET("/senseproposalslistform/:commonid", query.SenseAndProposalsListFormRouteHandler)
}