// demo opens an in-memory SQLite database, runs the ent schema migration,
// creates a User row exercising every entkit column type, reads it back,
// and verifies the round-trip. No external services required.
//
//   go run ./cmd/demo
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"

	"github.com/ubgo/entkit/encryptedtype"
	"github.com/ubgo/entkit/examples/ent"
	"github.com/ubgo/entkit/examples/ent/user"
	"github.com/ubgo/entkit/jsonmap"
	"github.com/ubgo/entkit/passwordtype"
	"github.com/ubgo/jsonslice"
	"github.com/ubgo/jsontype"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// 1. Wire encryptedtype's AES key once at boot. The key must be
	//    16, 24, or 32 bytes. In real apps load from env or KMS.
	if err := encryptedtype.SetKey([]byte("0123456789abcdef0123456789abcdef")); err != nil {
		return fmt.Errorf("encryptedtype.SetKey: %w", err)
	}

	// 2. Open an in-memory SQLite — no external DB, perfect for examples.
	client, err := ent.Open(dialect.SQLite, "file:demo.db?mode=memory&cache=shared&_fk=1")
	if err != nil {
		return fmt.Errorf("ent.Open: %w", err)
	}
	defer client.Close()

	// 3. Run the auto-migration to create the users table.
	if err := client.Schema.Create(ctx); err != nil {
		return fmt.Errorf("schema.Create: %w", err)
	}
	fmt.Println("✓ schema migrated")

	// 4. Build a HashedPassword from plaintext via passwordtype.New.
	pw, err := passwordtype.New("hunter2")
	if err != nil {
		return fmt.Errorf("passwordtype.New: %w", err)
	}

	// 5. Create a User with every entkit column type populated.
	created, err := client.User.Create().
		SetEmail("alice@example.com").
		SetMetadata(jsonmap.JsonMap{
			"stripe_id": "cus_alice",
			"tier":      float64(2),
		}).
		SetTags(jsonslice.JsonSlice{"admin", "billing", float64(42)}).
		SetProfile(jsontype.JSON(`{"bio":"hello world","theme":"dark"}`)).
		SetPassword(pw).
		SetAPISecret(encryptedtype.New("sk_test_xyz123abc")).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	fmt.Printf("✓ created user id=%d email=%s\n", created.ID, created.Email)

	// 6. Read the user back through ent — round-trips through Value/Scan
	//    and re-hydrates each column type from its stored representation.
	loaded, err := client.User.Query().
		Where(user.EmailEQ("alice@example.com")).
		Only(ctx)
	if err != nil {
		return fmt.Errorf("query user: %w", err)
	}
	fmt.Printf("✓ loaded user id=%d\n", loaded.ID)

	// 7. Verify each column type round-tripped correctly.
	checkpoint("email", loaded.Email == "alice@example.com")
	checkpoint("metadata.tier == 2", loaded.Metadata["tier"] == float64(2))
	checkpoint("tags[0] == admin", loaded.Tags[0] == "admin")
	checkpoint("profile contains bio", string(loaded.Profile) != "" && containsKey(loaded.Profile, "bio"))

	// 8. Demonstrate the password lifecycle: Verify the original plaintext
	//    succeeds, a wrong plaintext fails, and the hash never appears in
	//    fmt or JSON output.
	checkpoint("password verifies (hunter2)", loaded.Password.Verify("hunter2"))
	checkpoint("wrong password rejected", !loaded.Password.Verify("password123"))
	checkpoint("password.String() redacted", fmt.Sprint(loaded.Password) == "[redacted]")
	pwJSON, _ := json.Marshal(loaded.Password)
	checkpoint("password JSON is null", string(pwJSON) == "null")

	// 9. Demonstrate encrypted-string lifecycle: Plain() returns the
	//    original plaintext, fmt prints "[encrypted]", and JSON is null.
	checkpoint("api_secret decrypted", loaded.APISecret.Plain() == "sk_test_xyz123abc")
	checkpoint("api_secret.String() redacted", fmt.Sprint(loaded.APISecret) == "[encrypted]")
	asJSON, _ := json.Marshal(loaded.APISecret)
	checkpoint("api_secret JSON is null", string(asJSON) == "null")

	// 10. Update path — change the password, verify ent persists it
	//     correctly, and the old password no longer verifies.
	newPw, _ := passwordtype.New("rosebud")
	updated, err := client.User.UpdateOne(loaded).
		SetPassword(newPw).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	checkpoint("new password verifies (rosebud)", updated.Password.Verify("rosebud"))
	checkpoint("old password rejected after change", !updated.Password.Verify("hunter2"))

	fmt.Println()
	if anyFailure {
		fmt.Fprintln(os.Stderr, "some assertions failed — see above")
		os.Exit(1)
	}
	fmt.Println("=== all checkpoints passed — entkit family works end-to-end through ent + SQLite ===")
	return nil
}

var anyFailure bool

func checkpoint(name string, ok bool) {
	mark := "✓"
	if !ok {
		mark = "✗"
		anyFailure = true
	}
	fmt.Printf("  %s %s\n", mark, name)
}

func containsKey(j jsontype.JSON, key string) bool {
	var m map[string]any
	if err := json.Unmarshal(j, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
}
