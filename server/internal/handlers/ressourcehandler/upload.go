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
	repo := currentRepo(c)
	// The path inside the repository, used for the files, and the full path, used for artifacts and permissions
	relPath, err := validateUploadPath(strings.TrimPrefix(c.Params("*"), repo.Name+"/"))
	if err != nil {
		d.logTo(c, repos, c.Params("*"))
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	fullPath := repo.Name + "/" + relPath

	allowed, err := auth.Authorize(c, repos.Grants, fullPath, repository.PermissionWrite)
	if err != nil {
		log.Printf("Checking permissions for %s failed: %v", fullPath, err)
		return fiber.ErrInternalServerError
	}
	if !allowed {
		d.logTo(c, repos, fullPath)
		return fiber.NewError(fiber.StatusForbidden, "no write permission for "+path.Dir(fullPath))
	}

	// Checked after authorization so only users with write access learn which artifacts are registered
	artifact, err := repos.Artifacts.FindByPath(c.Context(), fullPath)
	registered := err == nil
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Printf("Looking up artifact for %s failed: %v", fullPath, err)
		return fiber.ErrInternalServerError
	}
	// Checksums and metadata follow the files they belong to, so an unregistered artifact is only registered by a file
	if !registered && (!repo.AutoRegister || isIgnoredUpload(relPath)) {
		// Not 404: Maven reports that as "Could not find artifact" and drops the message
		return fiber.NewError(fiber.StatusForbidden,
			coordinatesOf(relPath)+" is not registered in repository "+repo.Name+", an admin has to register it first")
	}

	if isIgnoredUpload(relPath) {
		return c.SendStatus(fiber.StatusNoContent)
	}
	if registered {
		d.artifact, d.relPath = &artifact, fullPath
	}
	if version := path.Base(path.Dir(relPath)); !repo.VersionPolicy.Allows(version) {
		return fiber.NewError(fiber.StatusBadRequest, policyError(repo, version))
	}
	if !hasStoredExtension(relPath) {
		return fiber.NewError(fiber.StatusUnsupportedMediaType, "only "+strings.Join(storedExtensions, ", ")+" files can be uploaded")
	}
	// Registered only once the upload passed the checks, so a rejected one doesn't leave an artifact behind
	if !registered {
		if artifact, err = autoRegister(c, repos, repo, relPath); err != nil {
			return err
		}
		d.artifact, d.relPath = &artifact, fullPath
	}

	if err := root.MkdirAll(repo.Name, 0o755); err != nil {
		log.Printf("Creating the directory of repository %s failed: %v", repo.Name, err)
		return fiber.ErrInternalServerError
	}
	repoRoot, err := root.OpenRoot(repo.Name)
	if err != nil {
		log.Printf("Opening the directory of repository %s failed: %v", repo.Name, err)
		return fiber.ErrInternalServerError
	}
	defer repoRoot.Close()

	replace := artifact.AllowReleaseOverwrite || isSnapshotPath(relPath)
	if !replace {
		if _, err := repoRoot.Stat(relPath); err == nil {
			return errReleaseExists(fullPath)
		} else if !errors.Is(err, fs.ErrNotExist) {
			log.Printf("Checking for existing %s failed: %v", fullPath, err)
			return fiber.ErrInternalServerError
		}
	}

	var body io.Reader
	if c.Request().IsBodyStream() {
		body = c.Request().BodyStream()
	} else {
		body = bytes.NewReader(c.Body())
	}

	if err := writeFileAtomicWithHashes(repoRoot, relPath, body, replace); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return errReleaseExists(fullPath)
		}
		log.Printf("Upload of %s failed: %v", fullPath, err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to store file")
	}
	if err := updateMetadata(repoRoot, repo.Name, relPath); err != nil {
		log.Printf("Updating metadata for %s failed: %v", fullPath, err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to update metadata")
	}
	if info, err := repoRoot.Stat(relPath); err == nil {
		size := info.Size()
		d.size = &size
	}
	return c.SendStatus(fiber.StatusCreated)
}

// autoRegister registers the artifact of the upload path relPath in repo, which must allow it.
func autoRegister(c fiber.Ctx, repos repository.Repositories, repo repository.Repo, relPath string) (repository.Artifact, error) {
	groupID, artifactID, _ := strings.Cut(coordinatesOf(relPath), ":")
	if !repository.ValidGroupID(groupID) || !repository.ValidArtifactID(artifactID) {
		return repository.Artifact{}, fiber.NewError(fiber.StatusBadRequest, coordinatesOf(relPath)+" is not a valid groupId:artifactId")
	}

	artifact, err := repos.Artifacts.Create(c.Context(), repo.Name, groupID, artifactID, repo.DefaultPublic, false, nil)
	if errors.Is(err, repository.ErrConflict) {
		// Registered by a concurrent upload in the meantime
		artifact, err = repos.Artifacts.Get(c.Context(), repo.Name, groupID, artifactID)
	} else if err == nil {
		log.Printf("Registered %s:%s in repository %s on deploy by %s", groupID, artifactID, repo.Name, auth.CurrentUser(c).Username)
	}
	if err != nil {
		log.Printf("Registering %s:%s in repository %s failed: %v", groupID, artifactID, repo.Name, err)
		return repository.Artifact{}, fiber.ErrInternalServerError
	}
	return artifact, nil
}

// coordinatesOf returns groupId:artifactId of a valid upload path <group path>/<artifactId>/<version>/<file>.
func coordinatesOf(relPath string) string {
	segments := strings.Split(relPath, "/")
	n := len(segments)
	return strings.Join(segments[:n-3], ".") + ":" + segments[n-3]
}

func policyError(repo repository.Repo, version string) string {
	if repo.VersionPolicy == repository.PolicySnapshot {
		return "repository " + repo.Name + " only accepts SNAPSHOT versions, not " + version
	}
	return "repository " + repo.Name + " only accepts release versions, not " + version
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
