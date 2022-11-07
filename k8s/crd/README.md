# CRD 学习
## kubebuilder 
### 安装 kubebuilder
[kubebuilder 开源仓库](https://github.com/kubernetes-sigs/kubebuilder)
[kubebuilder 发版地址](https://github.com/kubernetes-sigs/kubebuilder/releases)
在**kubebuilder发版地址**找到我们需要的在kubebuilder版本，把它下载下来。
```bash
# 下载kubebuilder 二进制文件 注意 需要根据自己的系统内核和CPU架构进行下载
sudo curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/v3.3.0/kubebuilder_linux_amd64
# 修改权限
chmod +x kubebuilder
# 移动到bin目录
mv kubebuilder /usr/local/bin/
# 查看版本
kubebuilder version
Version: main.version{KubeBuilderVersion:"3.3.0", KubernetesVendor:"1.23.1", GitCommit:"47859bf2ebf96a64db69a2f7074ffdec7f15c1ec", BuildDate:"2022-01-18T17:03:29Z", GoOs:"linux", GoArch:"amd64"}
```
### 使用kubebuilder搭建CRD和Controller的框架

我们使用刚刚下载好的 *kubebuilder* 帮我们创建一个crd框架。

**注意** 创建之前 我们必须先要有一个k8s集群。

搭建k8s集群请参考:[k8s集群搭建]()。

同时也需要配置好go环境变量:[go环境变量配置]()。

```bash
# 在GOPATH 下创建一下项目
mkdir crd
cd crd
# 初始化项目
kubebuilder init --project-name crd --domain crd.test.zhousong.com

# 查看生成的文件
tree -L 2
```
```text
.
├── config
│   ├── default
│   ├── manager # 部署CRD所需要的yaml
│   ├── prometheus # 监控指标的配置
│   └── rbac # 部署所需的 rbac 授权 yaml
├── Dockerfile
├── go.mod
├── go.sum
├── hack
│   └── boilerplate.go.txt # go 头文件
├── main.go
├── Makefile
└── PROJEC
```

**注意**: 我使用的golang版本为1.18 直接使用 kubebuilder create api命令会出现 controller-gen:no such file or directory的错误。

我们需要修改一下Makefile来避免错误:
修改go-get-tool 使用go install 来替代 go get
```makefile
# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\  # 修改这里 go get => go install
}
endef
```
```bash
# 修该完成后 我们继续创建api
# 创建api
kubebuilder create api --group crd --version v1alpha1 --kind Kidtoy --plural kidtoys
```
