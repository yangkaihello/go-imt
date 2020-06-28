package imt

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// 获取IP地址
// 参数：IP地址
//"CN|北京|北京|None|CHINANET|1|None"
func LocationIP(ip string) ([]string,error) {
	client := http.Client{Timeout: 4 * time.Second}
	resp, err := client.Get(`http://api.map.baidu.com/location/ip?ak=MbLBTFTfMcv4f9gk6WlGk6oz6rT0VPad&ip=`+ip+`&coor=bd09ll`)
	if err != nil {
		return nil,err
	}
	defer resp.Body.Close()
	var buffer [512]byte
	result := bytes.NewBuffer(nil)
	for {
		n, err := resp.Body.Read(buffer[0:])
		result.Write(buffer[0:n])
		if err != nil && err == io.EOF {
			break
		}else{
			return nil,errors.New("接口异常")
		}
	}
	var s = &struct {
		Address string `json:"address"`
	}{}
	var resultString = result.String()
	json.Unmarshal([]byte(resultString),s)
	var address = strings.Split(s.Address,"|")
	return address,nil
}

func JsonEncode(v interface{}) string {
	var b,_ = json.Marshal(v)
	return string(b)
}

func RandInt(length int) int {
	if length > 9 || length < 1 {
		return 0
	}
	var number = 1
	var numberVerify int
	for i:=0; i<length ; i++ {
		number = number*10
	}
	number = number-1
	numberVerify = number/10
	rand.Seed(time.Now().UnixNano())
	randInt := rand.Intn(number)

	if randInt == 0 {
		randInt = 1
	}
	for i:=0; i<length ; i++ {
		if randInt > numberVerify {
			break
		}else{
			randInt = randInt*10
		}
	}
	return randInt
}

//验证0-9的字符集
func ASCIINumber(ByteDec byte) bool {
	if ByteDec >= 48 && ByteDec <= 57 {
		return true
	} else {
		return false
	}
}

//验证a-z A-Z的字符集
func ASCIILetter(ByteDec byte) bool {
	if (ByteDec >= 65 && ByteDec <= 90) || ( ByteDec >= 97 && ByteDec <= 122) {
		return true
	} else {
		return false
	}
}
