package helmstrip

import (
	"bytes"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestUnquote(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantQuoted []string // these values should remain quoted in output
		wantPlain  []string // these values should appear unquoted in output
	}{
		{
			name: "plain strings get unquoted",
			input: `key1: "hello"
key2: "some-string"
key3: "my-app.example.com"
`,
			wantPlain: []string{
				"key1: hello",
				"key2: some-string",
				"key3: my-app.example.com",
			},
		},
		{
			name: "booleans stay quoted",
			input: `key1: "true"
key2: "false"
key3: "yes"
key4: "no"
key5: "on"
key6: "off"
key7: "True"
key8: "FALSE"
key9: "Yes"
key10: "NO"
`,
			wantQuoted: []string{
				`"true"`, `"false"`, `"yes"`, `"no"`, `"on"`, `"off"`,
				`"True"`, `"FALSE"`, `"Yes"`, `"NO"`,
			},
		},
		{
			name: "nulls stay quoted",
			input: `key1: "null"
key2: "~"
key3: "Null"
key4: "NULL"
`,
			wantQuoted: []string{`"null"`, `"~"`, `"Null"`, `"NULL"`},
		},
		{
			name: "integers stay quoted",
			input: `key1: "123"
key2: "0"
key3: "-42"
key4: "0x1F"
key5: "0o17"
key6: "0b1010"
`,
			wantQuoted: []string{`"123"`, `"0"`, `"-42"`, `"0x1F"`, `"0o17"`, `"0b1010"`},
		},
		{
			name: "floats stay quoted",
			input: `key1: "1.5"
key2: "-3.14"
key3: "1e10"
key4: "1.5e-3"
key5: ".inf"
key6: "-.inf"
key7: ".nan"
key8: ".Inf"
key9: ".NaN"
`,
			wantQuoted: []string{
				`"1.5"`, `"-3.14"`, `"1e10"`, `"1.5e-3"`,
				`".inf"`, `"-.inf"`, `".nan"`, `".Inf"`, `".NaN"`,
			},
		},
		{
			name: "special leading characters stay quoted",
			input: `key1: ":8080"
key2: "- item"
key3: "# comment"
key4: "{foo}"
key5: "[bar]"
key6: "*anchor"
key7: "&anchor"
key8: "!tag"
key9: "%directive"
key10: "@mention"
key11: "? question"
key12: "> folded"
key13: "| literal"
`,
			wantQuoted: []string{
				`":8080"`, `"- item"`, `"# comment"`,
				`"{foo}"`, `"[bar]"`, `"*anchor"`, `"&anchor"`,
				`"!tag"`, `"%directive"`, `"@mention"`, `"? question"`,
				`"> folded"`, `"| literal"`,
			},
		},
		{
			name: "mid-string special sequences stay quoted",
			input: `key1: "foo: bar"
key2: "foo # comment"
key3: "hello, world"
`,
			wantQuoted: []string{`"foo: bar"`, `"foo # comment"`, `"hello, world"`},
		},
		{
			name: "empty string stays quoted",
			input: `key1: ""
`,
			wantQuoted: []string{`""`},
		},
		{
			name: "strings with whitespace stay quoted",
			input: `key1: " leading"
key2: "trailing "
key3: "has  double  spaces"
`,
			wantQuoted: []string{`" leading"`, `"trailing "`, `"has  double  spaces"`},
		},
		{
			name: "single-quoted strings also get unquoted when safe",
			input: `key1: 'hello'
key2: 'simple-string'
`,
			wantPlain: []string{
				"key1: hello",
				"key2: simple-string",
			},
		},
		{
			name: "single-quoted booleans stay quoted",
			input: `key1: 'true'
key2: 'false'
`,
			wantQuoted: []string{`'true'`, `'false'`},
		},
		{
			name: "empty string in sequence stays quoted",
			input: `apiGroups:
- ""
- apps
`,
			wantQuoted: []string{`""`},
			wantPlain:  []string{"- apps"},
		},
		{
			name: "already plain strings are left alone",
			input: `key1: hello
key2: some-string
`,
			wantPlain: []string{
				"key1: hello",
				"key2: some-string",
			},
		},
		{
			name: "strings with newlines stay quoted",
			input: `key1: "line1\nline2"
`,
			wantQuoted: []string{`"line1\nline2"`},
		},
		{
			name: "timestamps stay quoted",
			input: `key1: "2024-01-01"
key2: "2024-01-01T00:00:00Z"
`,
			wantQuoted: []string{`"2024-01-01"`, `"2024-01-01T00:00:00Z"`},
		},
		{
			name: "mixed safe and unsafe in same document",
			input: `safe1: "hello"
unsafe1: "true"
safe2: "world"
unsafe2: "123"
safe3: "my-app"
unsafe3: ":8080"
`,
			wantPlain:  []string{"safe1: hello", "safe2: world", "safe3: my-app"},
			wantQuoted: []string{`"true"`, `"123"`, `":8080"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var doc yaml.Node
			if err := yaml.NewDecoder(strings.NewReader(tt.input)).Decode(&doc); err != nil {
				t.Fatalf("decode: %v", err)
			}

			UnquoteScalars(&doc)

			var buf bytes.Buffer
			enc := yaml.NewEncoder(&buf)
			enc.SetIndent(2)
			if err := enc.Encode(&doc); err != nil {
				t.Fatalf("encode: %v", err)
			}
			enc.Close()
			out := buf.String()

			for _, want := range tt.wantQuoted {
				if !strings.Contains(out, want) {
					t.Errorf("expected %s to remain quoted in output:\n%s", want, out)
				}
			}
			for _, want := range tt.wantPlain {
				if !strings.Contains(out, want) {
					t.Errorf("expected %q to appear unquoted in output:\n%s", want, out)
				}
			}
		})
	}
}
