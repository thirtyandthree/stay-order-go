# Go 点餐项目

## 项目简介
这是一个基于 Go 语言开发的点餐系统，旨在提供一个简单、高效的点餐解决方案。该项目包含前端和后端部分，用户可以通过前端进行点餐操作，后端负责处理业务逻辑和数据存储。
真的可惜小程序代码被误删，后悔一辈子
## 项目功能
- 用户注册和登录
- 菜单浏览和搜索
-
- 创建订单和支付
- 订单历史记录查询
- 管理端功能（菜品管理、订单管理等）

## 技术栈
- 后端：Go, GORM
- 数据库：MySQL
- 前端：Vue.js, ElementUI
- 其他： Nginx

## 安装和运行

### 前端部分

1. 克隆前端项目代码：
    ```bash
    git clone https://github.com/thirtyandthree/Stay-order-management
    cd your-frontend-repo
    ```

2. 安装依赖：
    ```bash
    npm install
    ```

3. 运行开发环境：
    ```bash
    npm run serve
    ```

4. 构建生产环境：
    ```bash
    npm run build
    ```

### 后端部分

1. 克隆后端项目代码：
    ```bash
    git clone https://github.com/thirtyandthree/Stay-order-management.git
    cd your-backend-repo
    ```

2. 安装依赖：
    ```bash
    go mod tidy
    ```

3. 运行项目：
    ```bash
    go run main.go
    ```

## 配置

### 数据库配置

在 `config` 文件夹中的 `config.yaml` 文件中配置数据库连接信息：

```yaml
database:
  host: "localhost"
  port: 3306
  user: "yourusername"
  password: "yourpassword"
  name: "yourdbname"
