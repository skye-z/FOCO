package content

import "time"

type Subject struct {
	Id        string    `xorm:"'id' pk uuid" json:"id"`
	ExamId    string    `xorm:"'exam_id' notnull uuid" json:"exam_id"`
	Code      string    `xorm:"'code' notnull" json:"code"`
	Name      string    `xorm:"'name' notnull" json:"name"`
	SortOrder int       `xorm:"'sort_order' notnull default 0" json:"sort_order"`
	CreatedAt time.Time `xorm:"'created_at' notnull default now()" json:"created_at"`
	UpdatedAt time.Time `xorm:"'updated_at' notnull default now()" json:"updated_at"`
}

func (Subject) TableName() string { return "subjects" }

type Chapter struct {
	Id        string    `xorm:"'id' pk uuid" json:"id"`
	SubjectId string    `xorm:"'subject_id' notnull uuid" json:"subject_id"`
	Code      string    `xorm:"'code' notnull" json:"code"`
	Name      string    `xorm:"'name' notnull" json:"name"`
	SortOrder int       `xorm:"'sort_order' notnull default 0" json:"sort_order"`
	CreatedAt time.Time `xorm:"'created_at' notnull default now()" json:"created_at"`
	UpdatedAt time.Time `xorm:"'updated_at' notnull default now()" json:"updated_at"`
}

func (Chapter) TableName() string { return "chapters" }

type KnowledgePoint struct {
	Id          string    `xorm:"'id' pk uuid" json:"id"`
	ExamId      string    `xorm:"'exam_id' notnull uuid" json:"exam_id"`
	Code        string    `xorm:"'code' notnull" json:"code"`
	Name        string    `xorm:"'name' notnull" json:"name"`
	Description *string   `xorm:"'description'" json:"description"`
	Status      string    `xorm:"'status' notnull default 'active'" json:"status"`
	CreatedAt   time.Time `xorm:"'created_at' notnull default now()" json:"created_at"`
	UpdatedAt   time.Time `xorm:"'updated_at' notnull default now()" json:"updated_at"`
}

func (KnowledgePoint) TableName() string { return "knowledge_points" }

type Question struct {
	Id                        string    `xorm:"'id' pk uuid" json:"id"`
	ExamId                    string    `xorm:"'exam_id' notnull uuid" json:"exam_id"`
	SubjectId                 string    `xorm:"'subject_id' notnull uuid" json:"subject_id"`
	ChapterId                 *string   `xorm:"'chapter_id' uuid" json:"chapter_id"`
	AuthorId                  *string   `xorm:"'author_id' uuid" json:"author_id"`
	Status                    string    `xorm:"'status' notnull default 'draft'" json:"status"`
	CurrentPublishedVersionId *string   `xorm:"'current_published_version_id' uuid" json:"current_published_version_id"`
	CreatedAt                 time.Time `xorm:"'created_at' notnull default now()" json:"created_at"`
	UpdatedAt                 time.Time `xorm:"'updated_at' notnull default now()" json:"updated_at"`
}

func (Question) TableName() string { return "questions" }

type QuestionVersion struct {
	Id            string     `xorm:"'id' pk uuid" json:"id"`
	QuestionId    string     `xorm:"'question_id' notnull uuid" json:"question_id"`
	VersionNo     int        `xorm:"'version_no' notnull" json:"version_no"`
	Status        string     `xorm:"'status' notnull default 'draft'" json:"status"`
	QuestionType  string     `xorm:"'question_type' notnull" json:"question_type"`
	Difficulty    int        `xorm:"'difficulty' notnull" json:"difficulty"`
	Stem          string     `xorm:"'stem' jsonb notnull" json:"stem"`
	Options       *string    `xorm:"'options' jsonb" json:"options"`
	CorrectAnswer string     `xorm:"'correct_answer' jsonb notnull" json:"correct_answer"`
	Explanation   string     `xorm:"'explanation' jsonb notnull" json:"explanation"`
	PublishedAt   *time.Time `xorm:"'published_at'" json:"published_at"`
	PublishedBy   *string    `xorm:"'published_by' uuid" json:"published_by"`
	PublishNote   *string    `xorm:"'publish_note'" json:"publish_note"`
	ContentHash   *string    `xorm:"'content_hash'" json:"content_hash"`
	CreatedAt     time.Time  `xorm:"'created_at' notnull default now()" json:"created_at"`
	UpdatedAt     time.Time  `xorm:"'updated_at' notnull default now()" json:"updated_at"`
}

