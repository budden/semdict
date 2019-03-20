package user

import "time"

// RegistrationData is a transient struct containing data obtained from a /registrationformsubmit query
// as well as some of calculated data
type RegistrationData struct {
	Nickname          string
	Registrationemail string
	Password1         string
	Password2         string
	Salt              string
	Hash              string
	ConfirmationKey   string
	UserID            int32
}

// SDUserData is based on sduser table
type SDUserData struct {
	ID                    int32
	Nickname              string
	Registrationemail     string
	Salt                  string
	Hash                  string
	RegistrationTimestamp time.Time
}
