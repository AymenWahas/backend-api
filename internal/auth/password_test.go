package auth
//gofmt -w internal/auth/password_test.go
//go test ./internal/auth -v
import "testing"

func TestPasswordHashing(t *testing.T) {
	password := "MySecret123"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == password {
		t.Fatal("password must not be stored as plaintext")
	}

	if err := CheckPassword(password, hash); err != nil {
		t.Fatalf("CheckPassword() failed for correct password: %v", err)
	}

	if err := CheckPassword("WrongPassword", hash); err == nil {
		t.Fatal("CheckPassword() succeeded for incorrect password")
	}
}
