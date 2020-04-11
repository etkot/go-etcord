package types

type Msg interface {
	//Serialize() ([]byte, error)
	//String() string
}

type Error struct {
	Code    int16  `json:"id"`
	Message string `json:"message"`
}

type GetClientsRequest struct {
	Type      uint8    `json:"type"`
	ClientID  uint16   `json:"clientId,omitempty"`
	Count     uint16   `json:"count,omitempty"`
	ClientIDs []uint16 `json:"clientIds,omitempty"`
}

type GetClientsResponse struct {
	Count   uint16   `json:"count"`
	Clients []Client `json:"clients"`
}

type GetChannelsRequest struct {
}

type GetChannelsResponse struct {
	Count    uint16    `json:"count"`
	Channels []Channel `json:"channels"`
}
