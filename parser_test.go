// parser_test.go
package margo

import (
	"github.com/google/go-cmp/cmp"
	"github.com/iota-uz/margo/parser"
	"strings"
	"testing"
)

func TestParserBasic(t *testing.T) {
	input := `
\HeroV2
    Title: "Congratulations!"
    !Visible
    \ButtonPrimary
        Href: "https://example.com"
        See Demo
`
	nodes, err := parser.NewMargoParser(strings.TrimSpace(input)).Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 top-level node, got %d", len(nodes))
	}

	hero := nodes[0].(*parser.ComponentNode)
	if hero.Name != "HeroV2" {
		t.Errorf("Expected top-level node name to be HeroV2, got %s", hero.Name)
	}

	attrs := hero.Attributes()
	if len(attrs) != 2 {
		t.Fatalf("Expected 2 props for HeroV2, got %d", len(attrs))
	}

	if string(attrs[0].Name) != "Title" {
		t.Fatalf("Expected first prop name to be Title, got %s", attrs[0].Name)
	}

	if attrs[0].Value != "Congratulations!" {
		t.Errorf("Expected Title to be 'Congratulations!', got %v", attrs[0].Value)
	}

	if string(attrs[1].Name) != "Visible" {
		t.Fatalf("Expected second prop name to be Visible, got %s", attrs[1].Name)
	}

	if attrs[1].Value != true {
		t.Errorf("Expected Visible to be true, got %v", attrs[1].Value)
	}

	if len(hero.Children()) != 1 {
		t.Fatalf("Expected 1 child node under HeroV2, got %d", len(hero.Children()))
	}

	button := hero.Children()[0].(*parser.ComponentNode)
	if button.Name != "ButtonPrimary" {
		t.Errorf("Expected child node name to be ButtonPrimary, got %s", button.Name)
	}
	buttonAttrs := button.Attributes()
	if len(buttonAttrs) != 1 {
		t.Fatalf("Expected 1 prop for ButtonPrimary, got %d", len(buttonAttrs))
	}

	if string(buttonAttrs[0].Name) != "Href" {
		t.Fatalf("Expected prop name to be Href, got %s", buttonAttrs[0].Name)
	}

	if buttonAttrs[0].Value != "https://example.com" {
		t.Errorf("Expected Href to be 'https://example.com', got %v", buttonAttrs[0].Value)
	}

	if len(button.Children()) != 1 {
		t.Fatalf("Expected 1 text child under ButtonPrimary, got %d", len(button.Children()))
	}

	textNode := button.Children()[0].(*parser.TextNode)
	if textNode.Value != "See Demo" {
		t.Errorf("Expected text value 'See Demo', got %s", textNode.Value)
	}
}

func TestParserWithNestedComponents(t *testing.T) {
	input := `
\Features
    Title: "### A robust set of features"
    Description: "Based on community suggestions"
    \Card
        Title: "Blazing Fast Performance"
        Icon: 
            \IconLightning
                Size: "24"
                Variant: "DuoTone"
`

	nodes, err := parser.NewMargoParser(strings.TrimSpace(input)).Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 top-level node, got %d", len(nodes))
	}

	features := nodes[0].(*parser.ComponentNode)
	if features.Name != "Features" {
		t.Errorf("Expected top-level node name to be Features, got %s", features.Name)
	}

	attrs := features.Attributes()

	if len(attrs) != 2 {
		t.Fatalf("Expected 2 props for Features, got %d", len(attrs))
	}

	if string(attrs[0].Name) != "Title" {
		t.Fatalf("Expected first prop name to be Title, got %s", attrs[0].Name)
	}

	if attrs[0].Value != "### A robust set of features" {
		t.Errorf("Expected Title to be '### A robust set of features', got %v", attrs[0].Value)
	}

	if string(attrs[1].Name) != "Description" {
		t.Fatalf("Expected second prop name to be Description, got %s", attrs[1].Name)
	}

	if attrs[1].Value != "Based on community suggestions" {
		t.Errorf("Expected Description to be 'Based on community suggestions', got %v", attrs[1].Value)
	}

	if len(features.Children()) != 1 {
		t.Fatalf("Expected 1 Card child, got %d", len(features.Children()))
	}

	card := features.Children()[0].(*parser.ComponentNode)
	if card.Name != "Card" {
		t.Errorf("Expected Card node name to be Card, got %s", card.Name)
	}
	cardAttrs := card.Attributes()
	if len(cardAttrs) != 2 {
		t.Fatalf("Expected 2 props for Card, got %d", len(cardAttrs))
	}
	if string(cardAttrs[0].Name) != "Title" {
		t.Fatalf("Expected first prop name to be Title, got %s", cardAttrs[0].Name)
	}
	if cardAttrs[0].Value != "Blazing Fast Performance" {
		t.Errorf("Expected Card Title to be 'Blazing Fast Performance', got %v", cardAttrs[0].Value)
	}

	icon := cardAttrs[1].Value.(*parser.ComponentNode)
	if icon.Name != "IconLightning" {
		t.Errorf("Expected Icon node name to be IconLightning, got %s", icon.Name)
	}
	iconAttrs := icon.Attributes()
	if len(iconAttrs) != 2 {
		t.Fatalf("Expected 2 props for Icon, got %d", len(iconAttrs))
	}

	if string(iconAttrs[0].Name) != "Size" {
		t.Fatalf("Expected first prop name to be Size, got %s", iconAttrs[0].Name)
	}

	if iconAttrs[0].Value != "24" {
		t.Errorf("Expected Icon Size to be '24', got %v", iconAttrs[0].Value)
	}

	if string(iconAttrs[1].Name) != "Variant" {
		t.Fatalf("Expected second prop name to be Variant, got %s", iconAttrs[1].Name)
	}

	if iconAttrs[1].Value != "DuoTone" {
		t.Errorf("Expected Icon Variant to be 'DuoTone', got %v", iconAttrs[1].Value)
	}
}

