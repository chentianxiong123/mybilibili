//go:build vaapi

package transcoder

// AMD/Intel VAAPI 版。
// 编译: go build -tags vaapi -o transcoder-vaapi ./cmd/transcoder
// 开发机(AMD Radeon)用这个版本。
// hevc_vaapi 需要 hwupload + vaapi_device 指定 /dev/dri/renderD128。

func hwName() string  { return "vaapi" }
func hwAvailable() bool { return true }

func hwBuildArgs(srcFile, scale, crf, vaapiDev string, audio, hls []string) []string {
	return append([]string{
		"-i", srcFile,
		"-vf", "scale=" + scale + ",format=nv12,hwupload",
		"-vaapi_device", vaapiDev,
		"-c:v", "hevc_vaapi", "-tag:v", "hvc1", "-rc_mode", "CQP", "-qp", crf,
	}, append(audio, hls...)...)
}
