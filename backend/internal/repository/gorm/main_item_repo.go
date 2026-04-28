package gorm

import (
	"context"
	"fmt"

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

func (r *mainItemRepo) List(ctx context.Context, teamID uint, filter dto.MainItemFilter, page dto.Pagination) (*dto.PageResult[model.MainItem], error) {
	query := r.db.WithContext(ctx).Scopes(NotDeleted).Where("team_key = ?", teamID)

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

func (r *mainItemRepo) NextCode(ctx context.Context, teamID uint) (string, error) {
	var code string
	err := r.db.WithContext(ctx).Transaction(func(tx *gormlib.DB) error {
		// Increment counter first — real write acquires SQLite write lock, serializing concurrent calls.
		if err := tx.Exec("UPDATE pmw_teams SET item_seq = item_seq + 1 WHERE id = ?", teamID).Error; err != nil {
			return err
		}
		var team model.Team
		if err := tx.First(&team, teamID).Error; err != nil {
			return err
		}
		seq := team.ItemSeq

		// If items were inserted directly with a higher seq (e.g. migration), skip past them.
		var maxSeq *int
		subExpr := r.dialect.Substr(dbutil.ColCode, len(team.Code)+2)
		castExpr := r.dialect.CastInt(dbutil.NewColumnExpr(subExpr))
		if err := tx.Model(&model.MainItem{}).
			Where("team_key = ?", teamID).
			Select("MAX(" + castExpr + ")").
			Scan(&maxSeq).Error; err != nil {
			return err
		}
		if maxSeq != nil && uint(*maxSeq) >= seq {
			seq = uint(*maxSeq) + 1
			if err := tx.Exec("UPDATE pmw_teams SET item_seq = ? WHERE id = ?", seq, teamID).Error; err != nil {
				return err
			}
		}

		code = fmt.Sprintf("%s-%05d", team.Code, seq)
		return nil
	})
	return code, err
}

func (r *mainItemRepo) CountByTeam(ctx context.Context, teamID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Scopes(NotDeleted).Model(&model.MainItem{}).Where("team_key = ?", teamID).Count(&count).Error
	return count, err
}

func (r *mainItemRepo) ListNonArchivedByTeam(ctx context.Context, teamID uint) ([]model.MainItem, error) {
	var items []model.MainItem
	err := r.db.WithContext(ctx).Scopes(NotDeleted).
		Where("team_key = ? AND archived_at IS NULL", teamID).
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

func (r *mainItemRepo) ListByTeamAndStatus(ctx context.Context, teamID uint, status string) ([]model.MainItem, error) {
	var items []model.MainItem
	err := r.db.WithContext(ctx).Scopes(NotDeleted).
		Where("team_key = ? AND item_status = ?", teamID, status).
		Find(&items).Error
	return items, err
}
