package rooms

type CreateRoomRequest struct {
	Title string `json:"title"`
}

type CreateRoomResponse struct {
	RoomID  string `json:"roomId"`
	JoinURL string `json:"joinUrl,omitempty"`
}

type IssueTokenRequest struct {
	Name string `json:"name"`
}

type IssueTokenResponse struct {
	Token      string `json:"token"`
	LiveKitURL string `json:"livekitUrl,omitempty"`
	Message    string `json:"message,omitempty"`
}
