package transcoder

// softBuildArgs 构建 libx265 软编参数（所有版本共用，用于硬编失败回退）。
func softBuildArgs(srcFile, scale, crf string, audio, hls []string) []string {
	return append([]string{
		"-i", srcFile,
		"-vf", "scale=" + scale,
		"-c:v", "libx265", "-preset", "veryfast", "-tag:v", "hvc1", "-crf", crf,
	}, append(audio, hls...)...)
}
