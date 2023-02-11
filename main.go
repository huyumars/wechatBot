package main

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func fileName(prefix string) string {
	now := time.Now()
	return fmt.Sprintf("%s_%s_%s", prefix, now.Format("2006_01_02_15"), RandStringBytes(4))
}

func main() {
	storePath := os.Getenv("STORE_PATH")
	if len(storePath) == 0 {
		storePath = "."
	}
	fmt.Println("store path: ", storePath)

	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式，上面登录不上的可以尝试切换这种模式

	// 注册登陆二维码回调
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 登陆
	if err := bot.Login(); err != nil {
		fmt.Println(err)
		return
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 获取所有的好友
	friends, err := self.Friends()
	fmt.Println(friends, err)

	// 获取所有的群组
	groups, err := self.Groups()
	fmt.Println(groups, err)

	// receive message
	bot.MessageHandler = func(msg *openwechat.Message) {
		var response *http.Response
		var prefix = ""
		if msg.IsPicture() {
			response, err = msg.GetPicture()
			prefix = "IMG"
		}
		if response != nil {
			ct := response.Header.Get("Content-Type")
			pathSplit := strings.Split(ct, "/")
			ext := pathSplit[len(pathSplit)-1]
			location := path.Join(storePath, fileName(prefix)+"."+ext)
			out, err := os.Create(location)
			if err != nil {
				fmt.Printf("meet error when create file %s for %s", location, err.Error())
			}
			defer out.Close()
			io.Copy(out, response.Body)
			fmt.Printf("save file to %s", location)
		}
	}

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}
