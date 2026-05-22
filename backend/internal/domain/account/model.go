package account

import "time"

type Profile struct {
	Id          string    `xorm:"'id' pk uuid"`
	DisplayName string    `xorm:"'display_name' notnull"`
	AvatarUrl   *string   `xorm:"'avatar_url'"`
	Status      string    `xorm:"'status' notnull default 'active'"`
	CreatedAt   time.Time `xorm:"'created_at' notnull default now()"`
}

func (Profile) TableName() string { return "profiles" }

type UserRole struct {
	Id        string    `xorm:"'id' pk uuid"`
	UserId    string    `xorm:"'user_id' notnull uuid"`
	Role      string    `xorm:"'role' notnull"`
	GrantedBy *string   `xorm:"'granted_by' uuid"`
	GrantedAt time.Time `xorm:"'granted_at' notnull default now()"`
	CreatedAt time.Time `xorm:"'created_at' notnull default now()"`
}

func (UserRole) TableName() string { return "user_roles" }

type Exam struct {
	Id          string    `xorm:"'id' pk uuid"`
	Code        string    `xorm:"'code' notnull unique"`
	Name        string    `xorm:"'name' notnull"`
	Status      string    `xorm:"'status' notnull default 'active'"`
	Description *string   `xorm:"'description'"`
	CreatedAt   time.Time `xorm:"'created_at' notnull default now()"`
}

func (Exam) TableName() string { return "exams" }

type ExamEnrollment struct {
	Id        string     `xorm:"'id' pk uuid"`
	UserId    string     `xorm:"'user_id' notnull uuid"`
	ExamId    string     `xorm:"'exam_id' notnull uuid"`
	Status    string     `xorm:"'status' notnull"`
	StartedAt time.Time  `xorm:"'started_at' notnull"`
	PassedAt  *time.Time `xorm:"'passed_at'"`
	CreatedAt time.Time  `xorm:"'created_at' notnull default now()"`
}

func (ExamEnrollment) TableName() string { return "exam_enrollments" }

type Wallet struct {
	Id           string    `xorm:"'id' pk uuid"`
	UserId       string    `xorm:"'user_id' notnull uuid unique"`
	CoinsBalance int       `xorm:"'coins_balance' notnull default 0"`
	CreatedAt    time.Time `xorm:"'created_at' notnull default now()"`
	UpdatedAt    time.Time `xorm:"'updated_at' notnull default now()"`
}

func (Wallet) TableName() string { return "wallets" }

type Streak struct {
	Id               string     `xorm:"'id' pk uuid"`
	ExamEnrollmentId string     `xorm:"'exam_enrollment_id' notnull uuid unique"`
	UserId           string     `xorm:"'user_id' notnull uuid"`
	CurrentStreak    int        `xorm:"'current_streak' notnull default 0"`
	BestStreak       int        `xorm:"'best_streak' notnull default 0"`
	LastStudyAt      *time.Time `xorm:"'last_study_at'"`
	Status           string     `xorm:"'status' notnull default 'active'"`
	CreatedAt        time.Time  `xorm:"'created_at' notnull default now()"`
	UpdatedAt        time.Time  `xorm:"'updated_at' notnull default now()"`
}

func (Streak) TableName() string { return "streaks" }

type VirtualPet struct {
	Id               string     `xorm:"'id' pk uuid"`
	ExamEnrollmentId string     `xorm:"'exam_enrollment_id' notnull uuid unique"`
	Species          string     `xorm:"'species' notnull"`
	Level            int        `xorm:"'level' notnull default 1"`
	Xp               int        `xorm:"'xp' notnull default 0"`
	EvolutionStage   string     `xorm:"'evolution_stage' notnull"`
	MoodState        string     `xorm:"'mood_state' notnull"`
	Status           string     `xorm:"'status' notnull"`
	LastActiveAt     *time.Time `xorm:"'last_active_at'"`
	CreatedAt        time.Time  `xorm:"'created_at' notnull default now()"`
}

func (VirtualPet) TableName() string { return "virtual_pets" }

type ExamSummary struct {
	Id     string `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type ActiveEnrollment struct {
	Id       string `json:"id"`
	ExamId   string `json:"exam_id"`
	ExamCode string `json:"exam_code"`
	ExamName string `json:"exam_name"`
	Status   string `json:"status"`
}

type BootstrapInput struct {
	UserID      string
	DisplayName string
}

type CreateEnrollmentInput struct {
	UserID string
	ExamID string
}

type PlatformStats struct {
	TotalExams       int64 `json:"total_exams"`
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers7Days int64 `json:"active_users_7d"`
}

type AdminUserSummary struct {
	Id          string    `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	Roles       []string  `json:"roles"`
}
