package auth

import (
	"context"

	"xorm.io/xorm"
)

type XormRoleRepository struct {
	engine *xorm.Engine
}

func NewXormRoleRepository(engine *xorm.Engine) *XormRoleRepository {
	return &XormRoleRepository{engine: engine}
}

func (r *XormRoleRepository) ListRolesByUserID(ctx context.Context, userID string) ([]string, error) {
	var roles []struct {
		Role string `xorm:"'role'"`
	}
	err := r.engine.Context(ctx).Table("user_roles").
		Where("user_id = ?::uuid", userID).
		OrderBy("role asc").
		Cols("role").
		Find(&roles)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(roles))
	for _, r := range roles {
		result = append(result, r.Role)
	}
	return result, nil
}
