# Just Talk - Windows 使用与安装指南

本指南将帮助您在 Windows 系统上安装、配置并使用 **Just Talk** 语音输入工具。

---

## 1. 准备工作与依赖安装

Just Talk 在 Windows 系统上录音需要调用 **FFmpeg** 或 **SoX** 工具。我们强烈推荐安装 **FFmpeg**，因为它最稳定且支持默认音频设备的自动识别。

### 1.1 安装 FFmpeg（二选一）

#### 方法 A：使用 Windows 官方包管理器（最推荐，最简单）
如果您使用的是 Windows 10/11，系统自带了 `winget` 包管理器。请打开 **PowerShell** 运行以下命令：
```cmd
winget install Gnu.FFmpeg
```
安装完成后，**重启终端**，运行以下命令验证是否安装成功：
```cmd
ffmpeg -version
```
如果能输出版本信息，说明安装并自动配置环境变量成功。

#### 方法 B：官网直接下载并手动配置
1. 打开 FFmpeg 官方推荐的 Windows 构建包下载网站：[gyan.dev FFmpeg Builds](https://www.gyan.dev/ffmpeg/builds/)。
2. 在 **Git Master Builds** 或 **Release Builds** 下，下载 `ffmpeg-git-full.7z` 或 `ffmpeg-release-full.7z` 压缩包。
3. 解压下载的压缩包到您的电脑（例如 `C:\Program Files\ffmpeg`）。
4. 将解压出来的 `bin` 目录路径（例如 `C:\Program Files\ffmpeg\bin`）添加到系统的 **PATH 环境变量** 中：
   - 右键点击“此电脑” -> “属性” -> “高级系统设置” -> “环境变量”。
   - 在“系统变量”中找到 `Path`，双击编辑，点击“新建”，将 `bin` 路径复制进去，一路点击确认保存。
5. 重启 CMD 或 PowerShell，输入 `ffmpeg -version` 检查是否生效。

---

## 2. 获取并构建 Just Talk

如果您已经安装了 Go 语言开发环境，可以直接编译生成可执行文件：

```bash
# 进入项目目录
cd just-talk-go

# 编译 Windows 平台二进制文件 (纯 Go 实现，无需 CGO 编译器)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o build/just-talk.exe ./cmd/just-talk
```
编译完成后，生成的 `just-talk.exe` 将位于 `build/` 目录下。

---

## 3. 配置 Just Talk

在第一次启动或使用前，您需要在 Windows 上配置火山引擎（字节跳动）的 ASR 密钥。

### 3.1 配置文件路径
Windows 下的配置文件默认存放路径为：
```text
C:\Users\您的用户名\.config\just-talk\config.toml
```
如果该文件夹不存在，您可以手动创建 `.config\just-talk` 目录，并在其中新建一个 `config.toml` 文件。

### 3.2 配置文件模板
在 `config.toml` 中写入以下配置，并替换为您自己的火山引擎 ASR 密钥和大模型 API 密钥：

```toml
[voice]
enabled = true
# 运行模式：toggle (按一下开始，再按一下停止) 或 hold (按住录音，松开停止)
mode = "toggle"
# 触发快捷键 (Win 键在配置中写为 Super)
push_to_talk = "Alt+Super"
# 增益大小 (声音小可以适当调大，如 2 或 3)
gain = 1
# 自动上屏（自动模拟键盘粘贴到当前焦点输入框）
auto_submit = true
# 火山引擎 (Volcengine) ASR 配置
app_key = "您的火山引擎AppKey"
access_key = "您的火山引擎AccessKey"
resource_id = "volc.bigasr.sauc.duration"
language = "zh-CN"
# 自定义识别热词
hotwords = ["Windows", "PowerShell", "just-talk", "CMD"]

# 语音大模型优化配置 (火山方舟 - 豆包大模型)
[voice.correction]
# 是否开启语音大模型润色与纠错
enabled = true
# 火山方舟 API Key / 密钥 (通常以 "Bearer " 开头，或直接填入 ApiKey)
api_key = "您的火山方舟API密钥"
# 接入点 ID (Endpoint ID，如 ep-xxxxxxxx-xxxx)
endpoint_id = "您的火山方舟推理接入点ID"
# 火山方舟 chat completions 接口地址，默认为北京节点，无需修改
url = "https://ark.cn-beijing.volces.com/api/v3/chat/completions"
# 模型生成温度，默认 0.1 保证极低随机性，避免大模型“胡言乱语”
temperature = 0.1
# 最大 Token 限制，默认 0 代表根据输入文本长度自动动态计算
max_tokens = 0

[overlay]
# 开启状态提示胶囊（悬浮窗）
enabled = true
# 悬浮窗显示位置，默认为底部居中 (bottom-center)，在 Windows 下会自动避开任务栏
position = "bottom-center"
# 空闲时是否显示悬浮窗 (false 为仅在录音/优化/出错时显示，true 为始终显示)
idle_visible = false
# 悬浮窗缩放比例 (默认 1.0)
scale = 1.0
```

---

## 4. 运行与使用

### 4.1 环境健康度检查 (Doctor)
启动前可以使用以下命令检查录音命令和配置是否正常：
```cmd
just-talk.exe --doctor
```
如果提示 `✓ 录音工具安装` 且没有 `✗` 的错误，说明环境完全正常。

### 4.2 启动程序

#### 模式一：TUI 交互界面模式（默认）
直接在 PowerShell 或 CMD 中运行：
```cmd
just-talk.exe
```
程序会以精致的终端界面形式运行，实时展示状态、历史统计字数、实时听写的内容。退出时按 `q` 或 `Ctrl+C`。

#### 模式二：后台静默模式（推荐日常使用）
如果您不希望看到任何多余窗口，只想让它在后台默默提供语音输入服务，可以运行：
```cmd
just-talk.exe --no-tui
```

### 4.3 语音输入操作与特色效果
1. 打开任何文本输入框（如 PowerShell 窗口、CMD 窗口、微信输入框或浏览器输入框），确保光标处于聚焦闪烁状态。
2. 按下全局快捷键：`Alt+Win`（如果您配置的是 `Alt+Super`）。此时如果开启了悬浮窗，状态胶囊会显示为 **连接中...** 随后变为 **录音中...**（带有绿色的状态指示小圆点）。
3. 开始说话。在说话的同时，ASR 会将实时识别出的文字**实时流式上屏**到当前光标所在位置。
4. 说话完毕后，再次按下 `Alt+Win`（如果是 `toggle` 模式）或者松开按键（如果是 `hold` 模式）。
5. 此时：
   - **若未开启大模型优化**：最后一批文字将直接上屏并自动复制到剪贴板，语音输入结束，状态胶囊隐藏。
   - **若开启了大模型优化**：状态胶囊会转为蓝色指示的 **优化中...** 状态。系统将自动向光标处模拟发送对应次数的退格键（Backspace）清空刚才流式上屏的临时文本，并同时将整段文本提交给火山方舟大模型（豆包）进行润色与纠错。大模型返回优化后完美的文本后，一次性粘贴上屏，并写入剪贴板。整个过程仅需数百毫秒，保证最终录入语句通顺无错别字。

### 4.4 临时快捷键
- **取消录音**：在说话录音过程中，如果不想输入了，直接按键盘上的 `Esc` 键，即可立即取消本次录音，且不会输出任何文本。
- **一键重试**：如果发生网络连接断开或 ASR 引擎超时报错，按键盘上的 `R` 键可以一键清除错误并重新开始新一轮录制。

---

## 5. 网络连通性测试 (火山引擎 ASR 连接校验)

在配置火山引擎参数前，如果想验证您的 Windows 电脑是否能正常访问火山引擎的 ASR 服务器，可以使用以下方法在 **PowerShell** 或 **CMD** 中进行测试。

### 5.1 ASR 服务的目标地址
Just Talk 使用的 API WebSocket 地址为：
`wss://openspeech.bytedance.com/api/v3/sauc/bigmodel_async`

对应的 HTTPS 握手测试地址为：
`https://openspeech.bytedance.com/api/v3/sauc/bigmodel_async`

### 5.2 验证方法一：使用 PowerShell 测试 TCP 443 端口（最直观）
由于 WebSocket 协议是基于 TCP 并复用 HTTPS 的 `443` 端口进行通信的，我们可以直接测试与该域名的 443 端口连接性。

打开 **PowerShell** 运行：
```powershell
Test-NetConnection -ComputerName openspeech.bytedance.com -Port 443
```

**预期输出：**
观察输出结果中的 `TcpTestSucceeded`。如果显示 `True`，说明您与火山引擎服务器之间的 TCP 网络通路畅通：
```text
ComputerName     : openspeech.bytedance.com
RemoteAddress    : ...
RemotePort       : 443
TcpTestSucceeded : True
```

### 5.3 验证方法二：使用 `curl` 测试 SSL 握手与 TLS 通道
我们可以发起一个简单的 HTTP 头请求，测试 Windows 客户端与火山服务器之间的 SSL/TLS 握手以及证书信任链是否正常。

打开 **PowerShell** 或 **CMD** 运行：
```cmd
curl -I https://openspeech.bytedance.com/api/v3/sauc/bigmodel_async
```

**预期结果：**
因为该接口只接受 WebSocket 握手并且需要鉴权，直接请求会返回 `HTTP/1.1 404 Not Found` 或 `HTTP/1.1 400 Bad Request` 错误响应。**只要接收到类似下方的 HTTP 状态返回，就证明 SSL 握手完全成功，网络处于畅通状态**：
```text
HTTP/1.1 404 Not Found
Server: Tengine
Date: ...
Connection: keep-alive
```
*注：如果命令挂起超时，说明被局域网防火墙过滤；如果提示证书过期或不可信，说明本机的系统根证书链需要更新。*

---

## 6. Windows 平台特色高级功能详解

### 6.1 实时流式上屏 (Streaming Typer)
在 Windows 下，当 `auto_submit = true` 时，系统在录音过程中会启动一个实时的输入字符增量计算机制。每当 ASR 语音识别返回确定的文本切片时，系统仅将新识别的“字词增量”（Rune Delta）通过模拟键盘事件追加输入到当前光标闪烁位置，而不是等待录音彻底结束才把完整段落贴上去。这提供了如同本地中文输入法般的极其流畅、行云流水的打字体验。

### 6.2 语音大模型优化 (Doubao LLM Post-Processing Correction)
如果只依赖 ASR（语音转文字），经常会出现错别字、标点缺失、口语语气词（如“呃”、“啊”、“然后”）过多的问题。为了达到极致的上屏效果，Just Talk 引入了语音后处理优化机制：
- **原理**：录音结束时，系统首先统计刚才实时流式上屏的字符总数，然后以极快的速度模拟发送对应数量的 `Backspace` 键，将临时文字全部清空。与此同时，在后台异步发起对火山方舟大模型（豆包 LLM）的 HTTPS 请求。
- **参数控制**：
  - `temperature`：建议固定为 `0.1`，以收紧模型输出，彻底杜绝大模型的自由发挥与幻觉，仅作错字、标点与口语整理。
  - `max_tokens`：若为 `0`，系统会根据输入长度自动调整（`len * 2 + 100`），保证长难句也能完整输出且不会额外浪费 Token。
- **完美上屏**：收到优化结果后，系统会直接将其通过剪贴板或模拟粘贴的方式写入到光标处，同时覆盖最新的剪贴板。

### 6.3 磨砂半透明状态胶囊悬浮窗 (Translucent Window Overlay)
在 Windows 平台下，状态胶囊已全面支持原生悬浮窗：
- **磨砂半透明质感**：窗口采用 Windows 原生的分层窗口（Layered Window）技术实现，配色呈暗色毛玻璃半透明风格，背景不会遮挡屏幕后方的文字和内容，具有出色的视觉精致度。
- **全中文状态**：悬浮窗左侧配备了不同颜色指示的状态点，文字提示全面中文化：
  - `连接中...` (黄色指示点)
  - `录音中...` (绿色指示点)
  - `优化中...` (蓝色指示点，表明大模型正在润色纠错)
  - `出错了` (红色指示点)
- **防任务栏遮挡**：系统使用 Win32 API 中的 `SPI_GETWORKAREA` 动态获取当前 Windows 桌面的可用工作区（已扣除系统任务栏的高度和位置），在计算悬浮窗坐标时自动向上偏移，完美解决了状态胶囊被任务栏挡住的问题。
- **配置热加载**：在程序运行过程中，如果您通过编辑配置文件将 `[overlay].enabled` 改为 `false`，状态悬浮窗会立即在运行时销毁；重新改为 `true` 时也会即时重新创建，无需重启程序。

