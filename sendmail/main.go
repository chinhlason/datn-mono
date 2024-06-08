package sendmail

import (
	"bytes"
	"fmt"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"os"
)

type EmailStruct struct {
	Email string
	Token string
}

func SendMail(email string, token string, path string) error {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	var (
		emailEnv = os.Getenv("EMAIL")
		psw      = os.Getenv("PSW")
	)
	fmt.Println(emailEnv, psw)
	var body bytes.Buffer
	t, err := template.ParseFiles(path)
	if err != nil {
		return err
	}
	t.Execute(&body, EmailStruct{
		Email: email,
		Token: token,
	})

	m := gomail.NewMessage()
	m.SetHeader("From", "alex@example.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Reset Password!")
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer("smtp.gmail.com", 587, emailEnv, psw)

	// Send the email to Bob, Cora and Dan.
	if err = d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
