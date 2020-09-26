package redis

import (
	"grpcChat/utils"
	"testing"
)

func TestRedis(t *testing.T) {
	InitRedis()
	execptStr := "666"
	RedisSet("xiniao", execptStr)
	resStr, err := RedisGet("xiniao")
	if !utils.Check(err) {
		t.Fatal(err)
	}
	if execptStr != resStr {
		t.Fatalf("RedisGet execpted:%s, result:%s", execptStr, resStr)
	}

	execptList := []string{"a", "b", "c", "d"}
	for _, v := range execptList {
		err = RedisRPush("xiniaoList", v)
		if !utils.Check(err) {
			t.Fatal(err)
		}
	}
	resList, err := RedisGetList("xiniaoList")
	if !utils.Check(err) {
		t.Fatal(err)
	}
	if len(resList) != len(execptList) {
		t.Fatalf("RedisGetList execpted:%s, result:%s", execptList, resList)
	}
	for i, _ := range resList {
		if resList[i] != execptList[i] {
			t.Fatalf("RedisGetList execpted:%s, result:%s", execptList, resList)
		}
	}

	resList = RedisPopList("xiniaoList")
	if len(resList) != len(execptList) {
		t.Fatalf("RedisPopList execpted:%s, result:%s", execptList, resList)
	}
	for i, _ := range resList {
		if resList[i] != execptList[i] {
			t.Fatalf("RedisPopList execpted:%s, result:%s", execptList, resList)
		}
	}
}
