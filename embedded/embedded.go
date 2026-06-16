package embedded

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var distFS embed.FS

// FS は埋め込まれたアセット（spa/dist以下のファイル群）を提供する http.FileSystem を返します。
func FS() http.FileSystem {
	subFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(subFS)
}
