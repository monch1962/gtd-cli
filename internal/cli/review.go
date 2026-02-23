package cli

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/anomalyco/gtd-cli/internal/core"
	"github.com/anomalyco/gtd-cli/internal/jsonout"
)

var dailyReviewCmd = &cobra.Command{
	Use:   "daily-review",
	Short: "Daily review commands",
	Long:  "Commands for managing daily GTD reviews.",
}

var weeklyReviewCmd = &cobra.Command{
	Use:   "weekly-review",
	Short: "Weekly review commands",
	Long:  "Commands for managing weekly GTD reviews.",
}

var monthlyReviewCmd = &cobra.Command{
	Use:   "monthly-review",
	Short: "Monthly review commands",
	Long:  "Commands for managing monthly GTD reviews.",
}

var dailyReviewStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a daily review",
	Long: `Start a daily review session.

Creates a review session record and returns a review plan with items needing attention.

Example:
  gtd-cli daily-review start`,
	Run: func(cmd *cobra.Command, args []string) {
		startReview(cmd, core.ReviewTypeDaily, "gtd-cli daily-review start")
	},
}

var weeklyReviewStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a weekly review",
	Long: `Start a weekly review session.

Creates a review session record and returns a review plan with items needing attention.

Example:
  gtd-cli weekly-review start`,
	Run: func(cmd *cobra.Command, args []string) {
		startReview(cmd, core.ReviewTypeWeekly, "gtd-cli weekly-review start")
	},
}

var monthlyReviewStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a monthly review",
	Long: `Start a monthly review session.

Creates a review session record and returns a review plan with items needing attention.

Example:
  gtd-cli monthly-review start`,
	Run: func(cmd *cobra.Command, args []string) {
		startReview(cmd, core.ReviewTypeMonthly, "gtd-cli monthly-review start")
	},
}

