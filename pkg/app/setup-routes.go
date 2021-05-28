package app

import (
	"github.com/budden/semdict/pkg/query"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

func setupRoutes(engine *gin.Engine) {
	engine.GET("/", homePageHandler)
	engine.GET("/menu", menuPageHandler)

	engine.GET("/profile", query.ProfilePageHandler)
	engine.GET("/profileedit", query.ProfileEditPageHandler)
	engine.POST("/profileeditsubmit", query.ProfileEditSubmitPageHandler)

	engine.GET("/changepasswordform", user.ChangePasswordFormPageHandler)
	engine.POST("/changepasswordsubmit", user.ChangePasswordSubmitPageHandler)

	engine.GET("/wordsearchform", query.WordSearchFormRouteHandler)
	engine.GET("/wordsearchresultform", query.WordSearchResultRouteHandler)
	engine.GET("/wordsearchquery", query.WordSearchQueryRouteHandler)
	// FIXME add reference from proposal to the origin
	engine.GET("/sensebyidview/:senseid", query.SenseByIdViewDirHandler)

	engine.GET("/senseedit/:senseid", query.SenseEditDirHandler)
	engine.POST("/senseeditsubmit", query.SenseEditSubmitPostHandler)
	// sensedeleteconfirm returns a form which asks if you really want to delete the item,
	// and, if yes, posts /sensedelete request
	engine.GET("/sensedeleteconfirm/:senseid", query.SenseDeleteConfirmRequestHandler)
	// sensedelete deletes sense w/o confirmation
	engine.POST("/sensedelete/:senseid", query.SenseDeleteRequestHandler)

	// sensenewedit accepts an «oword» query parameter
	engine.POST("/sensenewedit", query.SenseNewEditRequestHandler)
	engine.GET("/sensenewedit", query.SenseNewEditRequestHandler)
	engine.POST("/sensenewsubmit", query.SenseNewSubmitPostHandler)

	engine.GET("/lwsnewedit/:senseid/:languageid", query.LwsNewEditRequestHandler)
	engine.POST("/lwsnewsubmit", query.LwsNewSubmitPostHandler)
	// it is not good that we must know senseid to edit lws, but seem to be not fatal
	engine.GET("/lwsedit/:senseid/:languageid/:lwsid", query.LwsEditGetHandler)
	engine.POST("/lwseditsubmit", query.LwsEditSubmitPostHandler)
	engine.GET("/lwsdeleteconfirm/:senseid/:languageid/:lwsid", query.LwsDeleteConfirmRequestHandler)
	engine.POST("/lwsdelete/:lwsid", query.LwsDeleteRequestHandler)

	engine.GET("/registrationform", user.RegistrationFormPageHandler)
	engine.POST("/registrationsubmit", user.RegistrationSubmitPostHandler)
	engine.GET("/registrationconfirmation", user.RegistrationConfirmationPageHandler)

	engine.GET("/restorepasswordform", user.RestorePasswordFormPageHandler)
	engine.POST("/restorepasswordsubmit", user.RestorePasswordSubmitPageHandler)

	engine.GET("/loginform", user.LoginFormPageHandler)
	engine.POST("/loginsubmit", user.LoginSubmitPostHandler) // FIXME rename handler
	engine.GET("/logout", user.Logout)
	engine.Static("/static", *TemplateBaseDir+"static")

}
