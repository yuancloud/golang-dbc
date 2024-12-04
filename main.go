package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang-dbc/pkg/device"
	"io"
	"net/http"
	"os"
)

func main() {
	ip := flag.String("ip", "", "ip address")
	port := flag.Int("port", 0, "port number")
	dbc := flag.String("dbc", "", "dbc filename")

	// 解析命令行参数
	flag.Parse()

	if *ip == "" {
		fmt.Println("ip is empty")
		return
	}

	if *port == 0 {
		fmt.Println("port is empty")
		return
	}

	if *dbc == "" {
		fmt.Println("dbc is empty")
		return
	}
	dbcContent, err := os.ReadFile(*dbc)
	if err != nil {
		fmt.Printf("read dbc[%s] error:%s\n", *dbc, err.Error())
		return
	}
	newDevice, err := device.NewDevice(*ip, *port, string(dbcContent))
	if err != nil {
		fmt.Printf("NewDevice error:%s\n", err.Error())
		return
	}
	err = newDevice.Start()
	gin.New().Run()
	if err != nil {
		fmt.Printf("Start error:%s\n", err.Error())
		return
	}
	r := gin.New()
	gin.DefaultWriter = io.MultiWriter(os.Stdout)
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/dbc-vars", func(c *gin.Context) {
		dbcVars := make(map[string]any)
		newDevice.DBCVars.Range(func(key, value any) bool {
			dbcVars[key.(string)] = value
			return true
		})
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "success",
			"data": dbcVars,
		})
	})
	srv := &http.Server{
		Addr:    ":8090",
		Handler: r,
	}
	if err = srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logrus.Fatalf("listen: %s\n", err)
	}
}
