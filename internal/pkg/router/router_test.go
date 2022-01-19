package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	router := Setup()

	w := httptest.NewRecorder()

	cases := []struct {
		name     string
		req      string
		wantCode int
		wantBody string
	}{
		{
			name:     "Success 1",
			req:      `{"username":"super.mario"}`,
			wantCode: http.StatusCreated,
			wantBody: `{"username":"super.mario"}`,
		},
		{
			name:     "Success 2",
			req:      `{"username":"Yoshi"}`,
			wantCode: http.StatusCreated,
			wantBody: `{"username":"Yoshi"}`,
		},
		{
			name:     "Fail on duplicate",
			req:      `{"username":"super.mario"}`,
			wantCode: http.StatusConflict,
			wantBody: `{"code":409,"message":"user with the same username already registered"}`,
		},
		{
			name:     "Fail on bad request",
			req:      `{"oh_no":"bad request!"}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"code":400,"message":"invalid request"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.req))
			assert.NoError(t, err)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
			w.Body.Reset()
		})
	}
}
