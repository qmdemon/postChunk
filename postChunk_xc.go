package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

//协程版本
//利用http post 分块传输代理

func main() {
	socket := flag.String("s", "127.0.0.1:1234", "设置带IP以及端口")
	flag.Parse()
	fmt.Printf("利用http分块传输协议(chunk)编写的小工具\n    -s 自定义IP以及端口\n    eg: sqlmap -r t.txt --proxy http://%s\n", *socket)

	listener, err := net.Listen("tcp", *socket)
	if err != nil {
		fmt.Println("net.listen err:", err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	ch_httpdata := make(chan string, 1)
	ch_det := make(chan string, 1)
	ch_sconn := make(chan net.Conn, 1)
	go listen(listener, ch_sconn, ch_httpdata, ch_det) //协程 listen 用于获取代理流量并进行post请求处理
	go forward(ch_sconn, ch_httpdata, ch_det)

	wg.Wait()
	//fmt.Println(httpdata)

	//cnt, err = conn.Write([]byte(httpdata))
}

func listen(listener net.Listener, ch_sconn chan net.Conn, ch_httpdata, ch_det chan string) {
	buf := make([]byte, 4096)
	for {

		sconn, err := listener.Accept()
		if err != nil {
			fmt.Println("listenner.accerp err:", err)
			return
		}

		cnt, err := sconn.Read(buf)
		if err != nil {
			fmt.Println("sconn.Read err:", err)
			return
		}
		//httpdata, det := chunk(string(buf[:cnt]))
		chunk2(string(buf[:cnt]), ch_httpdata, ch_det)

		ch_sconn <- sconn
	}
}

func forward(ch_sconn chan net.Conn, ch_httpdata, ch_det chan string) {
	for {
		sconn := <-ch_sconn
		httpdata := <-ch_httpdata
		det := <-ch_det

		dconn, err := net.Dial("tcp", det)
		if err != nil {
			fmt.Println("dconn.Dial err:", err)
			return
		}

		_, err = dconn.Write([]byte(httpdata))

		_, err = io.Copy(sconn, dconn)
		if err != nil {
			fmt.Printf("接收数据失败:%v\n", err)
			sconn.Close()
			dconn.Close()
			return
		}

		sconn.Close()
		dconn.Close()
	}
}

func chunk2(content string, ch_httpdata, ch_det chan string) {

	var method string
	var data string

	clist := strings.Split(content, "\r\n")

	var getdataint int
	var det string

	for index, str := range clist {

		str = strings.TrimSpace(str)

		//分离http 请求
		getrequest_method := strings.Index(str, "HTTP/")
		if getrequest_method != -1 {
			method = strings.Split(str, " ")[0]
		}
		//提取请求头
		if host := strings.Index(str, "Host"); host != -1 {
			det = strings.Split(str, ": ")[1]
			if det_int := strings.Index(det, ":"); det_int == -1 {
				det = det + ":80"
			}
		}

		// 分离出第一个换行 获取出post 的body
		if str == "" {
			getdataint = index + 1
			break

		}
		//设置并替换Transfer-Encoding: chunked 请求头
		if content_length := strings.Index(str, "Content-Length"); content_length != -1 && method == "POST" {
			clist[index] = "Transfer-Encoding: chunked"
		}

	}

	data = strings.Join(clist[getdataint:], "\r\n")

	var cdata string = "\r\n"

	j := 0
	i := 0

	if method == "POST" {
		for j < len(data) {
			time.Sleep(1)
			rand.Seed(time.Now().UnixNano()) // 设置随机数种子  纳秒
			i = j + rand.Intn(4) + 1
			if i > len(data) {
				i = len(data)
			}
			//rand.Seed(time.Now().UnixNano())
			cdata += strconv.Itoa(i-j) + ";" + string(RandLow2(rand.Intn(10))) + "\r\n" + data[j:i] + "\r\n"

			j = i

		}
		cdata += "0\r\n\r\n"

		//fmt.Println("chunk 成功：")
		//fmt.Println(cdata)

	}

	fmt.Printf("数据chunk成功\n")
	httpdata := strings.Join(clist[:getdataint], "\r\n")
	httpdata += cdata

	ch_httpdata <- httpdata
	ch_det <- det
	//return httpdata, det
}

// 随机生成字符串
func RandLow2(n int) []byte {
	var letters = []byte("abcdefghjkmnpqrstuvwxyz123456789")
	if n <= 0 {
		return []byte{}
	}
	b := make([]byte, n)
	arc := uint8(0)
	if _, err := rand.Read(b[:]); err != nil {
		return []byte{}
	}
	for i, x := range b {
		arc = x & 31
		b[i] = letters[arc]
	}
	return b
}
