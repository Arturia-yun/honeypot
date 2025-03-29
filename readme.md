# Honeypot 蜜罐系统

基于Go语言开发的分布式蜜罐系统，用于网络安全监测、攻击行为捕获与分析。

[English Version](#honeypot-system)

## 系统架构

```
Agent(流量捕获) --> Server(服务模拟) --> LogServer(日志分析) --> MongoDB(数据存储)
```

### 主要组件

1. **Agent 代理服务**
   - 网络流量捕获与分析
   - 非监听端口扫描检测
   - 高交互服务流量转发
   - 基于YAML的策略管理

2. **Server 高交互服务**
   - 多协议服务模拟(SSH/MySQL/Redis/Web)
   - 攻击者行为记录
   - 真实IP提取与追踪
   - 统一日志上报

3. **LogServer 日志服务**
   - 集中日志接收与存储
   - API认证与安全控制
   - MongoDB数据持久化
   - 攻击行为分析

## 快速开始

### 环境要求
- Go 1.24+
- MongoDB 4.0+
- Windows/Linux系统

### 安装步骤

1. 克隆仓库
```bash
git clone https://github.com/yourusername/honeypot.git
cd honeypot
```

2. 启动日志服务器
```bash
cd logServer
go run main.go
```

3. 启动高交互服务
```bash
cd server
go run main.go
```

4. 启动Agent代理
```bash
cd Agent
go run main.go
```
### 效果展示

![image](https://github.com/user-attachments/assets/5a30b544-ddd3-4ad4-8c52-8d7da406494b)

LogServer接受到MySQL和普通访问流量

![image](https://github.com/user-attachments/assets/64b4132f-2113-4ada-8df7-f133e8794258)
![image](https://github.com/user-attachments/assets/985e7b1d-d55d-4a4e-aac9-05ea0c67c9bd)


## 配置说明

### Agent配置 (`Agent/config/app.ini`)
```ini
[app]
name = honeypot-agent
version = 1.0.0

[log]
server = http://localhost:8083
level = info

[capture]
interface = \Device\NPF_{网卡ID}
```

### 策略配置 (`Agent/config/policy.yaml`)
```yaml
policy:
  - id: "default"
    white_ips: ["127.0.0.1", "192.168.1.1"]
    white_ports: ["22", "80", "443"]

service:
  - id: "ssh"
    service_name: "ssh"
    local_port: 22
    backend_host: "127.0.0.1"
    backend_port: 2222
  - id: "mysql"
    service_name: "mysql"
    local_port: 3306
    backend_host: "127.0.0.1"
    backend_port: 3366
```

### 日志服务器配置 (`logServer/pkg/config/config.go`)
```go
GlobalConfig.MongoDB.URI = "mongodb://localhost:27017"
GlobalConfig.MongoDB.Database = "honeypot"
GlobalConfig.APIKey = "honeypot-api-key-2024"
```

## 功能特性

### 网络流量捕获
- 基于gopacket实现的网络数据包捕获
- TCP五元组信息提取与分析
- 非监听端口扫描检测

### 高交互服务模拟
- **SSH服务**: 基于gliderlabs/ssh实现的SSH服务，记录认证尝试
- **MySQL服务**: 基于go-mysql-server实现的MySQL服务，支持SQL语句解析与记录
- *更新*: 新增文件读取漏洞模拟功能，在3307端口监听，可捕获攻击者的文件读取尝试并记录详细信息
- **Redis服务**: 基于redcon实现的Redis服务，记录命令执行
- **Web服务**: 基于Gin实现的HTTP服务，记录请求详情

### 日志分析与存储
- 统一的日志格式与API接口
- 基于MongoDB的持久化存储
- API密钥认证保障安全性

## API文档

### 日志上报接口

| 端点 | 方法 | 描述 | 认证 |
|------|------|------|------|
| `/api/packet/` | POST | 网络数据包日志上报 | API Key |
| `/api/:service/` | POST | 服务日志上报(ssh/mysql/redis/web) | API Key |

**请求头**:
```
X-API-Key: honeypot-api-key-2024
```

**示例请求体(SSH服务)**:
```json
{
  "time": "2024-03-20T15:30:00Z",
  "ip": "192.168.1.100",
  "service": "ssh",
  "username": "root",
  "password": "password123",
  "remote_addr": "192.168.1.100:12345"
}
```

## 数据结构

### 网络数据包日志
```go
type PacketLog struct {
    Time      time.Time `bson:"time" json:"time"`
    SrcIP     string    `bson:"src_ip" json:"src_ip"`
    SrcPort   string    `bson:"src_port" json:"src_port"`
    DstIP     string    `bson:"dst_ip" json:"dst_ip"`
    DstPort   string    `bson:"dst_port" json:"dst_port"`
    Protocol  string    `bson:"protocol" json:"protocol"`
    Service   string    `bson:"service,omitempty" json:"service,omitempty"`
    IsHTTP    bool      `bson:"is_http,omitempty" json:"is_http,omitempty"`
}
```

### 服务日志
```go
type ServiceLog struct {
    Time      time.Time              `bson:"time" json:"time"`
    IP        string                 `bson:"ip" json:"ip"`
    Service   string                 `bson:"service" json:"service"`
    Username  string                 `bson:"username,omitempty" json:"username,omitempty"`
    Password  string                 `bson:"password,omitempty" json:"password,omitempty"`
    Command   string                 `bson:"command,omitempty" json:"command,omitempty"`
    Data      map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
}
```

## 项目结构
```
honeypot/
├── Agent/                 # 代理服务
│   ├── config/            # 配置文件
│   ├── pkg/               # 功能包
│   │   ├── capture/       # 流量捕获
│   │   ├── forward/       # 流量转发
│   │   ├── logger/        # 日志模块
│   │   ├── models/        # 数据模型
│   │   ├── policy/        # 策略管理
│   │   └── vars/          # 全局变量
│   └── main.go            # 入口文件
├── logServer/             # 日志服务器
│   ├── pkg/               # 功能包
│   │   ├── api/           # API接口
│   │   ├── config/        # 配置模块
│   │   ├── db/            # 数据库连接
│   │   ├── middleware/    # 中间件
│   │   └── models/        # 数据模型
│   └── main.go            # 入口文件
└── server/                # 高交互服务
    ├── pkg/               # 功能包
    │   ├── logger/        # 日志模块
    │   ├── service/       # 服务模块
    │   │   ├── MySQL/     # MySQL服务
    │   │   ├── redisServer/# Redis服务
    │   │   ├── ssh/       # SSH服务
    │   │   └── web/       # Web服务
    │   └── util/          # 工具函数
    └── main.go            # 入口文件
```

## 贡献指南

欢迎提交Issue和Pull Request，请遵循以下规范：
1. Fork本仓库并创建特性分支
2. 提交前请确保代码通过测试
3. 提交信息请遵循规范格式

## 许可证

本项目采用MIT许可证

---

# Honeypot System

A distributed honeypot system developed in Go for network security monitoring, attack behavior capture and analysis.

## System Architecture

```
Agent(Traffic Capture) --> Server(Service Simulation) --> LogServer(Log Analysis) --> MongoDB(Data Storage)
```

### Main Components

1. **Agent Service**
   - Network traffic capture and analysis
   - Non-listening port scan detection
   - High-interaction service traffic forwarding
   - YAML-based policy management

2. **Server (High-interaction Services)**
   - Multi-protocol service simulation (SSH/MySQL/Redis/Web)
   - Attacker behavior recording
   - Real IP extraction and tracking
   - Unified log reporting

3. **LogServer**
   - Centralized log collection and storage
   - API authentication and security control
   - MongoDB data persistence
   - Attack behavior analysis

## Quick Start

### Requirements
- Go 1.24+
- MongoDB 4.0+
- Windows/Linux system

### Installation Steps

1. Clone the repository
```bash
git clone https://github.com/yourusername/honeypot.git
cd honeypot
```

2. Start the log server
```bash
cd logServer
go run main.go
```

3. Start the high-interaction services
```bash
cd server
go run main.go
```

4. Start the Agent
```bash
cd Agent
go run main.go
```

### Effect display
![image](https://github.com/user-attachments/assets/d5e94654-8c6d-417b-a495-0c0d5693b7da)

LogServer receives MySQL and normal access traffic

![image](https://github.com/user-attachments/assets/e9cefcae-d1de-43d7-bb8f-aa04eb63658b)
![image](https://github.com/user-attachments/assets/2aefd001-a923-4e3b-a732-3ee9e59a79c2)

## Configuration

### Agent Configuration (`Agent/config/app.ini`)
```ini
[app]
name = honeypot-agent
version = 1.0.0

[log]
server = http://localhost:8083
level = info

[capture]
interface = \Device\NPF_{network_card_ID}
```

### Policy Configuration (`Agent/config/policy.yaml`)
```yaml
policy:
  - id: "default"
    white_ips: ["127.0.0.1", "192.168.1.1"]
    white_ports: ["22", "80", "443"]

service:
  - id: "ssh"
    service_name: "ssh"
    local_port: 22
    backend_host: "127.0.0.1"
    backend_port: 2222
  - id: "mysql"
    service_name: "mysql"
    local_port: 3306
    backend_host: "127.0.0.1"
    backend_port: 3366
```

### Log Server Configuration (`logServer/pkg/config/config.go`)
```go
GlobalConfig.MongoDB.URI = "mongodb://localhost:27017"
GlobalConfig.MongoDB.Database = "honeypot"
GlobalConfig.APIKey = "honeypot-api-key-2024"
```

## Features

### Network Traffic Capture
- Network packet capture based on gopacket
- TCP 5-tuple information extraction and analysis
- Non-listening port scan detection

### High-interaction Service Simulation
- **SSH Service**: SSH service based on gliderlabs/ssh, recording authentication attempts
- **MySQL Service**: MySQL service based on go-mysql-server, supporting SQL statement parsing and recording
  - *Update*: Added file reading vulnerability simulation on port 3307, capturing and recording attackers' file reading attempts with detailed information
- **Redis Service**: Redis service based on redcon, recording command execution
- **Web Service**: HTTP service based on Gin, recording request details

### Log Analysis and Storage
- Unified log format and API interface
- MongoDB-based persistent storage
- API key authentication for security

## API Documentation

### Log Reporting Endpoints

| Endpoint | Method | Description | Authentication |
|----------|--------|-------------|----------------|
| `/api/packet/` | POST | Network packet log reporting | API Key |
| `/api/:service/` | POST | Service log reporting (ssh/mysql/redis/web) | API Key |

**Request Headers**:
```
X-API-Key: honeypot-api-key-2024
```

**Example Request Body (SSH Service)**:
```json
{
  "time": "2024-03-20T15:30:00Z",
  "ip": "192.168.1.100",
  "service": "ssh",
  "username": "root",
  "password": "password123",
  "remote_addr": "192.168.1.100:12345"
}
```

## Data Structures

### Network Packet Log
```go
type PacketLog struct {
    Time      time.Time `bson:"time" json:"time"`
    SrcIP     string    `bson:"src_ip" json:"src_ip"`
    SrcPort   string    `bson:"src_port" json:"src_port"`
    DstIP     string    `bson:"dst_ip" json:"dst_ip"`
    DstPort   string    `bson:"dst_port" json:"dst_port"`
    Protocol  string    `bson:"protocol" json:"protocol"`
    Service   string    `bson:"service,omitempty" json:"service,omitempty"`
    IsHTTP    bool      `bson:"is_http,omitempty" json:"is_http,omitempty"`
}
```

### Service Log
```go
type ServiceLog struct {
    Time      time.Time              `bson:"time" json:"time"`
    IP        string                 `bson:"ip" json:"ip"`
    Service   string                 `bson:"service" json:"service"`
    Username  string                 `bson:"username,omitempty" json:"username,omitempty"`
    Password  string                 `bson:"password,omitempty" json:"password,omitempty"`
    Command   string                 `bson:"command,omitempty" json:"command,omitempty"`
    Data      map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
}
```

## Project Structure
```
honeypot/
├── Agent/                 # Agent service
│   ├── config/            # Configuration files
│   ├── pkg/               # Function packages
│   │   ├── capture/       # Traffic capture
│   │   ├── forward/       # Traffic forwarding
│   │   ├── logger/        # Logging module
│   │   ├── models/        # Data models
│   │   ├── policy/        # Policy management
│   │   └── vars/          # Global variables
│   └── main.go            # Entry file
├── logServer/             # Log server
│   ├── pkg/               # Function packages
│   │   ├── api/           # API interfaces
│   │   ├── config/        # Configuration module
│   │   ├── db/            # Database connection
│   │   ├── middleware/    # Middleware
│   │   └── models/        # Data models
│   └── main.go            # Entry file
└── server/                # High-interaction services
    ├── pkg/               # Function packages
    │   ├── logger/        # Logging module
    │   ├── service/       # Service module
    │   │   ├── MySQL/     # MySQL service
    │   │   ├── redisServer/# Redis service
    │   │   ├── ssh/       # SSH service
    │   │   └── web/       # Web service
    │   └── util/          # Utility functions
    └── main.go            # Entry file
```

## Contribution Guidelines

Issues and Pull Requests are welcome. Please follow these guidelines:
1. Fork this repository and create a feature branch
2. Ensure your code passes tests before submitting
3. Follow the standard commit message format

## License

This project is licensed under the MIT License
