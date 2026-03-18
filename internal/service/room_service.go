package service

import (
	"context"
	"errors"
	"my-live/internal/db"
	"my-live/internal/model"
	"my-live/internal/request"

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
func (s *RoomService) CreateRoom(ctx context.Context, hostID uint, req request.CreateRoomReq) (*model.Room, error) {

	if req.Title == "" {
		return nil, errors.New("房间标题不能为空")
	}

	if req.IsPrivate && req.Password == "" {
		return nil, errors.New("私密房间必须设置密码")
	}

	room := &model.Room{
		Title:     req.Title,
		HostID:    int64(hostID),
		IsPrivate: req.IsPrivate,
		Password:  req.Password,
	}

	if err := db.DB.WithContext(ctx).Create(room).Error; err != nil {
		return nil, err
	}

	return room, nil
}

// GetAllRooms 获取当前所有可见房间列表
func (s *RoomService) GetAllRooms(ctx context.Context) ([]model.Room, error) {
	var rooms []model.Room
	err := db.DB.WithContext(ctx).Where("status IN ?", []string{"waiting", "live"}).
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
func (s *RoomService) GetRoom(ctx context.Context, roomID uint) (*model.Room, error) {
	var room model.Room
	if err := db.DB.WithContext(ctx).First(&room, roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}
	return &room, nil
}

// UpdateRoom 主播更新房间信息
func (s *RoomService) UpdateRoom(ctx context.Context, roomID int64, hostID int64, title, announcement string, isPrivate bool, password string) error {
	var room model.Room
	if err := s.db.First(&room, roomID).Error; err != nil {
		return errors.New("房间不存在")
	}
	if room.HostID != hostID {
		return errors.New("只有主播可以操作")
	}

	updates := map[string]interface{}{
		"title":        title,
		"announcement": announcement,
		"is_private":   isPrivate,
	}
	if isPrivate && password != "" {
		updates["password"] = password
	} else {
		updates["password"] = ""
	}

	return db.DB.WithContext(ctx).Model(&room).Updates(updates).Error
}

// BanUser 禁言用户
func (s *RoomService) BanUser(ctx context.Context, roomID int64, hostID int64, targetUserID int64) error {
	var room model.Room
	if err := s.db.First(&room, roomID).Error; err != nil || room.HostID != hostID {
		return errors.New("只有主播可以禁言")
	}
	return nil // 实际禁言逻辑可后续加 Redis 黑名单
}

// KickUser 踢出用户
func (s *RoomService) KickUser(ctx context.Context, roomID int64, hostID int64, targetUserID int64) error {
	var room model.Room
	if err := s.db.First(&room, roomID).Error; err != nil || room.HostID != hostID {
		return errors.New("只有主播可以踢人")
	}
	return nil // 实际踢人在 WS 层处理
}