func (QuestionVersion) TableName() string { return "question_versions" }

type Exam struct {
	Id               string     `xorm:"'id' pk uuid" json:"id"`
	Code             string     `xorm:"'code' notnull unique" json:"code"`
	Name             string     `xorm:"'name' notnull" json:"name"`
	Status           string     `xorm:"'status' notnull default 'active'" json:"status"`
	Description      *string    `xorm:"'description'" json:"description"`
	NextExamDate     *time.Time `xorm:"'next_exam_date'" json:"next_exam_date"`
	NextNextExamDate *time.Time `xorm:"'next_next_exam_date'" json:"next_next_exam_date"`
	CreatedAt        time.Time  `xorm:"'created_at' notnull default now()" json:"created_at"`
	UpdatedAt        time.Time  `xorm:"'updated_at' notnull default now()" json:"updated_at"`
}

func (Exam) TableName() string { return "exams" }

type KnowledgePointEdge struct {
	Id                     string   `xorm:"'id' pk uuid" json:"id"`
	ExamId                 string   `xorm:"'exam_id' notnull uuid" json:"exam_id"`
	FromKnowledgePointId   string   `xorm:"'from_knowledge_point_id' notnull uuid" json:"from_knowledge_point_id"`
	ToKnowledgePointId     string   `xorm:"'to_knowledge_point_id' notnull uuid" json:"to_knowledge_point_id"`
	EdgeType               string   `xorm:"'edge_type' notnull" json:"edge_type"`
	Weight                 *float64 `xorm:"'weight'" json:"weight"`
	CreatedAt              time.Time `xorm:"'created_at' notnull default now()" json:"created_at"`
}

func (KnowledgePointEdge) TableName() string { return "knowledge_point_edges" }

type ExamTreeNode struct {
	Id               string         `json:"id"`
	Code             string         `json:"code"`
	Name             string         `json:"name"`
	Type             string         `json:"type"`
	NextExamDate     string         `json:"next_exam_date,omitempty"`
	NextNextExamDate string         `json:"next_next_exam_date,omitempty"`
	CountdownDays    *int           `json:"countdown_days,omitempty"`
	Children         []ExamTreeNode `json:"children,omitempty"`
}

type QuestionCard struct {
	Id           string   `json:"id"`
	ExamId       string   `json:"exam_id"`
	SubjectId    string   `json:"subject_id"`
	SubjectName  string   `json:"subject_name"`
	ChapterId    *string  `json:"chapter_id"`
	ChapterName  *string  `json:"chapter_name"`
	Status       string   `json:"status"`
	QuestionType string   `json:"question_type"`
	Difficulty   int      `json:"difficulty"`
	VersionNo    int      `json:"version_no"`
	VersionId    string   `json:"version_id"`
	StemPreview  string   `json:"stem_preview"`
	PublishedVersionNo *int `json:"published_version_no"`
	DraftVersionNo     *int `json:"draft_version_no"`
	HasUnpublishedDraft bool `json:"has_unpublished_draft"`
}

type ExamTreeFilter struct {
	ExamId         string
	SubjectId      string
	ChapterId      string
	KnowledgePoint string
	Difficulty     string
	Status         string
}

type QuestionVersionDetail struct {
	VersionId       string   `json:"version_id"`
	QuestionId      string   `json:"question_id"`
	ExamId          string   `json:"exam_id"`
	SubjectId       string   `json:"subject_id"`
	ChapterId       *string  `json:"chapter_id"`
	QuestionType    string   `json:"question_type"`
	Difficulty      int      `json:"difficulty"`
	VersionNo       int      `json:"version_no"`
	Status          string   `json:"status"`
	Stem            string   `json:"stem"`
	Options         string   `json:"options"`
	CorrectAnswer   string   `json:"correct_answer"`
	Explanation     string   `json:"explanation"`
	KnowledgePointIds []string `json:"knowledge_point_ids"`
}

type QuestionVersionSummary struct {
	VersionId    string     `json:"version_id"`
	QuestionId   string     `json:"question_id"`
	VersionNo    int        `json:"version_no"`
	Status       string     `json:"status"`
	PublishedAt  *time.Time `json:"published_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PublishNote  *string    `json:"publish_note"`
	IsCurrent    bool       `json:"is_current"`
	IsPublished  bool       `json:"is_published"`
}

type KnowledgeGraphNode struct {
	Id          string `json:"id"`
	Type        string `json:"type"`
	RefId       string `json:"ref_id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	Group       string `json:"group,omitempty"`
}

