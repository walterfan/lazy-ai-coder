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
	root      string   // absolute path to assets directory
	ossRoots  []string // absolute paths to additional OSS skill directories
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

// AddOSSRoot registers an additional directory tree that contains SKILL.md files.
// The category for discovered skills is derived from the top-level folder name.
func (l *Loader) AddOSSRoot(absPath string) {
	l.ossRoots = append(l.ossRoots, absPath)
}

// safePath ensures path is under loader.root (or an oss root) and has no path traversal
func (l *Loader) safePath(rel string) (string, bool) {
	rel = filepath.Clean(rel)
	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", false
	}

	// Try assets root first
	abs := filepath.Join(l.root, rel)
	abs = filepath.Clean(abs)
	rootClean := filepath.Clean(l.root)
	if abs == rootClean || strings.HasPrefix(abs, rootClean+string(os.PathSeparator)) {
		return abs, true
	}

	// Try each oss root
	for _, ossRoot := range l.ossRoots {
		abs = filepath.Join(ossRoot, rel)
		abs = filepath.Clean(abs)
		clean := filepath.Clean(ossRoot)
		if abs == clean || strings.HasPrefix(abs, clean+string(os.PathSeparator)) {
			return abs, true
		}
	}

	return "", false
}

// resolveAbsPath resolves an absolute path if it falls under any known root.
// Used for reading files whose absolute path is already known from discovery.
func (l *Loader) resolveAbsPath(abs string) bool {
	abs = filepath.Clean(abs)
	rootClean := filepath.Clean(l.root)
	if strings.HasPrefix(abs, rootClean+string(os.PathSeparator)) {
		return true
	}
	for _, ossRoot := range l.ossRoots {
		clean := filepath.Clean(ossRoot)
		if strings.HasPrefix(abs, clean+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

// ListCommands returns all .md files under assets/commands (recursive, category from subfolder)
func (l *Loader) ListCommands() ([]Item, error) {
	root := filepath.Join(l.root, "commands")
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
		if !strings.HasSuffix(strings.ToLower(filepath.Base(path)), ".md") {
			return nil
		}
		rel, err := filepath.Rel(l.root, path)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		parts := strings.Split(rel, "/")
		category := ""
		if len(parts) >= 3 {
			category = parts[1]
		}
		item, err := l.itemFromFile(rel, TypeCommand, filepath.Base(path), category)
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

// ListSkills returns all SKILL.md files under assets/skills and all OSS roots (recursive)
func (l *Loader) ListSkills() ([]Item, error) {
	var out []Item

	// 1. Scan assets/skills (category = subfolder name, e.g. "walterfan")
	assetsSkillsDir := filepath.Join(l.root, "skills")
	if _, err := os.Stat(assetsSkillsDir); err == nil {
		_ = filepath.Walk(assetsSkillsDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Base(path) != "SKILL.md" {
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
				category = parts[1]
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
	}

	// 2. Scan each OSS root for SKILL.md files
	for _, ossRoot := range l.ossRoots {
		category := filepath.Base(ossRoot)
		_ = filepath.Walk(ossRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Base(path) != "SKILL.md" {
				return nil
			}
			// Skip hidden directories (e.g. .git)
			relToOss, _ := filepath.Rel(ossRoot, path)
			for _, seg := range strings.Split(filepath.ToSlash(relToOss), "/") {
				if strings.HasPrefix(seg, ".") {
					return nil
				}
			}
			skillName := filepath.Base(filepath.Dir(path))
			if skillName == "." || skillName == category {
				skillName = "Skill"
			}
			// Build a virtual path: oss/<category>/...relative to ossRoot
			rel := filepath.ToSlash(filepath.Join("oss", category, relToOss))
			item, err := l.itemFromFileAbs(path, rel, TypeSkill, skillName, category)
			if err != nil {
				return nil
			}
			out = append(out, item)
			return nil
		})
	}

	return out, nil
}

func (l *Loader) itemFromFile(rel, assetType, name, category string) (Item, error) {
	abs, ok := l.safePath(rel)
	if !ok {
		return Item{}, os.ErrInvalid
	}
	return l.itemFromFileAbs(abs, rel, assetType, name, category)
}

func (l *Loader) itemFromFileAbs(abs, rel, assetType, name, category string) (Item, error) {
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

// ReadFile returns the full content of an asset by relative path.
// Supports paths under assets/ (e.g. "commands/bug.fix.md") and
// virtual oss paths (e.g. "oss/anthropics-skills/skills/pdf/SKILL.md").
func (l *Loader) ReadFile(relPath string) ([]byte, string, error) {
	relPath = filepath.ToSlash(filepath.Clean(relPath))

	// Handle oss/ virtual paths: oss/<category>/<rest>
	if strings.HasPrefix(relPath, "oss/") {
		parts := strings.SplitN(relPath, "/", 3)
		if len(parts) < 3 {
			return nil, "", os.ErrInvalid
		}
		category := parts[1]
		rest := parts[2]
		for _, ossRoot := range l.ossRoots {
			if filepath.Base(ossRoot) == category {
				abs := filepath.Clean(filepath.Join(ossRoot, rest))
				if !strings.HasPrefix(abs, filepath.Clean(ossRoot)+string(os.PathSeparator)) {
					return nil, "", os.ErrPermission
				}
				data, err := os.ReadFile(abs)
				if err != nil {
					return nil, "", err
				}
				return data, filepath.Base(abs), nil
			}
		}
		return nil, "", os.ErrNotExist
	}

	// Normal assets/ path
	abs, ok := l.safePath(relPath)
	if !ok {
		return nil, "", os.ErrPermission
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, "", err
	}
	return data, filepath.Base(abs), nil
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
// skillDir is the relative path to the skill folder under the assets root
// (e.g. "skills/walterfan/my-skill") or under an OSS root (e.g. "oss/gstack/learn").
func (l *Loader) ZipSkillFolder(skillDir string) ([]byte, string, error) {
	var abs string
	skillDir = filepath.ToSlash(filepath.Clean(skillDir))

	if strings.HasPrefix(skillDir, "oss/") {
		parts := strings.SplitN(skillDir, "/", 3)
		if len(parts) < 3 {
			return nil, "", os.ErrInvalid
		}
		found := false
		for _, ossRoot := range l.ossRoots {
			if filepath.Base(ossRoot) == parts[1] {
				abs = filepath.Clean(filepath.Join(ossRoot, parts[2]))
				if !strings.HasPrefix(abs, filepath.Clean(ossRoot)+string(os.PathSeparator)) {
					return nil, "", os.ErrPermission
				}
				found = true
				break
			}
		}
		if !found {
			return nil, "", os.ErrNotExist
		}
	} else {
		var ok bool
		abs, ok = l.safePath(skillDir)
		if !ok {
			return nil, "", os.ErrPermission
		}
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
