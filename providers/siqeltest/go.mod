module github.com/wrapped-owls/testereiro/providers/siqeltest

go 1.25

replace github.com/wrapped-owls/testereiro/puppetest => ../../puppetest

require (
	github.com/stretchr/testify v1.11.1
	github.com/vinovest/sqlx v1.7.1
	github.com/wrapped-owls/testereiro/puppetest v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/muir/sqltoken v0.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
