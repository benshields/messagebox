package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/benshields/messagebox/internal/pkg/config"
	"github.com/benshields/messagebox/internal/pkg/db"
	"github.com/benshields/messagebox/internal/pkg/models"
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
			name:         "Success with unicode",
			req:          `{"username":"ɑɛϴИДѱ"}`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"username":"ɑɛϴИДѱ"}`,
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
		{
			name:         "Fail on too long",
			req:          `{"username":"012345678901234567890123456789012"}`,
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

func TestGetMailbox(t *testing.T) {
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
	INSERT INTO users (name) VALUES ('toad');
	INSERT INTO users (name) VALUES ('shy.guy');
	INSERT INTO groups (name) VALUES ('green');
	INSERT INTO groups (name) VALUES ('GOATs');
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,2);
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,3);
	INSERT INTO user_groups (group_id, user_id) VALUES (-2,2);
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, 2, 'hello', 'user');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, -1, 'hello', 'group');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (1, 3, 1, 're: hello', 'use');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (1, 3, 1, 're: hello', 'user*');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (2, 3, -1, 're: hello', 'group');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (2, 2, -1, 're: hello', 'group again');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, 5, 'hi', 'shy guy');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 5, 2, 'hi yoshi', 'from shy guy');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 4, -2, 'hi GOATs', 'from toad');
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
			name:         "Success with no messages",
			req:          "toad",
			expectedCode: http.StatusOK,
			expectedBody: `[]`,
		},
		{
			name:         "Success with 1 direct message",
			req:          "shy.guy",
			expectedCode: http.StatusOK,
			expectedBody: `[
				{"id":7,"sender":"super.mario","recipient":{"username":"shy.guy"},"subject":"hi","body":"shy guy","sentAt":"2019-09-03T17:12:42Z"}]`,
		},
		{
			name:         "Success with 2 direct messages",
			req:          "super.mario",
			expectedCode: http.StatusOK,
			expectedBody: `[
				{"id":3,"re":1,"sender":"luigi","recipient":{"username":"super.mario"},"subject":"re: hello","body":"use","sentAt":"2019-09-03T17:12:42Z"},
				{"id":4,"re":1,"sender":"luigi","recipient":{"username":"super.mario"},"subject":"re: hello","body":"user*","sentAt":"2019-09-03T17:12:42Z"}]`,
		},
		{
			name:         "Success with 2 direct messages & 4 groups messages from 2 groups",
			req:          "Yoshi",
			expectedCode: http.StatusOK,
			expectedBody: `[
				{"id":1,"sender":"super.mario","recipient":{"username":"Yoshi"},"subject":"hello","body":"user","sentAt":"2019-09-03T17:12:42Z"},
				{"id":2,"sender":"super.mario","recipient":{"groupname":"green"},"subject":"hello","body":"group","sentAt":"2019-09-03T17:12:42Z"},
				{"id":5,"re":2,"sender":"luigi","recipient":{"groupname":"green"},"subject":"re: hello","body":"group","sentAt":"2019-09-03T17:12:42Z"},
				{"id":6,"re":2,"sender":"Yoshi","recipient":{"groupname":"green"},"subject":"re: hello","body":"group again","sentAt":"2019-09-03T17:12:42Z"},
				{"id":8,"sender":"shy.guy","recipient":{"username":"Yoshi"},"subject":"hi yoshi","body":"from shy guy","sentAt":"2019-09-03T17:12:42Z"},
				{"id":9,"sender":"toad","recipient":{"groupname":"GOATs"},"subject":"hi GOATs","body":"from toad","sentAt":"2019-09-03T17:12:42Z"}]`,
		},
		{
			name:         "Fail on no username",
			req:          "",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
		{
			name:         "Fail on missing username",
			req:          "bowser",
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"user with given username does not exist"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/users/"+tt.req+"/mailbox", nil)
			assert.NoError(t, err)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedCode == http.StatusOK {
				// set expectedBody.sentAt to actual value
				var expected []models.Message
				err := json.Unmarshal([]byte(tt.expectedBody), &expected)
				assert.NoError(t, err)

				var actual []models.Message
				err = json.Unmarshal(rec.Body.Bytes(), &actual)
				assert.NoError(t, err)

				for _, exp := range expected {
					found := false
					for _, act := range actual {
						if exp.ID == act.ID {
							found = true
							exp.SentAt = act.SentAt
							assert.Equal(t, exp, act)
						}
					}
					assert.Truef(t, found, "expected message with ID %d not found\nexpected: %v\nactual  : %v", exp.ID, expected, actual)
				}
			} else {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}

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
		{
			name: "Fail on bad request",
			req: `{"groupname":"012345678901234567890123456789012",
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
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,2);
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,3);
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
			name: "Success with no body",
			req: `{
				"sender": "super.mario",
				"recipient": {
				  "groupname": "green"
				},
				"subject": "PR For MessageBox"
			  }`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"id":3,"sender":"super.mario","recipient":{"groupname":"green"},"subject":"PR For MessageBox","sentAt":"2019-09-03T17:12:42Z"}`,
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
		{
			name: "Fail on subject too long (256)",
			req: `{
				"sender": "super.mario",
				"recipient": {
				  "username": "luigi"
				},
				"subject": "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345",
				"body": "I have the first version of messagebox ready to review."
			  }`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
		{
			name: "Fail on body too long (2001)",
			req: `{
				"sender": "super.mario",
				"recipient": {
				  "username": "luigi"
				},
				"subject": "body too long",
				"body": "012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
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

func TestGetMessage(t *testing.T) {
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
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,2);
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,3);
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, 2, 'hello', 'user');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, -1, 'hello', 'group');
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
			name:         "Success with user recipient",
			req:          "1",
			expectedCode: http.StatusOK,
			expectedBody: `{"id":1,"sender":"super.mario","recipient":{"username":"Yoshi"},"subject":"hello","body":"user","sentAt":"2019-09-03T17:12:42Z"}`,
		},
		{
			name:         "Success with group recipient",
			req:          "2",
			expectedCode: http.StatusOK,
			expectedBody: `{"id":2,"sender":"super.mario","recipient":{"groupname":"green"},"subject":"hello","body":"group","sentAt":"2019-09-03T17:12:42Z"}`,
		},
		{
			name:         "Fail on no id",
			req:          "",
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"no route found"}`,
		},
		{
			name:         "Fail on missing id",
			req:          "3",
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"message ID does not exist"}`,
		},
		{
			name:         "Fail on bad request",
			req:          "abc",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/messages/"+tt.req, nil)
			assert.NoError(t, err)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedCode == http.StatusOK {
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

func TestCreateReply(t *testing.T) {
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
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,2);
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,3);
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, 2, 'hello', 'user');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, -1, 'hello', 'group');
	COMMIT;`
	SeedDB(t, database, seed)

	router := Setup()

	cases := []struct {
		name         string
		reqID        string
		req          string
		expectedCode int
		expectedBody string
	}{
		{
			name:  "Success with reply to user",
			reqID: "1",
			req: `{
				"sender": "luigi",
				"subject": "re: hello",
				"body": "user"
			  }`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"id":3,"re":1,"sender":"luigi","recipient":{"username":"super.mario"},"subject":"re: hello","body":"user","sentAt":"2019-09-03T17:12:42Z"}`,
		},
		{
			name:  "Success with reply to group",
			reqID: "2",
			req: `{
				"sender": "luigi",
				"subject": "re: hello",
				"body": "group"
			  }`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"id":4,"re":2,"sender":"luigi","recipient":{"groupname":"green"},"subject":"re: hello","body":"group","sentAt":"2019-09-03T17:12:42Z"}`,
		},
		{
			name:  "Success with no body",
			reqID: "2",
			req: `{
				"sender": "luigi",
				"subject": "re: hello"
			  }`,
			expectedCode: http.StatusCreated,
			expectedBody: `{"id":5,"re":2,"sender":"luigi","recipient":{"groupname":"green"},"subject":"re: hello","sentAt":"2019-09-03T17:12:42Z"}`,
		},
		{
			name:         "Fail on no id",
			reqID:        "",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
		{
			name:  "Fail on missing id",
			reqID: "42",
			req: `{
				"sender": "luigi",
				"subject": "re: hello",
				"body": "group"
			  }`,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"sender or message ID does not exist"}`,
		},
		{
			name:  "Fail on missing sender",
			reqID: "2",
			req: `{
				"sender": "bowser",
				"subject": "re: hello",
				"body": "group"
			  }`,
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"sender or message ID does not exist"}`,
		},
		{
			name:  "Fail on bad request",
			reqID: "1",
			req: `{
				"sender": "luigi",
				"oh_no": "no subject!",
				"body": "user"
			  }`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
		{
			name:  "Fail on subject too long (256)",
			reqID: "1",
			req: `{
				"sender": "luigi",
				"subject": "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345",
				"body": "user"
			  }`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
		{
			name:  "Fail on bad request",
			reqID: "1",
			req: `{
				"sender": "luigi",
				"subject": "body too long",
				"body": "012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
			  }`,
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/messages/"+tt.reqID+"/replies", bytes.NewBufferString(tt.req))
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

func TestGetReplies(t *testing.T) {
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
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,2);
	INSERT INTO user_groups (group_id, user_id) VALUES (-1,3);
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, 2, 'hello', 'user');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (0, 1, -1, 'hello', 'group');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (1, 3, 1, 're: hello', 'use');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (1, 3, 1, 're: hello', 'user*');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (2, 3, -1, 're: hello', 'group');
	INSERT INTO messages (re, sender, recipient, subject, body) VALUES (2, 2, -1, 're: hello', 'group again');
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
			name:         "Success with original user recipient",
			req:          "1",
			expectedCode: http.StatusOK,
			expectedBody: `[
				{"id":3,"re":1,"sender":"luigi","recipient":{"username":"super.mario"},"subject":"re: hello","body":"use","sentAt":"2019-09-03T17:12:42Z"},
				{"id":4,"re":1,"sender":"luigi","recipient":{"username":"super.mario"},"subject":"re: hello","body":"user*","sentAt":"2019-09-03T17:12:42Z"}]`,
		},
		{
			name:         "Success with original group recipient",
			req:          "2",
			expectedCode: http.StatusOK,
			expectedBody: `[
				{"id":5,"re":2,"sender":"luigi","recipient":{"groupname":"green"},"subject":"re: hello","body":"group","sentAt":"2019-09-03T17:12:42Z"},
				{"id":6,"re":2,"sender":"Yoshi","recipient":{"groupname":"green"},"subject":"re: hello","body":"group again","sentAt":"2019-09-03T17:12:42Z"}]`,
		},
		{
			name:         "Success with original user recipient but no replies",
			req:          "3",
			expectedCode: http.StatusOK,
			expectedBody: `[]`,
		},
		{
			name:         "Success with original group recipient but no replies",
			req:          "5",
			expectedCode: http.StatusOK,
			expectedBody: `[]`,
		},
		{
			name:         "Fail on no id",
			req:          "",
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"code":400,"message":"invalid request"}`,
		},
		{
			name:         "Fail on missing id",
			req:          "42",
			expectedCode: http.StatusNotFound,
			expectedBody: `{"code":404,"message":"message ID does not exist"}`,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/messages/"+tt.req+"/replies", nil)
			assert.NoError(t, err)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			assert.Equal(t, tt.expectedCode, rec.Code)

			if tt.expectedCode == http.StatusOK {
				// set expectedBody.sentAt to actual value
				var expected []models.Message
				err := json.Unmarshal([]byte(tt.expectedBody), &expected)
				assert.NoError(t, err)

				var actual []models.Message
				err = json.Unmarshal(rec.Body.Bytes(), &actual)
				assert.NoError(t, err)

				for _, exp := range expected {
					found := false
					for _, act := range actual {
						if exp.ID == act.ID {
							found = true
							exp.SentAt = act.SentAt
							assert.Equal(t, exp, act)
						}
					}
					assert.Truef(t, found, "expected message with ID %d not found\nexpected: %v\nactual  : %v", exp.ID, expected, actual)
				}
			} else {
				assert.Equal(t, tt.expectedBody, rec.Body.String())
			}

			rec.Body.Reset()
		})
	}
}
