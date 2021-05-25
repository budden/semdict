package query

import (
	"net/http"
	"strconv"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/sddb"
	"github.com/budden/semdict/pkg/shared"
	"github.com/budden/semdict/pkg/user"
	"github.com/gin-gonic/gin"
)

// данные для БД
type profileEditSubmitDataType struct {
	Sduserid    int32
	Tlanguageid *int64
}

func ProfileEditSubmitPageHandler(c *gin.Context) {
	sduserID := user.GetSDUserIdOrZero(c)
	if sduserID > 0 {
		d := &profileEditSubmitDataType{}
		extractDataFromProfileEditSubmitRequest(c, d)
		d.Sduserid = sduserID
		writeProfileInDb(d)
		c.Redirect(http.StatusFound, "/profile")
		return
	}
	c.HTML(http.StatusOK, "general.t.html", shared.GeneralTemplateParams{Message: "Зарегистрируйтесь или войдите в систему."})
}

func extractDataFromProfileEditSubmitRequest(c *gin.Context, d *profileEditSubmitDataType) {
	if id, err := strconv.ParseInt(c.PostForm("tlanguage"), 10, 64); err == nil && id > 0 {
		d.Tlanguageid = &id
	}
}

func writeProfileInDb(d *profileEditSubmitDataType) {
	err := sddb.WithTransaction(func(trans *sddb.TransactionType) (err error) {
		rows, err1 := trans.Tx.NamedQuery(
			`SELECT count(1) FROM sduser_profile WHERE id = :sduserid;`,
			d)
		if err1 != nil {
			return
		}
		var n int64
		rows.Next()
		if err1 := rows.Scan(&n); err1 == nil && n > 0 {
			rows.Close()
			_, err = trans.Tx.NamedExec(
				`UPDATE sduser_profile SET favorite_tlanguageid = :tlanguageid WHERE id = :sduserid;`,
				d)
		} else {
			rows.Close()
			_, err = trans.Tx.NamedExec(
				`INSERT INTO  sduser_profile(id, favorite_tlanguageid)VALUES (:sduserid, :tlanguageid);`,
				d)
			return
		}
		return
	})
	apperror.Panic500AndErrorIf(err, "Не удалось обновить профиль, извините")
	return
}
