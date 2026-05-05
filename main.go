package main

import (
	"net/http"
	"os"
	"os/exec"

	"github.com/KevinZonda/GoX/pkg/panicx"
	"github.com/gin-gonic/gin"
)

func initWdir() {
	panicx.NotNilErr(os.MkdirAll("/lzcapp/var/frp", 0o755))
}

func runFrpc() {
	exec.Command("/app/frp/frpc", "-c", "/lzcapp/var/frp/frpc.toml").Start()
}

func setFrpcConfig(text string) error {
	return os.WriteFile("/lzcapp/var/frp/frpc.toml", []byte(text), 0o644)
}

type PureRequest struct {
	Text string `json:"text"`
}

func main() {
	initWdir()
	r := gin.Default()
	r.POST("/api/frpc/config", func(c *gin.Context) {
		var config PureRequest
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := setFrpcConfig(config.Text); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	r.GET("/api/frpc/config", func(c *gin.Context) {
		config, err := os.ReadFile("/lzcapp/var/frp/frpc.toml")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"text": string(config)})
	})

	r.POST("/api/frpc/run", func(c *gin.Context) {
	})
	r.Run(":3000")
}