func startReview(cmd *cobra.Command, reviewType core.ReviewType, command string) {
	app, err := newApp(rootCmd.Version)
	if err != nil {
		writeError(cmd, command, jsonout.ErrInternal, err.Error(), nil)
		return
	}
	defer app.Store.Close()

	now := app.Clock.Now()
	review := &core.ReviewSession{
		ID:        app.IDGen.NewID("rev_"),
		Type:      reviewType,
		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := app.Store.Reviews().Create(context.Background(), review); err != nil {
		writeError(cmd, command, jsonout.ErrInternal, err.Error(), nil)
		return
	}

	inboxStatus := core.TaskStatusInbox
	inboxResult, err := app.Store.Tasks().List(context.Background(), core.TaskFilter{Status: &inboxStatus, Limit: core.DefaultReviewLimit})
	if err != nil {
		writeError(cmd, command, jsonout.ErrInternal, err.Error(), nil)
		return
	}

	waitingStatus := core.TaskStatusWaiting
	waitingResult, err := app.Store.Tasks().List(context.Background(), core.TaskFilter{Status: &waitingStatus, Limit: core.DefaultReviewLimit})
	if err != nil {
		writeError(cmd, command, jsonout.ErrInternal, err.Error(), nil)
		return
	}

	activeStatus := core.ProjectStatusActive
	projects, err := app.Store.Projects().List(context.Background(), core.ProjectFilter{Status: &activeStatus, Limit: core.DefaultLimit})
	if err != nil {
		writeError(cmd, command, jsonout.ErrInternal, err.Error(), nil)
		return
	}

	var staleProjects []core.Project
	var projectsWithoutNext []string
	for _, p := range projects {
		nextTasks, err := app.Store.Tasks().GetNext(context.Background(), p.ID, "", 1)
		if err != nil || len(nextTasks) == 0 {
			staleProjects = append(staleProjects, p)
			projectsWithoutNext = append(projectsWithoutNext, p.ID)
		} else {
			staleThreshold := now.Add(-time.Duration(core.StaleProjectDays) * 24 * time.Hour)
			if p.UpdatedAt.Before(staleThreshold) {
				staleProjects = append(staleProjects, p)
			}
		}
	}

	ticklerStatus := core.TaskStatusTickler
	ticklerResult, err := app.Store.Tasks().List(context.Background(), core.TaskFilter{Status: &ticklerStatus, Limit: core.DefaultReviewLimit})
	if err != nil {
		writeError(cmd, command, jsonout.ErrInternal, err.Error(), nil)
		return
	}
	var ticklersDue []core.Task
	for _, t := range ticklerResult.Items {
		if t.TickleAt != nil && (t.TickleAt.Before(now) || t.TickleAt.Equal(now)) {
			ticklersDue = append(ticklersDue, t)
		}
	}

	checklist := getChecklist(reviewType)

	stats := map[string]int{
		"inbox_count":      inboxResult.Count,
		"waiting_count":    waitingResult.Count,
		"stale_projects":   len(staleProjects),
		"ticklers_due":     len(ticklersDue),
		"active_projects":  len(projects),
		"projects_no_next": len(projectsWithoutNext),
	}

	data := map[string]any{
		"review":    review,
		"checklist": checklist,
		"attention": map[string]any{
			"inbox": map[string]any{
				"count": inboxResult.Count,
				"items": inboxResult.Items,
			},
			"stale_projects": map[string]any{
				"count": len(staleProjects),
				"items": staleProjects,
			},
			"waiting": map[string]any{
				"count": waitingResult.Count,
				"items": waitingResult.Items,
			},
			"ticklers_due": map[string]any{
				"count": len(ticklersDue),
				"items": ticklersDue,
			},
		},
		"stats": stats,
	}

	writeSuccess(cmd, command, data)
}

func getChecklist(reviewType core.ReviewType) []map[string]any {
	switch reviewType {
	case core.ReviewTypeDaily:
		return []map[string]any{
			{"id": "daily_1", "title": "Review calendar", "description": "Check calendar for upcoming events and commitments"},
			{"id": "daily_2", "title": "Review inbox", "description": "Process any new items in inbox"},
			{"id": "daily_3", "title": "Review ticklers", "description": "Check for tickler items due today"},
			{"id": "daily_4", "title": "Review next actions", "description": "Identify tasks to work on today"},
		}
	case core.ReviewTypeWeekly:
		return []map[string]any{
			{"id": "weekly_1", "title": "Get clear", "description": "Collect and process all loose papers and notes"},
			{"id": "weekly_2", "title": "Get current", "description": "Review all project lists and next actions"},
			{"id": "weekly_3", "title": "Review waiting for", "description": "Check on delegated items and follow up"},
			{"id": "weekly_4", "title": "Review someday/maybe", "description": "Check for items to move to active projects"},
			{"id": "weekly_5", "title": "Review calendar", "description": "Review past and upcoming calendar entries"},
			{"id": "weekly_6", "title": "Review ticklers", "description": "Process tickler file for the week"},
		}
	case core.ReviewTypeMonthly:
		return []map[string]any{
			{"id": "monthly_1", "title": "Review goals", "description": "Review long-term goals and visions"},
			{"id": "monthly_2", "title": "Review areas of focus", "description": "Assess progress in each area of responsibility"},
			{"id": "monthly_3", "title": "Review someday/maybe", "description": "Consider activating or deleting items"},
			{"id": "monthly_4", "title": "Review projects", "description": "Ensure all projects have next actions"},
			{"id": "monthly_5", "title": "Review systems", "description": "Evaluate and improve GTD system setup"},
		}
	default:
		return nil
	}
}

func init() {
	rootCmd.AddCommand(dailyReviewCmd)
	rootCmd.AddCommand(weeklyReviewCmd)
	rootCmd.AddCommand(monthlyReviewCmd)

	dailyReviewCmd.AddCommand(dailyReviewStartCmd)
	weeklyReviewCmd.AddCommand(weeklyReviewStartCmd)
	monthlyReviewCmd.AddCommand(monthlyReviewStartCmd)
}
