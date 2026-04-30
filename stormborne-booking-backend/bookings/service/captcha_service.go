package service

import (
	"fmt"
	"math"
	"sync"
	"time"
	"log"
	
	"github.com/google/uuid"
	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"
)

const (
	captchaExpiry   = 5 * time.Minute
	verifyTolerance = 10.0
)

type captchaEntry struct {
	targetX   float64
	expiresAt time.Time
}

type CaptchaServiceInterface interface {
	Generate() (captchaID, masterImage, tileImage string, tileX, tileY, tileWidth, tileHeight int, err error)
	Verify(captchaID string, sliderPositionX float64) bool
}

type captchaService struct {
	captcha slide.Captcha
	store   map[string]captchaEntry
	mu      sync.Mutex
}

func NewCaptchaService() CaptchaServiceInterface {
	builder := slide.NewBuilder()
	builder.SetOptions(slide.WithEnableGraphVerticalRandom(true))

	imgs, err := imagesv2.GetImages()
	if err != nil {
		log.Fatalf("captcha: failed to load background images: %v", err)
	}

	graphs, err := tiles.GetTiles()
	if err != nil {
		log.Fatalf("captcha: failed to load tile images: %v", err)
	}

	newGraphs := make([]*slide.GraphImage, 0, len(graphs))
	for _, g := range graphs {
		newGraphs = append(newGraphs, &slide.GraphImage{
			OverlayImage: g.OverlayImage,
			MaskImage:    g.MaskImage,
			ShadowImage:  g.ShadowImage,
		})
	}

	builder.SetResources(
		slide.WithGraphImages(newGraphs),
		slide.WithBackgrounds(imgs),
	)

	return &captchaService{
		captcha: builder.Make(),
		store:   make(map[string]captchaEntry),
	}
}

func (s *captchaService) Generate() (captchaID, masterImage, tileImage string, tileX, tileY, tileWidth, tileHeight int, err error) {
	captData, genErr := s.captcha.Generate()
	if genErr != nil {
		err = fmt.Errorf("captcha: generate failed: %w", genErr)
		return
	}

	block := captData.GetData()
	if block == nil {
		err = fmt.Errorf("captcha: nil block data")
		return
	}

	masterImage, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		err = fmt.Errorf("captcha: master image encode failed: %w", err)
		return
	}

	tileImage, err = captData.GetTileImage().ToBase64()
	if err != nil {
		err = fmt.Errorf("captcha: tile image encode failed: %w", err)
		return
	}

	captchaID = uuid.New().String()
	tileX = block.X
	tileY = block.Y
	tileWidth = block.Width
	tileHeight = block.Height

	s.mu.Lock()
	s.store[captchaID] = captchaEntry{
		targetX:   float64(block.X),
		expiresAt: time.Now().Add(captchaExpiry),
	}
	s.mu.Unlock()

	go s.purgeExpired()

	return
}

func (s *captchaService) Verify(captchaID string, sliderPositionX float64) bool {
	s.mu.Lock()
	entry, exists := s.store[captchaID]
	if exists {
		delete(s.store, captchaID) 
	}
	s.mu.Unlock()

	if !exists {
		return false
	}

	if time.Now().After(entry.expiresAt) {
		return false
	}

	return math.Abs(sliderPositionX-entry.targetX) <= verifyTolerance
}

func (s *captchaService) purgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, entry := range s.store {
		if now.After(entry.expiresAt) {
			delete(s.store, id)
		}
	}
}