package captcha

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ReCaptcha struct {
	Secret string
	Client *http.Client
}

type rcResp struct {
	Success bool `json:"success"`
}

func (r *ReCaptcha) Verify(token, ip, _ string) bool {
	if token == "" || r.Secret == "" {
		return false
	}
	cl := r.Client
	if cl == nil {
		cl = &http.Client{Timeout: 6 * time.Second}
	}
	form := url.Values{}
	form.Set("secret", r.Secret)
	form.Set("response", token)
	if ip != "" {
		form.Set("remoteip", ip)
	}
	req, _ := http.NewRequestWithContext(context.Background(),
		http.MethodPost, "https://www.google.com/recaptcha/api/siteverify",
		strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := cl.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	var out rcResp
	_ = json.NewDecoder(res.Body).Decode(&out)
	return out.Success
}
