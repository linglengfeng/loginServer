package mailer

import (
	"fmt"
	"loginServer/config"
	"loginServer/src/mailer/emailqq"
	"loginServer/src/mailer/sendgrid"
)

const (
	way_qq       = "qq"
	way_sendgrid = "sendgrid"
)

var (
	mailer    any
	mailercfg = config.Cfg[config.Mailer]
	mailerway = mailercfg.GetString("mailer")
)

func Start() {
	switch mailerway {
	case way_qq:
		mailer = emailqq.Start(mailercfg.GetString("qq.ip"), mailercfg.GetString("qq.username"), mailercfg.GetString("qq.authcode"), mailercfg.GetInt("qq.port"))
	case way_sendgrid:
		mailer = sendgrid.Start(mailercfg.GetString("sendgrid.api_key"), mailercfg.GetString("sendgrid.username"))
	default:
		break
	}
}

func Send(toemail, subject, plain string) error {
	switch mailerway {
	case way_qq:
		return emailqq.Send(toemail, subject, plain)
	case way_sendgrid:
		return sendgrid.Send(toemail, subject, plain)
	default:
		return fmt.Errorf("SendLoginEmail, toemail:%v", toemail)
	}
}

func SendLoginEmail(toemail, code string) error {
	subject := "Welcome to Game!"
	plain := "Welcome to Game! To login, you’ll need to use your email and verify code:" + code
	return Send(toemail, subject, plain)
}
