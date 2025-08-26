package captcha

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Turnstile struct {
	Secret string
	Client *http.Client
}

type tsResp struct {
	Success bool `json:"success"`
}

func (t *Turnstile) Verify(token, ip, _ string) bool {
	if token == "" || t.Secret == "" {
		return false
	}
	cl := t.Client
	if cl == nil {
		cl = &http.Client{Timeout: 6 * time.Second}
	}
	form := url.Values{}
	form.Set("secret", t.Secret)
	form.Set("response", token)
	if ip != "" {
		form.Set("remoteip", ip)
	}
	req, _ := http.NewRequestWithContext(context.Background(),
		http.MethodPost, "https://challenges.cloudflare.com/turnstile/v0/siteverify",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := cl.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	var out tsResp
	_ = json.NewDecoder(res.Body).Decode(&out)
	return out.Success
}
