FROM registry.lazycat.cloud/lzc/lzcapp:3.20.3

WORKDIR /app
COPY main /app/app
COPY frp /app/frp
COPY run.sh /app/run.sh
RUN chmod +x /app/run.sh /app/app
COPY web /app/web

EXPOSE 3000
CMD ["/app/run.sh"]
