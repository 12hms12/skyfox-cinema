package service

import (
	"context"

	"net/url"
	"skyfox/bookings/constants"
	"skyfox/bookings/dto/request"
	"skyfox/bookings/dto/response"
	"skyfox/bookings/model"
	ce "skyfox/error"
	"strings"
	"time"
)

type AvatarRepository interface {
	ListAll(ctx context.Context) ([]model.PredefinedAvatar, error)
	ListByGender(ctx context.Context, gender string) ([]model.PredefinedAvatar, error)
	GetByID(ctx context.Context, id int64) (*model.PredefinedAvatar, error)
	GetRandomByGender(ctx context.Context, gender string) (*model.PredefinedAvatar, error)
	ExistsByURL(ctx context.Context, url string) (bool, error)
	BulkInsertIfNotExists(ctx context.Context, avatars []model.PredefinedAvatar) error
	GetUserAvatar(ctx context.Context, userID uint) (avatarURL string, avatarType string, err error)
	UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string, avatarType string) error
}

type ProfilePictureDeleter interface {
	DeleteByUserID(ctx context.Context, userID uint) error
}

type avatarService struct {
	repo    AvatarRepository
	pfpRepo ProfilePictureDeleter
}

func NewAvatarService(repo AvatarRepository, opts ...ProfilePictureDeleter) *avatarService {
	svc := &avatarService{repo: repo}
	if len(opts) > 0 {
		svc.pfpRepo = opts[0]
	}
	return svc
}

var (
	Predefined = "predefined"
	Uploaded   = "uploaded"
)

func (s *avatarService) UpdateAvatar(ctx context.Context, userID uint, req request.UpdateAvatarRequest) (*response.AvatarResponse, error) {
	hasID := req.PredefinedAvatarID > 0
	hasURL := strings.TrimSpace(req.AvatarURL) != ""

	if hasID == hasURL {
		return nil, ce.ErrPickEitherIDOrURL()
	}

	var finalURL string
	var finalType string

	if hasID {
		item, err := s.repo.GetByID(ctx, req.PredefinedAvatarID)
		if err != nil {
			return nil, err
		}
		if item == nil {
			return nil, ce.ErrAvatarNotFound()
		}
		finalURL = item.URL
		finalType = Predefined
	} else {
		if !isValidURL(req.AvatarURL) {
			return nil, ce.ErrInvalidAvatarURL()
		}
		typ := strings.ToLower(strings.TrimSpace(req.AvatarType))
		if typ != Uploaded && typ != Predefined {
			return nil, ce.ErrInvalidAvatarType()
		}
		finalURL = req.AvatarURL
		finalType = typ
	}

	if finalType == Predefined && s.pfpRepo != nil {
		_ = s.pfpRepo.DeleteByUserID(ctx, userID)
	}

	if err := s.repo.UpdateUserAvatar(ctx, userID, finalURL, finalType); err != nil {
		return nil, err
	}

	return &response.AvatarResponse{
		UserID:     userID,
		AvatarURL:  finalURL,
		AvatarType: finalType,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *avatarService) GetUserAvatar(ctx context.Context, userID uint) (*response.AvatarResponse, error) {
	avatarURL, avatarType, err := s.repo.GetUserAvatar(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &response.AvatarResponse{
		UserID:     userID,
		AvatarURL:  avatarURL,
		AvatarType: avatarType,
	}, nil
}

func (s *avatarService) AssignRandomAvatar(ctx context.Context, userID uint, gender string) (*response.AvatarResponse, error) {
	g := strings.ToLower(strings.TrimSpace(gender))

	avatar, err := s.repo.GetRandomByGender(ctx, g)
	if err != nil {
		return nil, err
	}
	if avatar == nil {
		avatar, err = s.repo.GetRandomByGender(ctx, "")
		if err != nil {
			return nil, err
		}
	}
	if avatar == nil {
		return nil, ce.ErrAvatarNotFound()
	}

	if err := s.repo.UpdateUserAvatar(ctx, userID, avatar.URL, Predefined); err != nil {
		return nil, err
	}

	return &response.AvatarResponse{
		UserID:     userID,
		AvatarURL:  avatar.URL,
		AvatarType: Predefined,
		UpdatedAt:  time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (s *avatarService) ListPredefinedAvatars(ctx context.Context, gender string) (*response.PredefinedAvatarListResponse, error) {
	var items []model.PredefinedAvatar
	var err error

	g := strings.TrimSpace(strings.ToLower(gender))
	switch g {
	case constants.EMPTY, constants.MALE, constants.FEMALE, constants.NEUTRAL:
		// ok
	default:
		// if unknown gender is passed, just fallback to all
		g = ""
	}

	if g == "" {
		items, err = s.repo.ListAll(ctx)
	} else {
		items, err = s.repo.ListByGender(ctx, g)
	}
	if err != nil {
		return nil, err
	}

	out := make([]response.PredefinedAvatarResponse, 0, len(items))
	for _, it := range items {
		out = append(out, response.PredefinedAvatarResponse{
			ID:     it.ID,
			Gender: it.Gender,
			URL:    it.URL,
		})
	}
	return &response.PredefinedAvatarListResponse{
		Items: out,
		Total: len(out),
	}, nil
}

func isValidURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}
