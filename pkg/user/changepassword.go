package user

import (
	"net/http"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"
	"github.com/gin-gonic/gin"
)

// ChangePasswordFormPageHandler renders a /changepasswordform page
func ChangePasswordFormPageHandler(c *gin.Context) {
	sduserID := GetSDUserIdOrZero(c)
	if sduserID > 0 {
		c.HTML(http.StatusOK, "changepasswordform.t.html", nil)
		return
	}
	email := c.Query("email")
	confirmationkey := c.Query("confirmationkey")
	if email != "" && confirmationkey != "" {
		d := &changePasswordData{
			Email:           email,
			Confirmationkey: confirmationkey,
		}
		processCheckConfirmationCodeWithDb(d)
		c.HTML(http.StatusOK, "changepasswordform.t.html", d)
		return
	}
	c.HTML(http.StatusOK, "general.t.html", shared.GeneralTemplateParams{Message: "Register or login."})
}

// ChangePasswordSubmitPageHandler processes a /changepasswordsubmit form post request
func ChangePasswordSubmitPageHandler(c *gin.Context) {
	pwd := c.PostForm("password1")
	if pwd != c.PostForm("password2") {
		apperror.Panic500If(apperror.ErrDummy, "Passwords don't match")
	}
	passwordErr := validatePassword(pwd)
	if passwordErr != nil {
		apperror.Panic500If(apperror.ErrDummy, "%s", passwordErr.Error())
	}
	salt, hash := SaltAndHashPassword(pwd)

	sduserID := GetSDUserIdOrZero(c)
	if sduserID > 0 {

		if err := processChangePasswordSubmitWithDb(&changePasswordData{
			Sduserid: sduserID,
			Salt:     salt,
			Hash:     hash,
		}); err == nil {
			c.Redirect(http.StatusFound, "/profile")
		} else {
			c.HTML(http.StatusOK,
				"general.t.html",
				shared.GeneralTemplateParams{Message: err.Error()})
		}
		return
	}
	email := c.PostForm("email")
	confirmationkey := c.PostForm("confirmationkey")
	if email != "" && confirmationkey != "" {
		if err := processChangePasswordSubmitWithDb(&changePasswordData{
			Salt:            salt,
			Hash:            hash,
			Email:           email,
			Confirmationkey: confirmationkey,
		}); err == nil {
			c.HTML(http.StatusOK,
				"general.t.html",
				shared.GeneralTemplateParams{Message: "Password is changed."})

		} else {
			c.HTML(http.StatusOK,
				"general.t.html",
				shared.GeneralTemplateParams{Message: err.Error()})
		}
		return
	}
	c.HTML(http.StatusOK, "general.t.html", shared.GeneralTemplateParams{Message: "Register or login."})
}

type changePasswordData struct {
	Sduserid int32
	Salt     string
	Hash     string

	Email           string
	Confirmationkey string
}

func processChangePasswordSubmitWithDb(d *changePasswordData) error {
	return sddb.WithTransaction(func(trans *sddb.TransactionType) (err error) {
		sddb.CheckDbAlive()
		if d.Sduserid > 0 {
			_, err = trans.Tx.NamedExec(
				`UPDATE sduser SET salt = :salt, hash = :hash  WHERE id = :sduserid;`,
				d)
		} else {
			rows, err1 := trans.Tx.NamedQuery(
				`select count(1)
from registrationattempt
WHERE nickname = (SELECT nickname FROM sduser WHERE registrationemail = :email)
  AND registrationemail = :email
  AND confirmationkey = :confirmationkey;`,
				d)
			if err1 != nil {
				return
			}
			var n int64
			rows.Next()
			if err1 := rows.Scan(&n); err1 == nil && n > 0 {
				rows.Close()
				_, err = trans.Tx.NamedExec(
					`DELETE FROM registrationattempt WHERE registrationemail = :email;`,
					d)
				if err != nil {
					return
				}
				_, err = trans.Tx.NamedExec(
					`UPDATE sduser SET salt = :salt, hash = :hash  WHERE registrationemail = :email;`,
					d)
			} else {
				rows.Close()
				return
			}
		}
		if err == nil {
			sddb.CheckDbAlive()
			err = trans.Tx.Commit()
		}
		return
	})
}

func processCheckConfirmationCodeWithDb(d *changePasswordData) {
	reply, err1 := sddb.NamedReadQuery(
		`
select count(1)
from registrationattempt
WHERE nickname = (SELECT nickname FROM sduser WHERE registrationemail = :email)
  AND registrationemail = :email
  AND confirmationkey = :confirmationkey;
`, &d)
	apperror.Panic500AndErrorIf(err1, "Failed confirmation link")
	defer sddb.CloseRows(reply)()
	for reply.Next() {
		var r int64
		err1 = reply.Scan(&r)
		if err1 == nil && r > 0 {
			return
		}
		apperror.Panic500AndErrorIf(err1, "Failed confirmation link")
	}
}
