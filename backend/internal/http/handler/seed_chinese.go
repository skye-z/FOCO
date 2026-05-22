package handler

import (
	"net/http"

	"xorm.io/xorm"
)

type SeedChineseHandler struct {
	engine *xorm.Engine
}

func NewSeedChineseHandler(engine *xorm.Engine) *SeedChineseHandler {
	return &SeedChineseHandler{engine: engine}
}

func (h *SeedChineseHandler) Run(w http.ResponseWriter, r *http.Request) {
	if h.engine == nil {
		http.Error(w, "db unavailable", http.StatusInternalServerError)
		return
	}
	sess := h.engine.NewSession().Context(r.Context())
	defer sess.Close()

	results := make([]map[string]string, 0)

	type u struct{ table, id, name, desc string }
	for _, item := range []u{
		{"exams", "60000000-0000-0000-0000-000000000101", "CFA 一级", "特许金融分析师一级考试"},
		{"subjects", "60000000-0000-0000-0000-000000000201", "定量方法", ""},
		{"subjects", "60000000-0000-0000-0000-000000000202", "道德与职业准则", ""},
		{"subjects", "60000000-0000-0000-0000-000000000203", "财务报表分析", ""},
		{"chapters", "60000000-0000-0000-0000-000000000301", "货币时间价值", ""},
		{"chapters", "60000000-0000-0000-0000-000000000302", "独立性与客观性", ""},
		{"chapters", "60000000-0000-0000-0000-000000000303", "存货与长期资产", ""},
		{"knowledge_points", "60000000-0000-0000-0000-000000000401", "折现与现值", "理解折现率和时间如何影响现值"},
		{"knowledge_points", "60000000-0000-0000-0000-000000000402", "独立性与客观性应对", "识别收到可能威胁独立性的礼物时应采取的正确行动"},
		{"knowledge_points", "60000000-0000-0000-0000-000000000403", "IFRS存货减值", "IFRS下存货减值对财务报表的影响"},
	} {
		var err error
		if item.desc != "" {
			_, err = sess.Exec("UPDATE "+item.table+" SET name = $1, description = $2 WHERE id = $3::uuid", item.name, item.desc, item.id)
		} else {
			_, err = sess.Exec("UPDATE "+item.table+" SET name = $1 WHERE id = $2::uuid", item.name, item.id)
		}
		s := "ok"
		if err != nil {
			s = err.Error()
		}
		results = append(results, map[string]string{"table": item.table, "id": item.id, "name": item.name, "status": s})
	}

	type stem struct{ id, text string }
	for _, s := range []stem{
		{"60000000-0000-0000-0000-000000000601", "保持未来现金流不变，折现率上升时，现值会如何变化？"},
		{"60000000-0000-0000-0000-000000000602", "一位投资组合经理在审查委托前收到了券商送来的昂贵体育比赛门票。以下哪项是最佳的第一步应对措施？"},
		{"60000000-0000-0000-0000-000000000603", "在IFRS下，存货减值最直接减少的是以下哪个项目？"},
	} {
		_, err := sess.Exec(
			"UPDATE question_versions SET stem = $1::jsonb, options = $2::jsonb, correct_answer = $3::jsonb, explanation = $4::jsonb WHERE id = $5::uuid",
			`{"text":"`+s.text+`"}`,
			`{"A":"选项A","B":"选项B","C":"选项C"}`,
			`{"answer":"B"}`,
			`{"text":"详细解析"}`,
			s.id,
		)
		st := "ok"
		if err != nil {
			st = err.Error()
		}
		results = append(results, map[string]string{"table": "question_versions", "id": s.id, "status": st})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": results, "meta": map[string]any{}, "error": nil})
}
