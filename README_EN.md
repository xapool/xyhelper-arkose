# xyhelper-arkose


Automatically obtain arkose tokens for automated testing.

## Notice
The BYPASS feature for project P is no longer available, and there is no reason provided. Please refrain from asking why.

## 1. Installation
Create the `docker-compose.yml` file:
```yaml
version: '3'
services:
  browser:
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
Start the service:
```bash
docker-compose up -d
```

## 2. Usage

### 2.1 Obtain Token
```bash
curl "http://SERVER_IP:8199/token"
```

### 2.2 Check Token Pool Capacity
```bash
curl "http://SERVER_IP:8199/ping"
```

### 2.3 Initiate Hanging
```bash
curl "http://SERVER_IP:8199/?delay=10"
```

## 3. Add Hanging Nodes
Create the `docker-compose.yml` file:
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
      - FORWARD_URL=https://chatarkose.xyhelper.cn/pushtoken # Replace with your own token pool address
    shm_size: 512m
```
Start the service:
```bash
docker-compose up -d
```

For multiple nodes, you can use the `docker-compose scale` command:
```bash
docker compose up -d --scale token-pusher=5
```

## 4. Manage Chrome

Login URL: https://SERVER_IP:6901

Username: kasm_user

Default Password: xyhelper

## 5. Public Nodes

Obtain token address: https://chatarkose.xyhelper.cn/token

Check token pool capacity: https://chatarkose.xyhelper.cn/ping