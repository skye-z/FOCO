package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"xorm.io/xorm"
)

type Repository interface {
	EnsureProfile(ctx context.Context, userID, displayName string) error
	EnsureUserRole(ctx context.Context, userID, role string) error
	ListRolesByUserID(ctx context.Context, userID string) ([]string, error)
	ListAdminUsers(ctx context.Context) ([]AdminUserSummary, error)
	GrantRole(ctx context.Context, userID, role, grantedBy string) error
	DisableProfile(ctx context.Context, userID string) error
	ListActiveExams(ctx context.Context) ([]ExamSummary, error)
	GetActiveEnrollmentByUserID(ctx context.Context, userID string) (*ActiveEnrollment, error)
	EnsureEnrollment(ctx context.Context, userID, examID string) (*ActiveEnrollment, error)
	CountExams(ctx context.Context) (int64, error)
	CountProfiles(ctx context.Context) (int64, error)
	CountActiveUsersSince(ctx context.Context, since time.Time) (int64, error)
}

func (r *XormRepository) CountExams(ctx context.Context) (int64, error) {
	return r.engine.Context(ctx).Where("status = 'active'").Count(&Exam{})
}

func (r *XormRepository) CountProfiles(ctx context.Context) (int64, error) {
	return r.engine.Context(ctx).Where("status = 'active'").Count(&Profile{})
}

func (r *XormRepository) CountActiveUsersSince(ctx context.Context, since time.Time) (int64, error) {
	var result []struct {
		Cnt int64 `xorm:"cnt"`
	}
	err := r.engine.Context(ctx).
		Table("exam_enrollments").
		Select("count(distinct user_id) as cnt").
		Where("created_at >= ?", since).
		Find(&result)
	if err != nil {
		return 0, err
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Cnt, nil
}

type XormRepository struct {
	engine *xorm.Engine
}

func NewRepository(engine *xorm.Engine) *XormRepository {
	return &XormRepository{engine: engine}
}

func (r *XormRepository) EnsureProfile(ctx context.Context, userID, displayName string) error {
	existing := &Profile{Id: userID}
	has, err := r.engine.Context(ctx).Get(existing)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	if displayName == "" {
		displayName = "Learner"
	}
	_, err = r.engine.Context(ctx).Insert(&Profile{
		Id:          userID,
		DisplayName: displayName,
		Status:      "active",
		CreatedAt:   time.Now(),
	})
	return err
}

func (r *XormRepository) EnsureUserRole(ctx context.Context, userID, role string) error {
	existing := &UserRole{UserId: userID, Role: role}
	has, err := r.engine.Context(ctx).Get(existing)
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	_, err = r.engine.Context(ctx).Insert(&UserRole{
		Id:        uuid.New().String(),
		UserId:    userID,
		Role:      role,
		GrantedAt: time.Now(),
		CreatedAt: time.Now(),
	})
	return err
}

func (r *XormRepository) ListRolesByUserID(ctx context.Context, userID string) ([]string, error) {
	var roles []UserRole
	err := r.engine.Context(ctx).Cols("role").Where("user_id = ?::uuid", userID).OrderBy("role asc").Find(&roles)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(roles))
	for _, r := range roles {
		result = append(result, r.Role)
	}
	return result, nil
}

func (r *XormRepository) ListAdminUsers(ctx context.Context) ([]AdminUserSummary, error) {
	var rows []adminUserRow
	err := r.engine.Context(ctx).SQL(listAdminUsersSQL).Find(&rows)
	if err != nil {
		return nil, err
	}

	return buildAdminUserSummaries(rows), nil
}

