package resend

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

const resendURL = "https://api.resend.com/emails"

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

type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
	HTML    string   `json:"html"`
}

// Send envia o digest por e-mail via Resend.
func (c *Client) Send(subject, digestText string) error {
	from := fmt.Sprintf("%s <%s>", c.cfg.EmailFromName, c.cfg.EmailFrom)

	payload := resendPayload{
		From:    from,
		To:      c.cfg.EmailTo,
		Subject: subject,
		Text:    digestText,
		HTML:    email.TextToHTML(digestText),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("serializar payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, resendURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("criar request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.cfg.ResendAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("enviar e-mail: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Resend status %d: %s", resp.StatusCode, string(raw))
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
