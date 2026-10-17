package services

import (
	"net/url"
	"os"
)

// Opt in per isolated proxy instance; retain the legacy transport by default.
func MimoChatURL(ph string) string {
	path := "bot/chat"
	if os.Getenv("MIMO_CHAT_PROTOCOL") == "react" {
		path = "react/chat"
	}
	return "https://aistudio.xiaomimimo.com/open-apis/" + path + "?xiaomichatbot_ph=" + url.QueryEscape(ph)
}
func MimoDialogURL(ph string) string {
	path := "chat/dialog/list"
	if os.Getenv("MIMO_CHAT_PROTOCOL") == "react" {
		path = "react/dialog/list"
	}
	return "https://aistudio.xiaomimimo.com/open-apis/" + path + "?xiaomichatbot_ph=" + url.QueryEscape(ph)
}
