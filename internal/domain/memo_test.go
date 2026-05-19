package domain

import "testing"

func TestGetNameParentTokens(t *testing.T) {
	tokens, err := GetNameParentTokens("users/12/memos/34", "users/", "memos/")
	if err != nil {
		t.Fatalf("GetNameParentTokens returned error: %v", err)
	}
	if len(tokens) != 2 || tokens[0] != "12" || tokens[1] != "34" {
		t.Fatalf("unexpected tokens: %v", tokens)
	}
}

func TestGetNameParentTokensRejectsInvalidInput(t *testing.T) {
	cases := []string{
		"users/12/memos",
		"wrong/12",
		"users/",
	}

	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			if _, err := GetNameParentTokens(input, "users/"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestExtractMemoUIDFromName(t *testing.T) {
	got, err := ExtractMemoUIDFromName("memos/99")
	if err != nil {
		t.Fatalf("ExtractMemoUIDFromName returned error: %v", err)
	}
	if got != "99" {
		t.Fatalf("expected uid 99, got %q", got)
	}
}
