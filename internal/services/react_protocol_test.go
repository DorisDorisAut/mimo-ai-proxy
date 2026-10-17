package services

import (
	"strings"
	"testing"
)

func TestReactEndpoints(t *testing.T) {
	for _, protocol := range []string{"", "react"} {
		t.Run(protocol, func(t *testing.T) {
			t.Setenv("MIMO_CHAT_PROTOCOL", protocol)
			path := "bot/chat"
			history := "chat/dialog/list"
			if protocol == "react" {
				path = "react/chat"
				history = "react/dialog/list"
			}
			if !strings.Contains(MimoChatURL("a&b= c"), path+"?xiaomichatbot_ph=a%26b%3D+c") {
				t.Fatal("incorrect chat endpoint/escaping")
			}
			if !strings.Contains(MimoDialogURL("a&b= c"), history+"?xiaomichatbot_ph=a%26b%3D+c") {
				t.Fatal("incorrect history endpoint/escaping")
			}
		})
	}
}
