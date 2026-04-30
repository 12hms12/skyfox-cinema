package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCaptchaService_GenerateReturnsCaptchaID(t *testing.T) {
	svc := NewCaptchaService()

	captchaID, masterImage, tileImage, tileX, tileY, tileWidth, tileHeight, err := svc.Generate()

	assert.NoError(t, err)
	assert.NotEmpty(t, captchaID)
	assert.NotEmpty(t, masterImage)
	assert.NotEmpty(t, tileImage)
	assert.Greater(t, tileX, 0)
	assert.GreaterOrEqual(t, tileY, 0)
	assert.Greater(t, tileWidth, 0)
	assert.Greater(t, tileHeight, 0)
}

func TestCaptchaService_VerifySuccess(t *testing.T) {
	svc := NewCaptchaService()

	captchaID, _, _, tileX, _, _, _, err := svc.Generate()
	assert.NoError(t, err)

	ok := svc.Verify(captchaID, float64(tileX))
	assert.True(t, ok)
}

func TestCaptchaService_VerifyWithinTolerance(t *testing.T) {
	svc := NewCaptchaService()

	captchaID, _, _, tileX, _, _, _, err := svc.Generate()
	assert.NoError(t, err)

	ok := svc.Verify(captchaID, float64(tileX)+4)
	assert.True(t, ok)
}

func TestCaptchaService_VerifyFailsOutsideTolerance(t *testing.T) {
	svc := NewCaptchaService()

	captchaID, _, _, tileX, _, _, _, err := svc.Generate()
	assert.NoError(t, err)

	ok := svc.Verify(captchaID, float64(tileX)+50)
	assert.False(t, ok)
}

func TestCaptchaService_VerifyFailsForUnknownID(t *testing.T) {
	svc := NewCaptchaService()

	ok := svc.Verify("non-existent-id", 100)
	assert.False(t, ok)
}

func TestCaptchaService_VerifyIsSingleUse(t *testing.T) {
	svc := NewCaptchaService()

	captchaID, _, _, tileX, _, _, _, err := svc.Generate()
	assert.NoError(t, err)

	svc.Verify(captchaID, float64(tileX))

	ok := svc.Verify(captchaID, float64(tileX))
	assert.False(t, ok)
}

func TestCaptchaService_VerifyFailsForExpiredCaptcha(t *testing.T) {
	svc := NewCaptchaService()
	concrete := svc.(*captchaService)

	captchaID, _, _, tileX, _, _, _, err := svc.Generate()
	assert.NoError(t, err)

	concrete.mu.Lock()
	entry := concrete.store[captchaID]
	entry.expiresAt = time.Now().Add(-1 * time.Minute)
	concrete.store[captchaID] = entry
	concrete.mu.Unlock()

	ok := svc.Verify(captchaID, float64(tileX))
	assert.False(t, ok)
}