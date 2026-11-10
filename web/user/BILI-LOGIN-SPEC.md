# B 站登录面板特征提取（passport.bilibili.com/login）

> 方法：webclone-skill 官方 CLI（playwright adapter + 系统 Chrome + 全新空 profile **无登录态**）
> 快照：`/home/a1/bili-align/snap-login/`（24 资源，2.4MB，0 失败）
> 所有值来自快照真实 CSS（Data Before Code），非猜测。视口 1920×1080 / dpr1 / locale zh-CN。

## 页面整体布局
- `.login_wp`：**width 900px**，margin 0 auto（水平居中），**padding-top 82px**，padding-bottom 108px
- `.login__main`：display flex（左扫码区 + 中分隔线 + 右表单区 三栏）
- 页面背景：白底（`.login_mask` background #fff）

## 三栏结构
```
.login_wp (900px 居中)
└─ .login__main (flex)
   ├─ .login-scan.main__left   ← 左：扫码登录区
   ├─ .main__middle-line        ← 中：竖分隔线
   └─ .main__right              ← 右：密码/短信登录表单
```

### 左栏：扫码登录（`.login-scan.main__left`）
- `.login-scan__qrcode`（二维码框）：**158×158**，border `0.855491px solid #e3e5e7`，**border-radius 6.84393px**，padding 8，margin-bottom 8
  - 内部二维码 140×140（canvas/img）
  - `.qrcode__tip`：140×140，background `#ffffffe6`，居中 flex
  - `.qrcode__tip img`：20×20，border-radius 30px，padding 14，margin-bottom 8
  - `.qrcode__tip span`：11px/500/lh16，color #61666d
  - `.qrcode__loading`：140×140，background `#f7f7f7` + 加载 gif
- `.login-scan__tv-img`：70×44，TV 动画 gif，margin-left -80px
- `.login-scan__txt`：386×90，p 文字 14px/400/lh19 #61666d 居中
- `.login-scan__txt_bg`：386×115，2233 背景图
- `.scan-tips-icon`：386×202，扫码提示图（绝对定位 top-44 left-114）

### 中分隔线（`.main__middle-line`）
- **width 1px，height 260px**，background `#ddd`，margin `44px 48px 0`

### 右栏：登录表单（`.main__right`）
- 容器：display flex，justify-content center，align-items center
- `.tabs_wp`（tab 栏）：height 20，margin-bottom 24，flex 居中
  - div：18px/500/lh100%，color `#9499a0`
  - 奇数项（可点 tab）：min-width 72，cursor pointer
  - 偶数项（分隔）：1×20 background `#ddd`，margin auto 20
  - `.tab_active`：color `#00a1d6 !important`（激活 tab 品牌蓝）
- `.tab__form`（账号密码表单）：**400×90**，border `1px solid #e3e5e7`，**border-radius 8px**，14px/400/lh20
  - `.form__item`：**100%×44**，padding `12px 20px`，flex 居中
    - div 间距 margin-right 20
    - input：maxlength 32，placeholder `请输入账号`/`请输入密码`，14px/400/lh20
    - `input::placeholder`：color `#c9ccd0`
  - `.form__separator-line`：100%×0，border-bottom `1px solid #e3e5e7`（两输入框分隔）

### 登录按钮（核心 CTA）
- `.btn_wp`（按钮容器）：**400px**，margin-top 20，flex space-between，14px/400/lh40
- `.btn_primary`（主登录按钮）：**194×40**，background `#00a1d6`，**border-radius 8px**，color `#fff`
  - `.disabled`：opacity .5，cursor not-allowed，背景 `#00a1d6 !important`
- `.btn__primary_big`：width 400px !important（大按钮变体）
- `.btn_other`（次要按钮，如"其他方式"）：194×40，background `#fff`，border `1px solid #e3e5e7`，radius 8，color `#18191c`

### 第三方登录（`.third-party-login-wrapper`）
- margin-top 24
- `.title`：13px/lh16，color `#999`，margin-bottom 12，居中
- `.sns`：flex，justify-content center，align-items flex-start
- `.sns__item`：**width 72**，padding `0 2px`，flex 居中，cursor pointer
- `.sns__icon`：**28×28**，margin-bottom 2，cover（微信/微博/QQ 各自背景图）
- `.sns__name`：12px/lh16，居中

### 协议区（`.login-protocol`）
- **margin-top 40**，12px/400/lh17，color `#9499a0`
- p：flex 居中；label：margin `0 6px`
- 文案：「未注册过哔哩哔哩的手机号，我们将自动帮你注册账号」+「登录即代表你同意 用户协议 与 隐私政策」

### 忘记密码浮层（`.forget-tip`）
- 332×min-height 120，background #fff，radius 5，box-shadow `0 1px 5px #00000036`
- 项：padding `11px 18px`，cursor pointer

## 品牌色速查
| 用途 | 值 |
|---|---|
| 主品牌蓝（激活 tab/链接/登录按钮） | `#00a1d6` |
| 备用蓝（hover/分页激活） | `#00aeec` |
| 主文字 | `#18191c` |
| 次文字/协议 | `#9499a0` |
| 弱文字/占位符 | `#c9ccd0` |
| 边框/分隔 | `#e3e5e7` |
| 分隔线灰 | `#ddd` |
| 三方标题灰 | `#999` |
| 扫码区文字 | `#61666d` |

## 关键尺寸速查
| 元素 | 尺寸 | 圆角 |
|---|---|---|
| 整个登录区 | 900px 居中 | — |
| 二维码框 | 158×158 | 6.84px |
| 二维码内芯 | 140×140 | — |
| 表单框 | 400×90 | 8px |
| 输入项 | 100%×44 | — |
| 登录按钮 | 194×40 | 8px |
| 大按钮 | 400×40 | 8px |
| 分隔竖线 | 1×260 | — |
| 三方图标 | 28×28 | — |
| 三方项 | 72 宽 | — |

## 快照质量问题
- 组件自动抽取置信度低（JS 1.9MB 被截断到 100KB），但**纯 CSS 值完整可靠**（本 spec 全部取自 CSS 真值，不依赖 JS）
- 登录态确认：全新 profile 无 cookie，页面展示「扫码登录 / 密码登录 / 短信登录」完整未登录面板
