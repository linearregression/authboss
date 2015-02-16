package expire

import (
	"net/http"
	"testing"
	"time"

	"gopkg.in/authboss.v0"
	"gopkg.in/authboss.v0/internal/mocks"
)

func TestExpire_Touch(t *testing.T) {
	authboss.NewConfig()
	session := mocks.NewMockClientStorer()

	if _, ok := session.Get(UserLastAction); ok {
		t.Error("It should not have been set")
	}
	Touch(session)
	if dateStr, ok := session.Get(UserLastAction); !ok || len(dateStr) == 0 {
		t.Error("It should have been set")
	} else if date, err := time.Parse(time.RFC3339, dateStr); err != nil {
		t.Error("Date is malformed:", dateStr)
	} else if date.After(time.Now().UTC()) {
		t.Error("The time is set in the future.")
	}
}

func TestExpire_BeforeAuth(t *testing.T) {
	authboss.NewConfig()
	authboss.Cfg.ExpireAfter = time.Hour
	expire := &Expire{}
	session := mocks.NewMockClientStorer()

	ctx := mocks.MockRequestContext()
	ctx.SessionStorer = session

	if err := expire.BeforeAuth(ctx); err != nil {
		t.Error("There's no user in session, should be no-op.")
	}

	session.Values[authboss.SessionKey] = "moo"
	session.Values[UserLastAction] = "cow"
	if err := expire.BeforeAuth(ctx); err != nil {
		t.Error("There's a malformed date, this should not error, just fix it:", err)
	}
	if dateStr, ok := session.Get(UserLastAction); !ok || len(dateStr) == 0 {
		t.Error("It should have been set")
	} else if date, err := time.Parse(time.RFC3339, dateStr); err != nil {
		t.Error("Date is malformed:", dateStr)
	} else if date.After(time.Now().UTC()) {
		t.Error("The time is set in the future.")
	}

	session.Values[UserLastAction] = time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	if err := expire.BeforeAuth(ctx); err != ErrExpired {
		t.Error("The user should have been expired, got:", err)
	}

	if _, ok := session.Values[authboss.SessionKey]; ok {
		t.Error("The user session should have been expired.")
	}
}

type testHandler bool

func (t *testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*t = true
}

func TestExpire_Middleware(t *testing.T) {
	authboss.NewConfig()
	session := mocks.NewMockClientStorer()
	session.Values = map[string]string{
		authboss.SessionKey: "username",
	}
	maker := func(w http.ResponseWriter, r *http.Request) authboss.ClientStorer { return session }

	handler := new(testHandler)
	touch := Middleware(maker, handler)

	touch.ServeHTTP(nil, nil)
	if !*handler {
		t.Error("Expected middleware's chain to be called.")
	}

	if dateStr, ok := session.Get(UserLastAction); !ok || len(dateStr) == 0 {
		t.Error("It should have been set")
	} else if date, err := time.Parse(time.RFC3339, dateStr); err != nil {
		t.Error("Date is malformed:", dateStr)
	} else if date.After(time.Now().UTC()) {
		t.Error("The time is set in the future.")
	}
}
