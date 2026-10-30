package ressourcehandler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"os"
	"path"
	"strings"

	"github.com/gofiber/fiber/v3"
)

var ignoredSuffixes = []string{".md5", ".sha1", ".sha256", ".sha512"}

var storedExtensions = []string{".jar", ".pom"}

func newUploadHandler(root *os.Root) fiber.Handler {
	return func(c fiber.Ctx) error {
		relPath, err := validateUploadPath(c.Params("*"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		if isIgnoredUpload(relPath) {
			return c.SendStatus(fiber.StatusNoContent)
		}
		if !hasStoredExtension(relPath) {
			return fiber.NewError(fiber.StatusUnsupportedMediaType, "only .jar and .pom files can be uploaded")
		}

		var body io.Reader
		if c.Request().IsBodyStream() {
			body = c.Request().BodyStream()
		} else {
			body = bytes.NewReader(c.Body())
		}

		if err := writeFileAtomic(root, relPath, body); err != nil {
			log.Printf("Upload of %s failed: %v", relPath, err)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to store file")
		}
		return c.SendStatus(fiber.StatusCreated)
	}
}

// validateUploadPath checks that p looks like <group path>/<artifactId>/<version>/<file> and returns it cleaned
func validateUploadPath(p string) (string, error) {
	if p == "" || strings.Contains(p, "\\") {
		return "", errors.New("invalid path")
	}
	segments := strings.Split(p, "/")
	for _, s := range segments {
		if s == "" || s == "." || s == ".." || strings.HasPrefix(s, ".") {
			return "", errors.New("invalid path")
		}
	}

	name := segments[len(segments)-1]
	if strings.HasPrefix(name, "maven-metadata.xml") {
		return path.Join(segments...), nil
	}

	if len(segments) < 4 {
		return "", errors.New("path must be <group>/<artifactId>/<version>/<file>")
	}
	artifactID := segments[len(segments)-3]
	version := segments[len(segments)-2]
	// Snapshot files are named <artifactId>-<base>-<timestamp>-<build>, not <artifactId>-<base>-SNAPSHOT
	prefix := artifactID + "-" + strings.TrimSuffix(version, "SNAPSHOT")
	if !strings.HasPrefix(name, prefix) {
		return "", fmt.Errorf("file name must start with %q", prefix)
	}
	return path.Join(segments...), nil
}

func isIgnoredUpload(p string) bool {
	if strings.HasPrefix(path.Base(p), "maven-metadata.xml") {
		return true
	}
	for _, suffix := range ignoredSuffixes {
		if strings.HasSuffix(p, suffix) {
			return true
		}
	}
	return false
}

func hasStoredExtension(p string) bool {
	for _, ext := range storedExtensions {
		if strings.HasSuffix(p, ext) {
			return true
		}
	}
	return false
}

func writeFileAtomic(root *os.Root, relPath string, body io.Reader) error {
	dir := path.Dir(relPath)
	if err := root.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpPath := path.Join(dir, fmt.Sprintf(".%s.%d.upload", path.Base(relPath), rand.Uint64()))
	tmp, err := root.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}

	_, err = io.Copy(tmp, body)
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = root.Rename(tmpPath, relPath)
	}
	if err != nil {
		_ = root.Remove(tmpPath)
	}
	return err
}
