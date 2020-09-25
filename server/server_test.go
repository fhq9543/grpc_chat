package main

import (
	"grpcChat/utils"
	"regexp"
	"strings"
	"testing"
)

func TestMPCheck_Check(t *testing.T) {

	ss := "@qing dsadsadsa"
	// 判断是否为私聊。私聊格式   @用户名 消息
	reg := regexp.MustCompile(`^@(.+)\s`)
	if reg != nil {
		list := reg.FindAllStringSubmatch(ss, -1)
		utils.Debug(list)
		if len(list) > 0 && len(list[0]) >= 2 {
			utils.Debug(strings.TrimPrefix(ss, list[0][0]))
		}
	}
}
