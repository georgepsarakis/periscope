package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"gorm.io/gorm"

	"github.com/georgepsarakis/periscope/repository/rdbms"
)

func cacheKeyProject(id string) string {
	return fmt.Sprintf("project:%s", id)
}

const CharsetAlphanumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
const CharsetLettersLowercase = "abcdefghijklmnopqrstuvwxyz"

func RandomString(charset string, length uint) string {
	b := make([]byte, length)
	s := len(charset) - 1
	for i := range b {
		b[i] = charset[rand.Intn(s)]
	}
	return string(b)
}

func (r *Repository) ProjectCreate(ctx context.Context, name string) (Project, error) {
	project := Project{
		Name:     name,
		PublicID: RandomString(CharsetLettersLowercase, 8),
	}
	key := ProjectIngestionAPIKey{
		Key: RandomString(CharsetAlphanumeric, 36),
	}
	err := r.database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if res := r.database.WithContext(ctx).Create(&project); res.Error != nil {
			return res.Error
		}
		key.ProjectID = project.ID
		if res := r.database.WithContext(ctx).Create(&key); res.Error != nil {
			return res.Error
		}
		project.ProjectIngestionAPIKeys = []ProjectIngestionAPIKey{key}
		return nil
	})
	if err != nil {
		return Project{}, err
	}
	return project, nil
}

func (r *Repository) ProjectFindByID(ctx context.Context, id uint) (Project, error) {
	project := Project{}
	tx := r.database.WithContext(ctx).Preload("ProjectIngestionAPIKeys").First(&project, id)
	if tx.Error != nil {
		return Project{}, tx.Error
	}
	return project, nil
}

func (r *Repository) ProjectFindByPublicID(ctx context.Context, publicID string) (Project, error) {
	project := Project{}
	projectCacheKey := []byte(cacheKeyProject(publicID))
	v, err := r.cache.Get(projectCacheKey)
	if err != nil || len(v) == 0 {
		tx := r.database.WithContext(ctx).Preload("ProjectIngestionAPIKeys").First(&project, r.database.Where("public_id = ?", publicID))
		if tx.Error != nil {
			return Project{}, tx.Error
		}
		s, err := json.Marshal(project)
		if err != nil {
			return Project{}, err
		} else if err := r.cache.Set(projectCacheKey, s, 3600); err != nil {
			return Project{}, err
		}
	} else {
		var p Project
		if err := json.Unmarshal(v, &p); err != nil {
			return Project{}, err
		}
		project = p
	}
	return project, nil
}

type ProjectAlertDestinationCreateInput rdbms.ProjectAlertDestination

func (r *Repository) ProjectAlertDestinationCreate(ctx context.Context, ad ProjectAlertDestinationCreateInput) error {
	tx := r.dbExecutor(ctx)
	a := rdbms.ProjectAlertDestination{
		AlertDestinationTypeID: ad.AlertDestinationTypeID,
		ProjectID:              ad.ProjectID,
		Configuration:          ad.Configuration,
	}
	if r := tx.Create(&a); r.Error != nil {
		return r.Error
	}
	return nil
}

func (r *Repository) AlertDestinationTypeFindAll(ctx context.Context) ([]AlertDestinationType, error) {
	tx := r.dbExecutor(ctx)

	var destinationTypeList []rdbms.AlertDestinationType
	res := tx.Model(&rdbms.AlertDestinationType{}).Find(&destinationTypeList)
	if res.Error != nil {
		return nil, res.Error
	}
	adt := make([]AlertDestinationType, 0, len(destinationTypeList))
	for _, d := range destinationTypeList {
		adt = append(adt, AlertDestinationType{
			BaseModel: BaseModel{
				ID:        d.ID,
				CreatedAt: d.CreatedAt,
				UpdatedAt: d.UpdatedAt,
			},
			Title: d.Title,
			Key:   d.Key,
		})
	}
	return adt, nil
}