func TestParserWithMultilineText(t *testing.T) {
	input := `
\Description
    Text:
        This is a multiline description
        that spans multiple lines.
    MoreText: "More single-line text."
`

	nodes, err := parser.NewMargoParser(input).Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 top-level node, got %d", len(nodes))
	}

	desc := nodes[0].(*parser.ComponentNode)
	if desc.Name != "Description" {
		t.Errorf("Expected top-level node name to be Description, got %s", desc.Name)
	}

	attrs := desc.Attributes()
	if len(attrs) != 2 {
		t.Fatalf("Expected 2 props for Description, got %d", len(attrs))
	}
	expectedText := "This is a multiline description\nthat spans multiple lines."
	if string(attrs[0].Name) != "Text" {
		t.Fatalf("Expected first prop name to be Text, got %s", attrs[0].Name)
	}
	if attrs[0].Value != expectedText {
		t.Errorf(
			"Mismatch (-expected +actual):\n%s",
			cmp.Diff(expectedText, attrs[0].Value),
		)
	}

	if string(attrs[1].Name) != "MoreText" {
		t.Fatalf("Expected second prop name to be MoreText, got %s", attrs[1].Name)
	}

	expectedMoreText := "More single-line text."
	if attrs[1].Value != expectedMoreText {
		t.Errorf(
			"Mismatch (-expected +actual):\n%s",
			cmp.Diff(expectedMoreText, attrs[1].Value),
		)
	}
}

func TestParserWithNavigation(t *testing.T) {
	input := `
\Header
    \Link
        Href: "/features"
        Features
    \Link
        Href: "/services"
        Services
    \Link
        Href: "/about"
        About
    \Link
        Href: "/contact"
        Contact
    \Link
        Href: "/docs"
        Docs
\Slot`

	nodes, err := parser.NewMargoParser(input).Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("Expected 2 top-level nodes, got %d", len(nodes))
	}

	// Test Header component
	header := nodes[0].(*parser.ComponentNode)
	if header.Name != "Header" {
		t.Errorf("Expected first node to be Header, got %s", header.Name)
	}

	if len(header.Attributes()) != 0 {
		t.Fatalf("Expected no props for Header, got %d", len(header.Attributes()))
	}

	// Test navigation links
	links := header.Children()
	if len(links) != 5 {
		t.Fatalf("Expected 5 Link children under Header, got %d", len(links))
	}

	// Expected link data
	expectedLinks := []struct {
		href  string
		label string
	}{
		{"/features", "Features"},
		{"/services", "Services"},
		{"/about", "About"},
		{"/contact", "Contact"},
		{"/docs", "Docs"},
	}

	for i, link := range links {
		linkNode := link.(*parser.ComponentNode)
		if linkNode.Name != "Link" {
			t.Errorf("Expected Link component at index %d, got %s", i, linkNode.Name)
		}

		attrs := linkNode.Attributes()
		if href := attrs[0].Value; href != expectedLinks[i].href {
			t.Errorf("Expected Href %s at index %d, got %s", expectedLinks[i].href, i, href)
		}

		if len(linkNode.Children()) != 1 {
			t.Fatalf("Expected 1 text child under Link at index %d, got %d", i, len(linkNode.Children()))
		}

		textNode := linkNode.Children()[0].(*parser.TextNode)
		if textNode.Value != expectedLinks[i].label {
			t.Errorf("Expected text %s at index %d, got %s", expectedLinks[i].label, i, textNode.Value)
		}
	}

	// Test Slot component
	slot := nodes[1].(*parser.ComponentNode)
	if slot.Name != "Slot" {
		t.Errorf("Expected second node to be Slot, got %s", slot.Name)
	}
}

