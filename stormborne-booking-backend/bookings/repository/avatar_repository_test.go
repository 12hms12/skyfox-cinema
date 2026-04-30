package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"skyfox/bookings/model"
	"skyfox/bookings/repository"
	"skyfox/integration_test/db"
)

func TestAvatarRepository_ListAll(t *testing.T) {

	database := db.GetDB()
	err := database.GormDB().AutoMigrate(&model.PredefinedAvatar{})
	assert.NoError(t, err)

	ctx := context.Background()
	repo := repository.NewAvatarRepository(database)

	tests := []struct {
		name      string
		seed      []model.PredefinedAvatar
		wantCount int
	}{
		{
			name: "should return all predefined avatars ordered by ID when multiple avatars exist in database",
			seed: []model.PredefinedAvatar{
				{ID: 1, Gender: "male", URL: "url-m1"},
				{ID: 2, Gender: "female", URL: "url-f1"},
				{ID: 3, Gender: "male", URL: "url-m2"},
			},
			wantCount: 3,
		},
		{
			name:      "should return empty list when no predefined avatars exist in database",
			seed:      []model.PredefinedAvatar{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			database.GormDB().Exec("DELETE FROM predefined_avatars")

			if len(tt.seed) > 0 {
				database.GormDB().Create(&tt.seed)
			}

			avatars, err := repo.ListAll(ctx)

			assert.NoError(t, err)
			assert.Len(t, avatars, tt.wantCount)
		})
	}
}

func TestAvatarRepository_ListByGender(t *testing.T) {

	database := db.GetDB()
	err := database.GormDB().AutoMigrate(&model.PredefinedAvatar{})
	assert.NoError(t, err)

	ctx := context.Background()
	repo := repository.NewAvatarRepository(database)

	tests := []struct {
		name      string
		seed      []model.PredefinedAvatar
		gender    string
		wantCount int
	}{
		{
			name: "should return only male avatars when gender filter is male",
			seed: []model.PredefinedAvatar{
				{Gender: "male", URL: "url-m1"},
				{Gender: "female", URL: "url-f1"},
				{Gender: "male", URL: "url-m2"},
			},
			gender:    "male",
			wantCount: 2,
		},
		{
			name: "should return only female avatars when gender filter is female",
			seed: []model.PredefinedAvatar{
				{Gender: "male", URL: "url-m1"},
				{Gender: "female", URL: "url-f1"},
			},
			gender:    "female",
			wantCount: 1,
		},
		{
			name: "should return empty list when no avatars match the specified gender",
			seed: []model.PredefinedAvatar{
				{Gender: "male", URL: "url-m1"},
			},
			gender:    "female",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			database.GormDB().Exec("DELETE FROM predefined_avatars")
			database.GormDB().Create(&tt.seed)

			avatars, err := repo.ListByGender(ctx, tt.gender)

			assert.NoError(t, err)
			assert.Len(t, avatars, tt.wantCount)
		})
	}
}

func TestAvatarRepository_GetByID(t *testing.T) {

	database := db.GetDB()
	err := database.GormDB().AutoMigrate(&model.PredefinedAvatar{})
	assert.NoError(t, err)

	ctx := context.Background()
	repo := repository.NewAvatarRepository(database)

	tests := []struct {
		name    string
		seed    []model.PredefinedAvatar
		id      int64
		wantNil bool
	}{
		{
			name: "should return predefined avatar when the given ID exists in database",
			seed: []model.PredefinedAvatar{
				{ID: 10, Gender: "female", URL: "url-f1"},
			},
			id:      10,
			wantNil: false,
		},
		{
			name:    "should return nil when the given avatar ID does not exist in database",
			seed:    []model.PredefinedAvatar{},
			id:      999,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			database.GormDB().Exec("DELETE FROM predefined_avatars")
			database.GormDB().Create(&tt.seed)

			avatar, err := repo.GetByID(ctx, tt.id)

			assert.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, avatar)
			} else {
				assert.NotNil(t, avatar)
			}
		})
	}
}

