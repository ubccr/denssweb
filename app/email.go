// Copyright 2017 DENSSWeb Authors. All rights reserved.
//
// This file is part of DENSSWeb.
//
// DENSSWeb is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// DENSSWeb is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with DENSSWeb.  If not, see <http://www.gnu.org/licenses/>.

package app

import (
	"bytes"
	"errors"
	"fmt"
	"mime/quotedprintable"
	"net/smtp"
	"net/textproto"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func quotedBody(body []byte) ([]byte, error) {
	var buf bytes.Buffer
	w := quotedprintable.NewWriter(&buf)
	_, err := w.Write(body)
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (a *AppContext) SendEmail(toEmail, status, jobURL string, jid int64) error {
	if !viper.GetBool("enable_notifications") {
		log.Info("Attempting to send email but notifications are turned off")
		return nil
	}

	if len(viper.GetString("email_from")) == 0 {
		return errors.New("Invalid from address. Please configure a from address before sending email")
	}

	log.WithFields(log.Fields{
		"email": toEmail,
	}).Info("Sending email")

	text := fmt.Sprintf(`
DENSSWeb Job %d

Status: %s

To view your job please visit the following URL:

    %s

Cheers!
	`, jid, status, jobURL)

	qtext, err := quotedBody([]byte(text))
	if err != nil {
		return err
	}

	header := make(textproto.MIMEHeader)
	header.Set("Mime-Version", "1.0")
	header.Set("Date", time.Now().Format(time.RFC1123Z))
	header.Set("To", toEmail)
	header.Set("Subject", fmt.Sprintf("[DENSSWeb] Job %d - %s", jid, status))
	header.Set("From", viper.GetString("email_from"))
	header.Set("Content-Type", "text/plain; charset=UTF-8")
	header.Set("Content-Transfer-Encoding", "quoted-printable")

	c, err := smtp.Dial(fmt.Sprintf("%s:%d", viper.GetString("smtp_host"), viper.GetInt("smtp_port")))
	if err != nil {
		return err
	}
	defer c.Close()

	c.Mail(viper.GetString("email_from"))
	c.Rcpt(toEmail)

	wc, err := c.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	var buf bytes.Buffer
	for k, vv := range header {
		for _, v := range vv {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
	}
	fmt.Fprintf(&buf, "\r\n")

	if _, err = buf.WriteTo(wc); err != nil {
		return err
	}
	if _, err = wc.Write(qtext); err != nil {
		return err
	}

	return nil
}
