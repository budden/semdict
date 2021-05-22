package user

import (
	"fmt"
	"html"
	"net/http"
	"net/url"

	"github.com/budden/semdict/pkg/sddb"

	"github.com/budden/semdict/pkg/apperror"
	"github.com/budden/semdict/pkg/shared"
	"github.com/gin-gonic/gin"
)

// RegistrationFormPageHandler отображает страницу /registrationform
func RegistrationFormPageHandler(c *gin.Context) {
	EnsureNotLoggedIn(c)
	c.HTML(http.StatusOK,
		"registrationform.t.html",
		shared.GeneralTemplateParams{Message: "Форма поиска"})
}

// RegistrationSubmitPostHandler обрабатывает пост-запрос формы отправки регистрации
func RegistrationSubmitPostHandler(c *gin.Context) {
	EnsureNotLoggedIn(c)
	var rd RegistrationData
	rd.Nickname = c.PostForm("nickname")
	rd.Registrationemail = c.PostForm("registrationemail")
	rd.Password1 = c.PostForm("password1")
	rd.Password2 = c.PostForm("password2")
	appErr := doRegistrationSubmit(c, &rd)
	if appErr == nil {
		c.HTML(http.StatusOK,
			"general.t.html",
			shared.GeneralTemplateParams{
				Message: "Проверьте свою электронную почту на наличие кода подтверждения, который будет действителен в течение 10 минут"})
	} else {
		c.HTML(http.StatusOK,
			"general.t.html",
			shared.GeneralTemplateParams{Message: appErr.Message})
	}
}

func doRegistrationSubmit(c *gin.Context, rd *RegistrationData) (apperr *apperror.AppErr) {
	validateRegistrationData(rd)
	apperr = processRegistrationSubmitWithDb(rd)
	if apperr == nil {
		// sendConfirmationEmail выдаёт только 500 в случае неудачи
		sendConfirmationEmail(c, rd)
	}
	return apperr
}

func validateRegistrationData(rd *RegistrationData) {
	if !isNicknameInValidFormat(rd.Nickname) {
		apperror.Panic500If(apperror.ErrDummy, "Ник недействителен")
	}
	if rd.Password1 != rd.Password2 {
		apperror.Panic500If(apperror.ErrDummy, "Пароли не совпадают")
	}
	passwordErr := validatePassword(rd.Password1)
	if passwordErr != nil {
		apperror.Panic500If(apperror.ErrDummy, "%s", passwordErr.Error())
	}
	if !isEmailInValidFormat(rd.Registrationemail) {
		apperror.Panic500If(apperror.ErrDummy, "Электронная почта недействительна")
	}
}

func sendConfirmationEmail(c *gin.Context, rd *RegistrationData) {
	scd := shared.SecretConfigData
	// TODO: если нет файлов сертификатов, используйте http an7
	confirmationLinkBase := shared.SitesProtocol() + "//" + scd.SiteRoot + shared.SitesPort() + "/registrationconfirmation"
	parameters := url.Values{"ник": {rd.Nickname}, "ключ подтверждения": {rd.ConfirmationKey}}
	u, err := url.Parse(confirmationLinkBase)
	apperror.GracefullyExitAppIf(err, "Невозможно разобрать базовый URL для ссылки подтверждения")
	u.RawQuery = parameters.Encode()
	confirmationLink := u.String()
	body := fmt.Sprintf(
		"Здравствуйте, %s!\nЧтобы активировать свой аккаунт, перейдите по ссылке активации: <a href=%s>%s</a>",
		// FIXME должен ли Nickname нуждаться в html-экранировании?
		html.EscapeString(rd.Nickname),
		confirmationLink, confirmationLink)

	err = SendEmail(
		rd.Registrationemail,
		"Добро пожаловать в семантический словарь!",
		body)

	if err != nil {
		// Мы предполагаем, что неспособность отправить электронное письмо может быть вызвана
		// временными проблемами в сети
		apperror.Panic500AndLogAttackIf(err, c, "Не удалось отправить подтверждение по электронной почте")
	}

	noteRegistrationConfirmationEMailSentWithDb(rd)
	return
}

