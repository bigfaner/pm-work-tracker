package gorm

import (
	"context"
	stderrors "errors"

	gormlib "gorm.io/gorm"

	"pm-work-tracker/backend/internal/model"
	"pm-work-tracker/backend/internal/pkg/errors"
	"pm-work-tracker/backend/internal/pkg/repo"
	"pm-work-tracker/backend/internal/repository"
)

type userRepo struct {
	db *gormlib.DB
}

// NewGormUserRepo creates a GORM-backed UserRepo.
func NewGormUserRepo(db *gormlib.DB) repository.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) FindByID(ctx context.Context, id uint) (*model.User, error) {
	return repo.FindByID[model.User](r.db, ctx, id)
}

func (r *userRepo) FindByBizKey(ctx context.Context, bizKey int64) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("biz_key = ?", bizKey).First(&user).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if stderrors.Is(err, gormlib.ErrRecordNotFound) {
			return nil, errors.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) List(ctx context.Context) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) FindByIDs(ctx context.Context, ids []uint) (map[uint]*model.User, error) {
	return repo.FindByIDs[model.User](r.db, ctx, ids)
}

func (r *userRepo) ListFiltered(ctx context.Context, search string, offset, limit int) ([]*model.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.User{})
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR display_name LIKE ?", pattern, pattern)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []*model.User
	if err := query.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *userRepo) SearchAvailable(ctx context.Context, teamID uint, search string, limit int) ([]*model.User, error) {
	query := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id NOT IN (?)", r.db.Table("pmw_team_members").Select("user_key").Where("team_key = ?", teamID))
	if search != "" {
		pattern := "%" + search + "%"
		query = query.Where("username LIKE ? OR display_name LIKE ?", pattern, pattern)
	}
	var users []*model.User
	if err := query.Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
