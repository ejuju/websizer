package httputils

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMeasure(t *testing.T) {
	t.Parallel()

	httpTestFS := fstest.MapFS{
		// Sample HTML page that imports some assets
		"index.html": &fstest.MapFile{
			// 118 bytes (without counting tabs and line returns)
			Data: []byte(strings.NewReplacer("\t", "", "\n", "").Replace(`
			<html>
				<head>
					<link href="index.css" />
					<link href="index.js" />
					<link href="./favicon.ico" />
				</head>
				<h1>Hello</h1>
			</html>
			`)),
		},
		// Sample CSS
		"index.css": &fstest.MapFile{
			Data: []byte("body{overflow-x:hidden}"), // 23 bytes
		},
		// Sample JS
		"index.js": &fstest.MapFile{
			Data: []byte("const yes=true;"), // 15 bytes
		},
		// Sample favicon
		"favicon.ico": &fstest.MapFile{
			Data: []byte{0xFF, 0xFF, 0xFF, 0xFF}, // 4 bytes
		},
	}
	httpTestServer := httptest.NewServer(http.FileServer(http.FS(httpTestFS)))
	defer httpTestServer.Close()

	t.Run("measures page size correctly", func(t *testing.T) {
		size, err := GetPageSize(httpTestServer.URL + "/")
		if err != nil {
			t.Fatal(err)
		}
		wantTotalSize := 118 + 23 + 15 + 4
		gotTotalSize := size.Total()
		if gotTotalSize != wantTotalSize {
			t.Fatalf("want %d but got %d", wantTotalSize, gotTotalSize)
		}
	})
}
