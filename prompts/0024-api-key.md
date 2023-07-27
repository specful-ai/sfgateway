use api key

Change `gateway/main.go` to read the environment variable `OPENAI_API_KEY` and
use its value to add a header `Authorization: Bearer $OPENAI_API_KEY` to
the headers copied from the client request to the backend request.
