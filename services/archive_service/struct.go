package main

import (
	"time"

	bili "github.com/114514ns/BiliClient"
)

type Config struct {
	Port      int
	ClickDSL  string
	Thread    int
	Cookie    string
	HttpProxy string
	ProxyUser string
	ProxyPass string
}

type Dynamic struct {
	Top         bool
	UName       string
	UID         int64
	Face        string
	Images      string
	Type        string
	Title       string
	Text        string
	ID          int64
	BV          string
	Comments    int
	Like        int
	Forward     int
	CommentID   int64
	CommentType int
	CreateAt    time.Time
	ForwardFrom int64
	RawResponse string
	Forwarded   bool

	View     int
	Danmakus int
}

type ReplyList struct {
	OID         int64
	Pb          []byte
	Count       int
	ServerCount int
	Typo        int
	Versions    string
}

type Video struct {
	bili.Video
	RawDanmaku string
}

type Danmaku struct {
	Aid         int64
	Cid         int64
	RawDanmaku  string
	Count       int
	ServerCount int
	Versions    string
}

type UserRule struct {
}
