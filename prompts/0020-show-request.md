show detailed request

Change `gateway/show.go` to parse the request column as a JSON and render its details as a nested description list.

The request column looks like this:

```json
{
  "model": "gpt-3.5-turbo-16k",
  "messages": [
    { "role": "user", "content": "..." },
    { "role": "assistant", "content": "..." }
  ],
  "temperature": 0
}
```
