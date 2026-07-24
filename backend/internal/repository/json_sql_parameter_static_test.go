package repository

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestJSONBuildObjectSQLPlaceholdersAreExplicitlyTyped(t *testing.T) {
	root := filepath.Clean("..")
	var findings []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "json_sql_parameter_static_test.go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for lineNo, line := range strings.Split(string(content), "\n") {
			lower := strings.ToLower(line)
			if !strings.Contains(lower, "jsonb_build_object") && !strings.Contains(lower, "json_build_object") {
				continue
			}
			if placeholders := untypedDirectJSONPlaceholders(line); len(placeholders) > 0 {
				findings = append(findings, filepath.ToSlash(path)+":"+strconv.Itoa(lineNo+1)+" uses "+strings.Join(placeholders, ", "))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) > 0 {
		t.Fatalf("json/jsonb_build_object SQL placeholders used as direct key/value arguments must be explicitly cast:\n%s", strings.Join(findings, "\n"))
	}
}

func TestUntypedDirectJSONPlaceholders(t *testing.T) {
	tests := []struct {
		name string
		line string
		want []string
	}{
		{
			name: "direct value placeholder must be cast",
			line: "jsonb_build_object('entry_type', $6)",
			want: []string{"$6"},
		},
		{
			name: "typed direct placeholders and inferred expressions are allowed",
			line: "jsonb_build_object($2::text, COALESCE(extra, '{}'::jsonb) -> ($2::text), 'quota_used', COALESCE((extra->>'quota_used')::numeric, 0) + $1)",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := untypedDirectJSONPlaceholders(tt.line)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func untypedDirectJSONPlaceholders(line string) []string {
	var placeholders []string
	lower := strings.ToLower(line)
	for _, fn := range []string{"jsonb_build_object", "json_build_object"} {
		for offset := 0; ; {
			idx := strings.Index(lower[offset:], fn)
			if idx < 0 {
				break
			}
			start := offset + idx + len(fn)
			open := strings.IndexByte(line[start:], '(')
			if open < 0 {
				break
			}
			open += start
			close := closingParen(line, open)
			if close < 0 {
				break
			}
			placeholders = append(placeholders, untypedDirectJSONPlaceholdersInArgs(line[open+1:close])...)
			offset = close + 1
		}
	}
	return placeholders
}

func untypedDirectJSONPlaceholdersInArgs(args string) []string {
	var placeholders []string
	for i := 0; i < len(args); i++ {
		if args[i] != '$' {
			continue
		}
		j := i + 1
		for j < len(args) && args[j] >= '0' && args[j] <= '9' {
			j++
		}
		if j == i+1 {
			continue
		}

		prev := previousNonSpace(args, i)
		if prev != '(' && prev != ',' {
			continue
		}
		next := strings.TrimSpace(args[j:])
		if strings.HasPrefix(next, "::") {
			continue
		}
		placeholders = append(placeholders, args[i:j])
	}
	return placeholders
}

func closingParen(s string, open int) int {
	depth := 0
	inSingleQuote := false
	for i := open; i < len(s); i++ {
		switch s[i] {
		case '\'':
			inSingleQuote = !inSingleQuote
		case '(':
			if !inSingleQuote {
				depth++
			}
		case ')':
			if !inSingleQuote {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}
	return -1
}

func previousNonSpace(s string, before int) byte {
	for i := before - 1; i >= 0; i-- {
		if s[i] != ' ' && s[i] != '\t' {
			return s[i]
		}
	}
	return 0
}
