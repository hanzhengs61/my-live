package service

import (
	"errors"
	"my-live/internal/db"
	"my-live/internal/model"
	"my-live/internal/request"
	"my-live/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoomService struct {
	db *gorm.DB
}

func NewRoomService() *RoomService {
	return &RoomService{
		db: db.DB,
	}
}

var (
	ErrRoomNotFound     = errors.New("房间不存在")
	ErrInvalidPassword  = errors.New("私密房间密码错误")
	ErrRoomAlreadyExist = errors.New("房间已存在（标题重复）") // 可选
)

// CreateRoom 创建房间
func (s *RoomService) CreateRoom(ctx *gin.Context, hostID uint, req request.CreateRoomReq) (*model.Room, error) {

	if req.Title == "" {
		return nil, errors.New("房间标题不能为空")
	}

	if req.IsPrivate && req.Password == "" {
		return nil, errors.New("私密房间必须设置密码")
	}

	room := &model.Room{
		Title:       req.Title,
		Description: req.Description,
		HostID:      hostID,
		IsPrivate:   req.IsPrivate,
		Password:    req.Password,
		Status:      "waiting",
	}

	if err := db.DB.WithContext(ctx.Request.Context()).Create(room).Error; err != nil {
		return nil, err
	}

	return room, nil
}

// GetAllRooms 获取当前所有可见房间列表
func (s *RoomService) GetAllRooms(ctx *gin.Context) ([]model.Room, error) {
	var rooms []model.Room
	err := db.DB.WithContext(ctx.Request.Context()).Where("status IN ?", []string{"waiting", "live"}).
		Order("created_at desc").
		Find(&rooms).Error
	return rooms, err
}

// VerifyJoinRoom 验证用户是否有权限加入该房间
func (s *RoomService) VerifyJoinRoom(roomID uint, inputPassword string) (*model.Room, error) {

	var room model.Room
	if err := db.DB.First(&room, roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	// 如果是私密房间，校验密码
	if room.IsPrivate && room.Password != inputPassword {
		return nil, ErrInvalidPassword
	}

	return &room, nil
}

// GetRoom 获取房间信息
func (s *RoomService) GetRoom(ctx *gin.Context, roomID uint) (*model.Room, error) {
	var room model.Room
	if err := db.DB.WithContext(ctx.Request.Context()).First(&room, roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(ctx, 404, "房间不存在")
		}
		response.Error(ctx, 500, "查询失败")
	}
	return &room, nil
}
