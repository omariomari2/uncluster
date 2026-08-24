package converter

import (
	"strings"
	"testing"
)

func convertForTest(t *testing.T, source string) string {
	t.Helper()

	got, err := ConvertToJSX(source, "", "", nil, nil)
	if err != nil {
		t.Fatalf("ConvertToJSX() error = %v", err)
	}
	return got
}

func TestConvertToJSXWrapsInlineStyleInAJSXObject(t *testing.T) {
	got := convertForTest(t, `<div style="color: red; font-size: 12px"></div>`)
	want := `style={{color: 'red', fontSize: '12px'}}`

	if !strings.Contains(got, want) {
		t.Fatalf("ConvertToJSX() style output missing %q; got:\n%s", want, got)
	}
}

func TestConvertToJSXTreatsPresenceOnlyBooleanAttributesAsTrue(t *testing.T) {
	got := convertForTest(t, `<input disabled checked>`)

	for _, want := range []string{"disabled={true}", "checked={true}"} {
		if !strings.Contains(got, want) {
			t.Errorf("ConvertToJSX() boolean output missing %q; got:\n%s", want, got)
		}
	}
}

func TestConvertToJSXEscapesAttributeAndTextSyntax(t *testing.T) {
	got := convertForTest(t, `<div title="say &quot;hi&quot; &amp; {x}">Hello {name} &amp; goodbye</div>`)

	for _, want := range []string{
		`title="say &quot;hi&quot; &amp; &#123;x&#125;"`,
		`Hello &#123;name&#125; &amp; goodbye`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("ConvertToJSX() escaped output missing %q; got:\n%s", want, got)
		}
	}
}

func TestConvertToJSXPreservesSpacesAroundInlineElements(t *testing.T) {
	got := convertForTest(t, `<p>Hello <strong>world</strong> again</p>`)
	want := `<p>Hello <strong>world</strong> again</p>`

	if !strings.Contains(got, want) {
		t.Fatalf("ConvertToJSX() inline text missing %q; got:\n%s", want, got)
	}
}
