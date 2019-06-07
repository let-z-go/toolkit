package uuid

import (
	"testing"
)

func TestGenerateUUID4(t *testing.T) {
	for i := 0; i < 1024; i++ {
		u, e := GenerateUUID4()

		if e != nil {
			t.Fatalf("%v", e)
		}

		s := u.String()

		if c := s[14]; c != '4' {
			t.Errorf("%v", s)
		}

		if c := s[19]; c != '8' && c != '9' && c != 'a' && c != 'b' {
			t.Errorf("%v", s)
		}
	}
}

func TestGenerateUUID4Fast(t *testing.T) {
	for i := 0; i < 1024; i++ {
		u, e := GenerateUUID4Fast()

		if e != nil {
			t.Fatalf("%v", e)
		}

		s := u.String()

		if c := s[14]; c != '4' {
			t.Errorf("%v", s)
		}

		if c := s[19]; c != '8' && c != '9' && c != 'a' && c != 'b' {
			t.Errorf("%v", s)
		}
	}
}
