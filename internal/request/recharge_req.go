package request

type UpdateRechargeReq struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}
