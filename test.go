package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  go run main.go server     - 啟動UDP Server")
		fmt.Println("  go run main.go client1    - 啟動UDP Client 1")
		fmt.Println("  go run main.go client2    - 啟動UDP Client 2")
		fmt.Println("  go run main.go proxy1     - 啟動UDP Proxy 1")
		fmt.Println("  go run main.go proxy2     - 啟動UDP Proxy 2")
		fmt.Println("  go run main.go all        - 啟動所有組件")
		fmt.Println("  go run main.go all -q     - 啟動所有組件（靜默模式，只顯示Client結果）")
		return
	}

	command := os.Args[1]
	quietMode := len(os.Args) > 2 && os.Args[2] == "-q"

	switch command {
	case "server":
		runServer()
	case "client1":
		runClient1()
	case "client2":
		runClient2()
	case "proxy1":
		runProxy1()
	case "proxy2":
		runProxy2()
	case "all":
		runAll(quietMode)
	default:
		fmt.Printf("未知命令: %s\n", command)
	}
}

func runServer() {
	fmt.Println("啟動UDP Server...")
	cmd := exec.Command("go", "run", "server/server.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runClient1() {
	fmt.Println("啟動UDP Client 1...")
	cmd := exec.Command("go", "run", "client/client1.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runClient2() {
	fmt.Println("啟動UDP Client 2...")
	cmd := exec.Command("go", "run", "client/client2.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runProxy1() {
	fmt.Println("啟動UDP Proxy 1...")
	cmd := exec.Command("go", "run", "proxy/proxy1.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runProxy2() {
	fmt.Println("啟動UDP Proxy 2...")
	cmd := exec.Command("go", "run", "proxy/proxy2.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func runAll(quietMode bool) {
	if !quietMode {
		fmt.Println("啟動所有組件...")
	}
	
	// 啟動Client 1
	if !quietMode {
		fmt.Println("啟動UDP Client 1...")
	}
	client1Cmd := exec.Command("go", "run", "client/client1.go")
	if quietMode {
		client1Cmd.Stdout = os.Stdout
		client1Cmd.Stderr = os.Stderr
	} else {
		client1Cmd.Stdout = os.Stdout
		client1Cmd.Stderr = os.Stderr
	}
	go client1Cmd.Run()
	
	// 啟動Client 2
	if !quietMode {
		fmt.Println("啟動UDP Client 2...")
	}
	client2Cmd := exec.Command("go", "run", "client/client2.go")
	if quietMode {
		client2Cmd.Stdout = os.Stdout
		client2Cmd.Stderr = os.Stderr
	} else {
		client2Cmd.Stdout = os.Stdout
		client2Cmd.Stderr = os.Stderr
	}
	go client2Cmd.Run()
	
	// 等待一下讓Clients啟動
	time.Sleep(1 * time.Second)
	
	// 啟動Proxy 1
	if !quietMode {
		fmt.Println("啟動UDP Proxy 1...")
	}
	proxy1Args := []string{"run", "proxy/proxy1.go"}
	if quietMode {
		proxy1Args = append(proxy1Args, "-q")
	}
	proxy1Cmd := exec.Command("go", proxy1Args...)
	if quietMode {
		proxy1Cmd.Stdout = nil
		proxy1Cmd.Stderr = nil
	} else {
		proxy1Cmd.Stdout = os.Stdout
		proxy1Cmd.Stderr = os.Stderr
	}
	go proxy1Cmd.Run()
	
	// 啟動Proxy 2
	if !quietMode {
		fmt.Println("啟動UDP Proxy 2...")
	}
	proxy2Args := []string{"run", "proxy/proxy2.go"}
	if quietMode {
		proxy2Args = append(proxy2Args, "-q")
	}
	proxy2Cmd := exec.Command("go", proxy2Args...)
	if quietMode {
		proxy2Cmd.Stdout = nil
		proxy2Cmd.Stderr = nil
	} else {
		proxy2Cmd.Stdout = os.Stdout
		proxy2Cmd.Stderr = os.Stderr
	}
	go proxy2Cmd.Run()
	
	// 等待一下讓Proxies啟動
	time.Sleep(1 * time.Second)
	
	// 啟動Server
	if !quietMode {
		fmt.Println("啟動UDP Server...")
	}
	serverArgs := []string{"run", "server/server.go"}
	if quietMode {
		serverArgs = append(serverArgs, "-q")
	}
	serverCmd := exec.Command("go", serverArgs...)
	if quietMode {
		serverCmd.Stdout = nil
		serverCmd.Stderr = nil
	} else {
		serverCmd.Stdout = os.Stdout
		serverCmd.Stderr = os.Stderr
	}
	serverCmd.Run()
}
