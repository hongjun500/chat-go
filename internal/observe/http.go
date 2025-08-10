package observe

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed static
var embeddedStatic embed.FS

// StartHTTP 启动一个最简 HTTP 服务，提供 /healthz，并托管 / 静态页面
func StartHTTP(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintln(w, "ok")
	})
	mux.Handle("/metrics", promhttp.Handler())
	// 静态资源（用于 WS 测试页面）
	sub, _ := fs.Sub(embeddedStatic, "static")
	mux.Handle("/", http.FileServer(http.FS(sub)))
	return http.ListenAndServe(addr, mux)
}
