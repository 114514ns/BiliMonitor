package main

import (
	"time"

	bili "github.com/114514ns/BiliClient"
)

type Config struct {
	Port                int
	ClickDSL            string
	Thread              int
	Cookie              string
	HttpProxy           string
	ProxyUser           string
	ProxyPass           string
	StreamAgentEndPoint string
	StreamCookie        string
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
	CreatedAt   time.Time
}

type Reply struct {
	bili.Reply
	Alive bool
	Seens []time.Time
}

type Video struct {
	bili.Video
	UpdatedAt  time.Time
	RawDanmaku string
}

type Danmaku struct {
	Aid         int64
	Cid         int64
	RawDanmaku  string
	Count       int
	ServerCount int
	Versions    string
	CreatedAt   time.Time
}

type Collection struct {
	UID       int64
	Name      string
	ID        int
	Items     string
	Desc      string
	CreatedAt time.Time
}

type UserRule struct {
	RefreshFeedsDelay int //每次刷新动态的间隔，单位为分钟
	ActiveCount       int //每次更新动态的时候，同时更新前面多少条动态的全量信息
	FullDelay         int //每次全量刷新的间隔
}
