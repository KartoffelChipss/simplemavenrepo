package badge

import (
	"fmt"
	"html"
	"math"
)

const (
	ColorBlue   = "#007ec6"
	ColorOrange = "#fe7d37"
	ColorGrey   = "#9f9f9f"
)

func Render(label, message, color string) []byte {
	labelText, messageText := textWidth(label), textWidth(message)
	labelWidth, messageWidth := labelText+10, messageText+10
	width := labelWidth + messageWidth
	title := html.EscapeString(label + ": " + message)
	label, message = html.EscapeString(label), html.EscapeString(message)

	return fmt.Appendf(nil, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s">`+
		`<title>%s</title>`+
		`<linearGradient id="s" x2="0" y2="100%%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>`+
		`<clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>`+
		`<g clip-path="url(#r)"><rect width="%d" height="20" fill="#555"/><rect x="%d" width="%d" height="20" fill="%s"/><rect width="%d" height="20" fill="url(#s)"/></g>`+
		`<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">`+
		`<text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>`+
		`<text x="%d" y="140" transform="scale(.1)" textLength="%d">%s</text>`+
		`<text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>`+
		`<text x="%d" y="140" transform="scale(.1)" textLength="%d">%s</text>`+
		`</g></svg>`,
		width, title, title,
		width,
		labelWidth, labelWidth, messageWidth, html.EscapeString(color), width,
		labelWidth*5, labelText*10, label,
		labelWidth*5, labelText*10, label,
		labelWidth*10+messageWidth*5, messageText*10, message,
		labelWidth*10+messageWidth*5, messageText*10, message,
	)
}

func textWidth(s string) int {
	var w float64
	for _, r := range s {
		w += runeWidth(r)
	}
	return int(math.Ceil(w))
}

func runeWidth(r rune) float64 {
	switch {
	case r == ' ', r == '.', r == ',', r == ':', r == ';', r == '|', r == '!', r == '\'':
		return 3.9
	case r == 'i', r == 'l', r == 'j':
		return 3.1
	case r == 'f', r == 't', r == 'r', r == 'I', r == '(', r == ')', r == '[', r == ']', r == '/':
		return 4.6
	case r == '-':
		return 4.6
	case r == 'm', r == 'w':
		return 10.7
	case r == 'M', r == 'W':
		return 11.7
	case r >= 'a' && r <= 'z':
		return 6.7
	case r >= 'A' && r <= 'Z':
		return 7.6
	case r >= '0' && r <= '9':
		return 7.0
	}
	return 7.5
}
