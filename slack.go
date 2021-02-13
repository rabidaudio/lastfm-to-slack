package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
	"github.com/slack-go/slack"
)

var Scopes = []string{"users.profile:read", "users.profile:write"}

type SlackClient struct {
	client *slack.Client
	auth   *slack.OAuthV2Response
}

func Authenticate(clientID, clientSecret string) (*SlackClient, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:58513")
	defer listener.Close()

	if err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}
	redirectURI := "http://" + listener.Addr().String() + "/authorized"
	state, err := genState()
	if err != nil {
		return nil, fmt.Errorf("gen state: %w", err)
	}
	errors := make(chan error)
	responses := make(chan *slack.OAuthV2Response)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			w.WriteHeader(400)
			w.Write([]byte("no code provided"))
			return
		}
		rs := r.URL.Query().Get("state")
		if state != rs {
			w.WriteHeader(401)
			w.Write([]byte("state didn't match"))
			return
		}
		sc := http.Client{}
		res, err := slack.GetOAuthV2Response(&sc, clientID, clientSecret, code, redirectURI)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("problem authenticating"))
			log.Printf("oauth error: %v", err)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("Authenticated. You can close this window."))
		responses <- res
	})
	go func() {
		if err = http.Serve(listener, handler); err != nil {
			errors <- fmt.Errorf("start server: %w", err)
		}
		debugPrintf("listening on %v", redirectURI)
	}()

	u := oauthRequestURL(clientID, state, redirectURI)
	fmt.Println("To get started, grant permissions with this link:")
	fmt.Println()
	fmt.Println(u)
	fmt.Println()
	_ = browser.OpenURL(u) // if the browser fails to open, that's okay

	select {
	case res := <-responses:
		sc := slack.New(res.AuthedUser.AccessToken)
		return &SlackClient{client: sc, auth: res}, nil
	case err := <-errors:
		return nil, err
	}
}

func genState() (string, error) {
	var bytes [20]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes[:]), nil
}

func oauthRequestURL(clientID string, state, redirectURI string) string {
	u, _ := url.Parse("https://slack.com/oauth/v2/authorize")
	q := u.Query()
	q.Add("user_scope", strings.Join(Scopes, ","))
	q.Add("client_id", clientID)
	q.Add("state", state)
	q.Add("redirect_uri", redirectURI)
	u.RawQuery = q.Encode()
	return u.String()
}

func (sc *SlackClient) SetStatus(icon, status string) error {
	expireTime := time.Now().Add(10 * time.Minute) // TODO: use the actual track length?
	return sc.client.SetUserCustomStatus(status, icon, expireTime.Unix())
}

func (sc *SlackClient) ClearStatus() error {
	return sc.client.UnsetUserCustomStatus()
}

func (sc *SlackClient) Status() (string, error) {
	profile, err := sc.client.GetUserProfile(sc.auth.AuthedUser.ID, false)
	if err != nil {
		return "", err
	}
	return profile.StatusText, nil
}
