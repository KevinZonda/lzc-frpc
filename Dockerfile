FROM registry.lazycat.cloud/lzc/lzcapp:3.20.3 AS builder

RUN apk add --no-cache go
WORKDIR /workspace

COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /workspace/app ./main.go

FROM registry.lazycat.cloud/lzc/lzcapp:3.20.3

WORKDIR /app
COPY --from=builder /workspace/app /app/app
COPY run.sh /app/run.sh
RUN chmod +x /app/run.sh /app/app
COPY web /app/web

EXPOSE 3000
CMD ["/app/run.sh"]
