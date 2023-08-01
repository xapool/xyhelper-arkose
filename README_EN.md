# xyhelper-arkose

[ENGLISH](README_EN.md)

Automatically obtain Arkose tokens for automated testing.

## Notice

The P project BYPASS feature is no longer available, and there is no reason for it. Please do not ask why.

## 1. Installation

Create the `docker-compose.yml` file:

```yaml
version: "3"
services:
  broswer:
    image: xyhelper/xyhelper-arkose:latest
    ports:
      - "8199:8199"
```

Start the container:

```bash
docker-compose up -d
```

## 2. Usage

### 2.1 Get Token

```bash
curl "http://server-IP:8199/token"
```

### 2.2 Get Token Pool Capacity

```bash
curl "http://server-IP:8199/ping"
```

### 2.3 Actively Hang

```bash
curl "http://server-IP:8199/?delay=10"
```

## 3. How to Generate Tokens

Use a compatible browser to visit `http://server-IP:8199/` and wait for the token to be generated.

## 4. Public Nodes

Get token address: https://chatarkose.xyhelper.cn/token

Check token pool capacity: https://chatarkose.xyhelper.cn/ping