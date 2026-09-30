//go:build unit

package server

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// SuperAI 的自定义功能通过少量"接线点"挂进上游代码：把自定义中间件、路由注册器
// 与 provider set 接进上游的文件。上游合并很容易把这些行冲掉——历史上
// router.go 的地区拦截注册就被冲掉过，admin.go / public.go 的自定义路由注册
// 也曾静默失效（接口 404 但没有任何编译或测试报错）。
//
// 这些接线点所在的文件都很大，无法在单测里完整构建路由器，所以在源码层守住它们。
func TestCustomFeatureWiringIsPresent(t *testing.T) {
	cases := []struct {
		file string
		want string
	}{
		{"router.go", "middleware2.RegionBlock(cfg.RegionBlockSettings)"},
		{"router.go", "routes.RegisterPublicRoutes(v1, h)"},
		{"routes/admin.go", "registerCustomAdminRoutes(admin, h)"},
		{"routes/admin.go", "registerCustomAdminSettingsRoutes(adminSettings, h)"},
		{"routes/user.go", "registerCustomUserRoutes(authenticated, h)"},
		{"routes/public.go", "registerCustomPublicRoutes(public, h)"},
	}

	for _, tc := range cases {
		source, err := os.ReadFile(tc.file)
		require.NoError(t, err, tc.file)
		require.Contains(t, string(source), tc.want,
			"%s 必须保留自定义接线 %q；若上游合并冲掉了它，请补回而不是删掉断言", tc.file, tc.want)
	}
}
