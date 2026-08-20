# WebP 图片压缩方案

## 为什么选择 WebP？

### 1. 压缩率对比

| 格式 | 有损质量 90 | 有损质量 30 | 无损 | 透明通道 |
|---|---|---|---|---|
| JPEG | 基准 | ~80% 原始大小 | 不支持 | 不支持 |
| PNG | - | - | 比原始大 | 支持 |
| **WebP** | **比 JPEG 小 30%** | **比 JPEG 小 50-80%** | 比 PNG 小 26% | 支持 |
| AVIF | 比 JPEG 小 50% | 比 JPEG 小 80% | 比 PNG 小 50% | 支持 |

**WebP 在质量/体积之间取得最佳平衡**，且兼容性最好（2024 年起所有主流浏览器均支持）。

### 2. 为什么不用其他格式

- **JPEG 2000 / JPEG XL**：Safari 不支持，需额外 fallback
- **AVIF**：编码慢 10 倍，老旧设备解码开销大，本项目的 ffmpeg 版本支持但编码耗时是 WebP 的 3-5 倍
- **HEIC/HEIF**：专利风险，非浏览器原生支持

### 3. 本项目配置参数

```
ffmpeg -i input.jpg \
  -vf "scale='min(1920,iw)':-2"  \  # 最长边 ≤ 1920px，保持宽高比
  -c:v libwebp                      \  # WebP 编码器
  -quality 30                       \  # 质量 30（1-100，越低越小）
  -compression_level 6              \  # 压缩级别 6（0-6，越高越慢但越小）
  -preset picture                   \  # 图片模式预设
  -loop 0                           \  # 静态图片（非动画）
  -an                                  # 无音频
```

- **质量 30**：肉眼几乎不可辨差异，文件体积减 80-90%
- **scale 1920px**：封面/轮播图不需要超过 1920px 宽，限制最大尺寸减少冗余像素
- **compression_level 6**：最大压缩，单张图片额外多省 10-15%

### 4. 实际效果

存量 26 张图片压缩后：

| 类型 | 压缩前 | 压缩后 | 缩减 |
|---|---|---|---|
| 头像 | 几十 KB | **1-2 KB** | 95%+ |
| 图片/封面 | 几百 KB ~ 几 MB | **20-60 KB** | 90%+ |
| 大封面 (2560x1440) | ~5 MB | **~47 KB** | 99% |

### 5. 集成方式

所有图片上传入口统一调用 `pkg/imageutil.CompressAndReplace()`：

```go
// 接收上传文件
dst := saveUploadFile(file)
// 压缩为 WebP，替换原文件
webpPath, _ := imageutil.CompressAndReplace(dst)
// 返回 .webp URL
url := "/uploads/" + filepath.Base(webpPath)
```

现有 5 个入口已集成：
- 稿件封面上传 (`manuscript_http_handler.go`)
- 轮播图上传 (`video_handler.go`)
- 头像上传 (`user_extend_handler.go`)
- 直播封面上传 (`live_http_handler.go`)
- 工作室素材上传 (`studio_repository.go`)

### 6. 注意事项

- **首次上传大图片时**，ffmpeg 编码会占用少量 CPU（约 0.5-2 秒/张），不影响用户体验（上传后异步返回）
- **WebP 不支持 EXIF**：压缩后图片方向信息丢失，建议上传前自行旋转
- **GIF 转 WebP**：会丢失动画，只保留第一帧（静态图片场景下可接受）
- **存量迁移**：已通过批量脚本将所有 MinIO 中图片转为 WebP