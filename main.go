package main

import (
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/KevinZonda/GoX/pkg/panicx"
	"github.com/gin-gonic/gin"
)

var (
	frpcCmd *exec.Cmd
	frpcMu  sync.Mutex
)

func initWdir() {
	panicx.NotNilErr(os.MkdirAll("/lzcapp/var/frp", 0o755))
}

func startFrpc() error {
	frpcMu.Lock()
	defer frpcMu.Unlock()
	if frpcCmd != nil && frpcCmd.Process != nil {
		frpcCmd.Process.Kill()
		frpcCmd.Wait()
		frpcCmd = nil
	}
	cmd := exec.Command("/app/frp/frpc", "-c", "/lzcapp/var/frp/frpc.toml")
	if err := cmd.Start(); err != nil {
		return err
	}
	frpcCmd = cmd
	return nil
}

func isFrpcRunning() bool {
	frpcMu.Lock()
	defer frpcMu.Unlock()
	if frpcCmd == nil || frpcCmd.Process == nil {
		return false
	}
	err := frpcCmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

func setFrpcConfig(text string) error {
	return os.WriteFile("/lzcapp/var/frp/frpc.toml", []byte(text), 0o644)
}

func getFrpcConfig() string {
	data, err := os.ReadFile("/lzcapp/var/frp/frpc.toml")
	if err != nil {
		return ""
	}
	return string(data)
}

type PureRequest struct {
	Text string `json:"text"`
}

func main() {
	initWdir()
	r := gin.Default()

	r.StaticFile("/", "./web/index.html")

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

	r.GET("/api/frpc/status", func(c *gin.Context) {
		status := "stopped"
		if isFrpcRunning() {
			status = "running"
		}
		c.JSON(http.StatusOK, gin.H{
			"status": status,
			"text":   getFrpcConfig(),
		})
	})

	r.POST("/api/frpc/run", func(c *gin.Context) {
		if err := startFrpc(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "started"})
	})

	r.Run(":3000")
}
