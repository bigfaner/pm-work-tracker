package gorm

import (
	"context"
	"fmt"
	"time"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/dto"
	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/dbutil"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/repository"
)

type mainItemRepo struct {
	db      *gormlib.DB
	dialect dbutil.Dialect
}

// NewGormMainItemRepo creates a GORM-backed MainItemRepo.
func NewGormMainItemRepo(db *gormlib.DB, dialect dbutil.Dialect) repository.MainItemRepo {
	return &mainItemRepo{db: db, dialect: dialect}
}

func (r *mainItemRepo) Create(ctx context.Context, item *model.MainItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *mainItemRepo) FindByID(ctx context.Context, id uint) (*model.MainItem, error) {
	return repo.FindByID[model.MainItem](r.db, ctx, id)
}

func (r *mainItemRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.MainItem, error) {
	var item model.MainItem
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key = ?", bizKey).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *mainItemRepo) Update(ctx context.Context, item *model.MainItem, fields map[string]interface{}) error {
	return repo.UpdateFields[model.MainItem](r.db, ctx, item, item.TeamKey, fields)
}

func (r *mainItemRepo) List(ctx context.Context, teamBizKey int64, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	query := r.db.WithContext(ctx).Scopes(NotDeleted).Where("team_key = ?", teamBizKey)

	// Filter out archived by default; include when filter.Archived is true
	if !filter.Archived {
		query = query.Where("archived_at IS NULL")
	}

	query = applyItemFilter(query, filter.Status, filter.Priority, filter.AssigneeKey, filter.IsKeyItem)

	var total int64
	if err := query.Model(&model.MainItem{}).Count(&total).Error; err != nil {
		return nil, err
	}

	// Apply pagination
	offset, p, ps := dto.ApplyPaginationDefaults(page.Page, page.PageSize)
	page.Page = p
	page.PageSize = ps

	var items []model.MainItem
	if err := query.Order("id DESC").Offset(offset).Limit(page.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	return &dto.PageResult[model.MainItem]{
		Items: items,
		Total: total,
		Page:  page.Page,
		Size:  page.PageSize,
	}, nil
}

//nolint:dupl // same transaction shape as subItemRepo.NextSubCode; core logic extracted to nextSeqInTx
func (r *mainItemRepo) NextCode(ctx context.Context, teamBizKey int64) (string, error) {
	var code string
	err := r.db.WithContext(ctx).Transaction(func(tx *gormlib.DB) error {
		// Increment counter first — real write acquires SQLite write lock, serializing concurrent calls.
		if err := tx.Exec("UPDATE pmw_teams SET item_seq = item_seq + 1 WHERE biz_key = ?", teamBizKey).Error; err != nil {
			return err
		}
		var team model.Team
		if err := tx.Where("biz_key = ?", teamBizKey).First(&team).Error; err != nil {
			return err
		}

		seq, err := nextSeqInTx(tx, r.dialect, "pmw_teams", "item_seq", "biz_key", teamBizKey, team.ItemSeq, team.Code, "team_key", teamBizKey, &model.MainItem{})
		if err != nil {
			return err
		}

		code = fmt.Sprintf("%s-%05d", team.Code, seq)
		return nil
	})
	return code, err
}

func (r *mainItemRepo) CountByTeam(ctx context.Context, teamBizKey int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Model(&model.MainItem{}).Where("team_key = ?", teamBizKey).Count(&count).Error
	return count, err
}

func (r *mainItemRepo) ListNonArchivedByTeam(ctx context.Context, teamBizKey int64) ([]model.MainItem, error) {
	var items []model.MainItem
	err := r.db.WithContext(ctx).Scopes(NotDeleted).
		Where("team_key = ? AND archived_at IS NULL", teamBizKey).
		Find(&items).Error
	return items, err
}

func (r *mainItemRepo) FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.MainItem, error) {
	return repo.FindByIDs[model.MainItem](r.db, ctx, ids)
}

func (r *mainItemRepo) FindByBizKeys(ctx context.Context, bizKeys []int64) (map[int64]*model.MainItem, error) {
	result := make(map[int64]*model.MainItem)
	if len(bizKeys) == 0 {
		return result, nil
	}
	var items []*model.MainItem
	if err := r.db.WithContext(ctx).Scopes(NotDeleted).Where("biz_key IN ?", bizKeys).Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		result[item.BizKey] = item
	}
	return result, nil
}

func (r *mainItemRepo) ListByTeamAndStatus(ctx context.Context, teamBizKey int64, status string) ([]model.MainItem, error) {
	var items []model.MainItem
	err := r.db.WithContext(ctx).Scopes(NotDeleted).
		Where("team_key = ? AND item_status = ?", teamBizKey, status).
		Find(&items).Error
	return items, err
}

func (r *mainItemRepo) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&model.MainItem{}).
		Where("id = ?", id).
		Updates(map[string]any{"deleted_flag": 1, "deleted_time": time.Now()}).Error
}

func (r *mainItemRepo) CascadeSoftDelete(ctx context.Context, mainItemID uint, subItemIDs []uint, histories []model.StatusHistory) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gormlib.DB) error {
		now := time.Now()
		// Soft-delete all sub-items
		if len(subItemIDs) > 0 {
			if err := tx.Model(&model.SubItem{}).
				Where("id IN ?", subItemIDs).
				Updates(map[string]any{"deleted_flag": 1, "deleted_time": now}).Error; err != nil {
				return err
			}
		}
		// Soft-delete the main item
		if err := tx.Model(&model.MainItem{}).
			Where("id = ?", mainItemID).
			Updates(map[string]any{"deleted_flag": 1, "deleted_time": now}).Error; err != nil {
			return err
		}
		// Insert status history records for audit
		for i := range histories {
			histories[i].CreateTime = now
			if err := tx.Create(&histories[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
