# xyhelper-arkose

[ENGLISH](README_EN.md)

自动获取arkose的token，用于自动化测试

## 通知
不再提供P项目BYPASS功能,没有原因,请不要问为什么

## 1. 安装
创建`docker-compose.yml`文件
```yaml
version: '3'
services:
  broswer:
    image: xyhelper/xyhelper-arkose-browser:latest
    ports:
      - "6901:6901"
      - "8199:3000"
    environment:
      - VNC_PW=xyhelper
      - LAUNCH_URL=http://localhost:3000
      - APP_ARGS=--ignore-certificate-errors
      - PORT=3000
    shm_size: 512m
```
启动
```bash
docker-compose up -d
```
## 2. 使用

### 2.1 获取token
```bash
curl "http://服务器IP:8199/token"
```

### 2.2 获取token池容量
```bash
curl "http://服务器IP:8199/ping"
```

### 2.3 主动挂机
```bash
curl "http://服务器IP:8199/?delay=10"
```

## 3. 增加挂机节点
创建`docker-compose.yml`文件
```yaml
version: '3'
services:
  token-pusher:
    image: xyhelper/xyhelper-arkose-browser:latest
    environment:
      - VNC_PW=xyhelper
      - LAUNCH_URL=http://localhost:3000
      - APP_ARGS=--ignore-certificate-errors
      - PORT=3000
      - FORWORD_URL=https://arkose.xyhelper.cn/pushtoken # 修改为自己的token池地址
    shm_size: 512m
```
启动
```bash
docker-compose up -d
```

多个节点可以使用`docker-compose scale`命令
```bash
docker-compose scale token-pusher=10
```

## 4. 管理chrome

登陆地址：https://服务器IP:6901

用户名：kasm_user

默认密码：xyhelper  

## 5. 公共节点

获取token地址：https://chatarkose.xyhelper.cn/token

查询token池容量：https://chatarkose.xyhelper.cn/ping