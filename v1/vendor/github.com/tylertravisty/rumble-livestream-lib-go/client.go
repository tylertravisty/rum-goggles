package rumblelivestreamlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"

	"github.com/robertkrimen/otto"
)

const (
	domain        = "rumble.com"
	urlWeb        = "https://" + domain
	urlAccount    = urlWeb + "/account/"
	urlGetSalts   = urlWeb + "/service.php?name=user.get_salts"
	urlUserLogin  = urlWeb + "/service.php?name=user.login"
	urlUserLogout = urlWeb + "/service.php?name=user.logout"
)

type Client struct {
	httpClient    *http.Client
	chatInfo      *ChatInfo
	chatStream    *ChatStream
	chatStreamMu  sync.Mutex
	ApiKey        string
	LiveStreamUrl string
}

func (c *Client) cookies() ([]*http.Cookie, error) {
	u, err := url.Parse(urlWeb)
	if err != nil {
		return nil, fmt.Errorf("error parsing domain: %v", err)
	}
	return c.httpClient.Jar.Cookies(u), nil
}

func (c *Client) PrintCookies() error {
	cookies, err := c.cookies()
	if err != nil {
		return pkgErr("error getting cookies", err)
	}
	fmt.Println("Cookies:", len(cookies))
	for _, cookie := range cookies {
		fmt.Println(cookie.String())
	}

	return nil
}

type NewClientOptions struct {
	Cookies       []*http.Cookie `json:"cookies"`
	ApiKey        string         `json:"stream_key"`
	LiveStreamUrl string         `json:"stream_url"`
}

func NewClient(opts NewClientOptions) (*Client, error) {
	cl, err := newHttpClient(opts.Cookies)
	if err != nil {
		return nil, pkgErr("error creating http client", err)
	}

	return &Client{httpClient: cl, ApiKey: opts.ApiKey, LiveStreamUrl: opts.LiveStreamUrl}, nil
}

func newHttpClient(cookies []*http.Cookie) (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating cookiejar: %v", err)
	}

	url, err := url.Parse(urlWeb)
	if err != nil {
		return nil, fmt.Errorf("error parsing domain: %v", err)
	}
	jar.SetCookies(url, cookies)

	return &http.Client{Jar: jar}, nil
}

type GetSaltsData struct {
	Salts []string `json:"salts"`
}

type GetSaltsResponse struct {
	Data GetSaltsData `json:"data"`
}

func (c *Client) Login(username string, password string) ([]*http.Cookie, error) {
	if c.httpClient == nil {
		return nil, pkgErr("", fmt.Errorf("http client is nil"))
	}

	salts, err := c.getSalts(username)
	if err != nil {
		return nil, pkgErr("error getting salts", err)
	}

	cookies, err := c.userLogin(username, password, salts)
	if err != nil {
		return nil, pkgErr("error logging in", err)
	}

	return cookies, nil
}

func (c *Client) getWebpage(url string) (*http.Response, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http Get request returned error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("http Get response status not %s: %s", http.StatusText(http.StatusOK), resp.Status)
	}

	return resp, nil
}

func (c *Client) getSalts(username string) ([]string, error) {
	u := url.URL{}
	q := u.Query()
	q.Add("username", username)
	body := q.Encode()
	resp, err := c.httpClient.Post(urlGetSalts, "application/x-www-form-urlencoded", strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("http Post request returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http Post response status not %s: %s", http.StatusText(http.StatusOK), resp.Status)
	}

	bodyB, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body bytes: %v", err)
	}

	var gsr GetSaltsResponse
	err = json.NewDecoder(strings.NewReader(string(bodyB))).Decode(&gsr)
	if err != nil {
		return nil, fmt.Errorf("error decoding response body from server: %v", err)
	}

	return gsr.Data.Salts, nil
}

type DataBool struct {
	Session bool `json:"session"`
}

type LoginResponseBool struct {
	Data DataBool `json:"data"`
}

type DataString struct {
	Session string `json:"session"`
}

type LoginResponseString struct {
	Data DataString `json:"data"`
}

func loginResponseSession(body []byte) (string, error) {
	bodyS := string(body)

	var lrs LoginResponseString
	err := json.NewDecoder(strings.NewReader(bodyS)).Decode(&lrs)
	if err == nil {
		return lrs.Data.Session, nil
	}

	var lrb LoginResponseBool
	err = json.NewDecoder(strings.NewReader(bodyS)).Decode(&lrb)
	if err == nil {
		return "false", nil
	}

	return "", fmt.Errorf("error decoding login response")
}

func (c *Client) userLogin(username string, password string, salts []string) ([]*http.Cookie, error) {
	hashes, err := generateHashes(password, salts)
	if err != nil {
		return nil, fmt.Errorf("error generating password hashes: %v", err)
	}

	u := url.URL{}
	q := u.Query()
	q.Add("username", username)
	q.Add("password_hashes", hashes)
	body := q.Encode()
	resp, err := c.httpClient.Post(urlUserLogin, "application/x-www-form-urlencoded", strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("http Post request returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http Post response status not %s: %s", http.StatusText(http.StatusOK), resp.Status)
	}

	bodyB, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body bytes: %v", err)
	}

	session, err := loginResponseSession(bodyB)
	if err != nil {
		return nil, fmt.Errorf("error getting login response session: %v", err)
	}

	if session == "false" {
		return nil, fmt.Errorf("failed to log in")
	}

	return resp.Cookies(), nil
}

func generateHashes(password string, salts []string) (string, error) {
	vm := otto.New()

	vm.Set("password", password)
	vm.Set("salt0", salts[0])
	vm.Set("salt1", salts[1])
	vm.Set("salt2", salts[2])

	_, err := vm.Run(md5)
	if err != nil {
		return "", fmt.Errorf("error running md5 javascript: %v", err)
	}

	value, err := vm.Get("hashes")
	if err != nil {
		return "", fmt.Errorf("error getting hashes value: %v", err)
	}

	hashes, err := value.ToString()
	if err != nil {
		return "", fmt.Errorf("error converting hashes value to string: %v", err)
	}

	return hashes, nil
}

func (c *Client) Logout() error {
	if c.httpClient == nil {
		return pkgErr("", fmt.Errorf("http client is nil"))
	}

	err := c.userLogout()
	if err != nil {
		return pkgErr("error logging out", err)
	}

	return nil
}

func (c *Client) userLogout() error {
	resp, err := c.httpClient.Get(urlUserLogout)
	if err != nil {
		return fmt.Errorf("http Get request returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http Get response status not %s: %s", http.StatusText(http.StatusOK), resp.Status)
	}

	return nil
}

type LoggedInResponseData struct {
	Username string `json:"username"`
}

type LoggedInResponseUser struct {
	ID       string `json:"id"`
	LoggedIn bool   `json:"logged_in"`
}

type LoggedInResponse struct {
	Data LoggedInResponseData `json:"data"`
	User LoggedInResponseUser `json:"user"`
}

func (c *Client) LoggedIn() (*LoggedInResponse, error) {
	resp, err := c.httpClient.Get(urlUserLogin)
	if err != nil {
		return nil, pkgErr("error getting login service", err)
	}
	defer resp.Body.Close()

	bodyB, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, pkgErr("error reading body bytes", err)
	}

	var lir LoggedInResponse
	err = json.NewDecoder(strings.NewReader(string(bodyB))).Decode(&lir)
	if err != nil {
		return nil, pkgErr("error un-marshaling response body", err)
	}

	return &lir, nil
}
