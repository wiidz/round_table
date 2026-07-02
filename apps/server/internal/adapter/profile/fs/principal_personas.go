package fs

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/profile"
)

const (
	principalPersonasFile = "personas.json"
	defaultPersonaID      = "pp_default"
	defaultPersonaTitle   = "默认档案"
)

func (s *Store) principalDir(id string) (string, error) {
	if err := validatePrincipalID(id); err != nil {
		return "", err
	}
	return filepath.Join(s.Root, "principals", id), nil
}

func (s *Store) principalPersonasPath(id string) (string, error) {
	dir, err := s.principalDir(id)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, principalPersonasFile), nil
}

func (s *Store) principalPersonaDir(principalID, personaID string) (string, error) {
	if err := validatePrincipalID(principalID); err != nil {
		return "", err
	}
	if err := validatePersonaID(personaID); err != nil {
		return "", err
	}
	dir, err := s.principalDir(principalID)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "personas", personaID), nil
}

func validatePersonaID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" || strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return profile.ErrNotFound
	}
	return nil
}

// EnsurePrincipalPersonas seeds manifest and migrates legacy root USER.md when needed.
func (s *Store) EnsurePrincipalPersonas(principalID string) (profile.PrincipalPersonaManifest, error) {
	if err := s.EnsurePrincipal(principalID); err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	manifest, err := s.loadPrincipalPersonaManifest(principalID)
	if err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	if len(manifest.Personas) > 0 {
		return manifest, nil
	}

	now := time.Now().UTC()
	persona := profile.PrincipalPersonaMeta{
		ID:        defaultPersonaID,
		Title:     defaultPersonaTitle,
		CreatedAt: now,
		UpdatedAt: now,
	}
	personaDir, err := s.principalPersonaDir(principalID, persona.ID)
	if err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	if err := os.MkdirAll(personaDir, 0o755); err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}

	legacyPath := filepath.Join(s.Root, "principals", principalID, profile.FileUser)
	if data, err := os.ReadFile(legacyPath); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := os.WriteFile(filepath.Join(personaDir, profile.FileUser), data, 0o644); err != nil {
			return profile.PrincipalPersonaManifest{}, err
		}
	} else if err := s.seedPersonaUserFromTemplate(principalID, persona.ID); err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}

	manifest = profile.PrincipalPersonaManifest{
		ActivePersonaID: persona.ID,
		Personas:        []profile.PrincipalPersonaMeta{persona},
	}
	if err := s.savePrincipalPersonaManifest(principalID, manifest); err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	if err := s.syncActivePrincipalUserFile(principalID, manifest); err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	return manifest, nil
}

func (s *Store) ListPrincipalPersonas(principalID string) (profile.PrincipalPersonaManifest, error) {
	return s.EnsurePrincipalPersonas(principalID)
}

func (s *Store) ReadPrincipalPersonaUserProfile(principalID, personaID string) (profile.UserProfile, error) {
	if _, err := s.EnsurePrincipalPersonas(principalID); err != nil {
		return profile.UserProfile{}, err
	}
	data, err := s.readPrincipalPersonaUser(principalID, personaID)
	if err != nil {
		return profile.UserProfile{}, err
	}
	return profile.ParseUserMD(string(data)), nil
}

func (s *Store) WritePrincipalPersonaUserProfile(principalID, personaID string, payload profile.UserProfile) (profile.UserProfile, error) {
	if _, err := s.EnsurePrincipalPersonas(principalID); err != nil {
		return profile.UserProfile{}, err
	}
	manifest, err := s.loadPrincipalPersonaManifest(principalID)
	if err != nil {
		return profile.UserProfile{}, err
	}
	if !personaExists(manifest, personaID) {
		return profile.UserProfile{}, profile.ErrNotFound
	}

	content := profile.RenderUserMD(payload)
	personaDir, err := s.principalPersonaDir(principalID, personaID)
	if err != nil {
		return profile.UserProfile{}, err
	}
	if err := os.MkdirAll(personaDir, 0o755); err != nil {
		return profile.UserProfile{}, err
	}
	if err := os.WriteFile(filepath.Join(personaDir, profile.FileUser), []byte(content), 0o644); err != nil {
		return profile.UserProfile{}, err
	}

	now := time.Now().UTC()
	for i := range manifest.Personas {
		if manifest.Personas[i].ID == personaID {
			manifest.Personas[i].UpdatedAt = now
			break
		}
	}
	if err := s.savePrincipalPersonaManifest(principalID, manifest); err != nil {
		return profile.UserProfile{}, err
	}
	if manifest.ActivePersonaID == personaID {
		if err := s.syncActivePrincipalUserFile(principalID, manifest); err != nil {
			return profile.UserProfile{}, err
		}
	}
	return profile.ParseUserMD(content), nil
}

