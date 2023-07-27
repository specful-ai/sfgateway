fix temperature type

Fix the following error when running `gateway/show.go`:

    Failed to unmarshal request: json: cannot unmarshal number 0.7 into Go struct field Request.temperature of type int
