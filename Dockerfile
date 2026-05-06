FROM registry.lazycat.cloud/lzc/lzcapp:3.20.3

WORKDIR /app
COPY main /app/app
COPY frp /app/frp
COPY web /app/web

COPY run-release.sh /app/run.sh
RUN chmod +x /app/run.sh /app/app

EXPOSE 3000
CMD ["sh", "/app/run.sh"]
