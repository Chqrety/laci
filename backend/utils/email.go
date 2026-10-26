package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"strings"

	"laci-backend/config"
)

// sendHTMLEmail mengirim email HTML via SMTP (mendukung port 465 SSL & 587 STARTTLS)
func sendHTMLEmail(toEmail, subject, htmlBody string) {
	if strings.TrimSpace(toEmail) == "" {
		return
	}

	smtpHost := config.GetEnv("SMTP_HOST", "")
	smtpPortStr := config.GetEnv("SMTP_PORT", "587")
	smtpUser := config.GetEnv("SMTP_USER", "")
	smtpPass := config.GetEnv("SMTP_PASS", "")
	smtpFrom := config.GetEnv("SMTP_FROM", smtpUser)

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		log.Printf("[MAIL] SMTP tidak dikonfigurasi. Lewati kirim email ke: %s (Subject: %s)", toEmail, subject)
		return
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		smtpPort = 587
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[MAIL ERROR] Panic saat mengirim email: %v", r)
			}
		}()

		headers := make(map[string]string)
		headers["From"] = smtpFrom
		headers["To"] = toEmail
		headers["Subject"] = subject
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"

		var message strings.Builder
		for k, v := range headers {
			message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		message.WriteString("\r\n")
		message.WriteString(htmlBody)

		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		addr := fmt.Sprintf("%s:%d", smtpHost, smtpPort)

		var sendErr error
		if smtpPort == 465 {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: false,
				ServerName:         smtpHost,
			}
			conn, err := tls.Dial("tcp", addr, tlsConfig)
			if err != nil {
				log.Printf("[MAIL ERROR] Gagal koneksi TLS 465 ke %s: %v", addr, err)
				return
			}
			defer conn.Close()

			client, err := smtp.NewClient(conn, smtpHost)
			if err != nil {
				log.Printf("[MAIL ERROR] Gagal membuat SMTP client: %v", err)
				return
			}
			defer client.Quit()

			if err = client.Auth(auth); err != nil {
				log.Printf("[MAIL ERROR] Gagal autentikasi SMTP: %v", err)
				return
			}
			if err = client.Mail(smtpUser); err != nil {
				log.Printf("[MAIL ERROR] Gagal set sender: %v", err)
				return
			}
			if err = client.Rcpt(toEmail); err != nil {
				log.Printf("[MAIL ERROR] Gagal set recipient: %v", err)
				return
			}
			w, err := client.Data()
			if err != nil {
				log.Printf("[MAIL ERROR] Gagal write data: %v", err)
				return
			}
			_, err = w.Write([]byte(message.String()))
			if err != nil {
				log.Printf("[MAIL ERROR] Gagal write message content: %v", err)
				return
			}
			sendErr = w.Close()
		} else {
			sendErr = smtp.SendMail(addr, auth, smtpUser, []string{toEmail}, []byte(message.String()))
		}

		if sendErr != nil {
			log.Printf("[MAIL ERROR] Gagal mengirim email ke %s: %v", toEmail, sendErr)
		} else {
			log.Printf("[MAIL SUCCESS] Email berhasil dikirim ke: %s (Subject: %s)", toEmail, subject)
		}
	}()
}

func renderEmailCard(heading, greeting, buttonText, buttonURL, footerNotice string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8" />
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background-color:#f8fafc;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;">
  <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color:#f8fafc;padding:48px 16px;">
    <tr>
      <td align="center">
        <table width="100%%" border="0" cellspacing="0" cellpadding="0" style="max-width:480px;background-color:#ffffff;border-radius:20px;border:1px solid #e2e8f0;overflow:hidden;box-shadow:0 4px 24px rgba(15,23,42,0.04);">
          <tr>
            <td style="padding:40px 36px 24px 36px;">
              <table border="0" cellspacing="0" cellpadding="0" style="margin-bottom:20px;">
                <tr>
                  <td align="center" style="width:44px;height:44px;background-color:#0A84DC;border-radius:12px;color:#ffffff;font-size:20px;font-weight:700;line-height:44px;">
                    L
                  </td>
                </tr>
              </table>
              <h1 style="margin:0 0 10px 0;color:#0f172a;font-size:20px;font-weight:700;">%s</h1>
              <p style="margin:0;color:#475569;font-size:14px;line-height:1.6;">
                %s
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding:0 36px 36px 36px;">
              <table width="100%%" border="0" cellspacing="0" cellpadding="0">
                <tr>
                  <td align="center" style="background-color:#0A84DC;border-radius:12px;">
                    <a href="%s" target="_blank" style="display:block;padding:14px 24px;color:#ffffff;text-decoration:none;font-size:14px;font-weight:600;text-align:center;border-radius:12px;">
                      %s
                    </a>
                  </td>
                </tr>
              </table>
              <p style="margin:20px 0 0 0;color:#94a3b8;font-size:12px;line-height:1.5;">
                %s
              </p>
            </td>
          </tr>
          <tr>
            <td style="background-color:#f8fafc;border-top:1px solid #e2e8f0;padding:16px 36px;text-align:center;">
              <p style="margin:0;color:#94a3b8;font-size:11px;">DOSCOM &bull; Dinus Open Source Community</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, heading, heading, greeting, buttonURL, buttonText, footerNotice)
}

// SendActivationEmail mengirim tautan aktivasi akun baru untuk membuat password mandiri
func SendActivationEmail(toEmail, nama, username, token string) {
	appURL := strings.TrimRight(config.GetEnv("APP_URL", "http://localhost:5173"), "/")
	activationURL := fmt.Sprintf("%s/activate?token=%s", appURL, token)

	log.Printf("[AUTH ACTIVATION] Email: %s | User: %s | Link: %s", toEmail, username, activationURL)

	greeting := fmt.Sprintf("Halo <b>%s</b> (@%s), akun kamu di sistem pelacakan surat DOSCOM telah disiapkan. Klik tombol di bawah ini untuk mengaktifkan akun dan membuat password baru kamu.", nama, username)
	body := renderEmailCard("Aktivasi Akun LACI DOSCOM", greeting, "Aktifkan Akun & Buat Password &rarr;", activationURL, "Tautan ini berlaku selama 24 jam. Jika kamu tidak merasa didaftarkan, abaikan email ini.")
	sendHTMLEmail(toEmail, "Aktivasi Akun & Buat Password - LACI DOSCOM", body)
}

// SendPasswordResetEmail mengirim tautan pemulihan kata sandi
func SendPasswordResetEmail(toEmail, nama, username, token string) {
	appURL := strings.TrimRight(config.GetEnv("APP_URL", "http://localhost:5173"), "/")
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", appURL, token)

	log.Printf("[AUTH RESET] Email: %s | User: %s | Link: %s", toEmail, username, resetURL)

	greeting := fmt.Sprintf("Halo <b>%s</b> (@%s), kami menerima permintaan untuk mereset kata sandi akun LACI DOSCOM kamu. Klik tombol di bawah untuk membuat kata sandi baru.", nama, username)
	body := renderEmailCard("Permintaan Reset Kata Sandi", greeting, "Reset Kata Sandi &rarr;", resetURL, "Tautan ini hanya berlaku selama 1 jam. Jika kamu tidak meminta reset password, abaikan email ini dan akunmu tetap aman.")
	sendHTMLEmail(toEmail, "Pemulihan Kata Sandi - LACI DOSCOM", body)
}
