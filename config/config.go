package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var ContainerLogTailNum = int64(100)
var ContainerLogShowAll bool = true
var K8sNamespace = "default"
var Settings *cli.EnvSettings
var ActionConfig *action.Configuration

var kubeconfig = "/Users/jaywan/.kube/config"

var Clientset *kubernetes.Clientset
var ReleaseName string

func debug(format string, v ...interface{}) {
	if Settings.Debug {
		format = fmt.Sprintf("[debug] %s\n", format)
		log.Output(2, fmt.Sprintf(format, v...))
	}
}

func CheckArg() {
	if len(flag.Args()) != 1 {
		flag.Usage = func() {
			fmt.Printf("Usage: %s [options] helmReleaseName\n", os.Args[0])
			flag.PrintDefaults()
		}
		flag.Usage()
		os.Exit(1)
	}
	ReleaseName = flag.Args()[0]
	Settings.KubeConfig = kubeconfig
	if err := ActionConfig.Init(Settings.RESTClientGetter(), K8sNamespace, os.Getenv("HELM_DRIVER"), debug); err != nil {
		panic(err.Error())
	}

}

type RESTClientGetter interface {
	ToRESTConfig() (*rest.Config, error)
	ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error)
	ToRESTMapper() (meta.RESTMapper, error)
}

type RestCli struct {
}

func (c *RestCli) ToRESTConfig() (*rest.Config, error) {
	return nil, nil
}

func (c *RestCli) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return nil, nil
}
func (c *RestCli) ToRESTMapper() (meta.RESTMapper, error) {
	return nil, nil
}

func init() {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	// 设定参数
	flag.StringVar(&kubeconfig, "kubeconfig", fmt.Sprintf("%s/.kube/config", currentUser.HomeDir), "kubeconfig file path")
	flag.Int64Var(&ContainerLogTailNum, "container-log-num", 100, "how many log line print if container is waiting")
	flag.BoolVar(&ContainerLogShowAll, "container-log-all", false, "print all log if container is waiting")
	flag.StringVar(&K8sNamespace, "namespace", "default", "kubernetes namespace")
	flag.Parse()

	Settings = cli.New()

	// 初始化配置
	ActionConfig = new(action.Configuration)
	helmDriver := os.Getenv("HELM_DRIVER")

	if err := ActionConfig.Init(Settings.RESTClientGetter(), Settings.Namespace(), helmDriver, debug); err != nil {
		log.Fatal(err)
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(fmt.Sprintf("打开kubeconfig文件时发生错误: %s", err.Error()))
	}
	// create the clientset
	Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Sprintf("初始化kubernetes连接时发生错误: %s", err.Error()))
	}

}