type KnowledgeGraphEdge struct {
	Id       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Type     string `json:"type"`
	Label    string `json:"label,omitempty"`
}

type KnowledgeGraphPayload struct {
	Nodes []KnowledgeGraphNode `json:"nodes"`
	Edges []KnowledgeGraphEdge `json:"edges"`
}

type QuestionVersionKnowledgePoint struct {
	QuestionVersionId string    `xorm:"'question_version_id' uuid" json:"question_version_id"`
	KnowledgePointId  string    `xorm:"'knowledge_point_id' uuid" json:"knowledge_point_id"`
	CreatedAt         time.Time `xorm:"'created_at'" json:"created_at"`
}

func (QuestionVersionKnowledgePoint) TableName() string { return "question_version_knowledge_points" }

type ContentPackagePayload struct {
	Exams               []Exam                     `json:"exams"`
	Subjects            []Subject                  `json:"subjects"`
	Chapters            []Chapter                  `json:"chapters"`
	KnowledgePoints     []KnowledgePoint           `json:"knowledge_points"`
	KnowledgePointEdges []KnowledgePointEdge       `json:"knowledge_point_edges"`
	Questions           []ContentPackageQuestion   `json:"questions"`
	InteractiveUnits    []ContentPackageInteractiveUnit `json:"interactive_units,omitempty"`
}

type ContentPackageInteractiveUnit struct {
	ID            string                             `json:"id"`
	ExamID        string                             `json:"exam_id"`
	SubjectID     string                             `json:"subject_id"`
	Title         string                             `json:"title"`
	Status        string                             `json:"status"`
	VersionNo     int                                `json:"version_no"`
	VersionID     string                             `json:"version_id"`
	PublishedAt   *time.Time                         `json:"published_at"`
	Steps         []ContentPackageInteractiveStep    `json:"steps"`
	CreatedAt     time.Time                          `json:"created_at"`
	UpdatedAt     time.Time                          `json:"updated_at"`
}

type ContentPackageInteractiveStep struct {
	ID                 string `json:"id"`
	StepNo             int    `json:"step_no"`
	WidgetType         string `json:"widget_type"`
	Content            string `json:"content"`
	InitialState       string `json:"initial_state"`
	AllowedActions     string `json:"allowed_actions"`
	EvaluationConfig   string `json:"evaluation_config"`
	FeedbackMap        string `json:"feedback_map"`
	HintPolicy         string `json:"hint_policy"`
	KnowledgePointIDs  string `json:"knowledge_point_ids"`
	KnowledgePointTags string `json:"knowledge_point_tags"`
}

type ContentPackageQuestion struct {
	Id                        string   `json:"id"`
	ExamId                    string   `json:"exam_id"`
	SubjectId                 string   `json:"subject_id"`
	ChapterId                 *string  `json:"chapter_id"`
	Status                    string   `json:"status"`
	CurrentPublishedVersionId *string  `json:"current_published_version_id"`
	VersionNo                 int      `json:"version_no"`
	QuestionType              string   `json:"question_type"`
	Difficulty                int      `json:"difficulty"`
	Stem                      string   `json:"stem"`
	Options                   *string  `json:"options"`
	CorrectAnswer             string   `json:"correct_answer"`
	Explanation               string   `json:"explanation"`
	PublishedAt               *time.Time `json:"published_at"`
	PublishedBy               *string  `json:"published_by"`
	PublishNote               *string  `json:"publish_note"`
	KnowledgePointIds         []string `json:"knowledge_point_ids"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

type ContentPackageImportReport struct {
	ExamsImported                         int      `json:"exams_imported"`
	SubjectsImported                      int      `json:"subjects_imported"`
	ChaptersImported                      int      `json:"chapters_imported"`
	KnowledgePointsImported               int      `json:"knowledge_points_imported"`
	KnowledgePointEdgesImported           int      `json:"knowledge_point_edges_imported"`
	QuestionsImported                     int      `json:"questions_imported"`
	QuestionVersionsImported              int      `json:"question_versions_imported"`
	QuestionVersionKnowledgePointsImported int      `json:"question_version_knowledge_points_imported"`
	InteractiveUnitsImported              int      `json:"interactive_units_imported"`
	InteractiveUnitStepsImported          int      `json:"interactive_unit_steps_imported"`
	ValidationErrors                      []string `json:"validation_errors"`
}
