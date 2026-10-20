package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"laci-backend/config"
)

type MailAccountPayload struct {
	ToEmail  string
	Nama     string
	NIM      string
	Username string
	Password string
	Divisi   string
}

// SendAccountCredentialsEmail mengirim detail akun ke email user secara asynchronous.
// Jika konfigurasi SMTP belum lengkap di .env, proses akan di-skip tanpa menimbulkan error pada request utama.
func SendAccountCredentialsEmail(payload MailAccountPayload) {
	if strings.TrimSpace(payload.ToEmail) == "" {
		return
	}

	smtpHost := config.GetEnv("SMTP_HOST", "")
	smtpPortStr := config.GetEnv("SMTP_PORT", "587")
	smtpUser := config.GetEnv("SMTP_USER", "")
	smtpPass := config.GetEnv("SMTP_PASS", "")
	smtpFrom := config.GetEnv("SMTP_FROM", smtpUser)
	appURL := config.GetEnv("APP_URL", "http://localhost:5173")

	if smtpHost == "" || smtpUser == "" || smtpPass == "" {
		log.Printf("[MAIL] SMTP tidak dikonfigurasi. Lewati kirim email ke: %s (Username: %s)", payload.ToEmail, payload.Username)
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

		nimText := ""
		if strings.TrimSpace(payload.NIM) != "" {
			nimText = fmt.Sprintf(" dengan NIM <b>%s</b>", payload.NIM)
		}

		divisiLabel := strings.ToUpper(payload.Divisi)
		if divisiLabel == "" {
			divisiLabel = "PENGURUS"
		}

		subject := "Akun LACI DOSCOM - Kredensial Masuk Sistem"
		body := fmt.Sprintf(`<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Akun LACI DOSCOM</title>
</head>
<body style="margin: 0; padding: 0; background-color: #f8fafc; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; -webkit-font-smoothing: antialiased; -moz-osx-font-smoothing: grayscale;">
  <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #f8fafc; padding: 48px 16px;">
    <tr>
      <td align="center">
        <!-- Main Card Container -->
        <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="max-width: 480px; background-color: #ffffff; border-radius: 20px; border: 1px solid #e2e8f0; overflow: hidden; box-shadow: 0 4px 24px rgba(15, 23, 42, 0.04);">
          <!-- Header Area -->
          <tr>
            <td style="padding: 40px 36px 24px 36px; text-align: left;">
              <!-- Brand mark -->
              <table role="presentation" border="0" cellspacing="0" cellpadding="0" style="margin-bottom: 20px;">
                <tr>
                  <td align="center" style="width: 44px; height: 44px; background-color: #0A84DC; border-radius: 12px; color: #ffffff; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif; font-size: 20px; font-weight: 700; line-height: 44px;">
                    L
                  </td>
                </tr>
              </table>

              <h1 style="margin: 0 0 10px 0; color: #0f172a; font-size: 22px; font-weight: 700; letter-spacing: -0.02em; line-height: 1.3;">
                Selamat Datang di LACI DOSCOM
              </h1>
              <p style="margin: 0; color: #475569; font-size: 14px; line-height: 1.6;">
                Halo <b style="color: #0f172a;">%s</b> (%s), akun sistem pelacakan surat kamu%s telah disiapkan oleh Sekretaris. Gunakan kredensial di bawah ini untuk masuk ke dashboard.
              </p>
            </td>
          </tr>

          <!-- Credential Box (Vertical SaaS Layout) -->
          <tr>
            <td style="padding: 0 36px 24px 36px;">
              <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0" style="background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 14px;">
                <tr>
                  <td style="padding: 24px 24px 20px 24px;">
                    <!-- Username Block -->
                    <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0">
                      <tr>
                        <td style="font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: #64748b; padding-bottom: 6px;">
                          Username
                        </td>
                      </tr>
                      <tr>
                        <td style="font-size: 16px; font-weight: 700; color: #0f172a; font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace; padding-bottom: 18px; border-bottom: 1px solid #e2e8f0;">
                          %s
                        </td>
                      </tr>
                      <!-- Password Block -->
                      <tr>
                        <td style="font-size: 11px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.06em; color: #64748b; padding-top: 18px; padding-bottom: 6px;">
                          Password Sementara
                        </td>
                      </tr>
                      <tr>
                        <td style="font-size: 16px; font-weight: 700; color: #0A84DC; font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', Menlo, Courier, monospace;">
                          %s
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Security Hint -->
          <tr>
            <td style="padding: 0 36px 28px 36px;">
              <p style="margin: 0; color: #64748b; font-size: 12px; line-height: 1.6;">
                *Demi keamanan akun, kamu disarankan untuk mengganti password ini melalui prompt saat pertama kali login atau lewat menu <b>Profil</b>.
              </p>
            </td>
          </tr>

          <!-- Primary CTA Button -->
          <tr>
            <td style="padding: 0 36px 40px 36px;">
              <table role="presentation" width="100%%" border="0" cellspacing="0" cellpadding="0">
                <tr>
                  <td align="center" style="background-color: #0A84DC; border-radius: 12px;">
                    <a href="%s" target="_blank" style="display: block; padding: 14px 24px; color: #ffffff; text-decoration: none; font-size: 14px; font-weight: 600; letter-spacing: -0.01em; text-align: center; border-radius: 12px;">
                      Masuk ke LACI DOSCOM &rarr;
                    </a>
                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Subtle Footer -->
          <tr>
            <td style="background-color: #f8fafc; border-top: 1px solid #e2e8f0; padding: 20px 36px; text-align: center;">
              <p style="margin: 0; color: #94a3b8; font-size: 11px; line-height: 1.6;">
                DOSCOM &bull; Dinus Open Source Community<br />
                Sistem Administrasi & Pelacakan Surat
              </p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, payload.Nama, divisiLabel, nimText, payload.Username, payload.Password, appURL)

		headers := make(map[string]string)
		headers["From"] = smtpFrom
		headers["To"] = payload.ToEmail
		headers["Subject"] = subject
		headers["MIME-Version"] = "1.0"
		headers["Content-Type"] = "text/html; charset=UTF-8"

		var message strings.Builder
		for k, v := range headers {
			message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
		message.WriteString("\r\n")
		message.WriteString(body)

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
				log.Printf("[MAIL ERROR] Gagal konek TLS SSL 465 ke %s: %v", addr, err)
				return
			}
			defer conn.Close()

			client, err := smtp.NewClient(conn, smtpHost)
			if err != nil {
				log.Printf("[MAIL ERROR] Gagal inisialisasi SMTP client: %v", err)
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
			if err = client.Rcpt(payload.ToEmail); err != nil {
				log.Printf("[MAIL ERROR] Gagal set recipient: %v", err)
				return
			}
			w, err := client.Data()
			if err != nil {
				log.Printf("[MAIL ERROR] Gagal kirim data: %v", err)
				return
			}
			_, err = w.Write([]byte(message.String()))
			if err != nil {
				log.Printf("[MAIL ERROR] Gagal write message: %v", err)
				return
			}
			err = w.Close()
			sendErr = err
		} else {
			// Port 587 / 25 dengan STARTTLS
			host, _, _ := net.SplitHostPort(addr)
			sendErr = smtp.SendMail(addr, auth, smtpUser, []string{payload.ToEmail}, []byte(message.String()))
			_ = host
		}

		if sendErr != nil {
			log.Printf("[MAIL ERROR] Gagal mengirim email ke %s: %v", payload.ToEmail, sendErr)
		} else {
			log.Printf("[MAIL SUCCESS] Kredensial akun berhasil dikirim ke: %s", payload.ToEmail)
		}
	}()
}
