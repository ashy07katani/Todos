package mail

import (
	"context"
	"fmt"
	"net/smtp"
)

type Mail struct {
	From     string `env:"EMAIL"`
	Password string `env:"PASSWORD"`
	Host     string `env:"HOST"`
	Port     string `env:"PORT"`
}

func (m *Mail) GetAuth() smtp.Auth {
	return smtp.PlainAuth("", m.From, m.Password, m.Host)
}

func (m *Mail) SendMail(ctx context.Context, a smtp.Auth, to []string, msg []byte) error {
	return smtp.SendMail(fmt.Sprintf("%s:%s", m.Host, m.Port), a, m.From, to, msg)
}

/*
func SendMail(addr string, a Auth, from string, to []string, msg []byte) error
*/
