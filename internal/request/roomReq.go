package request

type CreateRoomReq struct {
	Title       string `json:"title" binding:"required,max=100"` // 房间标题，必填，最长100字符
	Description string `json:"description"`                      // 描述，可选
	IsPrivate   bool   `json:"is_private"`                       // 是否私密房间
	Password    string `json:"password"`                         // 如果 is_private=true，则必须提供密码
}
