//go:build !nvenc && !vaapi

package transcoder

// 默认版（无 build tag）：软编 libx265。
// 编译: go build -o transcoder ./cmd/transcoder

func hwName() string  { return "soft" }
func hwAvailable() bool { return false }

func hwBuildArgs(srcFile, scale, crf, vaapiDev string, audio, hls []string) []string {
	return softBuildArgs(srcFile, scale, crf, audio, hls)
}
