// Package encryptedtypeent provides an ent.Field constructor for
// github.com/ubgo/entkit/encryptedtype.EncryptedString columns.
//
// IMPORTANT: callers must wire the encryption key at process boot via
// encryptedtype.SetKey before any Value/Scan happens. See the
// encryptedtype README.
package entencryptedtype

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"

	"github.com/ubgo/entkit/encryptedtype"
)

// Field returns an ent.Field for an EncryptedString column.
//
//   - Stored as TEXT (the AES-GCM ciphertext is base64url-encoded and
//     longer than the plaintext).
//   - Sensitive() set so the value is masked in ent's debug/print output.
//
// For Optional, Immutable, dialect-specific column type, etc., compose
// directly:
//
//	field.String("api_secret").
//	    GoType(encryptedtype.EncryptedString{}).
//	    Sensitive().
//	    Optional()
func Field(name string) ent.Field {
	return field.String(name).
		GoType(encryptedtype.EncryptedString{}).
		Sensitive()
}
