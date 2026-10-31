package fedex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	fedexapi "github.com/befship/fedex-sdk/internal/generated/fedexapi"
)

const (
	EnvSandbox    = "sandbox"
	EnvProduction = "production"
)

// Config holds FedEx OAuth credentials and runtime options.
type Config struct {
	ClientID      string
	ClientSecret  string
	AccountNumber string
	Environment   string // sandbox (default) | production
	BaseURL       string // optional override

	HTTPClient *http.Client
	Timeout    time.Duration
	MaxRetries int
	Logger     Logger
}

// Logger is a minimal structured logger hook.
type Logger interface {
	Debugf(format string, args ...any)
	Errorf(format string, args ...any)
}

// Creds is a compact credential bundle (adapter-friendly).
type Creds struct {
	ClientID      string
	ClientSecret  string
	AccountNumber string
	Environment   string
}

// Client is the hand-written FedEx facade (OAuth cache, retries, typed API access).
type Client struct {
	cfg   Config
	http  *http.Client
	token *tokenCache
}

// New builds a Client.
func New(cfg Config) *Client {
	if cfg.Environment == "" {
		cfg.Environment = EnvSandbox
	}
	switch strings.ToLower(cfg.Environment) {
	case "test", "cie", EnvSandbox:
		cfg.Environment = EnvSandbox
	case "prod", EnvProduction:
		cfg.Environment = EnvProduction
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	} else if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 2
	}
	hc := cfg.HTTPClient
	if hc == nil {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 60 * time.Second
		}
		hc = &http.Client{Timeout: timeout}
	}
	c := &Client{cfg: cfg, http: hc, token: newTokenCache()}
	c.ensureRetry()
	return c
}

// NewFromCreds builds a Client from a Creds bundle (adapter-friendly).
func NewFromCreds(cr Creds) *Client {
	return New(Config{
		ClientID:      cr.ClientID,
		ClientSecret:  cr.ClientSecret,
		AccountNumber: cr.AccountNumber,
		Environment:   cr.Environment,
	})
}

func (c *Client) Environment() string   { return c.cfg.Environment }
func (c *Client) HTTPClient() *http.Client { return c.http }
func (c *Client) AccountNumber() string { return c.cfg.AccountNumber }

// BaseURL returns the FedEx API host for env.
func BaseURL(env string) string {
	if env == EnvProduction {
		return "https://apis.fedex.com"
	}
	return "https://apis-sandbox.fedex.com"
}

func (c *Client) base() string {
	if c.cfg.BaseURL != "" {
		return strings.TrimRight(c.cfg.BaseURL, "/")
	}
	return BaseURL(c.cfg.Environment)
}

// Raw returns a generated client without auth editor.
func (c *Client) Raw() (*ClientWithResponses, error) {
	return fedexapi.NewClientWithResponses(c.base(), fedexapi.WithHTTPClient(c.http))
}

// AccessToken returns a cached or freshly fetched OAuth bearer token.
func (c *Client) AccessToken(ctx context.Context) (string, error) {
	if c.cfg.ClientID == "" || c.cfg.ClientSecret == "" {
		return "", fmt.Errorf("fedex: ClientID and ClientSecret are required")
	}
	key := c.cfg.Environment + ":" + c.cfg.ClientID
	if tok, ok := c.token.get(key); ok {
		return tok, nil
	}
	gc, err := c.Raw()
	if err != nil {
		return "", err
	}
	rsp, err := gc.OauthTokenWithFormdataBodyWithResponse(ctx, OauthTokenFormdataRequestBody{
		ClientId:     c.cfg.ClientID,
		ClientSecret: c.cfg.ClientSecret,
		GrantType:    ClientCredentials,
	})
	if err != nil {
		return "", &APIError{Op: "oauth", Message: "token request failed", Err: err}
	}
	if rsp.StatusCode() >= 300 || rsp.JSON200 == nil || rsp.JSON200.AccessToken == nil || *rsp.JSON200.AccessToken == "" {
		return "", &APIError{Op: "oauth", StatusCode: rsp.StatusCode(), Body: string(rsp.Body), Message: "token request rejected"}
	}
	tok := *rsp.JSON200.AccessToken
	sec := 0
	if rsp.JSON200.ExpiresIn != nil {
		sec = *rsp.JSON200.ExpiresIn
	}
	c.token.set(key, tok, sec)
	return tok, nil
}

// InvalidateToken clears the cached token for this client.
func (c *Client) InvalidateToken() {
	c.token.invalidate(c.cfg.Environment + ":" + c.cfg.ClientID)
}

// AuthedClient returns a generated client plus a Bearer RequestEditorFn.
func (c *Client) AuthedClient(ctx context.Context) (*ClientWithResponses, RequestEditorFn, error) {
	tok, err := c.AccessToken(ctx)
	if err != nil {
		return nil, nil, err
	}
	gc, err := c.Raw()
	if err != nil {
		return nil, nil, err
	}
	editor := RequestEditorFn(func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+tok)
		return nil
	})
	return gc, editor, nil
}

// LocaleParams returns x-customer-transaction-id and x-locale for request params.
func LocaleParams() (txid XCustomerTransactionId, locale XLocale) {
	txid = XCustomerTransactionId(uuid.NewString())
	locale = XLocale("en_US")
	return txid, locale
}

// CheckJSON maps a JSON-200 style response into a map.
func CheckJSON(method, rawURL string, status int, body []byte, json200 *map[string]interface{}) (map[string]any, error) {
	if status >= 300 {
		return nil, fmt.Errorf("%s %s: status=%d body=%s", method, rawURL, status, truncate(string(body), 800))
	}
	if json200 != nil {
		return *json200, nil
	}
	if len(body) == 0 {
		return nil, nil
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("decode %s: %w body=%s", rawURL, err, truncate(string(body), 400))
	}
	return m, nil
}
