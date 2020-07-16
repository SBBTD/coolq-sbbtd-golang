package main

import (
	"bufio"
	"encoding/json"
	"github.com/Tnze/CoolQ-Golang-SDK/cqp"
	"os"
	"regexp"
	"strings"
)

type monitorKey struct {
	Group int64  `json:"group"`
	QQ    int64  `json:"qq"`
	Key   string `json:"key"`
}

/*添加线报关键词
0-成功
1-有包含关系且已更新为新词
2-有包含关系新词较长无需更新
3-重复添加关键词
4-黑名单词汇
5-正则解析错误
default-内部错误 */
func addKeyword(list *[]monitorKey, group int64, qq int64, key string) int {
	if key == cmd["all"] {
		return 4
	}
	if key[0] == '/' {
		_, err := regexp.Compile(key)
		if err != nil {
			return 5
		}
	}
	for index, i := range keyList {
		if group == i.Group && qq == i.QQ {
			if i.Key == key {
				return 3
			} else if i.Key[0] != '/' && key[0] != '/' {
				if strings.Contains(i.Key, key) {
					(*list)[index].Key = key
					return 1
				} else if strings.Contains(key, i.Key) {
					return 2
				}
			}
		}
	}
	*list = append(*list, monitorKey{
		Group: group,
		QQ:    qq,
		Key:   key,
	})
	return 0
}

/*移除线报关键词
0-（all=false）单个成功
1-（all=true）全部成功
2-未找到要删除的词
default- 内部错误*/
func delKeyword(list *[]monitorKey, group int64, qq int64, key string, all bool) int {
	isDeleted := false
	for i := 0; i < len(*list); i++ {
		if (*list)[i].Group == group && (*list)[i].QQ == qq && ((*list)[i].Key == key || all) {
			if i == len(*list)-1 {
				*list = (*list)[:i]
			} else {
				*list = append((*list)[:i], (*list)[i+1:]...)
			}
			i--
			isDeleted = true
		}
	}
	if isDeleted {
		if !all {
			return 0
		} else {
			return 1
		}
	} else {
		return 2
	}
}

func saveKeyListToFile(list *[]monitorKey, file string) int {
	s, err := json.Marshal(*list)
	if err != nil {
		cqp.AddLog(cqp.Error, "线报", "保存失败："+err.Error())
		return -1
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		cqp.AddLog(cqp.Error, "线报", "保存失败："+err.Error())
		return -1
	}
	defer func() {
		if err = f.Close(); err != nil {
			cqp.AddLog(cqp.Error, "线报", "保存失败："+err.Error())
		}
	}()
	_, err = f.Write(s)
	if err != nil {
		cqp.AddLog(cqp.Error, "线报", "保存失败："+err.Error())
		return -1
	}
	return 0
}

func loadKeyListFromFile(list *[]monitorKey, file string) int {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		cqp.AddLog(cqp.Error, "线报", "加载失败："+err.Error())
		return -1
	}
	defer func() {
		if err = f.Close(); err != nil {
			cqp.AddLog(cqp.Error, "线报", "加载可能失败了："+err.Error())
		}
	}()
	s, _, _ := bufio.NewReader(f).ReadLine()
	err = json.Unmarshal(s, list)
	if err != nil {
		cqp.AddLog(cqp.Error, "线报", "加载可能失败了："+err.Error())
		return -1
	}
	return 0
}

func loadInt64ListFromFile(list *[]int64, file string) int {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		cqp.AddLog(cqp.Error, "线报", "加载失败："+err.Error())
		return -1
	}
	defer func() {
		if err = f.Close(); err != nil {
			cqp.AddLog(cqp.Error, "线报", "加载可能失败了："+err.Error())
		}
	}()
	s, _, _ := bufio.NewReader(f).ReadLine()
	err = json.Unmarshal(s, list)
	if err != nil {
		cqp.AddLog(cqp.Error, "线报", "加载可能失败了："+err.Error())
		return -1
	}
	return 0
}

func isInInt64List(list []int64, item int64) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}
	return false
}
