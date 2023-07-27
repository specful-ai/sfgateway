gateway

Create `gateway/main.go` and implement an HTTP server.
The server serves as a gateway / reverse proxy of an actual backend service.
There is only one handler to handle `/` and any path is accepted.
There is a command-line flag `-backend` which is the address of the backend service.
By default it should be `"https://api.openai.com"`.
When it receives a request, it must read all the incoming headers and the request body.
Then it makes an outgoing HTTP request to the backend, using the same headers and the body.
For example, if a request is received on `/whatever`, it should make a request to `<address_of_backend>/whatever`.
After it receives the response from the backend, it must write the response back to its client.
