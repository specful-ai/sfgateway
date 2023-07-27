show detailed response

Change `gateway/show.go` to parse the response column as a JSON and render its details as a nested description list.

The response column looks like this:

```json
{
  "id": "chatcmpl-7gib6PXeGPqHLRDvbYoeZSDShvHxY",
  "object": "chat.completion",
  "created": 1690415564,
  "model": "gpt-3.5-turbo-16k-0613",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "..."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 1274,
    "completion_tokens": 839,
    "total_tokens": 2113
  }
}
```
