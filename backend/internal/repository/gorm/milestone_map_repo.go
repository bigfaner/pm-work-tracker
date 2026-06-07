package gorm

import (
	"context"
	"strconv"
	"time"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/repository"
)

type milestoneMapRepo struct {
	db *gormlib.DB
}

// NewGormMilestoneMapRepo creates a GORM-backed MilestoneMapRepo.
func NewGormMilestoneMapRepo(db *gormlib.DB) repository.MilestoneMapRepo {
	return &milestoneMapRepo{db: db}
}

func (r *milestoneMapRepo) Create(ctx context.Context, m *model.MilestoneMap) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *milestoneMapRepo) FindByID(ctx context.Context, id uint) (*model.MilestoneMap, error) {
	return repo.FindByID[model.MilestoneMap](r.db, ctx, id)
}

func (r *milestoneMapRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.MilestoneMap, error) {
	var m model.MilestoneMap
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key = ?", bizKey).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *milestoneMapRepo) Update(ctx context.Context, m *model.MilestoneMap, fields map[string]interface{}) error {
	return repo.UpdateFields[model.MilestoneMap](r.db, ctx, m, m.TeamKey, fields)
}

func (r *milestoneMapRepo) List(ctx context.Context, teamBizKey int64, filter dto.MilestoneMapFilter, page dto.Pagination) (*dto.PageResult[model.MilestoneMap], error) {
	query := r.db.WithContext(ctx).Scopes(NotDeleted).Where("team_key = ?", teamBizKey)

	if filter.Name != nil && *filter.Name != "" {
		query = query.Where("map_name LIKE ?", "%"+*filter.Name+"%")
	}
	if filter.AssigneeKey != nil && *filter.AssigneeKey != "" {
		ak, err := strconv.ParseInt(*filter.AssigneeKey, 10, 64)
		if err != nil {
			query = query.Where("1 = 0")
		} else {
			query = query.Where("assignee_key = ?", ak)
		}
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("map_status = ?", *filter.Status)
	}

	var total int64
	if err := query.Model(&model.MilestoneMap{}).Count(&total).Error; err != nil {
		return nil, err
	}

	offset, p, ps := dto.ApplyPaginationDefaults(page.Page, page.PageSize)
	page.Page = p
	page.PageSize = ps

	var items []model.MilestoneMap
	if err := query.Order("id DESC").Offset(offset).Limit(page.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	return &dto.PageResult[model.MilestoneMap]{
		Items: items,
		Total: total,
		Page:  page.Page,
		Size:  page.PageSize,
	}, nil
}

func (r *milestoneMapRepo) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.MilestoneMap{}).
		Where("id = ? AND deleted_flag = 0", id).
		Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error
}
