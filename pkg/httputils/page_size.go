package httputils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"golang.org/x/net/html"
)

type PageSize map[string]int

func (ps PageSize) Total() int {
	out := 0
	for _, size := range ps {
		out += size
	}
	return out
}

func (ps PageSize) String() string {
	out := fmt.Sprintf("total page size: %.6f MB\n", float64(ps.Total())/1_000_000.0)
	for k, v := range ps {
		out += fmt.Sprintf("%12d %q\n", v, k)
	}
	return out
}

// Fetch a resource over HTTP and measure its size.
// If the resource is an HTML page, then it will add the size of the linked CSS, JS and favicon files
// declared in the <head> section.
func GetPageSize(srcURL string) (PageSize, error) {
	out := PageSize{}

	// Parse input URL
	pageURL, err := url.Parse(srcURL)
	if err != nil {
		return nil, fmt.Errorf("parse input URL: %w", err)
	}

	// Fetch file and get content length
	resp, err := http.Get(pageURL.String())
	if err != nil {
		return nil, fmt.Errorf("fetch URL: %q: %w", pageURL, err)
	}
	defer resp.Body.Close()

	// Read response body
	rawbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body for URL %q: %w", pageURL, err)
	}
	out[srcURL] = len(rawbody)

	// Measure imported files (CSS, JS, favicon, images, etc.) size if file is HTML
	if strings.Contains(resp.Header.Get("Content-Type"), "text/html;") {
		checkTagAttrs := [][2]string{
			{"link", "href"}, // CSS, JS, favicon
			{"img", "src"},
			{"video", "poster"},
			// {"video", "src"},
			// {"source", "src"},
			// {"source", "srcset"},
		}
		importedFilesURLs := []string{}
		for _, tagAttr := range checkTagAttrs {
			urls, err := extractHTMLTagAttribute(bytes.NewBuffer(rawbody), tagAttr[0], tagAttr[1])
			if err != nil {
				return nil, fmt.Errorf("extract %q %q on page %q: %w", tagAttr[0], tagAttr[1], pageURL, err)
			}
			importedFilesURLs = append(importedFilesURLs, urls...)
		}
		for _, linkURL := range importedFilesURLs {
			parsedLinkURL, err := url.Parse(linkURL)
			if err != nil {
				return nil, fmt.Errorf("parse %q: %w", linkURL, err)
			}
			linkDomainName := parsedLinkURL.Scheme + "://" + parsedLinkURL.Host
			linkURI := parsedLinkURL.Path
			// Re-write values if link URL is relative
			if !parsedLinkURL.IsAbs() {
				linkDomainName = pageURL.Scheme + "://" + pageURL.Host
				linkURI = path.Clean("/" + linkURL)
			}
			size, err := GetPageSize(linkDomainName + linkURI)
			if err != nil {
				return nil, fmt.Errorf("measure link: %w", err)
			}
			out[linkDomainName+linkURI] = size.Total()
		}
	}

	return out, nil
}

func extractHTMLTagAttribute(r io.Reader, tag string, attribute string) ([]string, error) {
	tokenizer := html.NewTokenizer(r)
	out := []string{}
	fmt.Println("extracting", tag, attribute)

	// Find desired tag instances
	for {
		_ = tokenizer.Next()
		tok := tokenizer.Token()
		err := tokenizer.Err()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if tok.Data != tag {
			continue
		}

		// Find desired attribute
		for _, attr := range tok.Attr {
			if attr.Key != attribute || attr.Val == "" {
				continue
			}
			out = append(out, string(attr.Val))
		}
	}

	return out, nil
}
