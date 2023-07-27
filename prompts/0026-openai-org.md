add -openai_org flag

Change `gateway/main.go` to add a command-line flag `-openai_org` and
use its value to add a header `OpenAI-Organization: ...` to
the headers copied from the client request to the backend request.
