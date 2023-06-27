# ChatGPT Proxy

本项目可以作为以下几种ChatGPT前端的代理：

- [pengzhile/pandora](https://github.com/pengzhile/pandora)
- [Chanzhaoyu/chatgpt-web](https://github.com/Chanzhaoyu/chatgpt-web)
- [moeakwak/chatgpt-web-share](https://github.com/moeakwak/chatgpt-web-share)
- ...

当前版本不依赖第三方服务，基于项目[noahcoolboy/funcaptcha](https://github.com/noahcoolboy/funcaptcha)的算法重写了一个基于Golang的版本，用来获取arkose_token

> **注意：当发现报错403时，请先尝试更新到最新版本**

## 运行

### 源码构建运行
```
go build
PORT=8080 ./chatgpt-proxy
```

### Docker运行

运行docker并映射到80端口
```
docker run -ti -p 80:8080 flyingpot/chatgpt-proxy:latest
```

## 免费部署

一些serverless提供商有免费额度，可以用来部署本项目，例如：

### Koyeb

点击下面的按钮一键部署

[![Deploy to Koyeb](https://www.koyeb.com/static/images/deploy/button.svg)](https://app.koyeb.com/deploy?type=docker&image=docker.io/flyingpot/chatgpt-proxy&name=chatgpt-proxy)

### Render

点击下面的按钮一键部署，缺点是免费版本冷启动比较慢

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/flyingpot/chatgpt-proxy)

### Vercel

当前Vercel上部署存在问题，暂时不可用

## 致谢

本项目受到了以下两个项目的启发：
- [acheong08/ChatGPT-Proxy-V4](https://github.com/acheong08/ChatGPT-Proxy-V4)
- [linweiyuan/go-chatgpt-api](https://github.com/linweiyuan/go-chatgpt-api)
