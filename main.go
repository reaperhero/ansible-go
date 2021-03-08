package main

import (
	"bufio"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type host struct {
	sshHost     string
	sshPassword string
}

func getHostList(hostType string) (hosts []host) {
	file, err := os.Open("./host.list")
	if err != nil {
		panic(err)
	}
	br := bufio.NewReader(file)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		readline := strings.Split(string(a), " ")
		if len(readline) != 2 {
			log.Fatal("主机信息有问题")
		}
		hosts = append(hosts, host{
			sshHost:     readline[0],
			sshPassword: readline[1],
		})
	}
	return
}

func main() {
	var command string
	flag.StringVar(&command, "cmd", "", "命令")
	flag.Parse()

	list := getHostList("")
	wg := sync.WaitGroup{}
	wg.Add(len(list))
	for _, h := range list {
		go func() {
			execCommand(h, command)
			wg.Done()
		}()
	}
	wg.Wait()
}

func execCommand(host host, command string) {
	//创建sshp登陆配置
	config := &ssh.ClientConfig{
		Timeout:         time.Second * 5, //ssh 连接time out 时间一秒钟, 如果ssh验证错误 会在一秒内返回
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		Auth:            []ssh.AuthMethod{ssh.Password(host.sshPassword)},
	}
	//dial 获取ssh client
	addr := fmt.Sprintf("%s:%d", host.sshHost, 22)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Println("connect err ", err)
		return
	}
	defer sshClient.Close()

	//创建ssh-session
	session, err := sshClient.NewSession()
	if err != nil {
		log.Println("create session err ", err)
	}
	defer session.Close()
	//执行远程命令
	combo, err := session.CombinedOutput(command)
	if err != nil {
		log.Println("exec command ", err)
		return
	}
	log.Println(host.sshHost + " 执行远程命令成功........")
	log.Println("命令输出:\n", string(combo))
}
