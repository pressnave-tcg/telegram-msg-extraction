package model

import "time"

type Group struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Kind  string `json:"kind"`
}

type Sender struct {
	ID       int64  `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Username string `json:"username,omitempty"`
	Kind     string `json:"kind,omitempty"`
}

type Attachment struct {
	Name     string `json:"name,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
}

type Message struct {
	Group      Group       `json:"group"`
	MessageID  int         `json:"message_id"`
	Date       time.Time   `json:"date"`
	Text       string      `json:"text"`
	Outgoing   bool        `json:"outgoing"`
	Sender     Sender      `json:"sender"`
	Attachment *Attachment `json:"attachment,omitempty"`
}
