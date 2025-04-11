package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
)

// dueTimeHandling is specific to Google tasks with a due time of 22:00:00Z.
// This seems to be the only indicator in the Google takeout dataset for whole
// day tasks. As there could be tasks that have a due time of 22:00:00Z defined,
// dueTimeHandling specifies how these tasks should be converted.
type dueTimeHandling string

const (
	// Convert Google task to whole day tasks.org task.
	dueTimeHandlingDay dueTimeHandling = "day"

	// Convert Google task to specific time tasks.org task.
	dueTimeHandlingTime dueTimeHandling = "time"

	// For each occurrence, show an interactive prompt.
	dueTimeHandlingAsk dueTimeHandling = "ask"
)

// String implements [github.com/spf13/pflag.Value].
func (d *dueTimeHandling) String() string {
	if d == nil {
		return "<nil>"
	}
	return string(*d)
}

// Set implements [github.com/spf13/pflag.Value].
func (d *dueTimeHandling) Set(s string) error {
	switch s {
	case string(dueTimeHandlingDay):
		*d = dueTimeHandlingDay
	case string(dueTimeHandlingTime):
		*d = dueTimeHandlingTime
	case string(dueTimeHandlingAsk):
		*d = dueTimeHandlingAsk
	default:
		return fmt.Errorf(`invalid value %q, valid values are: "day", "time", "ask"`, s)
	}
	return nil
}

// Type implements [github.com/spf13/pflag.Value].
func (d *dueTimeHandling) Type() string {
	return "string"
}

func transformGoogleTasksItem(item GoogleTasksItem, dtHandling dueTimeHandling) TaskOrgItemTask {
	var (
		priority     = 2
		dueTimeMagic = time.Second.Milliseconds() // See comment on DueDate.
	)

	result := TaskOrgItemTask{
		RemoteID:         item.ID,
		Priority:         priority,
		Title:            strings.TrimSpace(item.Title),
		Notes:            strings.TrimSpace(item.Notes),
		CreationDate:     item.Created.UnixMilli(),
		ModificationDate: item.Updated.UnixMilli(),
		DueDate:          item.Due.UnixMilli() + dueTimeMagic,
	}
	if item.Due.Hour() == 22 && item.Due.Minute() == 0 && item.Due.Second() == 0 {
		if dtHandling == dueTimeHandlingAsk {
			fmt.Fprintln(os.Stderr, "The following task could be interpreted to have no time.")
			fmt.Fprintln(os.Stderr, "  ID:   ", result.RemoteID)
			fmt.Fprintln(os.Stderr, "  Title:", result.Title)
			fmt.Fprintln(os.Stderr, "  Due:  ", item.Due)

			for {
				fmt.Fprint(os.Stderr, "Convert to next day task with no time (1), or same day task with time (2): ")

				var input string
				fmt.Scanln(&input)

				switch input {
				case "1":
					dtHandling = dueTimeHandlingDay
					goto useDay
				case "2":
					dtHandling = dueTimeHandlingTime
					// Already handled by the initialization.
					goto skipUseDay
				}
			}
		}

	useDay:
		if dtHandling == dueTimeHandlingDay {
			// Remove dueTimeMagic to indicate a whole day task. Add two hours to
			// reach 00:00Z of the next day.
			result.DueDate = result.DueDate + (2 * time.Hour).Milliseconds() - dueTimeMagic
		}

	skipUseDay:
		slog.Info("Resolved ambiguous task.", "used_handling", dtHandling)
	}

	if !item.Completed.IsZero() {
		result.CompletionDate = item.Completed.UnixMilli()
	}

	return result
}
