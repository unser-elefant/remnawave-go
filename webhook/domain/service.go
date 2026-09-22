package domain

type LoginAttempt struct {
	Username    string
	Ip          string
	UserAgent   string
	Description *string
	Password    *string
}
