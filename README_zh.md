# masque-vpn

基于 MASQUE (CONNECT-IP) 协议的教育用 VPN 实现。

**本项目专为教育目的和现代网络协议研究而设计。**

```bash
git clone https://github.com/iselt/masque-vpn.git
```

## 特性

- **现代协议**: 基于 QUIC 和 MASQUE CONNECT-IP (RFC 9484) 构建
- **自定义 MASQUE 实现**: 无外部依赖的自主协议实现
- **模块化架构**: 服务器分为专业化组件
- **REST API**: 用于管理和监控的全功能 API
- **双向 TLS 认证**: 基于证书的客户端-服务器认证
- **跨平台支持**: 支持 Windows、Linux 和 macOS
- **IP 池管理**: 自动客户端 IP 分配
- **Prometheus 监控**: 用于性能分析的详细指标
- **全面测试**: 单元测试、集成测试和负载测试

## 文档

为学生和研究人员提供的全面文档：

- **[学生指南 (EN)](docs/development/student-guide.md)**: 实验作业和研究项目
- **[学生指南 (RU)](docs/development/student-guide.ru.md)**: 实验作业和研究项目
- **[架构文档](docs/architecture/architecture.ru.md)**: 系统组件和数据流
- **[API 文档](docs/api/api.ru.md)**: 管理和监控 REST API
- **[入门指南](docs/development/getting-started.md)**: 构建和部署说明
- **[监控设置](docs/monitoring/setup.md)**: Prometheus 和 Grafana 配置

## 架构

系统由四个主要组件组成：
- **VPN 服务器**: 处理 MASQUE CONNECT-IP 请求和流量路由
- **VPN 客户端**: 通过 QUIC 连接服务器并隧道传输 IP 数据包
- **REST API 服务器**: 客户端管理和监控（端口 8080）
- **通用库**: 自定义 MASQUE 实现和工具

## 快速开始

### 1. 编译

```bash
# 编译服务器
cd vpn_server && go build -o vpn-server .

# 编译客户端
cd ../vpn_client && go build -o vpn-client .
```

### 2. 生成证书

```bash
cd cert

# Linux/macOS
./generate-test-certs.sh

# Windows
powershell -ExecutionPolicy Bypass -File generate-certs.ps1
```

### 3. 启动服务器（本地测试）

```bash
cd vpn_server
./vpn-server -c config.server.local.toml
```

服务器将在以下端口启动：
- MASQUE VPN: `127.0.0.1:4433`
- REST API: `127.0.0.1:8080`

### 4. 测试 API

```bash
# 健康检查
curl http://127.0.0.1:8080/health

# 服务器状态
curl http://127.0.0.1:8080/api/v1/status

# Prometheus 指标
curl http://127.0.0.1:8080/metrics
```

### 5. 启动客户端

```bash
cd vpn_client
# 编辑 config.client.toml: server_addr = "127.0.0.1:4433"
./vpn-client -c config.client.toml
```

### 6. 测试 MASQUE 协议

```bash
# 功能测试
go run test_masque_connection.go
```

## 配置

### 本地测试

本地测试使用提供的配置：

**服务器** (`config.server.local.toml`):
```toml
listen_addr = "127.0.0.1:4433"
assign_cidr = "10.0.0.0/24"
tun_name = ""  # 为简化而禁用 TUN
log_level = "debug"

[api_server]
listen_addr = "127.0.0.1:8080"
```

**客户端** (`config.client.toml`):
```toml
server_addr = "127.0.0.1:4433"
server_name = "masque-vpn-server"
insecure_skip_verify = true  # 用于测试证书
```

## REST API

API 提供管理和监控端点：

| 端点 | 描述 |
|------|------|
| `GET /health` | API 健康检查 |
| `GET /api/v1/status` | 服务器状态 |
| `GET /api/v1/clients` | 客户端列表 |
| `GET /api/v1/stats` | 服务器统计 |
| `GET /api/v1/config` | 服务器配置 |
| `GET /metrics` | Prometheus 指标 |

详情请参阅 [API 文档](docs/api/api.ru.md)。

## 技术细节

### 自定义 MASQUE 实现

项目包含自定义 MASQUE CONNECT-IP 实现：
- `common/masque_connectip.go` - MASQUE 客户端
- `common/masque_proxy.go` - IP 数据包代理功能
- `vpn_server/internal/server/masque_handler.go` - 服务器处理程序

### 模块化服务器架构

服务器分为专业化模块：
- `server.go` - 主服务器和初始化
- `api_server.go` - REST API 服务器
- `masque_handler.go` - MASQUE 请求处理程序
- `packet_processor.go` - TUN 设备数据包处理
- `metrics.go` - Prometheus 指标
- `tls_config.go` - TLS 配置

### 依赖项

- **Go 1.25** - 现代语言版本
- **QUIC**: [quic-go v0.57.1](https://github.com/quic-go/quic-go) - QUIC 协议
- **HTTP 框架**: [Gin](https://github.com/gin-gonic/gin) - 用于 REST API
- **指标**: [Prometheus client](https://github.com/prometheus/client_golang) - 指标
- **日志**: [Zap](https://github.com/uber-go/zap) - 结构化日志

## 测试

### 运行测试

```bash
# 单元测试
cd common && go test -v
cd vpn_server && go test -v ./...
cd vpn_client && go test -v

# 集成测试
cd tests/integration && go test -v

# 负载测试
cd tests/load && go test -v
```

## 教育用途

### 面向学生和研究人员

本项目专为学习以下内容而设计：
- **现代网络协议**: QUIC、HTTP/3、MASQUE
- **Go 编程**: 并发、网络、测试
- **系统设计**: 模块化架构、API 设计、监控
- **研究方法**: 性能分析、数据收集、文档编写

### 实验作业

1. **基础部署** - 构建和运行系统
2. **性能分析** - 测量 VPN 指标
3. **协议深入研究** - 学习 MASQUE 实现
4. **网络条件测试** - 模拟网络问题

详情请参阅[学生指南](docs/development/student-guide.md)。

## 贡献

本项目为教育目的而创建。欢迎以下贡献：
- 协议实现改进
- 跨平台兼容性
- 文档增强
- 错误修复
- 新测试场景

## 标准和 RFC

- **MASQUE CONNECT-IP**: [RFC 9484](https://datatracker.ietf.org/doc/html/rfc9484) - HTTP 中的 IP 代理
- **QUIC 传输**: [RFC 9000](https://datatracker.ietf.org/doc/html/rfc9000) - QUIC 传输
- **HTTP/3**: [RFC 9114](https://datatracker.ietf.org/doc/html/rfc9114) - HTTP/3 协议

## 支持

如有问题和合作：
- 原始仓库: [iselt/masque-vpn](https://github.com/iselt/masque-vpn)
- GitHub 讨论和问题
- 欢迎教育机构分叉和适配

## 其他语言

- [Русская документация](README.md) - 俄语文档
- [English Documentation](README_en.md) - 英语文档