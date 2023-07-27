handle no results

Change `gateway/main.go` to handle the no results case.
When handling `/_list`, if there are no requests in the database, show `No results` in the HTML response.
