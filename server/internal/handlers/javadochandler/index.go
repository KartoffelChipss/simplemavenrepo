package javadochandler

import (
	"bytes"
	"embed"
	"errors"
	"io"
	"io/fs"
	"log"
	"mime"
	"path"
	"regexp"
	"simplemavenrepo/internal/javadoc"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
)

//go:embed assets
var assets embed.FS

// Prefix is where javadoc is served. AssetsPrefix holds the stylesheet and script added to every javadoc page.
const (
	Prefix       = "/javadoc/"
	AssetsPrefix = Prefix + "_smr/"
)

// sandbox gives javadoc pages an opaque origin: their scripts run, but can't reach the web UI or the API as the
// viewer. Links to other sites open in a new tab outside the sandbox, see assets/bridge.js.
const sandbox = "sandbox allow-scripts allow-popups allow-popups-to-escape-sandbox"

var headEnd = regexp.MustCompile(`(?i)</head\s*>`)

const injected = `<link rel="stylesheet" type="text/css" href="` + AssetsPrefix + `theme.css">` +
	`<script type="text/javascript" src="` + AssetsPrefix + `bridge.js"></script>`

func RegisterJavadocRoutes(app *fiber.App, repoFS fs.FS, signer *javadoc.Signer) {
	assetsFS, err := fs.Sub(assets, "assets")
	if err != nil {
		log.Fatalf("Javadoc assets: %v", err)
	}
	app.Get(AssetsPrefix+"*", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentSecurityPolicy, sandbox)
		return c.Next()
	}, static.New("", static.Config{
		FS:              assetsFS,
		MaxAge:          300,
		NotFoundHandler: func(c fiber.Ctx) error { return fiber.ErrNotFound },
	}))

	app.Get(Prefix+":token/*", func(c fiber.Ctx) error {
		return serveFile(c, repoFS, signer)
	})
}

func serveFile(c fiber.Ctx, repoFS fs.FS, signer *javadoc.Signer) error {
	jarPath, ok := signer.Verify(c.Params("token"))
	if !ok {
		return fiber.NewError(fiber.StatusNotFound, "link expired, reload the page")
	}
	name := strings.TrimPrefix(path.Clean("/"+c.Params("*")), "/")
	if name == "" {
		name = "index.html"
	}

	jar, err := javadoc.OpenJar(repoFS, jarPath)
	if errors.Is(err, fs.ErrNotExist) {
		return fiber.ErrNotFound
	}
	if err != nil {
		log.Printf("Opening javadoc %s failed: %v", jarPath, err)
		return fiber.ErrInternalServerError
	}
	defer jar.Close()

	f, err := jar.Open(name)
	if err != nil {
		return fiber.ErrNotFound
	}
	defer f.Close()
	if info, err := f.Stat(); err != nil || info.IsDir() {
		return fiber.ErrNotFound
	}
	data, err := io.ReadAll(f)
	if err != nil {
		log.Printf("Reading %s from javadoc %s failed: %v", name, jarPath, err)
		return fiber.ErrInternalServerError
	}

	contentType := mime.TypeByExtension(path.Ext(name))
	if contentType == "" {
		contentType = fiber.MIMEOctetStream
	}
	if strings.HasPrefix(contentType, "text/html") {
		data = injectAssets(data)
	}

	c.Set(fiber.HeaderContentType, contentType)
	c.Set(fiber.HeaderContentSecurityPolicy, sandbox)
	c.Set(fiber.HeaderXContentTypeOptions, "nosniff")
	// The token in the URL must not reach the sites javadoc links to
	c.Set(fiber.HeaderReferrerPolicy, "no-referrer")
	c.Set(fiber.HeaderCacheControl, "private, max-age=3600")
	return c.Send(data)
}

func injectAssets(html []byte) []byte {
	loc := headEnd.FindIndex(html)
	if loc == nil {
		return html
	}
	var b bytes.Buffer
	b.Grow(len(html) + len(injected))
	b.Write(html[:loc[0]])
	b.WriteString(injected)
	b.Write(html[loc[0]:])
	return b.Bytes()
}
