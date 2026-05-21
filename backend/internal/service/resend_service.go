package service

import (
	"fmt"

	"backend/internal/config"

	"github.com/resend/resend-go/v3"
)

type ResendService struct {
	client    *resend.Client
	apiKey    string
	fromEmail string
	fromName  string
}

func NewResendService(cfg *config.Config) *ResendService {
	return &ResendService{
		client:    resend.NewClient(cfg.ResendAPIKey),
		apiKey:    cfg.ResendAPIKey,
		fromEmail: cfg.ResendFromEmail,
		fromName:  cfg.ResendFromName,
	}
}

func (r *ResendService) SendVerificationEmail(toEmail, toName, verifyLink string) error {
	if r.apiKey == "" {
		return fmt.Errorf("Resend API key not configured")
	}

	from := r.fromEmail
	if r.fromName != "" {
		from = fmt.Sprintf("%s <%s>", r.fromName, r.fromEmail)
	}

	htmlContent := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.button { display: inline-block; padding: 12px 24px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
				.footer { text-align: center; padding: 20px; font-size: 12px; color: #666; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h2>Verify Your Email Address</h2>
				</div>
				<div class="content">
					<p>Hello %s,</p>
					<p>Thank you for registering. Verify your email address by clicking the button below:</p>
					<p style="text-align: center;">
						<a href="%s" class="button">Verify Email Address</a>
					</p>
					<p>Or copy and paste this link into your browser:</p>
					<p>%s</p>
					<p>This link will expire in 24 hours.</p>
					<p>If you did not create an account, ignore this email.</p>
				</div>
				<div class="footer">
					<p>© 2026 KubeEvalHub. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>
	`, toName, verifyLink, verifyLink)

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{toEmail},
		Subject: "Verify Your Email Address",
		Html:    htmlContent,
	}

	_, err := r.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}