const listAdminUsersSQL = `
	SELECT
		COALESCE(auth_users.id, profiles.id)::text AS id,
		COALESCE(auth_users.email, '') AS email,
		COALESCE(NULLIF(profiles.display_name, ''), NULLIF(split_part(COALESCE(auth_users.email, ''), '@', 1), ''), 'User') AS display_name,
		COALESCE(
			NULLIF(profiles.status, ''),
			CASE
				WHEN auth_users.banned_until IS NOT NULL AND auth_users.banned_until > NOW() THEN 'disabled'
				ELSE 'active'
			END
		) AS status,
		COALESCE(profiles.created_at, auth_users.created_at) AS created_at,
		COALESCE(user_roles.role, '') AS role
	FROM auth.users AS auth_users
	FULL OUTER JOIN profiles ON profiles.id = auth_users.id
	LEFT JOIN user_roles ON user_roles.user_id = COALESCE(auth_users.id, profiles.id)
	WHERE COALESCE(auth_users.id, profiles.id) IS NOT NULL
	ORDER BY COALESCE(profiles.created_at, auth_users.created_at) DESC, user_roles.role ASC
`

type adminUserRow struct {
	Id          string    `xorm:"id"`
	Email       string    `xorm:"email"`
	DisplayName string    `xorm:"display_name"`
	Status      string    `xorm:"status"`
	CreatedAt   time.Time `xorm:"created_at"`
	Role        string    `xorm:"role"`
}

func buildAdminUserSummaries(rows []adminUserRow) []AdminUserSummary {
	byUser := make(map[string]*AdminUserSummary, len(rows))
	orderedIds := make([]string, 0, len(rows))
	for _, item := range rows {
		if strings.TrimSpace(item.Id) == "" {
			continue
		}
		summary, exists := byUser[item.Id]
		if !exists {
			summary = &AdminUserSummary{
				Id:          item.Id,
				Email:       item.Email,
				DisplayName: item.DisplayName,
				Status:      item.Status,
				CreatedAt:   item.CreatedAt,
				Roles:       []string{},
			}
			byUser[item.Id] = summary
			orderedIds = append(orderedIds, item.Id)
		}
		if item.Role != "" {
			summary.Roles = append(summary.Roles, item.Role)
		}
	}

	result := make([]AdminUserSummary, 0, len(orderedIds))
	for _, id := range orderedIds {
		result = append(result, *byUser[id])
	}
	return result
}

func (r *XormRepository) GrantRole(ctx context.Context, userID, role, grantedBy string) error {
	if role == "" {
		return fmt.Errorf("role is required")
	}

	existing := &UserRole{UserId: userID, Role: role}
	has, err := r.engine.Context(ctx).Get(existing)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	grantedByPtr := &grantedBy
	_, err = r.engine.Context(ctx).Insert(&UserRole{
		Id:        uuid.New().String(),
		UserId:    userID,
		Role:      role,
		GrantedBy: grantedByPtr,
		GrantedAt: time.Now(),
		CreatedAt: time.Now(),
	})
	return err
}

func (r *XormRepository) DisableProfile(ctx context.Context, userID string) error {
	_, err := r.engine.Context(ctx).
		ID(userID).
		Cols("status").
		Update(&Profile{Status: "disabled"})
	return err
}

func (r *XormRepository) ListActiveExams(ctx context.Context) ([]ExamSummary, error) {
	var exams []Exam
	err := r.engine.Context(ctx).Where("status = 'active'").OrderBy("created_at asc").Find(&exams)
	if err != nil {
		return nil, err
	}
	result := make([]ExamSummary, 0, len(exams))
	for _, e := range exams {
		result = append(result, ExamSummary{Id: e.Id, Code: e.Code, Name: e.Name, Status: e.Status})
	}
	return result, nil
}

func (r *XormRepository) GetActiveEnrollmentByUserID(ctx context.Context, userID string) (*ActiveEnrollment, error) {
	var enrollments []struct {
		ExamEnrollment `xorm:"extends"`
		Exam           `xorm:"extends"`
	}
	err := r.engine.Context(ctx).
		Table("exam_enrollments").
		Join("INNER", "exams", "exams.id = exam_enrollments.exam_id").
		Where("exam_enrollments.user_id = ?::uuid", userID).
		OrderBy("CASE WHEN exam_enrollments.status = 'in_progress' THEN 0 WHEN exam_enrollments.status = 'passed_manual' THEN 1 ELSE 2 END, exam_enrollments.created_at DESC").
		Limit(1).
		Find(&enrollments)
	if err != nil {
		return nil, err
	}
	if len(enrollments) == 0 {
		return nil, nil
	}
	e := enrollments[0]
	return &ActiveEnrollment{
		Id:       e.ExamEnrollment.Id,
		ExamId:   e.ExamEnrollment.ExamId,
		ExamCode: e.Exam.Code,
		ExamName: e.Exam.Name,
		Status:   e.ExamEnrollment.Status,
	}, nil
}

