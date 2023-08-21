> 本程序用于解析 Helm release，查找问题Pod或容器，并打印容器日志和Pod事件
### 逻辑
release 名字作为参数， 解析出该release所属的工作集、Job等包含pod的资源，并检查下属的Pod和容器状态是否正常
若资源所属的Pod状态或容器状态不ok， 则打印出容器日志和Pod警告事件
### 构建
`CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o helm-release-pod-log_linux-amd64`


### 用法
```shell
$ ./helm-release-pod-log_linux-amd64 -h
Usage of ./helm-release-pod-log_linux-amd64:
  -container-log-all
        print all log if container is waiting
  -container-log-num int
        how many log line print if container is waiting (default 100)
  -kubeconfig string
        kubeconfig file path (default "/root/.kube/config")
```
### 示例
```shell
$ ./helm-release-pod-log_linux-amd64 --kubeconfig=/your/kube/config --container-log-num=200 any-release
[WARN] wanjie-test2(Deployment)的 Pod wanjie-test2-7f4f46dfc-qfwsl 容器 wanjie-test2 状态为 CrashLoopBackOff
[INFO] --------------------------------------------------日志开始--------------------------------------------------
[INFO] Pod Name: [ wanjie-test2-7f4f46dfc-qfwsl ] | Container Name: [ wanjie-test2 ] | 打印长度: 最近 [ 100 ] 行
[ERROR] 2022-05-07T09:25:39.346164220Z Traceback (most recent call last):
2022-05-07T09:25:39.346211022Z   File "/workdir/run.py", line 1, in <module>
2022-05-07T09:25:39.346218739Z     from app import app
2022-05-07T09:25:39.346223057Z   File "/workdir/app/__init__.py", line 1, in <module>
2022-05-07T09:25:39.346230056Z     from flask import Flask
2022-05-07T09:25:39.346236346Z   File "/usr/local/lib/python3.9/site-packages/flask/__init__.py", line 14, in <module>
2022-05-07T09:25:39.346293927Z     from jinja2 import escape
2022-05-07T09:25:39.346309117Z ImportError: cannot import name 'escape' from 'jinja2' (/usr/local/lib/python3.9/site-packages/jinja2/__init__.py)

[INFO] --------------------------------------------------日志结束--------------------------------------------------

+----------------+------------------------------+--------------+------------+--------------------------------------------------------------------------------+
| 所属工作集/JOB |           POD名字            |   事件归因   |  近现时间  |                                    事件消息                                    |
+----------------+------------------------------+--------------+------------+--------------------------------------------------------------------------------+
| wanjie-test2   | wanjie-test2-7f4f46dfc-qfwsl | BackOff      | 2022-05-07 | Back-off restarting failed                                                     |
| (Deployment)   |                              |              | 17:25:51   | container                                                                      |
+                +                              +--------------+------------+--------------------------------------------------------------------------------+
|                |                              | 循环重启(34) |            | back-off 5m0s restarting failed container=wanjie-test2                         |
|                |                              |              |            | pod=wanjie-test2-7f4f46dfc-qfwsl_staging(947145ac-1427-44aa-809b-ba84028ce026) |
|                |                              |              |            | 请查阅容器日志进行排错                                                         |
+----------------+------------------------------+--------------+------------+--------------------------------------------------------------------------------+

```