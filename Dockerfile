FROM registry.lazycat.cloud/lzc/lzcapp:3.20.3

WORKDIR /app
COPY main /app/app
COPY frp /app/frp
COPY web /app/web

RUN chmod +x /app/app

EXPOSE 3000
CMD ["/app/app"]
