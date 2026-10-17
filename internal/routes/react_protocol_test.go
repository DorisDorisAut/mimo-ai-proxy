package routes

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"mimoproxy/internal/models"
	"mimoproxy/internal/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReactNonStream(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name, body      string
		status          int
		content, reason string
	}{
		{"answer", "event: message\ndata: {\"content\":\"OK\"}\n\nevent: finish\ndata: {\"content\":\"[DONE]\"}\n\n", 200, "OK", ""},
		{"reasoning", "event: message\ndata: {\"reasoningContent\":\"private reasoning\"}\n\nevent: message\ndata: {\"content\":\"OK\"}\n\n", 200, "OK", "private reasoning"},
		{"empty", "event: finish\ndata: {\"content\":\"[DONE]\"}\n\n", 502, "", ""},
		{"upstream-error", "event: error\ndata: {\"content\":\"sensitive provider detail\"}\n\n", 502, "", ""},
		{"typed-error", "event: message\ndata: {\"messageType\":\"error\",\"content\":\"sensitive provider detail\"}\n\n", 502, "", ""},
		{"late-error", "event: message\ndata: {\"content\":\"partial\"}\n\nevent: error\ndata: {\"content\":\"sensitive provider detail\"}\n\n", 502, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			processNonStream(c, strings.NewReader(tc.body), "test", "mimo-v2.6-flash", "fixture", "", "")
			if w.Code != tc.status {
				t.Fatalf("status %d", w.Code)
			}
			if strings.Contains(w.Body.String(), "sensitive provider detail") {
				t.Fatal("raw error leaked")
			}
			if tc.status == 200 {
				var d struct {
					Choices []struct {
						Message struct {
							Content string `json:"content"`
							Reason  string `json:"reasoning_content"`
						} `json:"message"`
					} `json:"choices"`
				}
				if err := json.Unmarshal(w.Body.Bytes(), &d); err != nil {
					t.Fatal(err)
				}
				if d.Choices[0].Message.Content != tc.content || d.Choices[0].Message.Reason != tc.reason {
					t.Fatalf("incorrect answer/reasoning: %s", w.Body)
				}
			}
		})
	}
}
func TestReactStreamingError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Header("Content-Type", "text/event-stream")
	processStream(c, strings.NewReader("event: error\ndata: {\"content\":\"sensitive provider detail\"}\n\n"), "test", "mimo-v2.6-flash", "", "")
	s := w.Body.String()
	if !strings.Contains(s, "upstream_error") || strings.Contains(s, "sensitive provider detail") || strings.Contains(s, "\"finish_reason\":\"stop\"") {
		t.Fatalf("incorrect failure stream: %s", s)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestReactStatelessCompanySeparation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("MIMO_CHAT_PROTOCOL", "react")
	t.Setenv("MIMO_STATELESS", "1")
	t.Setenv("SERVICE_TOKEN", "synthetic-token")
	t.Setenv("USER_ID", "synthetic-user")
	t.Setenv("XIAOMI_CHATBOT_PH", "synthetic-ph")
	old := services.GlobalHTTPClient
	defer func() { services.GlobalHTTPClient = old }()
	var conversations []string
	var queries []string
	services.GlobalHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/open-apis/chat/conversation/save" {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"code":0}`))}, nil
		}
		if r.URL.Path != "/open-apis/react/chat" {
			t.Fatalf("unexpected endpoint %s", r.URL.Path)
		}
		var payload models.MimoPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		conversations = append(conversations, payload.ConversationID)
		queries = append(queries, payload.Query)
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("event: message\ndata: {\"content\":\"OK\"}\n\nevent: finish\ndata: {\"content\":\"[DONE]\"}\n\n"))}, nil
	})}
	for _, company := range []string{"Synthetic Company A", "Synthetic Company B"} {
		body, _ := json.Marshal(map[string]interface{}{"user": "shared-untrusted-id", "model": "mimo-v2.6-flash", "stream": false, "messages": []map[string]string{{"role": "system", "content": company}, {"role": "user", "content": "same question"}}})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/v1/chat/completions", strings.NewReader(string(body)))
		c.Request.Header.Set("Content-Type", "application/json")
		handleChatCompletions(c)
		if w.Code != 200 {
			t.Fatalf("%d %s", w.Code, w.Body)
		}
	}
	if len(conversations) != 2 || conversations[0] == conversations[1] {
		t.Fatal("conversation reused across requests")
	}
	if !strings.Contains(queries[0], "Company A") || strings.Contains(queries[0], "Company B") || !strings.Contains(queries[1], "Company B") || strings.Contains(queries[1], "Company A") {
		t.Fatal("company context crossed")
	}
}
