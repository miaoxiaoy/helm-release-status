package main

import (
	"fmt"
	"helm-release-status/config"
	pod "helm-release-status/pods"
	"helm-release-status/release"
	"helm-release-status/utils"
	"os"

	"github.com/olekukonko/tablewriter"
)

// var currentUser *user.User

func main() {
	config.CheckArg()

	podTable := tablewriter.NewWriter(os.Stdout)
	podTable.SetHeader([]string{"所属工作集/Job", "Pod名字", "事件归因", "近现时间", "事件消息"})

	resEventTable := tablewriter.NewWriter(os.Stdout)
	resEventTable.SetHeader([]string{"资源名称", "资源类型", "事件归因", "近现时间", "事件消息"})

	// 分析出helm release 存在哪些资源， 并将这些资源的meta信息取出来
	_hasBadEvent := false
	resoureMetadataList := release.GetHelmReleaseResourcesMeta(config.ReleaseName)

	resTable := tablewriter.NewWriter(os.Stdout)
	resTable.SetHeader([]string{"资源名称", "资源类型", "命名空间", "ApiVersion"})

	for _, meta := range resoureMetadataList {
		//fmt.Printf("%+v\n", meta)
		//fmt.Printf("Checking Resource: %s --> %s", meta.Kind, meta.Metadata.Name)
		row := []string{meta.Metadata.Name, meta.Kind, meta.Metadata.Namespace, meta.ApiVersion}
		resTable.Append(row)
		if badPodList, err := pod.GetResourceBadPodList(
			meta.Metadata.Name, meta.Kind, meta.Metadata.Namespace, meta.ApiVersion); err != nil {
			fmt.Printf("ResourceMeta: %+v\n", meta)
			panic(err.Error())
		} else if len(badPodList) > 0 {
			//fmt.Printf("badPodList1: %+v", badPodList)
			pod.GetBadPodReason(meta, badPodList, podTable)
		} else { // pod都正常的情况下检查 resource 自己本身
			//fmt.Printf("badPodList2: %+v", badPodList)
			pod.GetBadResourceReason(meta, resEventTable)
		}
	}
	resTable.SetRowLine(true)
	resTable.Render()
	utils.DefaultLogger.Info("Helm release 资源集合如上表所示")

	if podTable.NumLines() > 0 {
		podTable.SetAutoMergeCells(true)
		podTable.SetRowLine(true)
		podTable.Render()
		_hasBadEvent = true
		utils.DefaultLogger.Error("发现异常 Pod/Container ! 请查阅上面的表格")
	}
	if resEventTable.NumLines() > 0 {
		resEventTable.SetAutoMergeCells(true)
		resEventTable.SetRowLine(true)
		resEventTable.Render()
		_hasBadEvent = true
		utils.DefaultLogger.Error("发现异常 工作集/Job/CronJob ! 请查阅上面的表格")
	}
	if _hasBadEvent {
		utils.DefaultLogger.Error("检测 helm release 资源对象状态结果为 [ 失败 ] !! 请检查")
		os.Exit(1)
	} else {
		utils.DefaultLogger.Success("检测 helm release 所属的所有资源对象和Pod状态结果为 [ 成功 ] !! 大吉大利，今晚吃鸡 :)")
		os.Exit(0)
	}

}
