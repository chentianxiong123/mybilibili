module mybilibili/live

go 1.26.5

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/lib/pq v1.12.3
	github.com/stretchr/testify v1.11.1
	golang.org/x/net v0.55.0
	mybilibili/pkg v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace mybilibili/pkg => ../../shared/pkg
