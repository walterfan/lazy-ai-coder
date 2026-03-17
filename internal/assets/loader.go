package assets

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

const (
	TypeCommand = "command"
	TypeRule    = "rule"
	TypeSkill   = "skill"
)

// Item represents a single asset file (command, rule, or skill)
type Item struct {
	Type     string `json:"type"`     // command, rule, skill
	Path     string `json:"path"`     // relative path under assets, e.g. commands/bug.fix.md
	Name     string `json:"name"`     // display name (filename or skill name)
	Snippet  string `json:"snippet"`  // first ~200 chars of content for preview
	Category string `json:"category"` // e.g. for rules: common, golang; for skills: awesome, anthropics
}

// Loader scans the assets directory for commands, rules, and skills
type Loader struct {
	root string // absolute path to assets directory
}

// NewLoader creates a loader with the given assets root (e.g. "assets" or "/app/assets")
func NewLoader(assetsRoot string) (*Loader, error) {
	if assetsRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		assetsRoot = filepath.Join(wd, "assets")
	}
	abs, err := filepath.Abs(assetsRoot)
	if err != nil {
		return nil, err
	}
	return &Loader{root: abs}, nil
}

// safePath ensures path is under loader.root and has no path traversal
func (l *Loader) safePath(rel string) (string, bool) {
	rel = filepath.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", false
	}
	abs := filepath.Join(l.root, rel)
	abs = filepath.Clean(abs)
	rootClean := filepath.Clean(l.root)
	if abs != rootClean && !strings.HasPrefix(abs, rootClean+string(os.PathSeparator)) {
		return "", false
	}
	return abs, true
}

// ListCommands returns all .md files under assets/commands
func (l *Loader) ListCommands() ([]Item, error) {
	dir := filepath.Join(l.root, "commands")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Item
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		rel := filepath.Join("commands", e.Name())
		item, err := l.itemFromFile(rel, TypeCommand, e.Name(), "")
		if err != nil {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// ListRules returns all .md and .mdc files under assets/rules (recursive)
func (l *Loader) ListRules() ([]Item, error) {
	root := filepath.Join(l.root, "rules")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}
	var out []Item
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		base := filepath.Base(path)
		if !strings.HasSuffix(strings.ToLower(base), ".md") && !strings.HasSuffix(strings.ToLower(base), ".mdc") {
			return nil
		}
		rel, err := filepath.Rel(l.root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		category := ""
		if len(parts) >= 2 {
			category = parts[1] // e.g. common, golang, frontend
		}
		item, err := l.itemFromFile(rel, TypeRule, base, category)
		if err != nil {
			return nil
		}
		out = append(out, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ListSkills returns all SKILL.md files under assets/skills (recursive)
func (l *Loader) ListSkills() ([]Item, error) {
	root := filepath.Join(l.root, "skills")
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return nil, nil
	}
	var out []Item
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) != "SKILL.md" {
			return nil
		}
		rel, err := filepath.Rel(l.root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		category := ""
		if len(parts) >= 2 {
			category = parts[1] // e.g. awesome, anthropics
		}
		skillName := "Skill"
		if len(parts) >= 3 {
			skillName = parts[len(parts)-2]
		}
		item, err := l.itemFromFile(rel, TypeSkill, skillName, category)
		if err != nil {
			return nil
		}
		out = append(out, item)
		return nil
	})
	return out, err
}

func (l *Loader) itemFromFile(rel, assetType, name, category string) (Item, error) {
	abs, ok := l.safePath(rel)
	if !ok {
		return Item{}, os.ErrInvalid
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return Item{}, err
	}
	content := string(data)
	snippet := firstN(content, 200)
	return Item{
		Type:     assetType,
		Path:     rel,
		Name:     name,
		Snippet:  snippet,
		Category: category,
	}, nil
}

func firstN(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return strings.TrimSpace(s)
	}
	runes := []rune(s)
	return strings.TrimSpace(string(runes[:n])) + "..."
}

// ReadFile returns the full content of an asset by relative path (e.g. commands/bug.fix.md)
func (l *Loader) ReadFile(relPath string) ([]byte, string, error) {
	relPath = filepath.ToSlash(filepath.Clean(relPath))
	abs, ok := l.safePath(relPath)
	if !ok {
		return nil, "", os.ErrPermission // path traversal or outside assets
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, "", err
	}
	base := filepath.Base(abs)
	return data, base, nil
}

// SkillFolderFromPath derives the skill folder relative path from a SKILL.md path.
// e.g. "skills/awesome/continuous-learning-v2/SKILL.md" -> "skills/awesome/continuous-learning-v2"
func SkillFolderFromPath(skillPath string) string {
	skillPath = filepath.ToSlash(filepath.Clean(skillPath))
	if strings.HasSuffix(skillPath, "/SKILL.md") {
		return strings.TrimSuffix(skillPath, "/SKILL.md")
	}
	return filepath.ToSlash(filepath.Dir(skillPath))
}

// ZipSkillFolder creates an in-memory zip of the entire skill directory.
// skillDir is the relative path to the skill folder under the assets root,
// e.g. "skills/awesome/continuous-learning-v2".
// Returns the zip bytes and a suggested filename like "continuous-learning-v2.zip".
func (l *Loader) ZipSkillFolder(skillDir string) ([]byte, string, error) {
	abs, ok := l.safePath(skillDir)
	if !ok {
		return nil, "", os.ErrPermission
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, "", err
	}
	if !info.IsDir() {
		return nil, "", os.ErrInvalid
	}

	zipName := filepath.Base(abs) + ".zip"

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	err = filepath.Walk(abs, func(path string, fi os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if fi.IsDir() {
			return nil
		}
		// skip hidden files
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		rel, err := filepath.Rel(abs, path)
		if err != nil {
			return err
		}
		// use the skill folder name as the zip root
		entryName := filepath.ToSlash(filepath.Join(filepath.Base(abs), rel))

		header, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}
		header.Name = entryName
		header.Method = zip.Deflate

		w, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})

	if err != nil {
		zw.Close()
		return nil, "", err
	}
	if err := zw.Close(); err != nil {
		return nil, "", err
	}

	return buf.Bytes(), zipName, nil
}

// Search filters items by query (name or snippet contains, case-insensitive)
func Search(items []Item, query string) []Item {
	if query == "" {
		return items
	}
	q := strings.ToLower(strings.TrimSpace(query))
	var out []Item
	for _, it := range items {
		if strings.Contains(strings.ToLower(it.Name), q) ||
			strings.Contains(strings.ToLower(it.Path), q) ||
			strings.Contains(strings.ToLower(it.Snippet), q) ||
			strings.Contains(strings.ToLower(it.Category), q) {
			out = append(out, it)
		}
	}
	return out
}
