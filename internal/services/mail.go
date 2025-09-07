package services

import (
	"bytes"
	"fmt"
	"html/template"
	"openhouse-2025-api/internal/config"
	"openhouse-2025-api/internal/models"
	"strconv"

	"gopkg.in/gomail.v2"
)

// MailService encapsulates email sending logic.
type MailService struct {
	dialer *gomail.Dialer
	from   string
}

func NewMailService(cfg *config.Config) *MailService {
	port, err := strconv.Atoi(cfg.SMTPPort)
	if err != nil {
		fmt.Println("Error converting string to int:", err)
		return nil
	}
	// Create a "dialer" to connect to the SMTP server.
	dialer := gomail.NewDialer(cfg.SMTPHost, port, cfg.SMTPUser, cfg.SMTPPass)
	return &MailService{
		dialer: dialer,
		from:   cfg.SMTPFrom,
	}
}

// TemplateData holds the dynamic data for our email templates.
type TemplateData struct {
	UserName         string
	UkmName          string
	IsAccepted       bool
	LinkGroupChatUKM string
}

// PaymentConfirmationData holds data for payment confirmation emails
type PaymentConfirmationData struct {
	UserName string
	UserNRP  string
	UkmName  string
}

// SendNotificationEmail is our equivalent of a Mailable class.
func (s *MailService) SendNotificationEmail(user *models.User, ukm *models.Ukm, status bool, linkChatUKM string) error {
	// 1. Define the recipient's email address
	recipient := fmt.Sprintf("%s@john.petra.ac.id", user.NRP)
	subject := fmt.Sprintf("Update mengenai pendaftaran untuk %s", ukm.Name)

	// 2. Parse the HTML templates
	tmpl, err := template.ParseFiles("internal/templates/notification_email.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// 3. Prepare the data to inject into the templates
	data := TemplateData{
		UserName:         user.Name,
		UkmName:          ukm.Name,
		IsAccepted:       status,
		LinkGroupChatUKM: linkChatUKM,
	}

	// 4. Execute the templates, injecting the data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// 5. Compose the email message
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	// 6. Send the email
	// .DialAndSend() connects to the server, sends the email, and closes the connection.
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendPaymentConfirmationEmailTemplate sends payment confirmation using HTML template file
func (s *MailService) SendPaymentConfirmationEmailTemplate(userNRP string, data PaymentConfirmationData) error {
	// 1. Define the recipient's email address
	recipient := fmt.Sprintf("%s@john.petra.ac.id", userNRP)
	subject := fmt.Sprintf("Konfirmasi Pendaftaran %s", data.UkmName)

	// 2. Parse the HTML templates
	tmpl, err := template.ParseFiles("internal/templates/notification_email_regist.html")
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// 3. The data is already passed as parameter, no need to redefine it

	// 4. Execute the templates, injecting the data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// 5. Compose the email message
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	// 6. Send the email
	// .DialAndSend() connects to the server, sends the email, and closes the connection.
	if err := s.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