func TestParser_ParseWithComplexProps(t *testing.T) {
	input := `
\DocsLayout
    Header:
        \Navbar
            \Link
                Href: "/features"
                Features
            \Link
                Href: "/services"
                Services
            \Link
                Href: "/blog"
                Blog
            \Link
                Href: "/contact"
                Contact
            \Link
                Href: "/docs"
                Docs
    \Slot
`

	nodes, err := parser.NewMargoParser(input).Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(nodes) != 1 {
		t.Fatalf("Expected 1 top-level node, got %d", len(nodes))
	}

	docsLayout := nodes[0].(*parser.ComponentNode)
	if docsLayout.Name != "DocsLayout" {
		t.Errorf("Expected top-level node name to be DocsLayout, got %s", docsLayout.Name)
	}

	if len(docsLayout.Attributes()) != 1 {
		t.Fatalf("Expected 1 prop for DocsLayout, got %d", len(docsLayout.Attributes()))
	}

	if len(docsLayout.Children()) != 1 {
		t.Fatalf("Expected 1 child node under DocsLayout, got %d", len(docsLayout.Children()))
	}

	attrs := docsLayout.Attributes()
	if len(attrs) != 1 {
		t.Fatalf("Expected 1 prop for DocsLayout, got %d", len(attrs))
	}
	if string(attrs[0].Name) != "Header" {
		t.Fatalf("Expected prop name to be Header, got %s", attrs[0].Name)
	}
	// Test Header component
	navbar := attrs[0].Value.(*parser.ComponentNode)
	if navbar.Name != "Navbar" {
		t.Errorf("Expected Header node name to be Navbar, got %s", navbar.Name)
	}

	if len(navbar.Children()) != 5 {
		t.Fatalf("Expected 5 Link children under Navbar, got %d", len(navbar.Children()))
	}

	if len(navbar.Attributes()) != 0 {
		t.Fatalf("Expected no props for Header, got %d", len(navbar.Attributes()))
	}

	// Test navigation links
	links := navbar.Children()
	expectedLinks := []struct {
		href  string
		label string
	}{
		{"/features", "Features"},
		{"/services", "Services"},
		{"/blog", "Blog"},
		{"/contact", "Contact"},
		{"/docs", "Docs"},
	}

	for i, link := range links {
		linkNode := link.(*parser.ComponentNode)
		if linkNode.Name != "Link" {
			t.Errorf("Expected Link component at index %d, got %s", i, linkNode.Name)
		}

		attrs := linkNode.Attributes()
		if href := attrs[0].Value; href != expectedLinks[i].href {
			t.Errorf("Expected Href %s at index %d, got %s", expectedLinks[i].href, i, href)
		}

		if len(linkNode.Children()) != 1 {
			t.Fatalf("Expected 1 text child under Link at index %d, got %d", i, len(linkNode.Children()))
		}

		textNode := linkNode.Children()[0].(*parser.TextNode)
		if textNode.Value != expectedLinks[i].label {
			t.Errorf("Expected text %s at index %d, got %s", expectedLinks[i].label, i, textNode.Value)
		}
	}
}

//func TestParserRogueCases(t *testing.T) {
//	tests := []struct {
//		name     string
//		input    string
//		wantErr  bool
//		errorMsg string
//	}{
//		{
//			name: "invalid indentation",
//			input: `
//\Header
//  \Link
//       Href: /about
//   About`,
//			wantErr:  true,
//			errorMsg: "inconsistent indentation",
//		},
//		{
//			name: "unclosed component",
//			input: `
//\Header
//    \Link
//        Href: /about
//        About
//    \AnotherLink
//        Href: /contact
//`,
//			wantErr:  true,
//			errorMsg: "unexpected EOF",
//		},
//		{
//			name: "invalid property format",
//			input: `
//\Header
//    \Link
//        Href:/about/no-space
//        About`,
//			wantErr:  true,
//			errorMsg: "invalid property format",
//		},
//		{
//			name: "duplicate properties",
//			input: `
//\Link
//    Href: /about
//    Href: /contact
//    About`,
//			wantErr:  true,
//			errorMsg: "duplicate property",
//		},
//		{
//			name: "empty component name",
//			input: `
//\
//    Title: Empty component`,
//			wantErr:  true,
//			errorMsg: "empty component name",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			_, err := NewMargoParser(tt.input).Parse()
//
//			if tt.wantErr {
//				if err == nil {
//					t.Errorf("Parse() expected error containing %q, got nil", tt.errorMsg)
//				} else if !strings.Contains(err.Error(), tt.errorMsg) {
//					t.Errorf("Parse() error %q does not contain %q", err.Error(), tt.errorMsg)
//				}
//			} else if err != nil {
//				t.Errorf("Parse() unexpected error: %v", err)
//			}
//		})
//	}
//}

func TestParserWithEmptyInput(t *testing.T) {
	input := ""
	nodes, err := parser.NewMargoParser(input).Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(nodes) != 0 {
		t.Fatalf("Expected no nodes for empty input, got %d", len(nodes))
	}
}
