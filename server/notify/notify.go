package notify

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/mail"
	"strconv"
	"time"

	"github.com/mibk/dali"
	"github.com/pdrhlik/deliberix/server/locale"
	gomail "gopkg.in/gomail.v2"
)

//go:generate go run gen_tmpls.go -o templates.go templates/*

const (
	typeEmailVerification = "email-verification"
	typePasswordReset     = "password-reset"
	typeMagicLink         = "magic-link"
)

type Notification interface {
	getUserID() uint
	getRcpt() *mail.Address
	getLang() string
	getType() string
}

// EmailVerification is queued when a user registers.
type EmailVerification struct {
	UserID   uint              `json:"-"`
	Email    *mail.Address     `json:"-"`
	Language string            `json:"-"`
	Link     string            `json:"link"`
	T        map[string]string `json:"-"`
}

func (n *EmailVerification) getUserID() uint        { return n.UserID }
func (n *EmailVerification) getRcpt() *mail.Address { return n.Email }
func (n *EmailVerification) getLang() string         { return n.Language }
func (n *EmailVerification) getType() string         { return typeEmailVerification }

// PasswordReset is queued when a user requests a password reset.
type PasswordReset struct {
	UserID   uint              `json:"-"`
	Email    *mail.Address     `json:"-"`
	Language string            `json:"-"`
	Link     string            `json:"link"`
	T        map[string]string `json:"-"`
}

func (n *PasswordReset) getUserID() uint        { return n.UserID }
func (n *PasswordReset) getRcpt() *mail.Address { return n.Email }
func (n *PasswordReset) getLang() string         { return n.Language }
func (n *PasswordReset) getType() string         { return typePasswordReset }

// MagicLink is queued when a user requests passwordless sign-in.
type MagicLink struct {
	UserID   uint              `json:"-"`
	Email    *mail.Address     `json:"-"`
	Language string            `json:"-"`
	Link     string            `json:"link"`
	T        map[string]string `json:"-"`
}

func (n *MagicLink) getUserID() uint        { return n.UserID }
func (n *MagicLink) getRcpt() *mail.Address { return n.Email }
func (n *MagicLink) getLang() string         { return n.Language }
func (n *MagicLink) getType() string         { return typeMagicLink }

// MailSender sends a composed email message.
type MailSender interface {
	Send(msg Message, lang string) error
}

type Message struct {
	To       *mail.Address
	Subject  string
	HTMLBody func(w io.Writer) error
}

type gomailSender struct {
	dialer   *gomail.Dialer
	fromAddr string
	fromName string
}

func NewGomailSender(d *gomail.Dialer, fromAddr, fromName string) MailSender {
	return &gomailSender{dialer: d, fromAddr: fromAddr, fromName: fromName}
}

func (s *gomailSender) Send(msg Message, lang string) error {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", s.fromAddr, s.fromName)
	m.SetAddressHeader("To", msg.To.Address, msg.To.Name)
	m.SetHeader("Subject", msg.Subject)
	if msg.HTMLBody != nil {
		m.AddAlternativeWriter("text/html", msg.HTMLBody)
	}
	return s.dialer.DialAndSend(m)
}

// NewDialer creates an SMTP dialer from individual config values.
func NewDialer(host string, port int, user, password string) *gomail.Dialer {
	return gomail.NewDialer(host, port, user, password)
}

// Service handles email notification queuing and sending.
type Service struct {
	db *dali.DB
}

func NewService(db *dali.DB) *Service {
	return &Service{db: db}
}

type notification struct {
	ID       uint      `db:"id,selectonly"`
	Email    string    `db:"email"`
	Name     string    `db:"name"`
	Language string    `db:"language"`
	UserID   uint      `db:"user_id"`
	Queued   time.Time `db:"queued"`
	Type     string    `db:"type"`
	RawData  []byte    `db:"data"`
}

func (s *Service) EnqueueEmail(n Notification) error {
	rcpt := n.getRcpt()
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	en := notification{
		Email:    rcpt.Address,
		Name:     rcpt.Name,
		Language: n.getLang(),
		UserID:   n.getUserID(),
		Queued:   time.Now(),
		Type:     n.getType(),
		RawData:  data,
	}
	q := s.db.Query(`INSERT INTO email_notification ?values`, en)
	_, err = q.Exec()
	return err
}

