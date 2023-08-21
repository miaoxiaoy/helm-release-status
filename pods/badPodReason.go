package pod

import (
	"context"
	"fmt"
	"helm-release-status/config"
	"helm-release-status/release"

	"github.com/olekukonko/tablewriter"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetBadPodReason(parentMate release.ReleaseMeta, podList []v1.Pod, table *tablewriter.Table) {
	for _, pod := range podList {
		events, err := config.Clientset.CoreV1().Events(pod.Namespace).List(
			context.TODO(),
			metav1.ListOptions{
				FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name),
				TypeMeta:      metav1.TypeMeta{Kind: "Pod"}},
		)
		if err != nil {
			panic(fmt.Sprintf("读取pod事件是发生错误=> %s", err.Error()))
		}
		for _, event := range events.Items {

			if event.Type != "Normal" {
				row := []string{fmt.Sprintf("%s\n(%s)", parentMate.Metadata.Name, parentMate.Kind), pod.Name,
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

		// 读取容器日志， 如果有容器
		_hasBadContainer := false
		for _, container := range pod.Status.ContainerStatuses {
			if container.State.Running == nil {
				_hasBadContainer = true
				reason := "未知"
				message := "未知"
				if container.State.Waiting != nil {
					logger.Warning(fmt.Sprintf("%s(%s)的 Pod %s 容器 %s 状态为 %s",
						parentMate.Metadata.Name, parentMate.Kind, pod.Name, container.Name, container.State.Waiting.Reason))
					switch container.State.Waiting.Reason {
					case "CrashLoopBackOff": //典型的应用崩溃起不来
						PrintContainerLog(pod.Name, pod.Namespace, container, config.ContainerLogTailNum)
						reason = fmt.Sprintf("循环重启(%d)", container.RestartCount)
						message = fmt.Sprintf("%s\n请查阅容器日志进行排错", container.State.Waiting.Message)
					case "ContainerCreating":
						reason = "容器创建中"
						message = fmt.Sprintf("%s; 请耐心等待容器启动", container.State.Waiting.Message)
					default:
						if container.State.Terminated != nil {
							reason = container.State.Terminated.Reason
							message = container.State.Terminated.Message
						}
					}
				}
				if container.State.Terminated != nil {
					switch container.State.Terminated.Reason {
					case "Error":
						PrintContainerLog(pod.Name, pod.Namespace, container, config.ContainerLogTailNum)
						reason = container.State.Terminated.Reason
						message = container.State.Terminated.Message
					default:
						reason = container.State.Terminated.Reason
						message = container.State.Terminated.Message
					}
				}
				if reason == "" {
					reason = "程序意外退出"
				}
				if message == "" {
					message = "请查看前面打印的容器日志进行问题定位"
				}
				if reason != "未知" && message != "未知" {
					row := []string{fmt.Sprintf("%s\n(%s)", parentMate.Metadata.Name, parentMate.Kind), pod.Name,
						reason, "", message}
					table.Rich(row, []tablewriter.Colors{
						{},
						{},
						{tablewriter.Normal, tablewriter.FgRedColor},
						{},
						{tablewriter.Normal, tablewriter.FgRedColor},
					})
				}

			} else { // 容器处于running
				if !container.Ready { // 但是未 ready 大概率是就绪探针还未通过
					row := []string{fmt.Sprintf("%s\n(%s)", parentMate.Metadata.Name, parentMate.Kind), pod.Name,
						"容器未就绪", "-", fmt.Sprintf("容器[ %s ]未就绪", container.Name)}
					table.Rich(row, []tablewriter.Colors{
						{},
						{},
						{tablewriter.Normal, tablewriter.FgRedColor},
						{},
						{tablewriter.Normal, tablewriter.FgRedColor},
					})
				}
			}
		}

		// 检查condtion
		for _, condition := range pod.Status.Conditions {
			if condition.Status == v1.ConditionFalse {
				reason := condition.Reason
				message := condition.Message
				if condition.Reason == "ContainersNotReady" && !_hasBadContainer {
					message = fmt.Sprintf("%s\n"+
						"经过检查容器运行中，但容器仍未就绪\n"+
						"请检查readinessProbe、livenessProbe相关探针配置是否合理？", condition.Message)
				}
				row := []string{fmt.Sprintf("%s\n(%s)", parentMate.Metadata.Name, parentMate.Kind), pod.Name,
					reason, condition.LastTransitionTime.Format("2006-01-02\n15:04:05"), message}
				table.Rich(row, []tablewriter.Colors{
					{},
					{},
					{tablewriter.Normal, tablewriter.FgRedColor},
					{},
					{tablewriter.Normal, tablewriter.FgRedColor},
				})
			}
		}

	}

}
