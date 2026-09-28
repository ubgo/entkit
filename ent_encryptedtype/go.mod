module github.com/ubgo/entkit/ent_encryptedtype

go 1.25.0

require (
	entgo.io/ent v0.14.5
	github.com/ubgo/entkit/encryptedtype v0.0.0
)

require (
	github.com/ubgo/crypt v0.2.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace github.com/ubgo/entkit/encryptedtype => ../encryptedtype
