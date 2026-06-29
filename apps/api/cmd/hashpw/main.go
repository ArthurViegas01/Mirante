// Command hashpw prints an Argon2id hash for a password so OWNER_PASSWORD_HASH
// can be seeded on the first deploy: the owner account is then created at boot
// and the first-run signup window never opens (F3). Run it offline and set the
// output as a secret; never commit the hash.
//
//	go run ./cmd/hashpw                 # prompts and reads the password from stdin
//	printf 's3cret' | go run ./cmd/hashpw
//
// Then: railway variables set OWNER_EMAIL=you@example.com OWNER_PASSWORD_HASH='<output>'
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lumni/mirante/internal/platform/auth"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "hashpw:", err)
		os.Exit(1)
	}
}

func run() error {
	fmt.Fprint(os.Stderr, "password: ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("read password: %w", err)
	}
	password := strings.TrimRight(line, "\r\n")
	if password == "" {
		return errors.New("empty password")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	fmt.Println(hash)
	return nil
}
