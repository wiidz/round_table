package fs

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/brief"
)

// Store implements brief.Port on the local filesystem.
type Store struct {
	Root      string
	Templates string
}

func NewStore(root, templates string) *Store {
	return &Store{Root: root, Templates: templates}
}

func (s *Store) ListTemplates() ([]brief.TemplateIndex, error) {
	byID := make(map[string]brief.TemplateIndex)

	if err := s.scanDir(s.Templates, brief.SourceBuiltin, byID); err != nil {
		return nil, err
	}
	if err := s.scanDir(s.Root, brief.SourceCustom, byID); err != nil {
		return nil, err
	}

	out := make([]brief.TemplateIndex, 0, len(byID))
	for _, item := range byID {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func (s *Store) ReadTemplate(id string) (brief.TemplateDetail, error) {
	if err := validateID(id); err != nil {
		return brief.TemplateDetail{}, err
	}

	path, source, err := s.resolvePath(id)
	if err != nil {
		return brief.TemplateDetail{}, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return brief.TemplateDetail{}, brief.ErrNotFound
		}
		return brief.TemplateDetail{}, err
	}

	doc, err := brief.ParseDocument(raw)
	if err != nil {
		return brief.TemplateDetail{}, err
	}

	info, _ := os.Stat(filepath.Dir(path))
	updated := time.Time{}
	if info != nil {
		updated = info.ModTime()
	}

	return brief.TemplateDetail{
		ID:          id,
		Title:       doc.Meta.Title,
		Description: doc.Meta.Description,
		Source:      source,
		Content:     string(raw),
		Document:    doc,
		Launch:      brief.DocumentToLaunch(doc),
		UpdatedAt:   updated,
	}, nil
}

func (s *Store) WriteTemplate(id string, content []byte) error {
	if err := validateID(id); err != nil {
		return err
	}
	if _, err := brief.ParseDocument(content); err != nil {
		return err
	}

	if s.isBuiltin(id) {
		return brief.ErrBuiltinReadonly
	}

	dir := filepath.Join(s.Root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, brief.FileBrief), content, 0o644)
}

func (s *Store) CreateTemplate(content []byte) (string, error) {
	doc, err := brief.ParseDocument(content)
	if err != nil {
		return "", err
	}
	base := brief.SlugTemplateID(doc.Meta.Title)
	id := brief.NextAvailableTemplateID(base, s.idTaken)
	if err := s.WriteTemplate(id, content); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Store) idTaken(id string) bool {
	if err := validateID(id); err != nil {
		return true
	}
	if s.isBuiltin(id) {
		return true
	}
	path := filepath.Join(s.Root, id, brief.FileBrief)
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func (s *Store) CloneFromMeetingDoc(meetingDoc string) (brief.LaunchDraft, error) {
	return brief.ParseMeetingDoc(meetingDoc)
}

func (s *Store) scanDir(root, source string, byID map[string]brief.TemplateIndex) error {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		id := entry.Name()
		if err := validateID(id); err != nil {
			continue
		}
		path := filepath.Join(root, id, brief.FileBrief)
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		doc, err := brief.ParseDocument(raw)
		if err != nil {
			continue
		}
		dir := filepath.Join(root, id)
		item := brief.TemplateIndex{
			ID:          id,
			Title:       doc.Meta.Title,
			Description: doc.Meta.Description,
			Source:      source,
			UpdatedAt:   dirModTime(dir),
		}
		if prev, ok := byID[id]; ok && prev.Source == brief.SourceCustom {
			continue
		}
		byID[id] = item
	}
	return nil
}

func (s *Store) resolvePath(id string) (string, string, error) {
	custom := filepath.Join(s.Root, id, brief.FileBrief)
	if _, err := os.Stat(custom); err == nil {
		return custom, brief.SourceCustom, nil
	}
	builtin := filepath.Join(s.Templates, id, brief.FileBrief)
	if _, err := os.Stat(builtin); err == nil {
		return builtin, brief.SourceBuiltin, nil
	}
	return "", "", brief.ErrNotFound
}

func (s *Store) isBuiltin(id string) bool {
	path := filepath.Join(s.Templates, id, brief.FileBrief)
	_, err := os.Stat(path)
	return err == nil
}

func validateID(id string) error {
	if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return brief.ErrInvalidID
	}
	return nil
}

func dirModTime(dir string) time.Time {
	info, err := os.Stat(dir)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
