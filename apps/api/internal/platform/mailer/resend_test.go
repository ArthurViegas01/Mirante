package mailer

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResendSend(t *testing.T) {
	var gotAuth, gotCT, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"abc-123"}`))
	}))
	defer srv.Close()

	m, err := NewResend("re_test_key", "Mirante <no-reply@mirante.dev>")
	require.NoError(t, err)
	m.endpoint = srv.URL

	require.NoError(t, m.Send(context.Background(), "user@example.com", "Redefinição", "texto", "<p>html</p>"))
	require.Equal(t, "Bearer re_test_key", gotAuth)
	require.Equal(t, "application/json", gotCT)
	require.Contains(t, gotBody, `"to":["user@example.com"]`)
	require.Contains(t, gotBody, `"subject":"Redefinição"`)
	require.Contains(t, gotBody, `"from":"Mirante <no-reply@mirante.dev>"`)
}

func TestResendSurfacesAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"The domain is not verified"}`))
	}))
	defer srv.Close()

	m, _ := NewResend("re_test_key", "no-reply@mirante.dev")
	m.endpoint = srv.URL
	err := m.Send(context.Background(), "u@e.com", "s", "t", "h")
	require.Error(t, err)
	require.Contains(t, err.Error(), "422")
	require.Contains(t, err.Error(), "not verified")
}

func TestNewResendValidation(t *testing.T) {
	_, err := NewResend("", "from@mirante.dev")
	require.Error(t, err)
	_, err = NewResend("re_key", "")
	require.Error(t, err)
}
