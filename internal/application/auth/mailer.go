package auth

import "context"

// Mailer delivers authentication emails (OTP and password-reset codes). Implementations live
// in the infrastructure layer. It may be nil on a use case, in which case no email is sent
// (the code is still persisted for verification).
type Mailer interface {
	SendOTP(ctx context.Context, email, fullName, otp string) error
	SendPasswordReset(ctx context.Context, email, code string) error
}
