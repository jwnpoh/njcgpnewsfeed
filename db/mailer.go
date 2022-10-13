package db

import (
	"os"
	"strconv"
	"time"

	"github.com/xhit/go-simple-mail/v2"
)

func SendMail(from, to, replyTo, subject, msg string) error {
	server := mail.NewSMTPClient()
	server.Host = os.Getenv("SMTP-HOST")
	server.Port, _ = strconv.Atoi(os.Getenv("SMTP-PORT"))
	server.Username = os.Getenv("SMTP-USER")
	server.Password = os.Getenv("SMTP-PASS")
	server.Encryption = mail.EncryptionTLS
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetReplyTo(replyTo).SetSubject(subject)
	email.SetBody(mail.TextPlain, msg)

	if err := email.Send(smtpClient); err != nil {
		return err
	}

	return nil
}
