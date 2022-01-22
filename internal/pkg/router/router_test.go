package router

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/benshields/messagebox/internal/pkg/config"
	"github.com/benshields/messagebox/internal/pkg/db"
)

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

	cases := []struct {
		name         string
		req          string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Success 1",
			req:          `{"username":"super.mario"}`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"username":"super.mario"}`,
		},
		{
			name:         "Success 2",
			req:          `{"username":"Yoshi"}`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"username":"Yoshi"}`,
		},
		{
			name:         "Fail on duplicate",
			req:          `{"username":"super.mario"}`,
			expectedCode: http.StatusConflict,
			expectedBody: `{"code":409,"message":"user with the same username already registered"}`,
		},
		{
			name:         "Fail on bad request",
			req:          `{"oh_no":"bad request!"}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(tt.req))
			assert.NoError(t, err)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
			rec.Body.Reset()
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

	cases := []struct {
		name         string
		req          string
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success 1 user",
			req: `{"groupname":"bros",
					"usernames": [
				  		"super.mario"
					]}`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"groupname":"bros","usernames":["super.mario"]}`,
		},
		{
			name: "Success 2 users",
			req: `{"groupname":"pals",
					"usernames": [
						"super.mario",
						"Yoshi"
					]}`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"groupname":"pals","usernames":["super.mario","Yoshi"]}`,
		},
		{
			name: "Fail on duplicate",
			req: `{"groupname":"bros",
					"usernames": [
						"super.mario",
						"luigi"
					]}`,
			expectedCode: http.StatusConflict,
			expectedBody: `{"code":409,"message":"group with the same groupname already registered"}`,
		},
		{
			name: "Fail on missing user",
			req: `{"groupname":"dinos",
					"usernames": [
						"Yoshi",
						"Barney"
					]}`,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"one or more users with given usernames do not exist"}`,
		},
		{
			name: "Fail on bad request",
			req: `{"oh_no":"no group name!",
			"usernames": [
				"luigi",
				"Yoshi"
			]}`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/groups", bytes.NewBufferString(tt.req))
			assert.NoError(t, err)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Equal(t, tt.expectedBody, rec.Body.String())
			rec.Body.Reset()
		})
	}
}

func TestCreateMessage(t *testing.T) {
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
	TRUNCATE messages RESTART IDENTITY CASCADE;
	TRUNCATE user_groups RESTART IDENTITY CASCADE;
	TRUNCATE groups RESTART IDENTITY CASCADE;
	TRUNCATE users RESTART IDENTITY CASCADE;
	INSERT INTO users (name) VALUES ('super.mario');
	INSERT INTO users (name) VALUES ('Yoshi');
	INSERT INTO users (name) VALUES ('luigi');
	INSERT INTO groups (name) VALUES ('green');
	INSERT INTO user_groups (group_id,user_id) VALUES (-1,2);
	INSERT INTO user_groups (group_id,user_id) VALUES (-1,3);
	COMMIT;`
	SeedDB(t, database, seed)

	router := Setup()

	cases := []struct {
		name         string
		req          string
		expectedCode int
		expectedBody string
	}{
		{
			name: "Success with user recipient",
			req: `{
				"sender": "super.mario",
				"recipient": {
				  "username": "luigi"
				},
				"subject": "PR For MessageBox",
				"body": "I have the first version of messagebox ready to review."
			  }`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"id":1,"sender":"super.mario","recipient":{"username":"luigi"},"subject":"PR For MessageBox","body":"I have the first version of messagebox ready to review.","sentAt":"2019-09-03T17:12:42Z"}`,
		},
		{
			name: "Success with group recipient",
			req: `{
				"sender": "super.mario",
				"recipient": {
				  "groupname": "green"
				},
				"subject": "PR For MessageBox",
				"body": "I have the first version of messagebox ready to review."
			  }`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"id":2,"sender":"super.mario","recipient":{"groupname":"green"},"subject":"PR For MessageBox","body":"I have the first version of messagebox ready to review.","sentAt":"2019-09-03T17:12:42Z"}`,
		},
		{
			name: "Fail on missing sender",
			req: `{
				"sender": "bowser",
				"recipient": {
				  "username": "luigi"
				},
				"subject": "PR For MessageBox",
				"body": "I have the first version of messagebox ready to review."
			  }`,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"sender or recipient does not exist"}`,
		},
		{
			name: "Fail on missing user recipient",
			req: `{
				"sender": "super.mario",
				"recipient": {
				  "username": "bowser"
				},
				"subject": "PR For MessageBox",
				"body": "I have the first version of messagebox ready to review."
			  }`,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"sender or recipient does not exist"}`,
		},
		{
			name: "Fail on missing group recipient",
			req: `{
				"sender": "super.mario",
				"recipient": {
				  "groupname": "red"
				},
				"subject": "PR For MessageBox",
				"body": "I have the first version of messagebox ready to review."
			  }`,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"sender or recipient does not exist"}`,
		},
		{
			name: "Fail on bad request",
			req: `{
				"oh_no": "no sender!",
				"recipient": {
				  "username": "luigi"
				},
				"subject": "PR For MessageBox",
				"body": "I have the first version of messagebox ready to review."
			  }`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/messages", bytes.NewBufferString(tt.req))
			assert.NoError(t, err)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedCode == http.StatusCreated {
				// set expectedBody.sentAt to actual value
				expected, err := simplejson.NewJson([]byte(tt.expectedBody))
				assert.NoError(t, err)
				actual, err := simplejson.NewFromReader(rec.Body)
				assert.NoError(t, err)
				actualSentAt := actual.Get("sentAt").MustString()
				assert.NotEmpty(t, actualSentAt)
				expected.Set("sentAt", actualSentAt)
				assert.Equal(t, expected, actual)
			} else {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}

			rec.Body.Reset()
		})
	}
}
