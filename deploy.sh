#!/bin/bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}  LiteWiki 一键部署${NC}"
echo -e "${GREEN}================================${NC}"
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ 未安装 Docker，请先安装：https://docs.docker.com/get-docker/${NC}"
    exit 1
fi

if ! docker compose version &> /dev/null && ! docker-compose version &> /dev/null; then
    echo -e "${RED}❌ 未安装 Docker Compose，请先安装${NC}"
    exit 1
fi

# 生成 .env
if [ ! -f .env ]; then
    echo -e "${YELLOW}📝 生成 .env 配置文件...${NC}"
    JWT_SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 64 /dev/urandom | od -An -tx1 | tr -d ' \n')
    # 只用字母+数字，避免 base64 特殊字符
    ADMIN_PWD=$(openssl rand -hex 12 2>/dev/null || head -c 12 /dev/urandom | od -An -tx1 | tr -d ' \n')

    cat > .env <<EOF
JWT_SECRET=${JWT_SECRET}
ADMIN_INIT_PASSWORD=${ADMIN_PWD}
HTTP_PORT=80
EOF

    # 设置 .env 权限为仅 owner 可读写
    chmod 600 .env

    echo -e "${GREEN}✅ .env 已生成${NC}"
else
    echo -e "${GREEN}✅ 检测到已有 .env 配置，跳过生成${NC}"
    ADMIN_PWD=$(grep ADMIN_INIT_PASSWORD .env | cut -d= -f2 | tr -d ' ')
fi

# 创建数据目录
mkdir -p data/db data/uploads data/frontend

# 构建并启动
echo -e "${YELLOW}🔨 构建镜像（首次需要几分钟）...${NC}"
docker compose build

echo -e "${YELLOW}🚀 启动服务...${NC}"
docker compose up -d

# 获取端口
PORT=$(grep HTTP_PORT .env 2>/dev/null | cut -d= -f2 | tr -d ' ')
PORT=${PORT:-80}

# 获取服务器 IP（优先公网 IP，fallback 到内网 IP）
SERVER_IP=$(curl -s --connect-timeout 3 ifconfig.me 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}' || echo "YOUR_SERVER_IP")

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}  ✅ LiteWiki 已启动！${NC}"
echo -e "${GREEN}================================${NC}"
echo ""
echo -e "  ${CYAN}📋 登录信息${NC}"
echo -e "  ─────────────────────────────"
echo -e "  👤 管理员账号: ${GREEN}admin${NC}"
echo -e "  🔑 管理员密码: ${GREEN}${ADMIN_PWD}${NC}"
echo ""
echo -e "  ${CYAN}🌐 访问地址${NC}"
echo -e "  ─────────────────────────────"
echo -e "  📖 管理后台: ${GREEN}http://${SERVER_IP}:${PORT}/admin/${NC}"
echo -e "  🏠 首页:     ${GREEN}http://${SERVER_IP}:${PORT}${NC}"
echo ""
echo -e "  ${CYAN}🛠 常用命令${NC}"
echo -e "  ─────────────────────────────"
echo -e "    查看日志: docker compose logs -f"
echo -e "    停止服务: docker compose down"
echo -e "    重启服务: docker compose restart"
echo -e "    更新版本: git pull && docker compose up -d --build"
echo ""
echo -e "  ${YELLOW}⚠️  请妥善保存以上密码，关闭后无法再次查看！${NC}"
echo ""
