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
	engine.GET("/languageproposalslistform/:languageid", query.LanguageProposalsListFormRouteHandler)
	engine.GET("/wordsearchquery", query.WordSearchQueryRouteHandler)
	// FIXME add reference from proposal to the origin
	engine.GET("/sensebyidview/:senseid", query.SenseByIdViewDirHandler)

	engine.GET("/sensebycommonidview/:commonid", query.SenseByCommonidViewDirHandler)
	engine.GET("/senseedit/:commonid/:proposalid", query.SenseEditDirHandler)
	engine.POST("/senseproposaldelete/:proposalid", query.SenseProposalDeleteRequestHandler)
	engine.POST("/senseproposaladdform", query.SenseProposalAddFormPageHandler)

	engine.GET("/registrationform", user.RegistrationFormPageHandler)
	engine.POST("/registrationsubmit", user.RegistrationSubmitPostHandler)
	engine.GET("/registrationconfirmation", user.RegistrationConfirmationPageHandler)

	engine.GET("/loginform", user.LoginFormPageHandler)
	engine.POST("/loginsubmit", user.LoginSubmitPostHandler) // FIXME rename handler
	engine.GET("/logout", user.Logout)
	engine.Static("/static", *TemplateBaseDir+"static")

	engine.POST("/senseeditsubmit", query.SenseEditSubmitPostHandler)
	engine.GET("/senseproposalslistform/:commonid", query.SenseAndProposalsListFormRouteHandler)
}
