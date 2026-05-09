package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/KevinZonda/GoX/pkg/panicx"
	"github.com/gin-gonic/gin"
)

// ---- process state ----

var (
	frpcCmd *exec.Cmd
	frpcMu  sync.Mutex
)

// ---- log buffer ----

const maxLogLines = 500

var (
	logLines []string
	logMu    sync.Mutex
)

func appendLog(line string) {
	logMu.Lock()
	defer logMu.Unlock()
	logLines = append(logLines, line)
	if len(logLines) > maxLogLines {
		logLines = logLines[len(logLines)-maxLogLines:]
	}
}

func clearLogs() {
	logMu.Lock()
	defer logMu.Unlock()
	logLines = nil
}

func getLogs() []string {
	logMu.Lock()
	defer logMu.Unlock()
	result := make([]string, len(logLines))
	copy(result, logLines)
	return result
}

func scanLines(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		appendLog(scanner.Text())
	}
}

// ---- frpc management ----

func initWdir() {
	panicx.NotNilErr(os.MkdirAll("/lzcapp/var/frp", 0o755))
}

func stopFrpc() {
	frpcMu.Lock()
	defer frpcMu.Unlock()
	if frpcCmd != nil && frpcCmd.Process != nil {
		frpcCmd.Process.Kill()
		frpcCmd = nil
	}
}

func startFrpc() error {
	frpcMu.Lock()
	defer frpcMu.Unlock()

	if frpcCmd != nil && frpcCmd.Process != nil {
		frpcCmd.Process.Kill()
		frpcCmd = nil
	}
	clearLogs()

	cmd := exec.Command("/app/frp/frpc", "-c", "/lzcapp/var/frp/frpc.toml")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	go scanLines(stdout)
	go scanLines(stderr)

	frpcCmd = cmd
	return nil
}

func isFrpcRunning() bool {
	frpcMu.Lock()
	defer frpcMu.Unlock()
	if frpcCmd == nil || frpcCmd.Process == nil {
		return false
	}
	return frpcCmd.Process.Signal(syscall.Signal(0)) == nil
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

// ---- HTTP ----

type PureRequest struct {
	Text string `json:"text"`
}

const DEFAULT_CONFIG = `serverAddr = "127.0.0.1"
serverPort = 7000

[[proxies]]
name = "test-tcp"
type = "tcp"
localIP = "127.0.0.1"
localPort = 22
remotePort = 6000`

func main() {
	gin.SetMode(gin.ReleaseMode)
	initWdir()
	if _, err := os.Stat("/lzcapp/var/frp/frpc.toml"); err == nil {
		startFrpc()
	} else if errors.Is(err, os.ErrNotExist) {
		err := setFrpcConfig(DEFAULT_CONFIG)
		if err != nil {
			log.Println("set default config error:", err)
		}
	}

	r := gin.Default()

	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/assets/codemirror.min.css", "./web/assets/codemirror.min.css")
	r.StaticFile("/assets/codemirror.min.js", "./web/assets/codemirror.min.js")
	r.StaticFile("/assets/toml.min.js", "./web/assets/toml.min.js")

	r.POST("/api/frpc/config", func(c *gin.Context) {
		var req PureRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := setFrpcConfig(req.Text); err != nil {
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
		c.JSON(http.StatusOK, gin.H{"status": status})
	})

	r.GET("/api/frpc/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"text": getFrpcConfig()})
	})

	r.POST("/api/frpc/run", func(c *gin.Context) {
		if err := startFrpc(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "started"})
	})

	r.POST("/api/frpc/stop", func(c *gin.Context) {
		stopFrpc()
		c.JSON(http.StatusOK, gin.H{"message": "stopped"})
	})

	r.POST("/api/frpc/restart", func(c *gin.Context) {
		if err := startFrpc(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "restarted"})
	})

	r.GET("/api/frpc/logs", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"lines": getLogs()})
	})

	r.Run(":3000")
}
