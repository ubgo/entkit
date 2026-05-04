module github.com/ubgo/entkit/examples

go 1.25.0

require (
	entgo.io/ent v0.14.6
	github.com/mattn/go-sqlite3 v1.14.44
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

require (
	ariga.io/atlas v0.36.2-0.20250730182955-2c6300d0a3e1 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/hcl/v2 v2.18.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/ubgo/crypt v0.0.0-20260504095124-aefc1abb0446 // indirect
	github.com/zclconf/go-cty v1.14.4 // indirect
	github.com/zclconf/go-cty-yaml v1.1.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/mod v0.34.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/ubgo/entkit/jsonmap => ../jsonmap

replace github.com/ubgo/entkit/passwordtype => ../passwordtype

replace github.com/ubgo/entkit/encryptedtype => ../encryptedtype

replace github.com/ubgo/entkit/ent_jsontype => ../ent_jsontype

replace github.com/ubgo/entkit/ent_jsonslice => ../ent_jsonslice

replace github.com/ubgo/entkit/ent_jsonmap => ../ent_jsonmap

replace github.com/ubgo/entkit/ent_passwordtype => ../ent_passwordtype

replace github.com/ubgo/entkit/ent_encryptedtype => ../ent_encryptedtype
