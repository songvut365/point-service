# Point Service

## Architecture
![architecture](docs/arch.jpg)

## Unit Test
You can run the tests using the following command:
```
go test ./... -cover
```

After running tests, you will see a coverage percentage, indicating the proportion of statements covered by the tests. In the example output:

```
ok      point-service/app/internal/handler      0.282s  coverage: 100.0% of statements
ok      point-service/app/internal/repository   0.244s  coverage: 97.2% of statements
ok      point-service/app/internal/service      0.250s  coverage: 100.0% of statements
```