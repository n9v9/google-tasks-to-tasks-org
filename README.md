# Google Tasks to Tasks.org

## What

Convert tasks stored in Google Tasks to tasks that can be imported into the Tasks.org app.

## How

1. Export your Google Tasks data via Google Takeout. Once downloaded, your export should contain a
   file called `Tasks.json`, from now on called `GOOGLE_TASKS.json`
2. Make a backup from your Tasks.org app. Once done, you should have a JSON file which we'll call
   `TASKS_ORG.json`.
3. Open `TASKS_ORG.json` to find out the UUID of the calendar you want your Google tasks to be
   imported into. The UUIDs can be obtained via the JSONPath `$.data.caldavCalendars[*].uuid`.
4. Run the tool:
   ```sh
   go run github.com/n9v9/google-tasks-to-tasks-org@latest convert \
      --calendar-uuid ... \
      --out converted.json \
      GOOGLE_TASKS.json TASKS_ORG.json
   ```

This will create the file `converted.json` which has the same structure as `TASKS_ORG.json` but in
addition contains the converted Google Tasks tasks.

To see the difference between your original and untouched `TASKS_ORG.json` file, and the newly
`converted.json` file, run:

```sh
go run github.com/n9v9/google-tasks-to-tasks-org@latest diff TASKS_ORG.json converted.json
```

If everything looks alright, import the `converted.json` via the "import backup" functionality into
your Tasks.org app.
