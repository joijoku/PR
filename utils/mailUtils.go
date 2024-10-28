package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	xmail "github.com/go-mail/mail"
	"github.com/google/uuid"
	"pr.net/shared"
)

var (
	host       = "smtp.gmail.com"
	username   = "dev.psi@praweda.id"
	password   = "PrawedaSarana2019"
	portNumber = "587"
)

type Sender struct {
	auth smtp.Auth
}

type Message struct {
	From         string
	To           []string
	Rcpt         []string
	CC           []string
	BCC          []string
	Subject      string
	Body         string
	Attachments  map[string][]byte
	htmlBody     bool
	Attachments2 []string
}

func New() *Sender {
	auth := smtp.PlainAuth("", username, password, host)
	return &Sender{auth}
}

func (s *Sender) Send(m *Message) error {
	return smtp.SendMail(fmt.Sprintf("%s:%s", host, portNumber), s.auth, username, m.Rcpt, m.ToBytes())
}

func NewMessage(s, b string, isHtmlBody bool) *Message {
	return &Message{Subject: s, Body: b, Attachments: make(map[string][]byte), htmlBody: isHtmlBody, Attachments2: make([]string, 0)}
}

func (m *Message) AddAttachment(src string) {
	m.Attachments2 = append(m.Attachments2, src)
}

func (m *Message) AttachFile(src string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	_, fileName := filepath.Split(src)
	m.Attachments[fileName] = b
	return nil
}

func (m *Message) ToBytes() []byte {
	buf := bytes.NewBuffer(nil)
	withAttachments := len(m.Attachments) > 0
	buf.WriteString(fmt.Sprintf("From: %s\n", m.From))
	buf.WriteString(fmt.Sprintf("Subject: %s\n", m.Subject))
	buf.WriteString(fmt.Sprintf("To: %s\n", strings.Join(m.To, ",")))
	if len(m.CC) > 0 {
		buf.WriteString(fmt.Sprintf("CC: %s\n", strings.Join(m.CC, ",")))
	}

	if len(m.BCC) > 0 {
		buf.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(m.BCC, ",")))
	}

	buf.WriteString("MIME-Version: 1.0\n")
	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	if withAttachments {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\n", boundary))
		buf.WriteString(fmt.Sprintf("--%s\n", boundary))

		//bodyBoundary := writer.Boundary()
		//buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\n", bodyBoundary))
		//buf.WriteString(fmt.Sprintf("--%s\n", bodyBoundary))
		if m.htmlBody {
			fileName := uuid.NewString() + ".html"
			err := CreateFile(os.Getenv("tmpFolder"), fileName, m.Body)
			shared.CheckErr(err)

			buf.WriteString("Content-Type: text/html; charset=UTF-8\n")
			buf.WriteString("Content-Transfer-Encoding: quoted-printable\n")

			t, _ := template.ParseFiles(os.Getenv("tmpFolder") + fileName)
			t.Execute(buf, nil)

			err = os.Remove(os.Getenv("tmpFolder") + fileName)
			shared.CheckErr(err)

		} else {
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\n")
			buf.WriteString("Content-Transfer-Encoding: quoted-printable\n")
			buf.WriteString(m.Body)
		}

		//buf.WriteString(fmt.Sprintf("\n--%s", bodyBoundary))
	} else {
		if m.htmlBody {
			fileName := uuid.NewString() + ".html"
			err := CreateFile(os.Getenv("tmpFolder"), fileName, m.Body)
			shared.CheckErr(err)

			//var body bytes.Buffer
			buf.WriteString("Content-Type: text/html; charset=UTF-8\n")
			//body.Write(buf.Bytes())

			t, _ := template.ParseFiles(os.Getenv("tmpFolder") + fileName)
			t.Execute(buf, nil)

			err = os.Remove(os.Getenv("tmpFolder") + fileName)
			shared.CheckErr(err)
			//return buf.Bytes()
		} else {
			buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\n", boundary))
			buf.WriteString(fmt.Sprintf("--%s\n", boundary))
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\n")
			buf.WriteString(m.Body)
			buf.WriteString(fmt.Sprintf("\n--%s--", boundary))
		}

		// if m.htmlBody {
		// 	buf.WriteString("Content-Type: text/html; charset=iso-8859-1\n")
		// } else {
		// 	buf.WriteString("Content-Type: text/plain; charset=UTF-8\n")
		// }
		// buf.WriteString("Content-Transfer-Encoding: quoted-printable\n")
	}

	if withAttachments {
		for k, v := range m.Attachments {
			buf.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
			buf.WriteString("Content-Transfer-Encoding: base64\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n", k))

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buf.Write(b)
			buf.WriteString(fmt.Sprintf("\n--%s", boundary))
		}

		buf.WriteString("--")
	}

	// log.Println(buf.String())
	return buf.Bytes()
}

func SendNewMail(m *Message) error {
	mail := xmail.NewMessage()
	mail.SetHeader("From", m.From)
	mail.SetHeader("To", m.To...)

	if len(m.CC) > 0 {
		mail.SetHeader("Cc", m.CC...)
	}

	if len(m.BCC) > 0 {
		mail.SetHeader("Bcc", m.BCC...)
	}

	mail.SetHeader("Subject", m.Subject)

	if m.htmlBody {
		mail.SetBody("text/html", m.Body)
	} else {
		mail.SetBody("text/plain", m.Body)
	}

	withAttachments := len(m.Attachments2) > 0

	if withAttachments {
		for _, v := range m.Attachments2 {
			mail.Attach(v)
		}
	}

	d := xmail.NewDialer(host, 587, username, password)

	return d.DialAndSend(mail)
}

func SetHost(val string) {
	host = val
}

func GetHost() string {
	return host
}

func SetUserName(val string) {
	username = val
}

func GetUserName() string {
	return username
}

func SetPassword(val string) {
	password = val
}

func GetPassword() string {
	return password
}

func SetPort(val string) {
	portNumber = val
}

func GetPort() string {
	return portNumber
}
