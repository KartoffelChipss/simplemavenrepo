package badge

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestRenderIsValidXML(t *testing.T) {
	svg := Render(`<label & "quotes">`, "1.0.0-SNAPSHOT", ColorOrange)
	d := xml.NewDecoder(strings.NewReader(string(svg)))
	for {
		_, err := d.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("invalid XML: %v\n%s", err, svg)
		}
	}
	if !strings.Contains(string(svg), "&lt;label &amp; &#34;quotes&#34;&gt;") {
		t.Errorf("label not escaped: %s", svg)
	}
}

func TestTextWidthGrowsWithText(t *testing.T) {
	if textWidth("1.0.0") >= textWidth("10.20.30") {
		t.Errorf("expected longer text to be wider")
	}
	if textWidth("") != 0 {
		t.Errorf("empty text should have no width")
	}
}
