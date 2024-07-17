//go:generate mockgen -source=email_template.go -destination=mock/email_template_mock.go -package=mock

package auth0

import (
	"context"

	"github.com/auth0/go-auth0/management"
)

type EmailTemplateAPI interface {
	// Create an email template.
	//
	// See: https://auth0.com/docs/api/management/v2#!/Email_Templates/post_email_templates
	Create(ctx context.Context, template *management.EmailTemplate, opts ...management.RequestOption) error

	// Read an email template by pre-defined name.
	//
	// These names are `verify_email`, `reset_email`, `welcome_email`,
	// `blocked_account`, `stolen_credentials`, `enrollment_email`, and
	// `mfa_oob_code`.
	//
	// The names `change_password`, and `password_reset` are also supported for
	// legacy scenarios.
	//
	// See: https://auth0.com/docs/api/management/v2#!/Email_Templates/get_email_templates_by_templateName
	Read(ctx context.Context, template string, opts ...management.RequestOption) (e *management.EmailTemplate, err error)

	// Update an email template.
	//
	// See: https://auth0.com/docs/api/management/v2#!/Email_Templates/patch_email_templates_by_templateName
	Update(ctx context.Context, template string, e *management.EmailTemplate, opts ...management.RequestOption) (err error)
}
