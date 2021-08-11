package login_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ory/kratos/internal"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ory/x/urlx"

	"github.com/ory/kratos/selfservice/flow"
	"github.com/ory/kratos/selfservice/flow/login"
	"github.com/ory/kratos/x"
)

func TestFakeFlow(t *testing.T) {
	var r login.Flow
	require.NoError(t, faker.FakeData(&r))

	assert.NotEmpty(t, r.ID)
	assert.NotEmpty(t, r.IssuedAt)
	assert.NotEmpty(t, r.ExpiresAt)
	assert.NotEmpty(t, r.RequestURL)
	assert.NotEmpty(t, r.Active)
	assert.NotNil(t, r.UI)
}

func TestNewFlow(t *testing.T) {
	conf, _ := internal.NewFastRegistryWithMocks(t)

	t.Run("type=browser", func(t *testing.T) {
		t.Run("case=regular flow creation without a session", func(t *testing.T) {
			r := login.NewFlow(conf, 0, "csrf", &http.Request{
				URL:  urlx.ParseOrPanic("/"),
				Host: "ory.sh", TLS: &tls.ConnectionState{},
			}, flow.TypeBrowser)
			assert.EqualValues(t, r.IssuedAt, r.ExpiresAt)
			assert.Equal(t, flow.TypeBrowser, r.Type)
			assert.False(t, r.Refresh)
			assert.Equal(t, "https://ory.sh/", r.RequestURL)
		})

		t.Run("case=regular flow creation", func(t *testing.T) {
			r := login.NewFlow(conf, 0, "csrf", &http.Request{
				URL:  urlx.ParseOrPanic("https://ory.sh/"),
				Host: "ory.sh"}, flow.TypeBrowser)
			assert.Equal(t, "https://ory.sh/", r.RequestURL)
		})
	})

	t.Run("type=api", func(t *testing.T) {
		t.Run("case=flow with refresh", func(t *testing.T) {
			r := login.NewFlow(conf, 0, "csrf", &http.Request{
				URL:  urlx.ParseOrPanic("/?refresh=true"),
				Host: "ory.sh"}, flow.TypeAPI)
			assert.Equal(t, r.IssuedAt, r.ExpiresAt)
			assert.Equal(t, flow.TypeAPI, r.Type)
			assert.True(t, r.Refresh)
			assert.Equal(t, "http://ory.sh/?refresh=true", r.RequestURL)
		})

		t.Run("case=flow without refresh", func(t *testing.T) {
			r := login.NewFlow(conf, 0, "csrf", &http.Request{
				URL:  urlx.ParseOrPanic("/"),
				Host: "ory.sh"}, flow.TypeAPI)
			assert.Equal(t, r.IssuedAt, r.ExpiresAt)
			assert.Equal(t, flow.TypeAPI, r.Type)
			assert.False(t, r.Refresh)
			assert.Equal(t, "http://ory.sh/", r.RequestURL)
		})
	})
}

func TestFlow(t *testing.T) {
	r := &login.Flow{ID: x.NewUUID()}
	assert.Equal(t, r.ID, r.GetID())

	t.Run("case=expired", func(t *testing.T) {
		for _, tc := range []struct {
			r     *login.Flow
			valid bool
		}{
			{
				r:     &login.Flow{ExpiresAt: time.Now().Add(time.Hour), IssuedAt: time.Now().Add(-time.Minute)},
				valid: true,
			},
			{r: &login.Flow{ExpiresAt: time.Now().Add(-time.Hour), IssuedAt: time.Now().Add(-time.Minute)}},
		} {
			if tc.valid {
				require.NoError(t, tc.r.Valid())
			} else {
				require.Error(t, tc.r.Valid())
			}
		}
	})
}

func TestGetType(t *testing.T) {
	for _, ft := range []flow.Type{
		flow.TypeAPI,
		flow.TypeBrowser,
	} {
		t.Run(fmt.Sprintf("case=%s", ft), func(t *testing.T) {
			r := &login.Flow{Type: ft}
			assert.Equal(t, ft, r.GetType())
		})
	}
}

func TestGetRequestURL(t *testing.T) {
	expectedURL := "http://foo/bar/baz"
	f := &login.Flow{RequestURL: expectedURL}
	assert.Equal(t, expectedURL, f.GetRequestURL())
}
