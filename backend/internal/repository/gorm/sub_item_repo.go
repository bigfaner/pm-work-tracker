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

type subItemRepo struct {
	db      *gormlib.DB
	dialect dbutil.Dialect
}

// NewGormSubItemRepo creates a GORM-backed SubItemRepo.
func NewGormSubItemRepo(db *gormlib.DB, dialect dbutil.Dialect) repository.SubItemRepo {
	return &subItemRepo{db: db, dialect: dialect}
}

func (r *subItemRepo) Create(ctx context.Context, item *model.SubItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *subItemRepo) FindByID(ctx context.Context, id uint) (*model.SubItem, error) {
	return repo.FindByID[model.SubItem](r.db, ctx, id)
}

func (r *subItemRepo) Update(ctx context.Context, item *model.SubItem, fields map[string]interface{}) error {
	return repo.UpdateFields[model.SubItem](r.db, ctx, item, item.TeamKey, fields)
}

func (r *subItemRepo) SoftDelete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.SubItem{}, id).Error
}

func (r *subItemRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.SubItem, error) {
	var item model.SubItem
	err := r.db.WithContext(ctx).Where("biz_key = ?", bizKey).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *subItemRepo) List(ctx context.Context, teamID uint, mainItemID uint, filter dto.SubItemFilter, page dto.Pagination) (*dto.PageResult[model.SubItem], error) {
	query := r.db.WithContext(ctx).Where("team_key = ?", teamID)

	if mainItemID > 0 {
		query = query.Where("main_item_key = ?", mainItemID)
	}

	query = applyItemFilter(query, filter.Status, filter.Priority, filter.AssigneeKey, filter.IsKeyItem)

	var total int64
	if err := query.Model(&model.SubItem{}).Count(&total).Error; err != nil {
		return nil, err
	}

	offset, p, ps := dto.ApplyPaginationDefaults(page.Page, page.PageSize)
	page.Page = p
	page.PageSize = ps

	var items []model.SubItem
	if err := query.Order("id DESC").Offset(offset).Limit(page.PageSize).Find(&items).Error; err != nil {
		return nil, err
	}

	return &dto.PageResult[model.SubItem]{
		Items: items,
		Total: total,
		Page:  page.Page,
		Size:  page.PageSize,
	}, nil
}

func (r *subItemRepo) ListByMainItem(ctx context.Context, mainItemID uint) ([]*model.SubItem, error) {
	var items []*model.SubItem
	err := r.db.WithContext(ctx).
		Where("main_item_key = ?", mainItemID).
		Find(&items).Error
	return items, err
}

func (r *subItemRepo) ListByTeam(ctx context.Context, teamID uint) ([]model.SubItem, error) {
	var items []model.SubItem
	err := r.db.WithContext(ctx).
		Where("team_key = ?", teamID).
		Find(&items).Error
	return items, err
}

// NextSubCode generates the next sub-item code for the given main item.
// It locks the main item row (SELECT FOR UPDATE) to serialize concurrent calls,
// reads the current MAX sub sequence, and returns "{mainCode}-{seq:02d}".
func (r *subItemRepo) NextSubCode(ctx context.Context, mainItemID uint) (string, error) {
	var code string
	err := r.db.WithContext(ctx).Transaction(func(tx *gormlib.DB) error {
		// Atomically increment the counter — real write forces SQLite write lock.
		if err := tx.Exec("UPDATE pmw_main_items SET sub_item_seq = sub_item_seq + 1 WHERE id = ?", mainItemID).Error; err != nil {
			return err
		}
		var mainItem model.MainItem
		if err := tx.First(&mainItem, mainItemID).Error; err != nil {
			return err
		}
		seq := mainItem.SubItemSeq

		// If sub-items were inserted directly with a higher seq, skip past them.
		var maxSeq *int
		subExpr := r.dialect.Substr(dbutil.ColCode, len(mainItem.Code)+2)
		castExpr := r.dialect.CastInt(dbutil.NewColumnExpr(subExpr))
		if err := tx.Model(&model.SubItem{}).
			Where("main_item_key = ?", mainItemID).
			Select("MAX(" + castExpr + ")").
			Scan(&maxSeq).Error; err != nil {
			return err
		}
		if maxSeq != nil && uint(*maxSeq) >= seq {
			seq = uint(*maxSeq) + 1
			if err := tx.Exec("UPDATE pmw_main_items SET sub_item_seq = ? WHERE id = ?", seq, mainItemID).Error; err != nil {
				return err
			}
		}

		code = fmt.Sprintf("%s-%02d", mainItem.Code, seq)
		return nil
	})
	return code, err
}
