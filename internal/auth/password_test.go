package auth

import "testing"

func TestPasswordWorkflow(t *testing.T) {
	gesla := []string{"geslo123", "xhrllo123-23", "passadmin3"}
	for _, geslo := range gesla {
		gesloHash, err := HashPassword(geslo)
		if err != nil {
			t.Errorf("Password hashing failed: %v", err)
		}
		skladnost, err := CheckPasswordHash(geslo, gesloHash)
		if err != nil {
			t.Errorf("CheckPasswordHash errored: %v", err)
		}
		if !skladnost {
			t.Errorf("Gesli se ne ujemata %v", skladnost)
		}
	}
}
