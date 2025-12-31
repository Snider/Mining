# Swagger Documentation

Interactive API documentation is available when running the Mining service.

## Accessing Swagger UI

When the service is running, you can access the Swagger UI at:

```
http://localhost:9090/api/v1/mining/swagger/index.html
```

## Starting the Service

=== "CLI"

    ```bash
    miner-ctrl serve
    ```

=== "Development"

    ```bash
    make dev
    ```

## OpenAPI Specification

The OpenAPI 3.0 specification is available at:

- **JSON**: `http://localhost:9090/api/v1/mining/swagger/doc.json`
- **YAML**: `http://localhost:9090/api/v1/mining/swagger/doc.yaml`

## Regenerating Docs

After modifying API endpoints, regenerate the Swagger documentation:

```bash
make docs
```

This runs `swag init` to parse the Go annotations and update the specification files.

## API Annotations

API documentation is generated from Go comments using [swaggo/swag](https://github.com/swaggo/swag). Example:

```go
// StartMiner godoc
// @Summary Start a miner
// @Description Start mining with a specific profile
// @Tags miners
// @Accept json
// @Produce json
// @Param profile_id path string true "Profile ID"
// @Success 200 {object} MinerResponse
// @Failure 400 {object} ErrorResponse
// @Router /miners/{profile_id}/start [post]
func (s *Service) StartMiner(c *gin.Context) {
    // ...
}
```
