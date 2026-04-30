package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/integration_test/db"
)

const (
	testPFPUserID         = uint(501)
	testPFPImageData      = "aGVsbG8="
	testPFPContentType    = "image/jpeg"
	testPFPContentTypePNG = "image/png"
	testPFPUpdatedData    = "d29ybGQ="
)

func TestProfilePictureRepository_Upsert(t *testing.T) {
	database := db.GetDB()
	assert.NoError(t, database.GormDB().AutoMigrate(&model.Customer{}, &model.ProfilePicture{}))

	ctx := context.Background()
	repo := repository.NewProfilePictureRepository(database)

	tests := []struct {
		name        string
		setup       func()
		upsert      model.ProfilePicture
		wantData    string
		wantContent string
	}{
		{
			name: "should insert a new profile picture when no record exists for the user",
			setup: func() {
				database.GormDB().Exec("DELETE FROM profile_pictures")
				database.GormDB().Exec("DELETE FROM online_customers")
				database.GormDB().Exec(
					"INSERT INTO online_customers (id, first_name, last_name, email, phone_number, country_code, age, gender, username, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
					testPFPUserID, "Test", "User", "pfp_insert@example.com", "+10000000001", "+1", 25, "male", "pfp_insert_user", "hashed",
				)
			},
			upsert:      model.ProfilePicture{UserID: testPFPUserID, ImageData: testPFPImageData, ContentType: testPFPContentType},
			wantData:    testPFPImageData,
			wantContent: testPFPContentType,
		},
		{
			name: "should overwrite the existing profile picture when a record already exists for the user",
			setup: func() {
				database.GormDB().Exec("DELETE FROM profile_pictures")
				database.GormDB().Exec("DELETE FROM online_customers")
				database.GormDB().Exec(
					"INSERT INTO online_customers (id, first_name, last_name, email, phone_number, country_code, age, gender, username, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
					testPFPUserID, "Test", "User", "pfp_update@example.com", "+10000000002", "+1", 25, "male", "pfp_update_user", "hashed",
				)
				database.GormDB().Create(&model.ProfilePicture{UserID: testPFPUserID, ImageData: "oldData", ContentType: testPFPContentType})
			},
			upsert:      model.ProfilePicture{UserID: testPFPUserID, ImageData: testPFPUpdatedData, ContentType: testPFPContentTypePNG},
			wantData:    testPFPUpdatedData,
			wantContent: testPFPContentTypePNG,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := repo.Upsert(ctx, tt.upsert)
			assert.NoError(t, err)

			result, err := repo.GetByUserID(ctx, tt.upsert.UserID)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.wantData, result.ImageData)
			assert.Equal(t, tt.wantContent, result.ContentType)
		})
	}
}

func TestProfilePictureRepository_GetByUserID(t *testing.T) {
	database := db.GetDB()
	assert.NoError(t, database.GormDB().AutoMigrate(&model.Customer{}, &model.ProfilePicture{}))

	ctx := context.Background()
	repo := repository.NewProfilePictureRepository(database)

	tests := []struct {
		name     string
		setup    func()
		userID   uint
		wantNil  bool
		wantData string
	}{
		{
			name: "should return the profile picture when a record exists for the given user ID",
			setup: func() {
				database.GormDB().Exec("DELETE FROM profile_pictures")
				database.GormDB().Exec("DELETE FROM online_customers")
				database.GormDB().Exec(
					"INSERT INTO online_customers (id, first_name, last_name, email, phone_number, country_code, age, gender, username, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
					testPFPUserID, "Test", "User", "pfp_get@example.com", "+10000000003", "+1", 25, "male", "pfp_get_user", "hashed",
				)
				database.GormDB().Create(&model.ProfilePicture{UserID: testPFPUserID, ImageData: testPFPImageData, ContentType: testPFPContentType})
			},
			userID:   testPFPUserID,
			wantData: testPFPImageData,
		},
		{
			name: "should return nil when no profile picture exists for the given user ID",
			setup: func() {
				database.GormDB().Exec("DELETE FROM profile_pictures")
			},
			userID:  99999,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			result, err := repo.GetByUserID(ctx, tt.userID)
			assert.NoError(t, err)
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantData, result.ImageData)
			}
		})
	}
}

func TestProfilePictureRepository_DeleteByUserID(t *testing.T) {
	database := db.GetDB()
	assert.NoError(t, database.GormDB().AutoMigrate(&model.Customer{}, &model.ProfilePicture{}))

	ctx := context.Background()
	repo := repository.NewProfilePictureRepository(database)

	tests := []struct {
		name         string
		setup        func()
		deleteUserID uint
		wantDeleted  bool
	}{
		{
			name: "should delete the profile picture when a record exists for the given user ID",
			setup: func() {
				database.GormDB().Exec("DELETE FROM profile_pictures")
				database.GormDB().Exec("DELETE FROM online_customers")
				database.GormDB().Exec(
					"INSERT INTO online_customers (id, first_name, last_name, email, phone_number, country_code, age, gender, username, password) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
					testPFPUserID, "Test", "User", "pfp_delete@example.com", "+10000000004", "+1", 25, "male", "pfp_delete_user", "hashed",
				)
				database.GormDB().Create(&model.ProfilePicture{UserID: testPFPUserID, ImageData: testPFPImageData, ContentType: testPFPContentType})
			},
			deleteUserID: testPFPUserID,
			wantDeleted:  true,
		},
		{
			name: "should not return an error when deleting a profile picture that does not exist",
			setup: func() {
				database.GormDB().Exec("DELETE FROM profile_pictures")
			},
			deleteUserID: 99998,
			wantDeleted:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			err := repo.DeleteByUserID(ctx, tt.deleteUserID)
			assert.NoError(t, err)

			if tt.wantDeleted {
				result, err := repo.GetByUserID(ctx, tt.deleteUserID)
				assert.NoError(t, err)
				assert.Nil(t, result)
			}
		})
	}
}
