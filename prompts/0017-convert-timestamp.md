convert timestamp to current timezone

Change `gateway/main.go` so that when handling `/_list`, convert the timestamp to the current timezone before formatting it.
