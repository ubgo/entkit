module github.com/ubgo/entkit/examples

go 1.25.0

require (
	entgo.io/ent v0.14.5
	github.com/ubgo/entkit/encryptedtype v0.0.0
	github.com/ubgo/entkit/ent_encryptedtype v0.0.0
	github.com/ubgo/entkit/ent_jsonmap v0.0.0
	github.com/ubgo/entkit/ent_jsonslice v0.0.0
	github.com/ubgo/entkit/ent_jsontype v0.0.0
	github.com/ubgo/entkit/ent_passwordtype v0.0.0
	github.com/ubgo/entkit/jsonmap v0.0.0
	github.com/ubgo/entkit/passwordtype v0.0.0
	github.com/ubgo/jsonslice v0.1.0
	github.com/ubgo/jsontype v0.1.1-0.20260501195444-6a263de9525b
)

replace github.com/ubgo/entkit/jsonmap => ../jsonmap
replace github.com/ubgo/entkit/passwordtype => ../passwordtype
replace github.com/ubgo/entkit/encryptedtype => ../encryptedtype
replace github.com/ubgo/entkit/ent_jsontype => ../ent_jsontype
replace github.com/ubgo/entkit/ent_jsonslice => ../ent_jsonslice
replace github.com/ubgo/entkit/ent_jsonmap => ../ent_jsonmap
replace github.com/ubgo/entkit/ent_passwordtype => ../ent_passwordtype
replace github.com/ubgo/entkit/ent_encryptedtype => ../ent_encryptedtype
