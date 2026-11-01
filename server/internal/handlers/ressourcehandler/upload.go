package ressourcehandler

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"log"
	"math/rand/v2"
	"os"
	"path"
	"simplemavenrepo/internal/auth"
	"simplemavenrepo/internal/repository"
	"strings"

	"github.com/gofiber/fiber/v3"
)

var ignoredSuffixes = []string{".md5", ".sha1", ".sha256", ".sha512"}

var storedExtensions = []string{".jar", ".pom", ".module", ".aar", ".klib", ".war", ".ear", ".zip", ".tar.gz"}

type deployment struct {
	artifact *repository.Artifact
	relPath  string
	size     *int64
}

// logTo makes the upload of p appear in the deploy log of its artifact, if p belongs to one.
func (d *deployment) logTo(c fiber.Ctx, repos repository.Repositories, p string) {
	if isIgnoredUpload(p) {
		return
	}
	artifact, err := repos.Artifacts.FindByPath(c.Context(), p)
	if err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			log.Printf("Looking up artifact for %s failed: %v", p, err)
		}
		return
	}
	d.artifact = &artifact
	d.relPath = p
}

func newUploadHandler(root *os.Root, repos repository.Repositories) fiber.Handler {
	return func(c fiber.Ctx) error {
		var d deployment
		err := upload(c, root, repos, &d)
		if d.artifact != nil {
			recordDeploy(c, repos, d, err)
		}
		return err
	}
}

// recordDeploy writes the outcome of an upload to the deploy log. A failure to log doesn't fail the upload.
func recordDeploy(c fiber.Ctx, repos repository.Repositories, d deployment, uploadErr error) {
	status, msg := fiber.StatusCreated, ""
	if uploadErr != nil {
		status, msg = fiber.StatusInternalServerError, "internal server error"
		var fe *fiber.Error
		if errors.As(uploadErr, &fe) {
			status, msg = fe.Code, fe.Message
		}
	}

	version, file, ok := strings.Cut(strings.TrimPrefix(d.relPath, d.artifact.Path+"/"), "/")
	if !ok {
		version, file = "", version
	}

	ctx := context.WithoutCancel(c.Context())
	if err := repos.Deploys.Record(ctx, d.artifact.Path, version, file, auth.CurrentUser(c).Username, status, msg, d.size); err != nil {
		log.Printf("Recording deploy of %s failed: %v", d.relPath, err)
	}
}

func upload(c fiber.Ctx, root *os.Root, repos repository.Repositories, d *deployment) error {
	relPath, err := validateUploadPath(c.Params("*"))
	if err != nil {
		d.logTo(c, repos, c.Params("*"))
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	allowed, err := auth.Authorize(c.Context(), repos.Grants, auth.CurrentUser(c), relPath, repository.PermissionWrite)
	if err != nil {
		log.Printf("Checking permissions for %s failed: %v", relPath, err)
		return fiber.ErrInternalServerError
	}
	if !allowed {
		d.logTo(c, repos, relPath)
		return fiber.NewError(fiber.StatusForbidden, "no write permission for "+path.Dir(relPath))
	}

	// Checked after authorization so only users with write access learn which artifacts are registered
	artifact, err := repos.Artifacts.FindByPath(c.Context(), relPath)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "no registered artifact contains "+relPath)
		}
		log.Printf("Looking up artifact for %s failed: %v", relPath, err)
		return fiber.ErrInternalServerError
	}

	if isIgnoredUpload(relPath) {
		return c.SendStatus(fiber.StatusNoContent)
	}
	d.artifact, d.relPath = &artifact, relPath
	if !hasStoredExtension(relPath) {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "only "+strings.Join(storedExtensions, ", ")+" files can be uploaded")
	}

	replace := artifact.AllowReleaseOverwrite || isSnapshotPath(relPath)
	if !replace {
		if _, err := root.Stat(relPath); err == nil {
			return errReleaseExists(relPath)
		} else if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("Checking for existing %s failed: %v", relPath, err)
			return fiber.ErrInternalServerError
		}
	}

	var body io.Reader
	if c.Request().IsBodyStream() {
		body = c.Request().BodyStream()
	} else {
		body = bytes.NewReader(c.Body())
	}

	if err := writeFileAtomicWithHashes(root, relPath, body, replace); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errReleaseExists(relPath)
		}
		log.Printf("Upload of %s failed: %v", relPath, err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to store file")
	}
	if err := updateMetadata(root, relPath); err != nil {
		log.Printf("Updating metadata for %s failed: %v", relPath, err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update metadata")
	}
	if info, err := root.Stat(relPath); err == nil {
		size := info.Size()
		d.size = &size
	}
	return c.SendStatus(fiber.StatusCreated)
}

func errReleaseExists(relPath string) error {
	return fiber.NewError(fiber.StatusConflict, relPath+" is already published and releases can't be overwritten")
}

// isSnapshotPath reports whether p is a file of a SNAPSHOT version, e.g. com/example/demo/1.0-SNAPSHOT/demo-1.0-20260101.120000-1.jar
func isSnapshotPath(p string) bool {
	return strings.HasSuffix(path.Base(path.Dir(p)), "-SNAPSHOT")
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

// writeFileAtomicWithHashes stores body at relPath along with its checksum files. Unless replace is set, it fails with
// fs.ErrExist if relPath already exists.
func writeFileAtomicWithHashes(root *os.Root, relPath string, body io.Reader, replace bool) error {
	hashes := map[string]hash.Hash{
		"md5":    md5.New(),
		"sha1":   sha1.New(),
		"sha256": sha256.New(),
		"sha512": sha512.New(),
	}

	var hashWriters []io.Writer
	for _, h := range hashes {
		hashWriters = append(hashWriters, h)
	}
	if err := writeFileAtomic(root, relPath, body, replace, hashWriters...); err != nil {
		return err
	}

	for ext, h := range hashes {
		sum := hex.EncodeToString(h.Sum(nil))
		if err := writeFileAtomic(root, relPath+"."+ext, strings.NewReader(sum), true); err != nil {
			return err
		}
	}
	return nil
}

func RemoveStaleUploads(root *os.Root) error {
	removed := 0
	err := fs.WalkDir(root.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.Type().IsRegular() && strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".upload") {
			if err := root.Remove(p); err != nil {
				return err
			}
			removed++
		}
		return nil
	})
	if removed > 0 {
		log.Printf("Removed %d unfinished upload(s)", removed)
	}
	return err
}

// writeFileAtomic writes body to a temporary file and moves it to relPath. Without replace, it links the file instead,
// which fails with fs.ErrExist if relPath already exists.
func writeFileAtomic(root *os.Root, relPath string, body io.Reader, replace bool, extra ...io.Writer) error {
	dir := path.Dir(relPath)
	if err := root.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpPath := path.Join(dir, fmt.Sprintf(".%s.%d.upload", path.Base(relPath), rand.Uint64()))
	tmp, err := root.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}

	_, err = io.Copy(io.MultiWriter(append([]io.Writer{tmp}, extra...)...), body)

	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		if replace {
			err = root.Rename(tmpPath, relPath)
		} else {
			// Unlike Rename, Link fails if relPath already exists
			err = root.Link(tmpPath, relPath)
		}
	}
	if err != nil || !replace {
		_ = root.Remove(tmpPath)
	}
	return err
}
