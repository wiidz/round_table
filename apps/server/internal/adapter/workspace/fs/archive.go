package fs

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"round_table/apps/server/internal/adapter/workspace"
)

const (
	maxArchiveFileBytes = 32 << 20 // 32 MiB per file
	maxArchiveFiles     = 256
)

// WriteMeetingArchive zips all files under the meeting workspace directory.
func (s *Store) WriteMeetingArchive(meetingID string, w io.Writer) error {
	dir, err := s.meetingDir(meetingID)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return workspace.ErrNotFound
	}

	zw := zip.NewWriter(w)
	defer zw.Close()

	count := 0
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if count >= maxArchiveFiles {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Size() > maxArchiveFileBytes {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "" || strings.Contains(rel, "..") {
			return nil
		}

		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = rel
		hdr.Method = zip.Deflate

		writer, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
		count++
		return nil
	})
	return err
}

// DeleteMeeting removes the meeting workspace directory tree.
func (s *Store) DeleteMeeting(meetingID string) error {
	dir, err := s.meetingDir(meetingID)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return workspace.ErrNotFound
	}
	return os.RemoveAll(dir)
}
