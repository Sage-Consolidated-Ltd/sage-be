package s3

import (
	"errors"
	"path/filepath"
	"strings"
)

type UploadPolicy struct {
	AllowedExt map[string]FileInfo
	MaxSize    int64
}

func (p *UploadPolicy) Validate(filename string, size int64) (FileInfo, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	info, ok := p.AllowedExt[ext]
	if !ok || !info.Allowed {
		return FileInfo{}, errors.New("file type not allowed")
	}

	if size <= 0 || size > p.MaxSize {
		return FileInfo{}, errors.New("invalid file size")
	}

	return info, nil
}
