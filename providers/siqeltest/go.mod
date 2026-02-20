module github.com/wrapped-owls/testereiro/providers/siqeltest

go 1.25

replace github.com/wrapped-owls/testereiro/puppetest => ../../puppetest

require (
	github.com/vinovest/sqlx v1.7.1
	github.com/wrapped-owls/testereiro/puppetest v0.0.0-00010101000000-000000000000
)

require github.com/muir/sqltoken v0.2.1 // indirect
