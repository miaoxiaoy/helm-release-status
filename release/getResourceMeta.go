package release

import (
	"fmt"
	"helm-release-status/config"
	"helm-release-status/utils"
	"os"
	"strings"

	"github.com/goccy/go-yaml"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"
)

type ReleaseMeta struct {
	ApiVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	DeployData struct { // 部署信息
		LatestDeploy time.Time
		Status       release.Status
		Reversion    int
	}
	Env     string // 环境暂没有使用到
	Cluster string // 集群名字，暂时没有使用
}

func GetHelmReleaseResourcesMeta(name string) []ReleaseMeta {
	//CheckArg()
	// 获取 release 的基本状态信息
	status := action.NewStatus(config.ActionConfig)
	release, err := status.Run(name)
	if err != nil {
		utils.DefaultLogger.Error(fmt.Sprintf("获取 Release %s 状态时发生错误=> %s\n", name, err))
		os.Exit(1)
	}

	strReaderForManifest := strings.NewReader(release.Manifest)
	yamlDecoderForManifest := yaml.NewDecoder(strReaderForManifest)
	meta := ReleaseMeta{}
	metaList := []ReleaseMeta{}
	for yamlDecoderForManifest.Decode(&meta) == nil {
		meta.DeployData.LatestDeploy = release.Info.LastDeployed
		meta.DeployData.Reversion = release.Version
		meta.DeployData.Status = release.Info.Status
		meta.Metadata.Namespace = release.Namespace
		metaList = append(metaList, meta)
	}
	// Job在s.Hooks 里面
	hookMetaList := []ReleaseMeta{}
	for _, hook := range release.Hooks {
		strReaderForHooks := strings.NewReader(hook.Manifest)
		yamlDecoderForHooks := yaml.NewDecoder(strReaderForHooks)
		hookMeta := ReleaseMeta{}
		for yamlDecoderForHooks.Decode(&hookMeta) == nil {
			hookMeta.DeployData.LatestDeploy = release.Info.LastDeployed
			hookMeta.DeployData.Reversion = release.Version
			hookMeta.DeployData.Status = release.Info.Status
			hookMeta.Metadata.Namespace = release.Namespace
			hookMetaList = append(hookMetaList, hookMeta)
		}
	}

	return append(metaList, hookMetaList...)
}
