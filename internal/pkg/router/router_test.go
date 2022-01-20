package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/benshields/messagebox/internal/pkg/config"
	"github.com/benshields/messagebox/internal/pkg/db"
)

func TestCreateUser(t *testing.T) {
	dbCfg := config.DatabaseConfiguration{
		DatabaseName: "messagebox",
		User:         "messagebox_user",
		Password:     "insecure",
		Host:         "0.0.0.0",
		Port:         "5432",
	}

	database, err := db.Setup(dbCfg, nil)
	if err != nil {
		t.Fatal("db.Setup() failed with:", err)
	}

	seed := `TRUNCATE users CASCADE;`
	SeedDB(t, database, seed)

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

func TestCreateGroup(t *testing.T) {
	dbCfg := config.DatabaseConfiguration{
		DatabaseName: "messagebox",
		User:         "messagebox_user",
		Password:     "insecure",
		Host:         "0.0.0.0",
		Port:         "5432",
	}

	database, err := db.Setup(dbCfg, nil)
	if err != nil {
		t.Fatal("db.Setup() failed with:", err)
	}

	seed := `BEGIN;
	TRUNCATE user_groups CASCADE;
	TRUNCATE groups CASCADE;
	TRUNCATE users CASCADE;
	INSERT INTO users (name) VALUES ('super.mario');
	INSERT INTO users (name) VALUES ('Yoshi');
	INSERT INTO users (name) VALUES ('luigi');
	COMMIT;`
	SeedDB(t, database, seed)

	router := Setup()

	w := httptest.NewRecorder()

	cases := []struct {
		name     string
		req      string
		wantCode int
		wantBody string
	}{
		{
			name: "Success 1 use",
			req: `{"groupname":"bros",
					"usernames": [
				  		"super.mario"
					]}`,
			wantCode: http.StatusCreated,
			wantBody: `{"groupname":"bros","usernames":["super.mario"]}`,
		},
		{
			name: "Success 2 users",
			req: `{"groupname":"pals",
					"usernames": [
						"super.mario",
						"Yoshi"
					]}`,
			wantCode: http.StatusCreated,
			wantBody: `{"groupname":"pals","usernames":["super.mario","Yoshi"]}`,
		},
		{
			name: "Fail on duplicate",
			req: `{"groupname":"bros",
					"usernames": [
						"super.mario",
						"luigi"
					]}`,
			wantCode: http.StatusConflict,
			wantBody: `{"code":409,"message":"group with the same groupname already registered"}`,
		},
		{
			name: "Fail on missing user",
			req: `{"groupname":"dinos",
					"usernames": [
						"Yoshi",
						"Barney"
					]}`,
			wantCode: http.StatusConflict,
			wantBody: `{"code":404,"message":"one or more users with given usernames do not exist"}`,
		},
		{
			name: "Fail on bad request",
			req: `{"oh_no":"no group name!",
			"usernames": [
				"luigi",
				"Yoshi"
			]}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"code":400,"message":"invalid request"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/groups", bytes.NewBufferString(tt.req))
			assert.NoError(t, err)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusCreated, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
			w.Body.Reset()
		})
	}
}

func SeedDB(t *testing.T, conn *gorm.DB, seed string) {
	reqs := strings.Split(strings.TrimSpace(seed), ";")
	t.Log("seeding db")
	for _, req := range reqs {
		if req != "" {
			if err := conn.Exec(req).Error; err != nil {
				t.Fatal("seed failed ", err)
			}
		}
	}
	t.Log("seeding complete, continuing test")
}
