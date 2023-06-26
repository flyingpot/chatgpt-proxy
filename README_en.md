# ChatGPT Proxy

This proxy can be used in [Pandora](https://github.com/pengzhile/pandora) project as a replacement for the default proxy.

## Building and running

### Build from source

```
go build
PORT=80 ./chatgpt-proxy
```

### Run with docker

```
docker run -ti -p 80:8080 flyingpot/chatgpt-proxy:latest
```

## Deploy

### Render

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy?repo=https://github.com/flyingpot/chatgpt-proxy)

### Vercel

There is still an issue with deployment on Vercel. So it's unsupported now.

## Acknowledgements

This project has referenced:
- [acheong08/ChatGPT-Proxy-V4](https://github.com/acheong08/ChatGPT-Proxy-V4)
- [linweiyuan/go-chatgpt-api](https://github.com/linweiyuan/go-chatgpt-api)
