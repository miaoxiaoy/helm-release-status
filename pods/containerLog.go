package pod

import (
	"bytes"
	"context"
	"fmt"
	"helm-release-status/config"
	"helm-release-status/release"
	"helm-release-status/utils"
	"io"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var logger *utils.Logger = utils.DefaultLogger
var noHeaderLogger *utils.Logger = utils.NoHeaderLogger

func PrintContainerLog(podName string, namespace string, containerStatus v1.ContainerStatus, lineNum int64) {
	//logger.Info(fmt.Sprintf("正在获取 %s(%s) 容器终端输出日志...", podName, containerStatus.Name))
	logger.Info("--------------------------------------------------日志开始--------------------------------------------------")
	logger.Info(fmt.Sprintf("Pod Name: [ %s ] | Container Name: [ %s ] | 打印长度: 最近 [ %d ] 行", podName, containerStatus.Name, lineNum))
	//fmt.Println()
	podLogOpts := v1.PodLogOptions{
		TailLines:  &lineNum,
		Previous:   containerStatus.RestartCount > 1,
		Container:  containerStatus.Name,
		Timestamps: true,
	}
	if config.ContainerLogShowAll {
		podLogOpts.TailLines = nil
	}
	req := config.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &podLogOpts)
	podLogs, err := req.Stream(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("打开Pod日志流时出错: %s", err.Error()))
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		panic(fmt.Sprintf("复制pod日志到io buffer时出错: %s", err.Error()))
	}
	str := buf.String()
	//fmt.Println(str)

	noHeaderLogger.Error(str)
	logger.Info("--------------------------------------------------日志结束--------------------------------------------------")
	fmt.Println()
}

func GetBadResourceReason(resourceMate release.ReleaseMeta, table *tablewriter.Table) {
	//table := tablewriter.NewWriter(os.Stdout)
	//table.SetHeader([]string{"资源名称", "资源类型", "事件归因", "近现时间", "事件消息"})
	// 根据resource的kind， apiVersion 获取resource的事件
	events, err := config.Clientset.CoreV1().Events(resourceMate.Metadata.Namespace).List(
		context.TODO(),
		metav1.ListOptions{
			FieldSelector: fmt.Sprintf("involvedObject.name=%s", resourceMate.Metadata.Name),
			TypeMeta:      metav1.TypeMeta{Kind: resourceMate.Kind}},
	)
	if err != nil {
		panic(fmt.Sprintf("读取 %s 事件是发生错误=> %s", resourceMate.Kind, err.Error()))
	}
	for _, event := range events.Items {
		if event.Type != "Normal" {
			row := []string{fmt.Sprintf("%s\n(%s)", resourceMate.Metadata.Name, resourceMate.Kind), resourceMate.Kind,
				event.Reason, event.LastTimestamp.Format("2006-01-02\n15:04:05"), event.Message}
			table.Rich(row, []tablewriter.Colors{
				{},
				{},
				{tablewriter.Normal, tablewriter.FgRedColor},
				{},
				{tablewriter.Normal, tablewriter.FgRedColor},
			})
			//table.Append(row)  table.Rich会自动append
		}
	}
	if resourceMate.Kind == "Job" {
		job, _ := config.Clientset.BatchV1().Jobs(resourceMate.Metadata.Namespace).Get(context.TODO(), resourceMate.Metadata.Name, metav1.GetOptions{})
		if job.Status.Active > 0 {
			row := []string{fmt.Sprintf("%s\n(%s)", resourceMate.Metadata.Name, resourceMate.Kind), resourceMate.Kind,
				"Job 未完成", job.Status.StartTime.Format("2006-01-02\n15:04:05"), "Job运行中，若Job执行时间过长，请检查代码逻辑是否合理。"}
			table.Rich(row, []tablewriter.Colors{
				{},
				{},
				{tablewriter.Normal, tablewriter.FgYellowColor},
				{},
				{tablewriter.Normal, tablewriter.FgYellowColor},
			})
		}

		if job.Spec.Completions != nil && job.Status.Succeeded < *job.Spec.Completions {
			_startTime := "未开始"
			if job.Status.StartTime != nil {
				_startTime = job.Status.StartTime.Format("2006-01-02\n15:04:05")
			}
			row := []string{fmt.Sprintf("%s\n(%s)", resourceMate.Metadata.Name, resourceMate.Kind), resourceMate.Kind,
				"Job 成功次数未达标", _startTime,
				fmt.Sprintf("Job执行成功次数未达到指定次数: %d", *job.Spec.Completions)}
			table.Rich(row, []tablewriter.Colors{
				{},
				{},
				{tablewriter.Normal, tablewriter.FgYellowColor},
				{},
				{tablewriter.Normal, tablewriter.FgYellowColor},
			})
		}

	}

}