func (r *XormRepository) EnsureEnrollment(ctx context.Context, userID, examID string) (*ActiveEnrollment, error) {
	session := r.engine.NewSession().Context(ctx)
	defer session.Close()

	if err := session.Begin(); err != nil {
		return nil, err
	}

	existing, err := r.findEnrollment(session, userID, examID)
	if err != nil {
		_ = session.Rollback()
		return nil, err
	}
	if existing != nil {
		if err := ensureRuntimeRows(session, userID, existing.Id); err != nil {
			_ = session.Rollback()
			return nil, err
		}
		if err := session.Commit(); err != nil {
			return nil, err
		}
		return existing, nil
	}

	enrollment := &ExamEnrollment{
		Id:        uuid.New().String(),
		UserId:    userID,
		ExamId:    examID,
		Status:    "in_progress",
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
	}
	_, err = session.Insert(enrollment)
	if err != nil {
		_ = session.Rollback()
		return nil, err
	}

	if err := ensureRuntimeRows(session, userID, enrollment.Id); err != nil {
		_ = session.Rollback()
		return nil, err
	}

	created, err := r.findEnrollment(session, userID, examID)
	if err != nil {
		_ = session.Rollback()
		return nil, fmt.Errorf("read back enrollment: %w", err)
	}

	if err := session.Commit(); err != nil {
		return nil, err
	}
	return created, nil
}

func (r *XormRepository) findEnrollment(session *xorm.Session, userID, examID string) (*ActiveEnrollment, error) {
	var results []struct {
		ExamEnrollment `xorm:"extends"`
		Exam           `xorm:"extends"`
	}
	err := session.Table("exam_enrollments").
		Join("INNER", "exams", "exams.id = exam_enrollments.exam_id").
		Where("exam_enrollments.user_id = ?::uuid AND exam_enrollments.exam_id = ?::uuid", userID, examID).
		Limit(1).
		Find(&results)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	e := results[0]
	return &ActiveEnrollment{
		Id:       e.ExamEnrollment.Id,
		ExamId:   e.ExamEnrollment.ExamId,
		ExamCode: e.Exam.Code,
		ExamName: e.Exam.Name,
		Status:   e.ExamEnrollment.Status,
	}, nil
}

func ensureRuntimeRows(session *xorm.Session, userID, enrollmentID string) error {
	existing := &Wallet{UserId: userID}
	has, err := session.Where("user_id = ?::uuid", userID).Get(existing)
	if err != nil {
		return err
	}
	if !has {
		_, err = session.Insert(&Wallet{
			Id:           uuid.New().String(),
			UserId:       userID,
			CoinsBalance: 0,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
		if err != nil {
			return err
		}
	}

	hasStreak, _ := session.Where("exam_enrollment_id = ?::uuid", enrollmentID).Get(&Streak{})
	if !hasStreak {
		_, err = session.Insert(&Streak{
			Id:               uuid.New().String(),
			ExamEnrollmentId: enrollmentID,
			UserId:           userID,
			CurrentStreak:    0,
			BestStreak:       0,
			Status:           "active",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		})
		if err != nil {
			return err
		}
	}

	hasPet, _ := session.Where("exam_enrollment_id = ?::uuid", enrollmentID).Get(&VirtualPet{})
	if !hasPet {
		_, err = session.Insert(&VirtualPet{
			Id:               uuid.New().String(),
			ExamEnrollmentId: enrollmentID,
			Species:          "moss",
			Level:            1,
			Xp:               0,
			EvolutionStage:   "base",
			MoodState:        "healthy",
			Status:           "active",
			CreatedAt:        time.Now(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
