package request

type SendGiftReq struct {
	RoomID int64 `json:"room_id" binding:"required"`
	GiftID int64 `json:"gift_id" binding:"required"`
	Count  int64 `json:"count" binding:"required,min=1"`
}
