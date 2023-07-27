# sfgateway

sfgateway serves as an API gateway and reverse proxy for OpenAI APIs, designed
to facilitate debugging, accounting, auditing, and monitoring when utilizing LLM
APIs in both development and production settings.

Current features include:

- API access for the internal network, without revealing the actual API key
- Persistence of requests and responses, serving debugging and auditing purposes
- Web-based inspection of request and response fields and nested structures

We have a roadmap of additional features we plan to incorporate:

- Integration with internal Single Sign-On (SSO) systems
- Statistics including byte/token sizes, latencies, and error rates
- Expanded support for more LLM backends

We welcome your contributions to add features that you find useful.

## Usage

Build:

    make

It produces an executable `sfgateway`, which takes a few command-line flags:

```
$ ./sfgateway --help
Usage of ./sfgateway:
  -api_key string
        API key
  -backend string
        address of the backend service (default "https://api.openai.com/v1")
  -db_file string
        path to the requests.db file (default "./requests.db")
  -listen_on string
        address to listen on (default ":8090")
  -openai_org string
        OpenAI Organization
```

All flags are optional and self-explanatory.

You can just run:

    ./sfgateway

And it's going to get the API key from the environment variable
`OPENAI_API_KEY`.

Alternatively, you can specify the API key with the `-api_key` flag:

    ./sfgateway -api_key "sk-hKsxS7FEbVx3iFJWh3nxRqMA1byLnVQT3B29QtyCm3iZflbk"

By default it listens on `*:8090` and you can open
[http://localhost:8090/\_list](http://localhost:8090/_list)
to see the list of requests.

To use sfgateway when using the official Python library:

```py
openai.api_base = "http://localhost:8090"
```

## Screenshots

![1](https://github.com/specful-ai/sfgateway/assets/196279/5a4c6bba-e938-4621-bf77-1866da808648)

![2](https://github.com/specful-ai/sfgateway/assets/196279/bf40fb65-d9d2-4aad-bcd2-57da80cb8115)
