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

var (
	errfile, _     = os.OpenFile("./error.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	successfile, _ = os.OpenFile("./success.log", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
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
	defer errfile.Close()
	list := getHostList("")
	wg := sync.WaitGroup{}
	wg.Add(len(list))
	for _, h := range list {
		func(host host) {
			execCommand(host, command)
			wg.Done()
		}(h)
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
		appendErrtoFile(host.sshHost + "执行报错" + err.Error())
		return
	}
	defer sshClient.Close()

	//创建ssh-session
	session, err := sshClient.NewSession()
	if err != nil {
		log.Println("create session err ", err)
		appendErrtoFile(host.sshHost + "执行报错" + err.Error())
	}
	defer session.Close()
	//执行远程命令
	combo, err := session.CombinedOutput(command)
	if err != nil {
		log.Println("exec command ", err)
		appendErrtoFile(host.sshHost + "执行报错" + err.Error())
		return
	}
	log.Println(host.sshHost + " 执行远程命令成功........")
	log.Println("命令输出:\n", string(combo))
	appendSuccesstoFile(host.sshHost + " 执行远程命令成功........\n" + string(combo))
}

func appendErrtoFile(erring string) {
	write := bufio.NewWriter(errfile)
	write.WriteString(erring)
	write.Flush()
}

func appendSuccesstoFile(succlog string) {
	write := bufio.NewWriter(successfile)
	write.WriteString(succlog)
	write.Flush()
}
