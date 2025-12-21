package main

import "time"

type Dash struct {
	Data struct {
		Dash0 struct {
			Video []struct {
				Link string `json:"base_url"`
			} `json:"video"`
			Audio []struct {
				Link   string   `json:"base_url"`
				Backup []string `json:"backupUrl"`
			} `json:"audio"`
		} `json:"dash"`
	} `json:"data"`
}
type CollectionList struct {
	Data struct {
		List []struct {
			Title string `json:"title"`
			ID    int    `json:"id"`
		}
	}
}
type CollectionMedias struct {
	Data struct {
		Medias []struct {
			Title string `json:"title"`
			BV    string `json:"bvid"`
		}
	}
}
type VideoResponse struct {
	Data struct {
		Cover     string `json:"pic"`
		Title     string `json:"title"`
		Duration  int    `json:"duration"`
		PublishAt int64  `json:"pubdate"`
		Desc      string `json:"desc"`
		Owner     struct {
			Mid  int64  `json:"mid"`
			Name string `json:"name"`
			Face string `json:"face"`
		} `json:"owner"`
		Pages []struct {
			Cid      int    `json:"cid"`
			Title    string `json:"part"`
			Duration int    `json:"duration"`
		}
	} `json:"data"`
}
type PlayListResponse struct {
	Data struct {
		Archives []struct {
			BV       string `json:"bvid"`
			CreateAt int    `json:"pubdate"`
			Cover    string `json:"pic"`
			Duration int    `json:"duration"`
			Title    string `json:"title"`
		} `json:"archives"`
		Meta struct {
			Name string `json:"name"`
		} `json:"meta"`
	} `json:"data"`
}
type UserDynamic struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Items []DynamicItem `json:"items"`
	} `json:"data"`
}
type DynamicItem struct {
	IDStr   string       `json:"id_str"`
	Orig    *DynamicItem `json:"orig"`
	Modules struct {
		ModuleDynamic struct {
			Major struct {
				Archive struct {
					Aid   string `json:"aid"`
					Badge struct {
						BgColor string      `json:"bg_color"`
						Color   string      `json:"color"`
						IconURL interface{} `json:"icon_url"`
						Text    string      `json:"text"`
					} `json:"badge"`
					Bvid  string `json:"bvid"`
					Cover string `json:"cover"`
					Desc  string `json:"desc"`
					Stat  struct {
						Danmaku string `json:"danmaku"`
						Play    string `json:"play"`
					} `json:"stat"`
					Title string `json:"title"`
					Type  int    `json:"type"`
				} `json:"archive"`
				Opus struct {
					Pics []struct {
						URL string `json:"url"`
					} `json:"pics"`
					Summary struct {
						Text string `json:"text"`
					} `json:"summary"`
				} `json:"opus"`
				Desc struct {
					Text string `json:"text"`
				} `json:"desc"`
				Type string `json:"type"`
			} `json:"major"`
			Topic interface{} `json:"topic"`
			Desc  struct {
				Nodes []struct {
					Text string `json:"text"`
				} `json:"rich_text_nodes"`
			} `json:"desc"`
		} `json:"module_dynamic"`
		ModuleAuthor struct {
			Name      string `json:"name"`
			Mid       int64  `json:"mid"`
			TimeStamp int64  `json:"pub_ts"`
		} `json:"module_author"`
	} `json:"modules"`
	Type string `json:"type"`
}
type FansList struct {
	Data struct {
		List []struct {
			Mid                string `json:"mid"`
			Attribute          int    `json:"attribute"`
			Uname              string `json:"uname"`
			Face               string `json:"face"`
			AttestationDisplay struct {
				Type int    `json:"type"`
				Desc string `json:"desc"`
			} `json:"attestation_display"`
		} `json:"list"`
	} `json:"data"`
	Ts        int64  `json:"ts"`
	RequestID string `json:"request_id"`
}
type AreaLiverListResponse struct {
	Data struct {
		More int8 `json:"has_more"`
		List []struct {
			Cover      string `json:"cover"`
			Room       int    `json:"roomid"`
			ParentArea string `json:"parent_name"`
			Area       string `json:"area_name"`
			Title      string `json:"title"`
			UName      string `json:"uname"`
			UID        int64  `json:"uid"`
			Watch      struct {
				Num int `json:"num"`
			} `json:"watched_show"`
		} `json:"list"`
	} `json:"data"`
}
type UserResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Data    struct {
		Card struct {
			Name   string `json:"name"`
			Face   string `json:"face"`
			Bio    string `json:"sign"`
			Verify struct {
				Content string `json:"desc"`
			} `json:"official_verify"`
		}
		Followers int `json:"follower"`
	} `json:"data"`
}
type LiveStreamResponse struct {
	Data struct {
		Time        int64 `json:"live_time"`
		PlayurlInfo struct {
			Playurl struct {
				Stream []struct {
					ProtocolName string `json:"protocol_name"`
					Format       []struct {
						FormatName string `json:"format_name"`
						Codec      []struct {
							CodecName string `json:"codec_name"`
							CurrentQn int    `json:"current_qn"`
							AcceptQn  []int  `json:"accept_qn"`
							BaseUrl   string `json:"base_url"`
							UrlInfo   []struct {
								Host      string `json:"host"`
								Extra     string `json:"extra"`
								StreamTtl int    `json:"stream_ttl"`
							} `json:"url_info"`
							HdrQn     interface{} `json:"hdr_qn"`
							DolbyType int         `json:"dolby_type"`
							AttrName  string      `json:"attr_name"`
							HdrType   int         `json:"hdr_type"`
						} `json:"codec"`
						MasterUrl string `json:"master_url"`
					} `json:"format"`
				} `json:"stream"`
			} `json:"playurl"`
		} `json:"playurl_info"`
	} `json:"data"`
}
type OnlineWatcherResponse struct {
	Data struct {
		Item []struct {
			UID   int64  `json:"uid"`
			Name  string `json:"name"`
			Face  string `json:"face"`
			Guard int8   `json:"guard_level"`
			Days  int16  `json:"days"`
			UInfo struct {
				Medal struct {
					Color string `json:"v2_medal_color_start"`
					Name  string `json:"name"`
					Level int8   `json:"level"`
				} `json:"medal"`
			} `json:"uinfo"`
		} `json:"item"`
		Count int `json:"count"`
	} `json:"data"`
}
type Watcher struct {
	UID   int64  `json:"uid"`
	Name  string `json:"name"`
	Face  string `json:"face"`
	Guard int8   `json:"guard_level"`
	Days  int16  `json:"days"`
	Score int
	Medal struct {
		Name          string `json:"medal_name"`
		Level         int8   `json:"level"`
		ColorDec      int    `json:"medal_color_start"`
		ColorInternal string `json:"v2_medal_color_start"`
		GuardLevel    int8
		Color         string
	} `json:"medal_info"`
}
type LiveStatusResponse struct {
	Cmd     string `json:"cmd"`
	Message string `json:"message"`
	Data    struct {
		LiveStatus int   `json:"live_status"`
		LiveTime   int64 `json:"live_time"`
	} `json:"data"`
}
type GuardListResponse struct {
	Data struct {
		List []GuardResponseItem `json:"list"`
		Top  []GuardResponseItem `json:"top3"`
		Info struct {
			Total int `json:"num"`
			Page  int `json:"page"`
		} `json:"info"`
	} `json:"data"`
}
type GuardResponseItem struct {
	Days int16 `json:"accompany"`
	Info struct {
		UID  int64 `json:"uid"`
		User struct {
			Name string `json:"name"`
			Face string `json:"face"`
		} `json:"base"`
		Medal struct {
			Name       string `json:"name"`
			Level      int8   `json:"level"`
			ColorDec   int    `json:"color_start"`
			GuardLevel int8   `json:"guard_level"`
			Color      string `json:"v2_medal_color_start"`
		} `json:"medal"`
	} `json:"uinfo"`
}
type LiveListResponse struct {
	Data struct {
		Result []struct {
			UName  string `json:"uname"`
			UID    int64  `json:"uid"`
			Living bool   `json:"is_live"`
			Face   string `json:"uface"`
			Room   int    `json:"roomid"`
			Area   string `json:"cate_name"`
		} `json:"result"`
	} `json:"data"`
}
type FansClubResponse struct {
	Message string `json:"message"`
	Data    struct {
		Item []struct {
			UID   int64  `json:"uid"`
			UName string `json:"name"`
			Score int    `json:"score"`
			Level int8   `json:"level"`
			Medal struct {
				Type int8   `json:"guard_level"`
				Name string `json:"name"`
			} `json:"uinfo_medal"`
		} `json:"item"`
		Num int `json:"num"`
	} `json:"data"`
}
type UserMapping struct {
	Hash  string
	UID   int64
	UName string
}

type Dynamic struct {
	Top         bool
	UName       string
	UID         int64
	Face        string
	Type        string
	Title       string
	Text        string
	ID          int64 `gorm:"primaryKey"`
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
	Images      string
}

// 留言
type Comment struct {
	Text        string
	Session     string
	CreatedAt   time.Time
	DisplayName string
	Self        bool
	ID          uint `gorm:"primaryKey"`
}

type PlaybackRepository struct {
	Type   string
	ListID int64
	UID    int64
}
