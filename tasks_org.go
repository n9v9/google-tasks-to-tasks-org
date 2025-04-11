package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
)

type TaskOrgItemAlarm struct {
	// Will always be 2, as that's what the backup I tested with contained.
	Type int `json:"type"`
}

type TaskOrgItemCaldavTasks struct {
	// Must be the ID of an existing calendar as identified by the following JSONPath
	// expression:
	//  $.data.caldavCalendars[*].uuid
	Calendar string `json:"calendar"`

	// Content does not matter, just has to be unique, see [discussion].
	//
	// [discussion]: https://github.com/tasks/tasks/discussions/2993#discussioncomment-10142080
	RemoteId string `json:"remoteId"`
}

type TaskOrgItemTask struct {
	// Content does not matter, just has to be unique, see [discussion]. Will be
	// equal to [GoogleTasksItem.ID].
	//
	// [discussion]: https://github.com/tasks/tasks/discussions/2993#discussioncomment-10142080
	RemoteID string `json:"remoteId"`

	Priority int    `json:"priority"`
	Title    string `json:"title"`
	Notes    string `json:"notes,omitempty"`

	// Unix timestamp.
	CreationDate int64 `json:"creationDate"`
	// Unix timestamp.
	ModificationDate int64 `json:"modificationDate"`
	// Unix timestamp.
	CompletionDate int64 `json:"completionDate,omitempty"`
	// Unix timestamp. For tasks with a due time set, a magic value of 1000 ms is
	// added, see [discussion]:
	//
	// [discussion]: https://github.com/tasks/tasks/discussions/2993#discussioncomment-10142080
	DueDate int64 `json:"dueDate"`
}

type TaskOrgItem struct {
	Task        TaskOrgItemTask          `json:"task"`
	Alarms      []TaskOrgItemAlarm       `json:"alarms"`
	CaldavTasks []TaskOrgItemCaldavTasks `json:"caldavTasks"`

	// Rest will be empty.
	Geofences   []any `json:"geofences"`
	Tags        []any `json:"tags"`
	Comments    []any `json:"comments"`
	Attachments []any `json:"attachments"`
}

type TaskOrgCaldavCalendar struct {
	UUID    string `json:"uuid"`
	Account string `json:"account"`
	Name    string `json:"name"`
}

type TaskOrgStructure struct {
	raw       map[string]json.RawMessage
	data      map[string]json.RawMessage
	Calendars []TaskOrgCaldavCalendar
	Tasks     []TaskOrgItem
}

func (t *TaskOrgStructure) UnmarshalJSON(bytes []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &raw); err != nil {
		return fmt.Errorf("unmarshal root: %w", err)
	}

	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw["data"], &data); err != nil {
		return fmt.Errorf("unmarshal data: %w", err)
	}
	delete(raw, "data")

	var caldavCalendars []TaskOrgCaldavCalendar
	if err := json.Unmarshal(data["caldavCalendars"], &caldavCalendars); err != nil {
		return fmt.Errorf("unmarshal caldav calendars: %w", err)
	}
	slog.Info("Read Tasks.org calendars.", "count", len(caldavCalendars))

	var tasks []TaskOrgItem
	if err := json.Unmarshal(data["tasks"], &tasks); err != nil {
		return fmt.Errorf("unmarshal tasks: %w", err)
	}
	// Delete because the tasks will be modified and then marshaled later on.
	delete(data, "tasks")
	slog.Info("Read Tasks.org tasks.", "count", len(tasks))

	t.raw = raw
	t.data = data
	t.Calendars = caldavCalendars
	t.Tasks = tasks

	return nil
}

func (t *TaskOrgStructure) MarshalJSON() ([]byte, error) {
	var err error

	data := maps.Clone(t.data)
	data["tasks"], err = json.Marshal(t.Tasks)
	if err != nil {
		return nil, fmt.Errorf("marshal tasks: %w", err)
	}

	raw := maps.Clone(t.raw)
	raw["data"], err = json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal data: %w", err)
	}

	return json.Marshal(raw)
}