// SendEmails runs a background loop that sends pending emails.
func (s *Service) SendEmails(ctx context.Context, sender MailSender) error {
	for {
		wait := time.Duration(3+rand.Intn(3)) * time.Second
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
			if err := s.sendPendingEmail(sender); err != nil {
				return err
			}
		}
	}
}

func (s *Service) sendPendingEmail(sender MailSender) error {
	q := s.db.Query(`SELECT * FROM email_notification WHERE done IS NULL LIMIT 1`)
	var en notification
	if err := q.One(&en); err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	translator := locale.GetTranslationFunc(en.Language)

	var msg Message
	msg.To = &mail.Address{Address: en.Email, Name: en.Name}

	switch en.Type {
	case typeEmailVerification:
		msg.Subject = translator.T("email.verification.title")
		var tmplData EmailVerification
		if err := json.Unmarshal(en.RawData, &tmplData); err != nil {
			return err
		}
		tmplData.Language = en.Language
		tmplData.T = map[string]string{
			"title":   translator.T("email.verification.title"),
			"heading": translator.T("email.verification.heading"),
			"info":    translator.T("email.verification.info"),
			"button":  translator.T("email.verification.button"),
			"info2":   translator.T("email.verification.info2"),
			"sent":    translator.T("email.sent"),
		}
		msg.HTMLBody = func(w io.Writer) error {
			return verificationHTML.Execute(w, tmplData)
		}
	case typePasswordReset:
		msg.Subject = translator.T("email.passwordReset.title")
		var tmplData PasswordReset
		if err := json.Unmarshal(en.RawData, &tmplData); err != nil {
			return err
		}
		tmplData.Language = en.Language
		tmplData.T = map[string]string{
			"title":   translator.T("email.passwordReset.title"),
			"heading": translator.T("email.passwordReset.heading"),
			"info":    translator.T("email.passwordReset.info"),
			"button":  translator.T("email.passwordReset.button"),
			"info2":   translator.T("email.passwordReset.info2"),
			"sent":    translator.T("email.sent"),
		}
		msg.HTMLBody = func(w io.Writer) error {
			return passwordResetHTML.Execute(w, tmplData)
		}
	case typeMagicLink:
		msg.Subject = translator.T("email.magicLink.title")
		var tmplData MagicLink
		if err := json.Unmarshal(en.RawData, &tmplData); err != nil {
			return err
		}
		tmplData.Language = en.Language
		tmplData.T = map[string]string{
			"title":   translator.T("email.magicLink.title"),
			"heading": translator.T("email.magicLink.heading"),
			"info":    translator.T("email.magicLink.info"),
			"button":  translator.T("email.magicLink.button"),
			"info2":   translator.T("email.magicLink.info2"),
			"sent":    translator.T("email.sent"),
		}
		msg.HTMLBody = func(w io.Writer) error {
			return magicLinkHTML.Execute(w, tmplData)
		}
	default:
		log.Printf("unknown notification type: %q (id=%d)", en.Type, en.ID)
		return s.logEmailSent(en, nil)
	}

	mailErr := sender.Send(msg, en.Language)
	return s.logEmailSent(en, mailErr)
}

func (s *Service) logEmailSent(en notification, mailErr error) error {
	var errMsg sql.NullString
	if mailErr != nil {
		errMsg = sql.NullString{String: mailErr.Error(), Valid: true}
	}
	q := s.db.Query(`UPDATE email_notification SET done = ?, error = ? WHERE id = ?`,
		time.Now(), errMsg, en.ID)
	_, err := q.Exec()
	return err
}

// SMTPConfigured returns true if SMTP host is set.
func SMTPConfigured(host string) bool {
	return host != ""
}

// ParseSMTPPort converts port string to int, defaults to 587.
func ParseSMTPPort(port string) int {
	p, err := strconv.Atoi(port)
	if err != nil {
		return 587
	}
	return p
}
