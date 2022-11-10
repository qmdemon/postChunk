package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

//单线程版本(好像有点问题)
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

	for {
		sconn, err := listener.Accept()
		if err != nil {
			fmt.Println("listenner.accerp err:", err)
			return
		}
		buf := make([]byte, 1024)
		cnt, err := sconn.Read(buf)
		if err != nil {
			fmt.Println("sconn.Read err:", err)
			return
		}
		httpdata, det := chunk(string(buf[:cnt]))

		dconn, err := net.Dial("tcp", det)
		if err != nil {
			fmt.Println("dconn.Dial err:", err)
			return
		}

		_, err = dconn.Write([]byte(httpdata))

		_, err = io.Copy(sconn, dconn)
		if err != nil {
			fmt.Printf("接收数据失败:%v\n", err)
			return
		}

		sconn.Close()
		dconn.Close()
	}

	//fmt.Println(httpdata)

	//cnt, err = conn.Write([]byte(httpdata))
}

func chunk(content string) (string, string) {

	var method string
	var data string

	clist := strings.Split(content, "\r\n")

	var getdataint int
	var det string

	for index, str := range clist {

		str = strings.TrimSpace(str)

		getrequest := strings.Index(str, "HTTP/")
		if getrequest != -1 {
			method = strings.Split(str, " ")[0]
		}
		if host := strings.Index(str, "Host"); host != -1 {
			det = strings.Split(str, ": ")[1]
			if det_int := strings.Index(det, ":"); det_int == -1 {
				det = det + ":80"
			}
		}
		if str == "" {
			getdataint = index + 1
			break

		}
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
			rand.Seed(time.Now().UnixNano())
			i = j + rand.Intn(4) + 1
			if i > len(data) {
				i = len(data)
			}

			cdata += strconv.Itoa(i-j) + ";" + string(RandLow(rand.Intn(10))) + "\r\n" + data[j:i] + "\r\n"

			j = i
		}
		cdata += "0\r\n\r\n"
		//fmt.Println("chunk 成功：")
		//fmt.Println(cdata)

	}

	httpdata := strings.Join(clist[:getdataint], "\r\n")
	httpdata += cdata
	return httpdata, det
}

func RandLow(n int) []byte {
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
