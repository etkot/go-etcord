package types

type Client struct {
	UserID uint16 `json:"userId"`
	Name   string `json:"name"`
}

type Channel struct {
	ChannelID uint16 `json:"channelId"`
	ParentID  uint   `json:"parentId"`
	Name      string `json:"name"`
	Type      uint8  `json:"type"`
}
