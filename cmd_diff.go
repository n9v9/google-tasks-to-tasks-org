package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/spf13/cobra"
)

type diffConfig struct {
	tasksOrgFile string
	newFile      string
}

func diffCmd() *cobra.Command {
	var cfg diffConfig
	return &cobra.Command{
		Use:   "diff [TASKS_ORG_FILE] [NEW_FILE]",
		Short: `Show the diff between the original TASKS_ORG_FILE and the NEW_FILE created by the "convert" command.`,
		Long: `This command can be used to verify the output of the "convert" command.

A simple "git diff --no-index" might not work, because the new file must not
necessarily be formatted the same as the original file is. The command works
by reading both files, parsing them as JSON, writing them formatted to memory,
and then comparing the formatted values, printing a colored diff.`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg.tasksOrgFile = args[0]
			cfg.newFile = args[1]

			run(cmd.Context(), func() error { return diffRun(cfg) })
		},
	}
}

func diffRun(cfg diffConfig) error {
	readFormattedJSON := func(file string) (string, error) {
		data, err := os.ReadFile(file)
		if err != nil {
			return "", fmt.Errorf("read file: %w", err)
		}
		var value map[string]any
		if err := json.Unmarshal(data, &value); err != nil {
			return "", fmt.Errorf("unmarshal json: %w", err)
		}
		formatted, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return "", fmt.Errorf("marshal indent json: %w", err)
		}
		return string(formatted), nil
	}

	original, err := readFormattedJSON(cfg.tasksOrgFile)
	if err != nil {
		return fmt.Errorf("read formatted json for %v: %w", cfg.tasksOrgFile, err)
	}

	newFile, err := readFormattedJSON(cfg.newFile)
	if err != nil {
		return fmt.Errorf("read formatted json for %v: %w", cfg.newFile, err)
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(original, newFile, false)
	fmt.Println(dmp.DiffPrettyText(diffs))

	return nil
}
