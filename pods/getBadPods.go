package pod

import (
	"context"
	"fmt"
	"helm-release-status/config"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func FilterBadPod(namespace string, label string, parentUid string) (badPodList []v1.Pod, err error) {

	podSet, err := config.Clientset.CoreV1().Pods(namespace).List(
		context.TODO(),
		metav1.ListOptions{
			LabelSelector: label,
		},
	)
	if err != nil {
		return badPodList, err
	}

	for _, pod := range podSet.Items {
		_found := false
		for _, ref := range pod.OwnerReferences {
			//fmt.Printf("ref.UID: %s  parentUid: %s\n", string(ref.UID), parentUid)
			if string(ref.UID) == parentUid {
				_found = true
				break
			}
		}

		if pod.Status.Phase == v1.PodSucceeded { //job的pod， 已执行完成的Pod，直接跳过
			continue
		}

		if _found { // pod与parentUid一致
			_hasBadCondition := false
			for _, c := range pod.Status.Conditions {
				if c.Status == v1.ConditionFalse {
					_hasBadCondition = true
				}
			}
			if _hasBadCondition {
				badPodList = append(badPodList, pod)
			}

		}
	}
	return badPodList, nil
}

func GetResourceBadPodList(name, kind, namespace, apiVersion string) (allBadPodList []v1.Pod, err error) {

	switch kind {
	case "Deployment":
		podTemplateHash := []map[string]string{}
		if rs, err := config.Clientset.AppsV1().ReplicaSets(namespace).List(
			context.TODO(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", name)},
		); err != nil {
			//panic(err.Error())
			return allBadPodList, err
		} else {
			// 获取期望pod数大于0的rs的label.pod-template-hash
			for _, r := range rs.Items {
				//fmt.Printf("RS---> %+v\n", r)
				if *r.Spec.Replicas > 0 {
					podData := map[string]string{}
					podData["rsName"] = r.Name
					podData["podTemplateHash"] = r.Labels["pod-template-hash"]
					podData["parentUID"] = string(r.UID)
					podTemplateHash = append(podTemplateHash, podData)
				}
			}
			//fmt.Printf("%+v\n", podTemplateHash)
			// 找出不正常的pod列表
			for _, podData := range podTemplateHash {

				lableString := fmt.Sprintf(
					"app.kubernetes.io/name=%s,pod-template-hash=%s", name, podData["podTemplateHash"])
				badPodList, err := FilterBadPod(namespace, lableString, podData["parentUID"])
				//fmt.Printf("HELLO123: %+v\n", len(badPodList))
				if err != nil {
					return allBadPodList, err
				}
				allBadPodList = append(allBadPodList, badPodList...)
				//fmt.Printf("HELLO456: %+v\n", len(allBadPodList))
			}

		}
		//fmt.Printf("HELL789: %+v\n", len(allBadPodList))
		return allBadPodList, err
	case "DaemonSet":
		if ds, err := config.Clientset.AppsV1().DaemonSets(namespace).List(
			context.TODO(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", name)},
		); err != nil {
			return allBadPodList, err
		} else {
			for _, d := range ds.Items {
				labelString := fmt.Sprintf("app.kubernetes.io/name=%s", name)
				badPodList, err := FilterBadPod(namespace, labelString, string(d.UID))
				if err != nil {
					return allBadPodList, nil
				}
				allBadPodList = append(allBadPodList, badPodList...)
			}

		}
		return allBadPodList, err
	case "StatefulSet":
		if ss, err := config.Clientset.AppsV1().StatefulSets(namespace).List(
			context.TODO(),
			metav1.ListOptions{LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", name)},
		); err != nil {
			return allBadPodList, err
		} else {
			for _, s := range ss.Items {
				labelString := fmt.Sprintf("app.kubernetes.io/name=%s", name)
				badPodList, err := FilterBadPod(namespace, labelString, string(s.UID))
				if err != nil {
					return allBadPodList, nil
				}
				allBadPodList = append(allBadPodList, badPodList...)
			}
		}
		return allBadPodList, err
	case "Job":
		if job, err := config.Clientset.BatchV1().Jobs(namespace).List(
			context.TODO(),
			//metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s,metedata.namespace=%s", name, namespace)},
			metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", name)},
		); err != nil {
			return allBadPodList, err
		} else {
			for _, s := range job.Items {
				labelString := fmt.Sprintf("controller-uid=%s,job-name=%s", s.UID, s.Name)
				badPodList, err := FilterBadPod(namespace, labelString, string(s.UID))
				if err != nil {
					return allBadPodList, nil
				}
				allBadPodList = append(allBadPodList, badPodList...)
			}
		}
		return allBadPodList, err
	case "CronJob":
		if apiVersion == "batch/v1" {
			cronjobs, err := config.Clientset.BatchV1().CronJobs(namespace).List(
				context.TODO(),
				metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", name)},
			)
			if err != nil {
				return allBadPodList, err
			}
			for _, s := range cronjobs.Items {
				labelString := fmt.Sprintf("controller-uid=%s,job-name=%s", s.Labels["controller-uid"], s.Name)
				badPodList, err := FilterBadPod(namespace, labelString, string(s.UID))
				if err != nil {
					return allBadPodList, nil
				}
				allBadPodList = append(allBadPodList, badPodList...)
			}
		} else if apiVersion == "batch/v1beta1" {
			cronjobs, err := config.Clientset.BatchV1beta1().CronJobs(namespace).List(
				context.TODO(),
				metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", name)},
			)
			if err != nil {
				return allBadPodList, err
			}
			for _, s := range cronjobs.Items {
				labelString := fmt.Sprintf("controller-uid=%s,job-name=%s", s.Labels["controller-uid"], s.Name)
				badPodList, err := FilterBadPod(namespace, labelString, string(s.UID))
				if err != nil {
					return allBadPodList, nil
				}
				allBadPodList = append(allBadPodList, badPodList...)
			}
		}

		return allBadPodList, nil
	}

	return allBadPodList, nil
}
