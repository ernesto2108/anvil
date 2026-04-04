# Post-Implementation Gate

After ANY code change to `.go` files, invoke the `/lint` skill before considering the task done. The lint skill runs both `golangci-lint` and `go test` and blocks on new violations. See the lint skill for details.
