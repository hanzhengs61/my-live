package service

import (
	"errors"
	"my-live/internal/db"
	"my-live/internal/model"
	"my-live/internal/request"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoomService struct {
}

func NewRoomService() *RoomService {
	return &RoomService{}
}

var (
	ErrRoomNotFound     = errors.New("房间不存在")
	ErrInvalidPassword  = errors.New("私密房间密码错误")
	ErrRoomAlreadyExist = errors.New("房间已存在（标题重复）") // 可选
)

// CreateRoom 创建房间
func (s *RoomService) CreateRoom(c *gin.Context, hostID uint, req request.CreateRoomReq) (*model.Room, error) {
	db := db.WithContext(c.Request.Context())

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

	if err := db.Create(room).Error; err != nil {
		return nil, err
	}

	return room, nil
}

// GetAllRooms 获取当前所有可见房间列表
func (s *RoomService) GetAllRooms(c *gin.Context) ([]model.Room, error) {
	db := db.WithContext(c.Request.Context())
	var rooms []model.Room
	err := db.Where("status IN ?", []string{"waiting", "live"}).
		Order("created_at desc").
		Find(&rooms).Error
	return rooms, err
}

// VerifyJoinRoom 验证用户是否有权限加入该房间
func (s *RoomService) VerifyJoinRoom(c *gin.Context, roomID uint, inputPassword string) (*model.Room, error) {
	db := db.WithContext(c.Request.Context())

	var room model.Room
	if err := db.First(&room, roomID).Error; err != nil {
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
