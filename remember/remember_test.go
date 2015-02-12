package remember

import (
	"bytes"
	"net/http"
	"testing"

	"gopkg.in/authboss.v0"
	"gopkg.in/authboss.v0/internal/mocks"
)

func TestInitialize(t *testing.T) {
	testConfig := authboss.NewConfig()

	r := &Remember{}
	err := r.Initialize(testConfig)
	if err == nil {
		t.Error("Expected error about token storers.")
	}

	testConfig.Storer = mocks.MockFailStorer{}
	err = r.Initialize(testConfig)
	if err == nil {
		t.Error("Expected error about token storers.")
	}

	testConfig.Storer = mocks.NewMockStorer()
	err = r.Initialize(testConfig)
	if err != nil {
		t.Error("Unexpected error:", err)
	}
}

func TestAfterAuth(t *testing.T) {
	storer := mocks.NewMockStorer()
	R.storer = storer
	cookies := mocks.NewMockClientStorer()
	session := mocks.NewMockClientStorer()

	req, err := http.NewRequest("POST", "http://localhost", bytes.NewBufferString("rm=true"))
	if err != nil {
		t.Error("Unexpected Error:", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ctx, err := authboss.ContextFromRequest(req)
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	ctx.SessionStorer = session
	ctx.CookieStorer = cookies
	ctx.User = authboss.Attributes{"username": "testuser"}

	R.AfterAuth(ctx)

	if _, ok := cookies.Values[ValueKey]; !ok {
		t.Error("Expected a cookie to have been set.")
	}
}

func TestNew(t *testing.T) {
	storer := mocks.NewMockStorer()
	R.storer = storer
	cookies := mocks.NewMockClientStorer()

	key := "tester"
	token, err := R.New(cookies, key)

	if err != nil {
		t.Error("Unexpected error:", err)
	}

	if len(token) == 0 {
		t.Error("Expected a token.")
	}

	if tok, ok := storer.Tokens[key]; !ok {
		t.Error("Expected it to store against the key:", key)
	} else if len(tok) != 1 || len(tok[0]) == 0 {
		t.Error("Expected a token to be saved.")
	}

	if token != cookies.Values[ValueKey] {
		t.Error("Expected a cookie set with the token.")
	}
}

func TestAuth(t *testing.T) {
	storer := mocks.NewMockStorer()
	R.storer = storer
	cookies := mocks.NewMockClientStorer()
	session := mocks.NewMockClientStorer()

	key := "tester"
	token, err := R.New(cookies, key)
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	outKey, err := R.Auth(cookies, session, token)
	if err != nil {
		t.Error("Unexpected error:", err)
	}

	if session.Values[authboss.HalfAuthKey] != "true" {
		t.Error("The user should have been half-authed.")
	}

	if session.Values[authboss.SessionKey] != key {
		t.Error("The user should have been logged in.")
	}

	if key != outKey {
		t.Error("Keys should have matched:", outKey)
	}
}
