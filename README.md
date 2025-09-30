# 扫描仪控制面板

这是一个基于Go语言和Gin框架开发的扫描仪控制面板，提供了现代化的Web界面来控制扫描仪设备。项目采用模块化架构设计，具备完整的扫描、预览和管理功能。可以用在arm、X86中用于直接调用USB扫描仪获取JPG格式图片，理论上支持兄弟和联想的打印扫描一体机，我测试过联想M7206。

## 🙏 致谢 / Acknowledgments

本项目在Brother扫描仪通信协议的实现上参考了 [brscanner](https://github.com/maciekmm/brscanner) 项目的优秀工作。感谢 [@maciekmm](https://github.com/maciekmm) 提供的Brother打印机通信协议实现，为本项目的USB通信功能奠定了重要基础。

This project references the excellent work of the [brscanner](https://github.com/maciekmm/brscanner) project for Brother scanner communication protocol implementation. Special thanks to [@maciekmm](https://github.com/maciekmm) for providing the Brother printer communication protocol implementation, which laid an important foundation for the USB communication functionality of this project.

特别感谢 / Special Thanks:
- **brscanner项目 / brscanner Project**: 提供了Brother设备通信协议的参考实现 / Provided reference implementation for Brother device communication protocols
- **协议解析 / Protocol Analysis**: 帮助理解USB扫描仪的命令格式和数据传输机制 / Helped understand USB scanner command formats and data transmission mechanisms
- **开源精神 / Open Source Spirit**: 为扫描仪设备的开源控制软件生态做出了贡献 / Contributed to the open source scanner device control software ecosystem

## ✨ 功能特性

### 🖥️ 用户界面
- **现代化UI设计**: 响应式布局，支持桌面和移动设备
- **美观的视觉效果**: 渐变背景、卡片设计、悬停动画
- **直观的操作界面**: 清晰的功能分区和状态提示

![pic](./doc/page-01.png)

### 🔌 设备管理
- **自动设备检测**: 支持USB扫描仪设备的自动识别
- **设备选择**: 可视化设备列表，支持多设备切换
- **设备信息显示**: 显示设备名称、厂商ID和产品ID

### ⚙️ 扫描功能
- **灵活的参数配置**: 支持DPI、扫描模式、尺寸等参数调整
- **实时扫描进度**: 带动画的进度条和状态显示
- **智能默认设置**: 预设常用的扫描参数

### 🖼️ 图像预览与操作
- **高质量图像预览**: Canvas渲染，支持高分辨率图像显示
- **完整的缩放功能**: 滚轮缩放、按钮缩放、智能适配视图
- **流畅的拖拽操作**: 支持鼠标和触摸拖拽移动图像
- **完美居中显示**: 图像自动居中，比例准确

### 📚 历史记录管理
- **扫描历史记录**: 自动保存扫描结果和参数信息
- **缩略图预览**: 历史记录的可视化缩略图展示
- **一键查看**: 快速预览历史扫描结果
- **批量清理**: 支持清空所有历史记录和文件

### 💾 数据管理
- **本地设置保存**: 自动保存用户偏好和设备配置
- **文件下载**: 支持扫描结果的直接下载
- **附件管理**: 服务端文件存储和管理

### 📱 响应式支持
- **多设备适配**: 自适应桌面、平板和手机屏幕
- **触摸友好**: 支持触摸设备的手势操作
- **性能优化**: 流畅的动画和交互体验

## 🛠️ 技术栈

### 后端架构
- **Go语言**: 高性能的系统级编程语言
- **Gin框架**: 轻量级HTTP Web框架
- **gousb库**: USB设备通信库
- **Go Modules**: 现代化的依赖管理

### 前端技术
- **HTML5**: 语义化标记和Canvas画布
- **CSS3**: 现代化样式设计（渐变、动画、响应式）
- **Vanilla JavaScript**: 原生JS实现，无第三方依赖
- **设计模式**: 单例、策略、观察者等模式

### 架构特点
- **模块化设计**: 清晰的代码分层和职责分离
- **RESTful API**: 标准化的HTTP接口设计
- **响应式布局**: 适配多种设备屏幕
- **状态管理**: 统一的前端状态管理机制

## 安装和运行

### 环境要求

- Go 1.24 或更高版本
- libusb开发库（用于USB通信）

在Ubuntu/Debian上安装libusb：
```bash
sudo apt-get update
sudo apt-get install libusb-1.0-0-dev pkg-config
```

在CentOS/RHEL上安装libusb：
```bash
sudo yum install libusbx-devel pkgconfig
```

在macOS上安装libusb：
```bash
brew install libusb
```

在Windows上，需要安装libusb-win32或使用WSL。

### 安装步骤

1. 克隆项目代码：
   ```bash
   git clone https://github.com/archfish/scanner
   cd scanner
   ```

2. 安装依赖：
   ```bash
   go mod tidy
   ```

3. 运行应用：
   ```bash
   go run main.go
   ```

   或者指定端口运行：
   ```bash
   go run main.go 5050
   ```

4. 访问应用：
   在浏览器中打开 `http://localhost:5050`

## 项目结构

```
scanner/
├── go.mod                  # Go模块定义
├── go.sum                  # Go模块校验和
├── cmd/
│   └── main.go             # 应用入口
├── src/
│   ├── scanner/            # 扫描仪核心逻辑
│   │   ├── device_state.go # 扫描仪状态数据
│   │   ├── device.go       # USB设备描述
│   │   ├── option.go       # 扫描选项定义
│   │   ├── protocol.go     # 扫描协议
│   └── web/                # Web相关代码
│       ├── api.go          # API基础功能
│       ├── request.go      # API请求参数
│       ├── scanner.go      # 扫描相关接口
│       ├── server.go       # Web服务器
│       └── index.html      # 前端页面
└── web/                    # 静态资源目录（可选）
```

## 📖 使用说明

### 基本操作流程

1. **启动应用**: 运行服务后访问Web界面 `http://localhost:5050`
2. **选择设备**: 从设备列表中选择要使用的扫描仪
3. **配置参数**: 根据需要调整DPI、尺寸等扫描参数
4. **开始扫描**: 点击"开始扫描"按钮执行扫描任务
5. **查看结果**: 在预览区域查看扫描结果，支持缩放和拖拽
6. **下载文件**: 点击下载按钮保存扫描结果到本地

### 图像预览操作

- **滚轮缩放**: 在图像上滚动鼠标滚轮进行缩放
- **拖拽移动**: 按住鼠标左键拖拽图像
- **适配视图**: 点击🖼️按钮让图像完整显示并居中
- **缩放控制**: 使用🔍+和🔍-按钮进行精确缩放
- **触摸操作**: 支持移动设备的触摸拖拽

### 历史记录管理

- **自动保存**: 每次扫描结果自动添加到历史记录
- **快速预览**: 点击历史记录中的"查看"按钮快速加载
- **批量清理**: 使用"清空"按钮删除所有历史记录和文件

## 🔧 API接口

### 获取设备列表
```http
GET /api/devices

Response:
{
  "Code": "0",
  "Msg": "success",
  "Data": [
    {
      "Name": "扫描仪设备名",
      "VendorID": "厂商ID",
      "ProductID": "产品ID"
    }
  ]
}
```

### 执行扫描任务
```http
POST /api/scan
Content-Type: application/json

{
  "device": {
    "Name": "扫描仪名称",
    "VendorID": "厂商ID",
    "ProductID": "产品ID"
  },
  "option": {
    "DPI": 400,
    "Mode": "CGRAY",
    "Top": 0,
    "Left": 0,
    "Width": 211.881,
    "Height": 355.567
  }
}

Response:
{
  "Code": "0",
  "Msg": "success",
  "Data": {
    "URL": "/api/attachments/scan_result.jpg"
  }
}
```

### 清空附件文件
```http
DELETE /api/attachments

Response:
{
  "Code": "0",
  "Msg": "success"
}
```

### 下载扫描结果
```http
GET /api/attachments/{filename}
```

## USB扫描仪支持

### 支持的设备

当前版本支持联想M7206扫描仪，可通过USB接口进行通信。

### USB通信实现

使用 `github.com/google/gousb` 库实现USB通信：
- 自动检测联想M7206设备（VendorID: 17ef, ProductID: 5629）
- 配置设备接口和端点
- 发送扫描命令并接收扫描数据

## 👨‍💻 开发指南

### 代码架构说明

#### 前端架构（设计模式）
- **DOMManager**: 单例模式管理DOM元素
- **StateManager**: 单例模式管理应用状态
- **DeviceManager**: 策略模式处理设备操作
- **ScanManager**: 命令模式处理扫描任务
- **ImageManager**: 观察者模式处理图像显示
- **CanvasController**: 责任链模式处理画布交互
- **UIManager**: 外观模式统一UI操作

#### 后端架构
- **模块化设计**: 清晰的包结构分离
- **接口抽象**: 便于扩展不同的扫描仪设备
- **错误处理**: 统一的错误响应格式
- **中间件支持**: CORS、日志等中间件

### 添加新功能

#### 1. 扫描参数扩展
```go
// 在 src/scanner/options.go 中添加新参数
type ScanOption struct {
    DPI    int     `json:"DPI"`
    Mode   string  `json:"Mode"`
    // 添加新参数
    Brightness int `json:"Brightness"`
    Contrast   int `json:"Contrast"`
}
```

#### 2. 新设备支持
```go
// 在 src/scanner/device.go 中添加设备定义
var SupportedDevices = []DeviceInfo{
    {VendorID: 0x17ef, ProductID: 0x5629}, // 联想M7206
    {VendorID: 0x1234, ProductID: 0x5678}, // 新设备
}
```

#### 3. 前端功能扩展
```javascript
// 在相应的管理器类中添加新方法
class NewFeatureManager {
    static handleNewFeature() {
        // 实现新功能逻辑
    }
}
```

### 前端定制指南

#### 样式修改
- **颜色主题**: 修改CSS变量定制主题色彩
- **布局调整**: 调整Grid和Flexbox布局
- **动画效果**: 自定义CSS动画和过渡效果

#### 功能扩展
- **新UI组件**: 遵循现有的卡片设计风格
- **交互增强**: 基于现有的事件处理模式
- **状态管理**: 通过StateManager统一管理状态

### 测试和调试

#### 前端调试
```javascript
// 开启调试模式
const DEBUG_MODE = true;

// 调试特定管理器
console.log(new StateManager().state);
```

#### 后端调试
```go
// 添加调试日志
log.Printf("扫描参数: %+v", scanOption)
```

### 部署建议

#### 开发环境
```bash
# 启用热重载
go install github.com/cosmtrek/air@latest
air
```

#### 生产环境
```bash
# 编译优化
go build -ldflags="-s -w" -o scanner main.go

# 使用反向代理
# 配置Nginx或Caddy作为反向代理
```

## ⚠️ 注意事项

### 开发环境
1. **USB权限**: 开发时可能需要sudo权限访问USB设备
2. **跨域问题**: 本地开发时注意CORS配置
3. **端口占用**: 确保5050端口未被其他应用占用

### 生产环境
1. **权限管理**: 避免使用root权限运行服务
2. **文件存储**: 配置适当的文件存储路径和权限
3. **安全考虑**: 添加必要的身份验证和授权机制
4. **性能优化**: 考虑文件大小限制和并发处理

### 硬件兼容性
1. **设备驱动**: 确保系统已安装相应的USB驱动
2. **USB协议**: 不同厂商的设备可能需要不同的通信协议
3. **测试验证**: 在目标硬件环境中充分测试

### 扩展性考虑
1. **设备插件化**: 考虑将设备支持做成插件架构
2. **配置外部化**: 将设备配置和参数外部化存储
3. **API版本控制**: 为API接口添加版本控制机制

## 家境清寒欢迎支持

![觉得有用支持一下吧](./doc/donation.png)

## 许可证

MIT License
