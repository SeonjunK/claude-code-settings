// Package storage provides file storage clients.
package storage

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// SessionFileClient handles session file I/O.
type SessionFileClient struct{}

// NewSessionFileClient creates a new session file client.
func NewSessionFileClient() *SessionFileClient {
	return &SessionFileClient{}
}

// ReadFrontmatter reads YAML frontmatter from a markdown file.
func (c *SessionFileClient) ReadFrontmatter(path string) (map[string]any, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	return ParseFrontmatter(data)
}

// ParseFrontmatter extracts YAML frontmatter and body from content.
func ParseFrontmatter(data []byte) (map[string]any, string, error) {
	content := string(data)

	// Check for frontmatter markers
	if !strings.HasPrefix(content, "---\n") {
		return nil, content, nil
	}

	// Find end of frontmatter
	endIndex := strings.Index(content[4:], "\n---\n")
	if endIndex == -1 {
		return nil, content, nil
	}

	frontmatterStr := content[4 : 4+endIndex]
	body := content[4+endIndex+5:]

	var frontmatter map[string]any
	if err := yaml.Unmarshal([]byte(frontmatterStr), &frontmatter); err != nil {
		return nil, "", err
	}

	return frontmatter, body, nil
}

// WriteFrontmatter writes YAML frontmatter to a markdown file.
func (c *SessionFileClient) WriteFrontmatter(path string, frontmatter map[string]any, body string) error {
	var buf bytes.Buffer
	buf.WriteString("---\n")

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(frontmatter); err != nil {
		return err
	}

	buf.WriteString("---\n")
	buf.WriteString(body)

	return os.WriteFile(path, buf.Bytes(), 0644)
}

// Exists checks if a file exists.
func (c *SessionFileClient) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// EnsureDir ensures a directory exists.
func (c *SessionFileClient) EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// ListSessions lists all session files in a directory.
func (c *SessionFileClient) ListSessions(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".local.md") {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// ReadStdin reads all input from stdin.
func ReadStdin() ([]byte, error) {
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice != 0 {
		// TTY - read line by line
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			return scanner.Bytes(), nil
		}
		return nil, scanner.Err()
	}

	// Pipe - read all
	return os.ReadFile("/dev/stdin")
}
