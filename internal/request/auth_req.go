package request

type CreateUserReq struct {
	Username string `json:"username" binding:"required,min=4,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname" binding:"required,min=2,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required,min=4,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}
