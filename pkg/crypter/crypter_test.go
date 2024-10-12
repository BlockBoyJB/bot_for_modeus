package crypter

import (
	"encoding/base64"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkCrypter_Encrypt(b *testing.B) {
	c := NewCrypter("hello_world")
	text := "some text for encrypt. Foo bar hello world abc"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := c.Encrypt(text)
		if err != nil {
			b.Fatalf("encryption error: %s", err)
		}
	}
}

func BenchmarkCrypter_Decrypt(b *testing.B) {
	c := NewCrypter("hello_world")
	defaultText := "some text for encrypt. Foo bar hello world abc"

	benchText, err := c.Encrypt(defaultText)
	if err != nil {
		b.Fatalf("encrypt error: %s", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		actualText, err := c.Decrypt(benchText)
		if err != nil {
			b.Fatalf("decrypt error: %s", err)
		}
		if actualText != defaultText {
			b.Errorf("dectypted text is not equal to default, expect %s, got: %s", defaultText, actualText)
		}
	}
}

func TestCrypter_Encrypt(t *testing.T) {
	c := NewCrypter("someSecretFoobar")

	testCases := []struct {
		testName  string
		text      string
		expectErr error
	}{
		{
			testName:  "correct test",
			text:      "some text",
			expectErr: nil,
		},
		{
			testName:  "empty text input",
			text:      "",
			expectErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			cipher, err := c.Encrypt(tc.text)

			assert.Equal(t, tc.expectErr, err)
			if tc.expectErr == nil {
				assert.NotEmpty(t, cipher)
			}
		})
	}
}

func TestCrypter_Decrypt(t *testing.T) {
	c := NewCrypter("someSecretFoobar")

	defaultCipher, err := c.Encrypt("defaultText")
	if err != nil {
		t.Fatalf("setup test error: %s", err)
	}

	testCases := []struct {
		testName     string
		text         string
		expectErr    error
		expectOutput string
	}{
		{
			testName:     "correct test",
			text:         defaultCipher,
			expectErr:    nil,
			expectOutput: "defaultText",
		},
		{
			testName:     "empty text input",
			text:         "",
			expectErr:    errors.New("cipher text too short"),
			expectOutput: "",
		},
		{
			testName:     "not a cipher",
			text:         defaultCipher[:defaultSaltBlockSize+1],
			expectErr:    base64.CorruptInputError(16),
			expectOutput: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.testName, func(t *testing.T) {
			decrypted, err := c.Decrypt(tc.text)

			assert.Equal(t, tc.expectErr, err)
			assert.Equal(t, tc.expectOutput, decrypted)
		})
	}
}
