package imageutil

import (
	"fmt"
	"path/filepath"
)

// 注：图片压缩已前移到前端（上传前浏览器转 WebP），后端不再依赖 ffmpeg，
// 因此容器镜像无需安装 ffmpeg（distroless 等最精简基底可直接运行）。

// CompressToWebP 原为 ffmpeg 转 WebP。前端上传前已转好 WebP，此处保留签名，
// 直接返回原路径，不再调用外部二进制。
func CompressToWebP(srcPath string) (string, error) {
	if srcPath == "" {
		return "", fmt.Errorf("imageutil: empty path")
	}
	return srcPath, nil
}

// CompressAndReplace 返回文件基本名；输入已是 WebP，无额外处理。
func CompressAndReplace(srcPath string) (string, error) {
	if srcPath == "" {
		return "", fmt.Errorf("imageutil: empty path")
	}
	return filepath.Base(srcPath), nil
}