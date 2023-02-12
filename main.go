package main

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
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

func getIntOr(val string, defaultVal int) int {
	if val != "" {
		ret, err := strconv.Atoi(val)
		if err != nil {
			panic("fail to convert int for " + val)
		}
		return ret
	}
	return defaultVal
}

func move(from string, to string) (err error) {
	var cmd *exec.Cmd
	cmd = exec.Command("mv", from, to)
	_, err = cmd.Output()
	return
}

func main() {
	storePath := os.Getenv("STORE_PATH")
	uid := getIntOr(os.Getenv("UID"), 0)
	gid := getIntOr(os.Getenv("GID"), 0)
	if len(storePath) == 0 {
		storePath = "."
	}
	fmt.Println("store path: ", storePath)

	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式，上面登录不上的可以尝试切换这种模式

	reloadStorage := openwechat.NewFileHotReloadStorage("storage.json")
	defer reloadStorage.Close()
	err := bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption())
	if err != nil {
		panic("login failed: " + err.Error())
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
			tmpFile := path.Join("/var/tmp", RandStringBytes(16))
			location := path.Join(storePath, fileName(prefix)+"."+ext)
			// create tmp file
			out, err := os.Create(tmpFile)
			if err != nil {
				fmt.Printf("meet error when create file %s for %s\n", tmpFile, err.Error())
			}
			defer out.Close()
			io.Copy(out, response.Body)
			fmt.Printf("save file to %s\n", tmpFile)
			// use move let system know file changes
			err = move(tmpFile, location)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("move %s to %s, let system knows\n", tmpFile, location)
			err = os.Chown(location, uid, gid)
			if err != nil {
				fmt.Println(err)
			}
		}
	}

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}
