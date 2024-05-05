module github.com/pqabelian/abelian-sdk-go

go 1.18

require (
	github.com/abesuite/abec v0.0.0-00010101000000-000000000000
	github.com/abesuite/abeutil v0.0.0-20231107022913-d6d3bf295938
)

require golang.org/x/crypto v0.1.0

replace github.com/abesuite/abec => github.com/pqabelian/abec v0.0.0-20231206045108-7db3092bc81c

replace github.com/abesuite/abeutil => github.com/pqabelian/abeutil v0.0.0-20231107022913-d6d3bf295938

replace github.com/cryptosuite/pqringct => github.com/pqabelian/pqringct v0.0.0-20231107022351-feb587470e43
