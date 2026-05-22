package profile

import (
	"cmp"
	"math"
	"slices"
	"time"
)

func buildArchive(walletBalance int, streak *StreakStats, sessions []PracticeSessionSummary) Archive {
	archive := Archive{
		CoinsBalance: walletBalance,
	}

	if streak != nil {
		archive.CurrentStreak = streak.CurrentStreak
		archive.BestStreak = streak.BestStreak
		archive.StreakStatus = streak.Status
	}

	var firstStudiedAt *time.Time
	var lastStudiedAt *time.Time

	for _, session := range sessions {
		archive.TotalSessions++
		archive.TotalAnswered += session.AnsweredCount
		archive.TotalCorrect += session.CorrectCount
		archive.TotalXP += session.XPEarned

		currentCreatedAt := session.CreatedAt
		if firstStudiedAt == nil || currentCreatedAt.Before(*firstStudiedAt) {
			ts := currentCreatedAt
			firstStudiedAt = &ts
		}
		if lastStudiedAt == nil || currentCreatedAt.After(*lastStudiedAt) {
			ts := currentCreatedAt
			lastStudiedAt = &ts
		}
	}

	archive.Accuracy = percentage(archive.TotalCorrect, archive.TotalAnswered)
	archive.FirstStudiedAt = firstStudiedAt
	archive.LastStudiedAt = lastStudiedAt

	return archive
}

func buildRecords(now time.Time, sessions []PracticeSessionSummary) Records {
	records := Records{
		RecentSessions: make([]RecordSession, 0, len(sessions)),
	}

	monthlyWindow := now.AddDate(0, 0, -30)

	for _, session := range sessions {
		record := RecordSession{
			ID:              session.ID,
			Status:          session.Status,
			TotalCount:      session.TotalCount,
			AnsweredCount:   session.AnsweredCount,
			CorrectCount:    session.CorrectCount,
			Accuracy:        percentage(session.CorrectCount, session.AnsweredCount),
			XPEarned:        session.XPEarned,
			CoinsEarned:     session.CoinsEarned,
			DurationMinutes: durationMinutes(session.TotalDurationSeconds),
			CreatedAt:       session.CreatedAt,
			CompletedAt:     session.CompletedAt,
		}
		records.RecentSessions = append(records.RecentSessions, record)

		if !session.CreatedAt.Before(monthlyWindow) {
			records.MonthlySessions++
			records.MonthlyAnswered += session.AnsweredCount
			records.MonthlyXP += session.XPEarned
		}
	}

	return records
}

func buildPortrait(results []KnowledgePointResult, overallAccuracy int) Portrait {
	type statAccumulator struct {
		id      string
		name    string
		attempt int
		correct int
	}

	statsByKnowledgePoint := map[string]*statAccumulator{}
	totalAttempts := 0

	for _, result := range results {
		item, ok := statsByKnowledgePoint[result.KnowledgePointID]
		if !ok {
			item = &statAccumulator{
				id:   result.KnowledgePointID,
				name: result.KnowledgePointName,
			}
			statsByKnowledgePoint[result.KnowledgePointID] = item
		}

		item.attempt++
		totalAttempts++
		if result.Correct {
			item.correct++
		}
	}

	stats := make([]KnowledgePointStat, 0, len(statsByKnowledgePoint))
	for _, item := range statsByKnowledgePoint {
		stats = append(stats, KnowledgePointStat{
			KnowledgePointID:   item.id,
			KnowledgePointName: item.name,
			Attempts:           item.attempt,
			CorrectCount:       item.correct,
			Accuracy:           percentage(item.correct, item.attempt),
		})
	}

	weakPoints := append([]KnowledgePointStat(nil), stats...)
	slices.SortFunc(weakPoints, func(a, b KnowledgePointStat) int {
		return cmp.Or(
			cmp.Compare(a.Accuracy, b.Accuracy),
			cmp.Compare(b.Attempts, a.Attempts),
			cmp.Compare(a.KnowledgePointName, b.KnowledgePointName),
		)
	})
	if len(weakPoints) > 5 {
		weakPoints = weakPoints[:5]
	}

	strongPoints := append([]KnowledgePointStat(nil), stats...)
	slices.SortFunc(strongPoints, func(a, b KnowledgePointStat) int {
		return cmp.Or(
			cmp.Compare(b.Accuracy, a.Accuracy),
			cmp.Compare(b.Attempts, a.Attempts),
			cmp.Compare(a.KnowledgePointName, b.KnowledgePointName),
		)
	})
	if len(strongPoints) > 5 {
		strongPoints = strongPoints[:5]
	}

	return Portrait{
		TotalAttempts:   totalAttempts,
		OverallAccuracy: overallAccuracy,
		WeakPoints:      weakPoints,
		StrongPoints:    strongPoints,
	}
}

func buildUserKnowledgeGraph(graph *KnowledgeGraph, results []KnowledgePointResult) KnowledgeGraph {
	if graph == nil {
		return KnowledgeGraph{Nodes: []KnowledgeGraphNode{}, Edges: []KnowledgeGraphEdge{}}
	}

	type masteryAccumulator struct {
		attempts int
		correct  int
	}
	byPoint := map[string]*masteryAccumulator{}
	for _, result := range results {
		item, ok := byPoint[result.KnowledgePointID]
		if !ok {
			item = &masteryAccumulator{}
			byPoint[result.KnowledgePointID] = item
		}
		item.attempts++
		if result.Correct {
			item.correct++
		}
	}

	nodes := make([]KnowledgeGraphNode, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		next := node
		if node.Type == "knowledge_point" {
			if mastery, ok := byPoint[node.RefID]; ok && mastery.attempts > 0 {
				value := percentage(mastery.correct, mastery.attempts)
				next.Mastery = &value
				next.Attempts = mastery.attempts
			}
		}
		nodes = append(nodes, next)
	}

	return KnowledgeGraph{
		Nodes: nodes,
		Edges: append([]KnowledgeGraphEdge(nil), graph.Edges...),
	}
}

func percentage(correct, total int) int {
	if total <= 0 {
		return 0
	}
	return int(math.Floor(float64(correct) * 100 / float64(total)))
}

func durationMinutes(seconds int) int {
	if seconds <= 0 {
		return 0
	}
	return int(math.Ceil(float64(seconds) / 60))
}
