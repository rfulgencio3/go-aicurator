package sendgrid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/seu-usuario/go-aicurator/internal/config"
	"github.com/seu-usuario/go-aicurator/internal/email"
)

const sendGridURL = "https://api.sendgrid.com/v3/mail/send"

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type sgPayload struct {
	Personalizations []personalization `json:"personalizations"`
	From             address           `json:"from"`
	Subject          string            `json:"subject"`
	Content          []content         `json:"content"`
}

type personalization struct {
	To []address `json:"to"`
}

type address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type content struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// Send envia o digest por e-mail via SendGrid.
func (c *Client) Send(subject, digestText string) error {
	var tos []address
	for _, e := range c.cfg.EmailTo {
		tos = append(tos, address{Email: e})
	}

	payload := sgPayload{
		Personalizations: []personalization{{To: tos}},
		From:             address{Email: c.cfg.EmailFrom, Name: c.cfg.EmailFromName},
		Subject:          subject,
		Content: []content{
			{Type: "text/plain", Value: digestText},
			{Type: "text/html", Value: email.TextToHTML(digestText)},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serializar payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, sendGridURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.SendGridAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("enviar e-mail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SendGrid status %d: %s", resp.StatusCode, string(raw))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
