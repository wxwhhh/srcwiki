<div align="center">

# 📚 SRCWiki

**轻量级 · 现代化 · 自托管知识库系统**

一个用 Go 和 Vue 3 构建的轻量级知识库/Wiki 系统，支持 Markdown 文档管理、多用户权限控制、全文搜索。

[![Docker](https://img.shields.io/badge/Docker-Ready-blue?style=flat-square&logo=docker)](https://docs.docker.com/)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org/)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

[功能特性](#-功能特性) · [快速开始](#-快速开始) · [部署方式](#-部署方式) · [技术架构](#-技术架构)

</div>

---

## ✨ 功能特性

| 模块 | 功能 |
|------|------|
| 📝 **文档管理** | Markdown 编辑器、版本历史、拖拽排序、批量导入（ZIP/GitHub） |
| 👥 **用户系统** | 多角色权限（管理员/编辑者/读者）、邀请码注册、JWT 认证 |
| 🔍 **全文搜索** | 中文分词、关键词高亮、实时搜索 |
| 📁 **分类系统** | 树状分类、拖拽排序、多级嵌套 |
| 🖼️ **媒体管理** | 图片上传、文件管理 |
| 📊 **操作审计** | 完整的操作日志记录 |
| 🔒 **安全防护** | 限流、验证码、CORS 白名单、XSS/SQL 注入防护 |
| 🐳 **一键部署** | Docker Compose 一键启动，自动生成密码 |

## 🚀 快速开始

### 前置要求

- [Docker](https://docs.docker.com/get-docker/) >= 20.10
- [Docker Compose](https://docs.docker.com/compose/install/) >= 2.0

### 一键部署

```bash
# 1. 克隆项目
git clone https://github.com/YOUR_USERNAME/srcwiki.git
cd srcwiki

# 2. 一键部署
chmod +x deploy.sh
./deploy.sh
```

部署完成后，终端会显示：

```
==================================
  ✅ SRCWiki 已启动！
==================================

  📋 登录信息
  ─────────────────────────────
  👤 管理员账号: admin
  🔑 管理员密码: <随机生成>

  🌐 访问地址
  ─────────────────────────────
  📖 管理后台: http://your-ip/admin/
  🏠 首页:     http://your-ip
```

<img width="2225" height="1072" alt="image" src="https://github.com/user-attachments/assets/adff598d-906b-420c-8a22-7644d864014c" />

<img width="2496" height="1266" alt="image" src="https://github.com/user-attachments/assets/85ac68cf-a4e6-4f68-ad11-5f2b407785d6" />

<img width="2519" height="1207" alt="image" src="https://github.com/user-attachments/assets/0e3c3035-099c-4f70-9259-084edcf754e7" />

<img width="2473" height="1190" alt="image" src="https://github.com/user-attachments/assets/0049f32f-5458-46fd-9abb-2077e14fb6c0" />


### 手动部署

```bash
# 构建并启动（首次运行会自动生成随机密码和密钥）
docker compose up -d --build
```

## ⚙️ 配置说明

编辑 `.env` 文件：

| 变量 | 说明 | 必填 |
|------|------|:----:|
| `JWT_SECRET` | JWT 签名密钥（至少 32 位随机字符串） | ✅ |
| `ADMIN_INIT_PASSWORD` | 管理员初始密码 | ✅ |
| `HTTP_PORT` | HTTP 监听端口（默认 `80`） | ❌ |

## 📂 项目结构

```
srcwiki/
├── backend/                # Go 后端
│   ├── config/            # 配置
│   ├── handlers/          # HTTP 处理器
│   ├── middleware/         # 中间件（认证、限流、CORS）
│   ├── models/            # 数据模型
│   ├── services/          # 业务逻辑
│   └── utils/             # 工具函数
├── frontend/               # Vue 前端
│   ├── src/
│   │   ├── views/         # 页面组件
│   │   ├── api/           # API 调用
│   │   └── layouts/       # 布局组件
│   └── dist/              # 构建产物
├── Caddyfile               # Caddy 反向代理配置
├── docker-compose.yml      # Docker Compose 编排
├── deploy.sh               # 一键部署脚本
└── .env.example            # 环境变量模板
```

## 🔧 常用命令

```bash
# 查看日志
docker compose logs -f

# 停止服务
docker compose down

# 重启服务
docker compose restart

# 更新版本
git pull && docker compose up -d --build

# 备份数据库
cp data/db/litewiki.db ./backup/litewiki-$(date +%Y%m%d).db
```

## 🏗️ 技术栈

<div align="center">

| 层级 | 技术 |
|------|------|
| **后端** | Go · Gin · SQLite · JWT |
| **前端** | Vue 3 · TypeScript · Element Plus |
| **代理** | Caddy |
| **容器** | Docker · Docker Compose |

</div>

## 📸 界面预览

<div align="center">

> 简约现代的界面设计，支持深色/浅色主题

</div>

## 🔒 安全特性

- ✅ JWT 认证 + RBAC 权限控制
- ✅ 登录/注册验证码 + IP 限流
- ✅ CORS 白名单 + XSS 防护
- ✅ SQL 参数化查询（防注入）
- ✅ 文件上传白名单 + 路径遍历防护
- ✅ 完整的操作审计日志

## 📄 License

[MIT](LICENSE) © [Your Name]

---

<div align="center">

**如果这个项目对你有帮助，请给个 ⭐ Star 支持一下！**

</div>
