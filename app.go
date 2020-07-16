package main

import (
	"fmt"
	"github.com/Tnze/CoolQ-Golang-SDK/cqp"
	"regexp"
	"strconv"
	. "strings"
)

var (
	keyList          []monitorKey
	monitorGroupList []int64
	blockQQList      []int64
	reg              *regexp.Regexp
	cmd              = map[string]string{
		"add":  "线报",
		"del":  "移除线报",
		"all":  "全部",
		"list": "当前线报",
	}
	groupName = map[int64]string{
		707965661: "debug群",
		635605854: "捡垃圾研究院",
		945583797: "哈佛捡垃圾",
		970458851: "King线报群",
		367943101: "三表哥线报群",
		699788908: "冷瞳活动分享",
		782790346: "Ben笨线报群",
		740897949: "疯子vip共享",
	}
)

//go:generate cqcfg -c .
// cqp: 名称: coolq-sbbtd
// cqp: 版本: 2.0.0:1
// cqp: 作者: 随便吧土豆
// cqp: 简介: sbbtd自用插件，线报等功能
func main() { /*此处应当留空*/ }

func init() {
	cqp.AppID = "tk.sbbtd.selfuse.app"
	cqp.Start = onStart
	cqp.GroupMsg = onGroupMsg
	cqp.DiscussMsg = onDiscussMsg
	cqp.GroupMemberIncrease = onGroupMemberIncrease
}

func onStart() int32 {
	loadInt64ListFromFile(&blockQQList, "settings_blockqq.json")
	loadInt64ListFromFile(&monitorGroupList, "settings_monitorgroup.json")
	loadKeyListFromFile(&keyList, "settings_monitorlist.json")
	reg = regexp.MustCompile(`[￥$][A-Za-z0-9]*?[￥$]|https?://[A-Za-z0-9/.]*|\[CQ:[A-Za-z0-9-_.]*.*?]`)
	return 0
}

func onGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	if isInInt64List(blockQQList, fromQQ) {
		return 1
	}

	if isInInt64List(monitorGroupList, fromGroup) {
		msgFmt := ToLower(reg.ReplaceAllLiteralString(msg, ""))
		for _, k := range keyList {
			isMatch := false
			if HasPrefix(k.Key, "/") {
				if t, err := regexp.MatchString(k.Key[1:], msgFmt); err == nil && t {
					isMatch = true
				}
			} else if Contains(msgFmt, k.Key) {
				isMatch = true
			}
			if isMatch {
				gName := groupName[fromGroup]
				if gName == "" {
					gName = strconv.FormatInt(fromGroup, 10)
				}
				repMsg := fmt.Sprintf("[CQ:at,qq=%d]\n线报“%s”来自%s\n%s", k.QQ, k.Key, gName, msg)
				cqp.SendGroupMsg(k.Group, repMsg)
			}
		}
	}

	repMsg := ""
	msgFmt := ReplaceAll(ReplaceAll(ReplaceAll(msg, "&amp;", "&"), "&#91;", "["), "&#93;", "]")
	if HasPrefix(msgFmt, cmd["add"]) {
		if t, err := regexp.MatchString(`\[CQ:[A-Za-z0-9-_.]*?.*?]`, msg); err == nil && !t {
			keyWord := msgFmt[len(cmd["add"]):]
			if keyWord[0] != '/' {
				keyWord = ToLower(keyWord)
			}
			switch addKeyword(&keyList, fromGroup, fromQQ, keyWord) {
			case 0:
				repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n已添加“%s”，将转发线报至本群并@你。注：如有字母均小写匹配，无需手动处理大小写", fromQQ, keyWord)
			case 1:
				repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n与已有关键词有包含关系，已更新为较短的新关键词", fromQQ)
			case 2:
				repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n与已有关键词存在包含关系，且已有关键词较短，无需添加", fromQQ)
			case 3:
				repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n已经添加过这个关键词了，请勿重复添加", fromQQ)
			case 4:
				repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n添加失败，黑名单词汇", fromQQ)
			case 5:
				repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n正则编译失败了，请检查语法。"+
					"另:1.前缀斜杠不是正则的内容；"+
					"2.不支持\\<number>形式的引用语法；"+
					"3.只支持go语言的正则特性，可以参考https://regex101.com/在左侧选择golang进行在线正则语法测试。", fromQQ)
			default:
				repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n添加失败，未知内部错误", fromQQ)
			}
			cqp.SendGroupMsg(fromGroup, repMsg)
			saveKeyListToFile(&keyList, "settings_monitorlist.json")
		} else {
			cqp.SendGroupMsg(fromGroup, fmt.Sprintf("[CQ:at,qq=%d]不要加奇奇怪怪的表情哦！", fromQQ))
		}
	} else if HasPrefix(msgFmt, cmd["del"]) {
		keyWord := msgFmt[len(cmd["del"]):]
		if keyWord[0] != '/' {
			keyWord = ToLower(keyWord)
		}
		switch delKeyword(&keyList, fromGroup, fromQQ, keyWord, keyWord == cmd["all"]) {
		case 0:
			repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n已移除你在当前群的线报关键词“%s”", fromQQ, keyWord)
		case 1:
			repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n已移除你在当前群的全部线报关键词", fromQQ)
		case 2:
			repMsg = fmt.Sprintf(
				"[CQ:at,qq=%d]\n未能找到你在当前群内追踪的线报关键词“%s”，可尝试发送“%s”来查询，或发送“%s”移除你在当前群内的所有追踪词",
				fromQQ, keyWord, cmd["list"], cmd["del"]+cmd["all"])
		default:
			repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n添加失败，未知内部错误", fromQQ)
		}
		cqp.SendGroupMsg(fromGroup, repMsg)
		saveKeyListToFile(&keyList, "settings_monitorlist.json")
	} else if HasPrefix(msgFmt, cmd["list"]) {
		isHasKey := false
		repMsg = fmt.Sprintf("[CQ:at,qq=%d]\n你在当前群的线报为：", fromQQ)
		for _, i := range keyList {
			if i.Group == fromGroup && i.QQ == fromQQ {
				repMsg += "\n" + ReplaceAll(ReplaceAll(ReplaceAll(i.Key, "&", "&amp;"), "[", "&#91;"), "]", "&#93;")
				isHasKey = true
			}
		}
		if isHasKey {
			cqp.SendGroupMsg(fromGroup, repMsg)
		} else {
			cqp.SendGroupMsg(fromGroup, fmt.Sprintf("[CQ:at,qq=%d]\n你在当前群没有追踪线报关键词。", fromQQ))
		}
	}
	return 0
}

func onDiscussMsg(subType, msgID int32, fromDiscuss, fromQQ int64, msg string, font int32) int32 {
	if isInInt64List(blockQQList, fromQQ) {
		return 1
	}
	return 0
}

func onGroupMemberIncrease(subType, sendTime int32, fromGroup, fromQQ, beingOperateQQ int64) int32 {
	cqp.SendGroupMsg(fromGroup, fmt.Sprintf("欢迎 [CQ:at,qq=%d] 加群！\nQQ：%d\n管理：%d", beingOperateQQ, beingOperateQQ, fromQQ))
	return 0
}
