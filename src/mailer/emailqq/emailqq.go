package emailqq

import (
	"gopkg.in/gomail.v2"
)

var (
	mailer *gomail.Dialer
	from   = ""
)

func Start(host, username, password string, port int) *gomail.Dialer {
	mailer = gomail.NewDialer(host, port, username, password)
	from = username
	return mailer
}

func Send(toemail, subject, plain string) error {
	m := gomail.NewMessage()
	//发送人
	m.SetHeader("From", from)
	//接收人
	m.SetHeader("To", toemail)
	//抄送人
	//m.SetAddressHeader("Cc", "xxx@qq.com", "xiaozhujiao")
	//主题
	m.SetHeader("Subject", subject)
	//内容
	m.SetBody("text/html", "<h1>"+plain+"</h1>")
	//附件
	//m.Attach("./myIpPic.png")
	// 发送邮件
	return mailer.DialAndSend(m)
}
