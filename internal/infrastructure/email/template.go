package email

import (
	"fmt"
	"html"
	"strings"
)

// codeEmailOptions configures the shared Fundly transactional-email layout used for
// both OTP verification and password-reset codes (see the "Verify Your Account" /
// "Reset Your Password" design).
type codeEmailOptions struct {
	BadgeText      string // e.g. "ACCOUNT VERIFICATION" / "PASSWORD RESET"
	Heading        string // e.g. "Verify Your Account"
	Subheading     string // small line under the heading
	IntroParagraph string // paragraph inside the white card
	Code           string // 6-digit code
	ExpiryMinutes  int
	Disclaimer     string // small text under the button
}

// renderCodeEmailHTML builds the branded Fundly HTML email matching the approved
// verification-email design: dark header, white card, boxed digits, expiry notice
// and a CTA button.
func renderCodeEmailHTML(o codeEmailOptions) string {
	digits := strings.Split(o.Code, "")
	var boxes strings.Builder
	for _, d := range digits {
		fmt.Fprintf(&boxes, `
					<td width="40" height="48" align="center" valign="middle" style="width:40px;height:48px;background:#eef2ff;border:2px solid #3b5bfd;border-radius:10px;font-family:Arial,Helvetica,sans-serif;font-size:22px;font-weight:bold;color:#2545d3;">%s</td>
					<td width="8" style="width:8px;font-size:0;line-height:0;">&nbsp;</td>`, html.EscapeString(d))
	}

	return fmt.Sprintf(`<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Fundly</title>
</head>
<body style="margin:0;padding:0;background:#0b1220;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0" style="background:#0b1220;padding:40px 16px;">
<tr>
<td align="center">
<table role="presentation" width="600" cellpadding="0" cellspacing="0" style="width:100%%;max-width:600px;">

<!-- Header -->
<tr>
<td style="padding:0 8px 24px 8px;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0">
<tr>
<td align="left" style="font-family:Arial,Helvetica,sans-serif;font-size:22px;font-weight:bold;color:#ffffff;">Fundly<span style="color:#3b82f6;">&#9679;</span></td>
<td align="right">
<span style="display:inline-block;font-family:Arial,Helvetica,sans-serif;font-size:11px;font-weight:bold;letter-spacing:0.5px;color:#93c5fd;background:#15213b;border:1px solid #2c3f66;border-radius:999px;padding:6px 14px;">%s</span>
</td>
</tr>
</table>
</td>
</tr>

<!-- Heading -->
<tr>
<td style="padding:0 8px 24px 8px;font-family:Arial,Helvetica,sans-serif;">
<div style="font-size:26px;line-height:32px;font-weight:bold;color:#ffffff;">%s</div>
<div style="font-size:14px;line-height:20px;color:#94a3b8;padding-top:6px;">%s</div>
</td>
</tr>

<!-- Card -->
<tr>
<td style="background:#ffffff;border-radius:16px;padding:32px;">
<table role="presentation" width="100%%" cellpadding="0" cellspacing="0">
<tr>
<td style="font-family:Arial,Helvetica,sans-serif;font-size:16px;font-weight:bold;color:#0f172a;padding-bottom:10px;">Welcome to Fundly,</td>
</tr>
<tr>
<td style="font-family:Arial,Helvetica,sans-serif;font-size:14px;line-height:21px;color:#475569;padding-bottom:24px;">%s</td>
</tr>
<tr>
<td align="center" style="padding-bottom:16px;">
<table role="presentation" cellpadding="0" cellspacing="0"><tr>%s
</tr></table>
</td>
</tr>
<tr>
<td align="center" style="font-family:Arial,Helvetica,sans-serif;font-size:12px;color:#64748b;padding-bottom:24px;">Can't see the boxes? Code: <b style="color:#0f172a;letter-spacing:2px;">%s</b></td>
</tr>
<tr>
<td style="background:#eff6ff;border-radius:10px;padding:14px 16px;font-family:Arial,Helvetica,sans-serif;font-size:13px;color:#1d4ed8;padding-bottom:14px;">
&#9201;&nbsp; This code will expire in %d minutes. Never share this code with anyone.
</td>
</tr>
<tr><td style="height:20px;line-height:20px;font-size:0;">&nbsp;</td></tr>
<tr>
<td align="center" style="background:#0f172a;border-radius:10px;">
<div style="font-family:Arial,Helvetica,sans-serif;font-size:14px;font-weight:bold;color:#ffffff;padding:14px 0;">%s</div>
</td>
</tr>
<tr><td style="height:18px;line-height:18px;font-size:0;">&nbsp;</td></tr>
<tr>
<td align="center" style="font-family:Arial,Helvetica,sans-serif;font-size:12px;line-height:18px;color:#64748b;">%s</td>
</tr>
</table>
</td>
</tr>

<!-- Footer -->
<tr>
<td align="center" style="padding-top:28px;font-family:Arial,Helvetica,sans-serif;font-size:12px;color:#64748b;">
&#128274;&nbsp; <span style="color:#94a3b8;">Fundly Security</span> &middot; Bank-grade 256-bit encryption protecting your financial data.
</td>
</tr>
<tr>
<td align="center" style="padding-top:12px;font-family:Arial,Helvetica,sans-serif;font-size:11px;line-height:17px;color:#475569;">
This automated message was sent to confirm activity on your Fundly account.<br>
&copy; 2026 Fundly Technologies Inc. All rights reserved.
</td>
</tr>

</table>
</td>
</tr>
</table>
</body>
</html>`,
		html.EscapeString(o.BadgeText),
		html.EscapeString(o.Heading),
		html.EscapeString(o.Subheading),
		html.EscapeString(o.IntroParagraph),
		boxes.String(),
		html.EscapeString(o.Code),
		o.ExpiryMinutes,
		"Enter Code in Fundly",
		html.EscapeString(o.Disclaimer),
	)
}