// rd.UserID is filled
func noteRegistrationConfirmationEMailSentWithDb(rd *RegistrationData) {
	err := sddb.WithTransaction(func(trans *sddb.TransactionType) (err1 error) {
		sddb.CheckDbAlive()
		_, err1 = trans.Tx.NamedExec(
			`select note_registrationconfirmation_email_sent(:nickname, :confirmationkey)`,
			rd)
		return
	})
	sddb.FatalDatabaseErrorIf(err, "Ошибка, помнящая, что электронная почта была отправлена, ошибка заключается в следующем %#v", err)
	return
}

var mapViolatedConstraintNameToMessage = map[string]string{
	"i_registrationattempt__confirmationkey":   "Вам повезло попасть в очень редкое столкновение случайных чисел. Пожалуйста, повторите попытку регистрации",
	"i_registrationattempt__registrationemail": "Кто-то уже пытается зарегистрироваться с тем же E-mail",
	"i_registrationattempt__nickname":          "Кто-то уже пытается зарегистрироваться с таким же псевдонимом",
	"i_sduser_registrationemail":               "Уже есть пользователь с таким же E-mail",
	"i_sdusernickname":                         "Уже существует пользователь с таким же ником"}

func deleteExpiredRegistrationAttempts(trans *sddb.TransactionType) error {
	tx := trans.Tx
	sddb.CheckDbAlive()
	_, err1 := tx.Exec("select delete_expired_registrationattempts()")
	// это не фатальная ошибка (редкий случай!)
	apperror.Panic500If(err1,
		"Не удалось зарегистрироваться. Пожалуйста, повторите попытку позже или свяжитесь с нами для получения помощи")
	sddb.CheckDbAlive()
	err1 = tx.Commit()
	sddb.FatalDatabaseErrorIf(err1,
		"Не удалось выполнить фиксацию после delete_expired_registrationattatts, ошибка = %#v",
		err1)
	return nil
}

// processRegistrationSubmitWithDb вставляет попытку регистрации в sdusers_db
// Если происходит какая-то "нормальная" ошибка, например, не уникальный псевдоним, он возвращается в dberror.
func processRegistrationSubmitWithDb(rd *RegistrationData) *apperror.AppErr {

	err := sddb.WithTransaction(deleteExpiredRegistrationAttempts)
	sddb.FatalDatabaseErrorIf(err,
		"Не удалось обойти delete_expired_registrationattempts, %#v",
		err)

	err = sddb.WithTransaction(func(trans *sddb.TransactionType) (err error) {
		rd.Salt, rd.Hash = SaltAndHashPassword(rd.Password1)
		rd.ConfirmationKey = GenNonce(20)
		sddb.CheckDbAlive()
		_, err = trans.Tx.NamedExec(
			`select add_registrationattempt(:nickname, :salt, :hash, :registrationemail, :confirmationkey)`,
			rd)
		if err == nil {
			sddb.CheckDbAlive()
			err = trans.Tx.Commit()
		}
		return
	})
	return handleRegistrationAttemptInsertError(err)
}

func handleRegistrationAttemptInsertError(err error) *apperror.AppErr {
	//xt := reflect.TypeOf(err1).Kind()
	/* if e, ok := err.(*pgx.Error); ok {
		if e.Code == PostgresqlErrorCodeUniqueViolation {
			message, found := mapViolatedConstraintNameToMessage[e.Constraint]
			if found {
				return apperror.NewAppErrf(message)
			}
		}
	} */
	sddb.FatalDatabaseErrorIf(err, "Непредвиденная ошибка в процессе отправки регистрации, %#v\n", err)
	return nil
}
