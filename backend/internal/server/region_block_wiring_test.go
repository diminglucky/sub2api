//go:build unit

package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// 地区访问限制只有在 SetupRouter 注册中间件后才会生效：配置与中间件单测都通过，
// 但漏掉这一行时功能会静默失效（曾经在合并 upstream 时被覆盖掉）。
// 这里在源码层守住接线，因为完整构建 SetupRouter 需要整套依赖注入图。
func TestSetupRouterRegistersRegionBlockMiddleware(t *testing.T) {
	source, err := os.ReadFile("router.go")
	require.NoError(t, err)
	require.Contains(t, string(source), "middleware2.RegionBlock(cfg.RegionBlockSettings)")
}
