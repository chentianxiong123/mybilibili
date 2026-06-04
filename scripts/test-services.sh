#!/bin/bash
# MyBiliBili 微服务启动测试脚本

echo "========== 第一步：检查中间件端口 =========="
echo "请确认以下服务已启动："
echo "  MySQL:      3306"
echo "  Redis:      6379"
echo "  Nacos:      8848"
echo "  RocketMQ:   9876 (NameServer)"
echo "  ES:         9200 (可选)"
echo ""

echo "========== 第二步：编译项目 =========="
cd /d/files/mybilibili-next/mybilibili-cloud
export JAVA_HOME="D:/Program Files (x86)/corretto-17"
export PATH="$JAVA_HOME/bin:$PATH"
echo "编译中..."
mvn clean install -DskipTests
if [ $? -ne 0 ]; then
    echo "编译失败！"
    exit 1
fi
echo "编译完成！"
echo ""

echo "========== 第三步：启动服务 =========="
echo "建议启动顺序（每个服务需要新窗口）："
echo ""
echo "窗口1 - 用户服务 (8081):"
echo "  cd mybilibili-user"
echo "  mvn spring-boot:run"
echo ""
echo "窗口2 - 视频服务 (8082):"
echo "  cd mybilibili-video"
echo "  mvn spring-boot:run"
echo ""
echo "窗口3 - 评论服务 (8085):"
echo "  cd mybilibili-comment"
echo "  mvn spring-boot:run"
echo ""
echo "窗口4 - 网关服务 (8080):"
echo "  cd mybilibili-gateway"
echo "  mvn spring-boot:run"
echo ""

echo "========== 验证方法 =========="
echo "1. 访问 Nacos 控制台检查服务注册:"
echo "   http://localhost:8848/nacos"
echo ""
echo "2. 测试 API:"
echo "   curl http://localhost:8080/api/user/check"
echo ""
echo "3. 启动前端:"
echo "   cd ../mybilibili-web && pnpm run dev"
echo ""

echo "========== 测试完成 =========="