func TestAvatarRepository_ExistsByURL(t *testing.T) {

	database := db.GetDB()
	err := database.GormDB().AutoMigrate(&model.PredefinedAvatar{})
	assert.NoError(t, err)

	ctx := context.Background()
	repo := repository.NewAvatarRepository(database)

	tests := []struct {
		name string
		seed []model.PredefinedAvatar
		url  string
		want bool
	}{
		{
			name: "should return true when the avatar URL already exists in database",
			seed: []model.PredefinedAvatar{
				{Gender: "male", URL: "url-m1"},
			},
			url:  "url-m1",
			want: true,
		},
		{
			name: "should return false when the avatar URL does not exist in database",
			seed: []model.PredefinedAvatar{
				{Gender: "male", URL: "url-m1"},
			},
			url:  "unknown",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			database.GormDB().Exec("DELETE FROM predefined_avatars")
			database.GormDB().Create(&tt.seed)

			exists, err := repo.ExistsByURL(ctx, tt.url)

			assert.NoError(t, err)
			assert.Equal(t, tt.want, exists)
		})
	}
}

func TestAvatarRepository_BulkInsertIfNotExists(t *testing.T) {

	database := db.GetDB()
	err := database.GormDB().AutoMigrate(&model.PredefinedAvatar{})
	assert.NoError(t, err)

	ctx := context.Background()
	repo := repository.NewAvatarRepository(database)

	tests := []struct {
		name      string
		initial   []model.PredefinedAvatar
		insert    []model.PredefinedAvatar
		wantCount int64
	}{
		{
			name:    "should insert all avatars when none of the URLs already exist in database",
			initial: []model.PredefinedAvatar{},
			insert: []model.PredefinedAvatar{
				{Gender: "male", URL: "bulk-m1"},
				{Gender: "female", URL: "bulk-f1"},
			},
			wantCount: 2,
		},
		{
			name: "should ignore duplicate avatar URLs when inserting batch with existing URLs",
			initial: []model.PredefinedAvatar{
				{Gender: "male", URL: "dup-url"},
			},
			insert: []model.PredefinedAvatar{
				{Gender: "male", URL: "dup-url"},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			database.GormDB().Exec("DELETE FROM predefined_avatars")

			database.GormDB().Create(&tt.initial)

			err := repo.BulkInsertIfNotExists(ctx, tt.insert)

			assert.NoError(t, err)

			var count int64
			database.GormDB().Model(&model.PredefinedAvatar{}).Count(&count)

			assert.Equal(t, tt.wantCount, count)
		})
	}
}

func TestAvatarRepository_GetRandomByGender(t *testing.T) {

	database := db.GetDB()
	err := database.GormDB().AutoMigrate(&model.PredefinedAvatar{})
	assert.NoError(t, err)

	ctx := context.Background()
	repo := repository.NewAvatarRepository(database)

	tests := []struct {
		name   string
		seed   []model.PredefinedAvatar
		gender string
	}{
		{
			name: "should return a random female avatar when female avatars exist",
			seed: []model.PredefinedAvatar{
				{Gender: "female", URL: "url-f1"},
			},
			gender: "female",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			database.GormDB().Exec("DELETE FROM predefined_avatars")
			database.GormDB().Create(&tt.seed)

			avatar, err := repo.GetRandomByGender(ctx, tt.gender)

			assert.NoError(t, err)
			assert.NotNil(t, avatar)
			assert.Equal(t, tt.gender, avatar.Gender)
		})
	}
}

func TestAvatarRepository_UserAvatarOperations(t *testing.T) {

	database := db.GetDB()
	err := database.GormDB().AutoMigrate(&model.Customer{})
	assert.NoError(t, err)

	ctx := context.Background()
	repo := repository.NewAvatarRepository(database)

	t.Run("should return avatar url and type when user has an avatar configured", func(t *testing.T) {

		database.GormDB().Exec("DELETE FROM customer")

		user := model.Customer{
			ID:         1,
			Email:      "user@example.com",
			AvatarURL:  "https://img.com/avatar.jpg",
			AvatarType: model.PREDEFINED,
		}

		database.GormDB().Create(&user)

		url, typ, err := repo.GetUserAvatar(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, "https://img.com/avatar.jpg", url)
		assert.Equal(t, string(model.PREDEFINED), typ)
	})

	t.Run("should update the avatar url and type for an existing user", func(t *testing.T) {

		database.GormDB().Exec("DELETE FROM customer")

		user := model.Customer{
			ID:    1,
			Email: "user@example.com",
		}

		database.GormDB().Create(&user)

		err := repo.UpdateUserAvatar(ctx, 1, "https://new.com/avatar.png", "uploaded")

		assert.NoError(t, err)

		var updated model.Customer
		database.GormDB().First(&updated, 1)

		assert.Equal(t, "https://new.com/avatar.png", updated.AvatarURL)
		assert.Equal(t, "uploaded", string(updated.AvatarType))
	})
}
