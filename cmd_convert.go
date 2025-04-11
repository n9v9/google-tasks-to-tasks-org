package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

type convertConfig struct {
	calendarUUID    string
	outFile         string
	googleTasksFile string
	tasksOrgFile    string
	dueTimeHandling dueTimeHandling
}

func convertCmd() *cobra.Command {
	var cfg convertConfig
	cfg.dueTimeHandling = dueTimeHandlingAsk

	cmd := &cobra.Command{
		Use:   "convert [flags] [GOOGLE_TASKS_FILE] [TASKS_ORG_FILE]",
		Short: `Convert tasks from the Google Tasks format into the tasks.org format`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.googleTasksFile = args[0]
			cfg.tasksOrgFile = args[1]

			ctx := cmd.Context()
			run(ctx, func() error { return convertRun(ctx, cfg) })
		},
	}

	cmd.Flags().StringVar(&cfg.outFile, "out", "-", `File to output data to, or "-" for stdout.`)
	cmd.Flags().StringVar(&cfg.calendarUUID, "calendar-uuid", "", `UUID of the calendar defined in the TASKS_ORG_FILE to use for new tasks.
Can be obtained via the JSONPath "$.data.caldavCalendars[*].uuid". (required)`)
	cmd.Flags().Var(&cfg.dueTimeHandling, "due-time-handling", `Google tasks with a time of 22:00:00Z can either mean a task with a due time of 22:00Z,
or a whole day task for the next day. This flag specifies how to handle these cases.
Valid values are: "day", "time", "ask" (default "ask")`)

	cmd.MarkFlagRequired("calendar-uuid")

	return cmd
}

func convertRun(ctx context.Context, cfg convertConfig) error {
	data, err := os.ReadFile(cfg.googleTasksFile)
	if err != nil {
		return fmt.Errorf("read google tasks source: %w", err)
	}

	var googleTasks GoogleTasksStructure
	if err := json.Unmarshal(data, &googleTasks); err != nil {
		return fmt.Errorf("unmarshal google tasks data: %w", err)
	}

	data, err = os.ReadFile(cfg.tasksOrgFile)
	if err != nil {
		return fmt.Errorf("read tasks.org source: %w", err)
	}

	var structure TaskOrgStructure
	if err := json.Unmarshal(data, &structure); err != nil {
		return fmt.Errorf("unmarshal tasks.org data: %w", err)
	}

	var validCalendarUUID bool
	for _, calendar := range structure.Calendars {
		if calendar.UUID == cfg.calendarUUID {
			validCalendarUUID = true
			break
		}
	}
	if !validCalendarUUID {
		return fmt.Errorf("invalid calendar uuid %q", cfg.calendarUUID)
	}

	existingByID := make(map[string]struct{}, len(structure.Tasks))
	for _, task := range structure.Tasks {
		existingByID[task.Task.RemoteID] = struct{}{}
	}

	var addedTasks int
	for _, googleTasksItem := range googleTasks.Items {
		task := transformGoogleTasksItem(googleTasksItem, cfg.dueTimeHandling)

		if _, ok := existingByID[task.RemoteID]; ok {
			slog.WarnContext(ctx, "Transformed task already exists in tasks.org file. Skipping.",
				"remote_id", task.RemoteID,
				"title", task.Title,
			)
			continue
		}

		structure.Tasks = append(structure.Tasks, TaskOrgItem{
			Task:   task,
			Alarms: []TaskOrgItemAlarm{{Type: 2}},
			CaldavTasks: []TaskOrgItemCaldavTasks{{
				Calendar: cfg.calendarUUID,
				RemoteId: uuid.New().String(),
			}},
			Geofences:   nil,
			Tags:        nil,
			Comments:    nil,
			Attachments: nil,
		})
		addedTasks++
	}

	data, err = json.Marshal(&structure)
	if err != nil {
		return fmt.Errorf("marshal result json: %w", err)
	}

	out := io.WriteCloser(os.Stdout)
	if cfg.outFile != "-" {
		out, err = os.Create(cfg.outFile)
		if err != nil {
			return fmt.Errorf("create out file: %w", err)
		}
	}
	defer func() {
		if err := out.Close(); err != nil {
			slog.ErrorContext(ctx, "Failed to close out file", "file", cfg.outFile, "error", err)
		}
	}()

	_, err = fmt.Fprintln(out, string(data))
	if err != nil {
		return fmt.Errorf("write result json to out file: %w", err)
	}
	slog.InfoContext(ctx, "Written output file containing existing and transformed tasks.",
		"out", cfg.outFile,
		"added_tasks", addedTasks,
	)

	return nil
}