func (s *Store) CreatePrincipalPersona(principalID, title string) (profile.PrincipalPersonaMeta, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return profile.PrincipalPersonaMeta{}, fmt.Errorf("profile: persona title required")
	}
	manifest, err := s.EnsurePrincipalPersonas(principalID)
	if err != nil {
		return profile.PrincipalPersonaMeta{}, err
	}

	id, err := newPersonaID(manifest)
	if err != nil {
		return profile.PrincipalPersonaMeta{}, err
	}
	now := time.Now().UTC()
	persona := profile.PrincipalPersonaMeta{
		ID:        id,
		Title:     title,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.seedPersonaUserFromTemplate(principalID, persona.ID); err != nil {
		return profile.PrincipalPersonaMeta{}, err
	}
	manifest.Personas = append(manifest.Personas, persona)
	if err := s.savePrincipalPersonaManifest(principalID, manifest); err != nil {
		return profile.PrincipalPersonaMeta{}, err
	}
	return persona, nil
}

func (s *Store) SetActivePrincipalPersona(principalID, personaID string) (profile.PrincipalPersonaManifest, error) {
	manifest, err := s.EnsurePrincipalPersonas(principalID)
	if err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	if !personaExists(manifest, personaID) {
		return profile.PrincipalPersonaManifest{}, profile.ErrNotFound
	}
	manifest.ActivePersonaID = personaID
	if err := s.savePrincipalPersonaManifest(principalID, manifest); err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	if err := s.syncActivePrincipalUserFile(principalID, manifest); err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	return manifest, nil
}

func (s *Store) ReadPrincipalDetailWithPersonas(id string) (profile.PrincipalDetail, error) {
	manifest, err := s.EnsurePrincipalPersonas(id)
	if err != nil {
		return profile.PrincipalDetail{}, err
	}
	active := manifest.ActivePersonaID
	if active == "" && len(manifest.Personas) > 0 {
		active = manifest.Personas[0].ID
	}
	userProfile, err := s.ReadPrincipalPersonaUserProfile(id, active)
	if err != nil {
		return profile.PrincipalDetail{}, err
	}
	return profile.PrincipalDetail{
		ID:              id,
		ActivePersonaID: active,
		Personas:        manifest.Personas,
		UserProfile:     userProfile,
	}, nil
}

func (s *Store) loadPrincipalPersonaManifest(principalID string) (profile.PrincipalPersonaManifest, error) {
	path, err := s.principalPersonasPath(principalID)
	if err != nil {
		return profile.PrincipalPersonaManifest{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return profile.PrincipalPersonaManifest{}, nil
		}
		return profile.PrincipalPersonaManifest{}, err
	}
	var manifest profile.PrincipalPersonaManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return profile.PrincipalPersonaManifest{}, fmt.Errorf("profile: parse personas: %w", err)
	}
	return manifest, nil
}

func (s *Store) savePrincipalPersonaManifest(principalID string, manifest profile.PrincipalPersonaManifest) error {
	path, err := s.principalPersonasPath(principalID)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *Store) readPrincipalPersonaUser(principalID, personaID string) ([]byte, error) {
	dir, err := s.principalPersonaDir(principalID, personaID)
	if err != nil {
		return nil, err
	}
	return s.read(filepath.Join(dir, profile.FileUser))
}

func (s *Store) seedPersonaUserFromTemplate(principalID, personaID string) error {
	dir, err := s.principalPersonaDir(principalID, personaID)
	if err != nil {
		return err
	}
	return s.ensureFromTemplates(dir, filepath.Join(s.Templates, "principals"))
}

func (s *Store) syncActivePrincipalUserFile(principalID string, manifest profile.PrincipalPersonaManifest) error {
	active := manifest.ActivePersonaID
	if active == "" {
		return nil
	}
	data, err := s.readPrincipalPersonaUser(principalID, active)
	if err != nil {
		return err
	}
	root := filepath.Join(s.Root, "principals", principalID, profile.FileUser)
	return os.WriteFile(root, data, 0o644)
}

func personaExists(manifest profile.PrincipalPersonaManifest, personaID string) bool {
	for _, p := range manifest.Personas {
		if p.ID == personaID {
			return true
		}
	}
	return false
}

func newPersonaID(manifest profile.PrincipalPersonaManifest) (string, error) {
	for i := 0; i < 8; i++ {
		var buf [4]byte
		if _, err := rand.Read(buf[:]); err != nil {
			return "", err
		}
		id := "pp_" + hex.EncodeToString(buf[:])
		if !personaExists(manifest, id) {
			return id, nil
		}
	}
	return "", fmt.Errorf("profile: generate persona id failed")
}
