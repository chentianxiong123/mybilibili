//go:build nvenc

package transcoder

// NVIDIA NVENC 版。
// 编译: go build -tags nvenc -o transcoder-nvenc ./cmd/transcoder
// 部署机有 NVIDIA GPU 时用这个版本。
// hevc_nvenc 不需要 hwupload，编码器自己处理 GPU 上传。
// -preset p1 最快，-cq 恒定质量(0-51，越低越高)。

func hwName() string  { return "nvenc" }
func hwAvailable() bool { return true }

func hwBuildArgs(srcFile, scale, crf, vaapiDev string, audio, hls []string) []string {
	return append([]string{
		"-i", srcFile,
		"-vf", "scale=" + scale,
		"-c:v", "hevc_nvenc", "-preset", "p1", "-tag:v", "hvc1", "-cq", crf,
	}, append(audio, hls...)...)
}
