# Kubectl 安装

[官方文档](https://kubernetes.io/zh/docs/tasks/tools/)

[kubectl for linux](https://kubernetes.io/zh/docs/tasks/tools/install-kubectl-linux/)

> 操作系统：centos7.5
> 命令行： bash

## 安装 Kubectl
```bash
#下载安装包 如果需要指定版本 使用版本号替换 $(curl -L -s https://dl.k8s.io/release/stable.txt) 即可
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
#验证可执行文件
#下载校验和
curl -LO "https://dl.k8s.io/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl.sha256"
#验证
echo "$(cat kubectl.sha256)  kubectl" | sha256sum --check 
# 输出 kubectl: OK 则验证通过
# 未通过重新下载即可

# 安装kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
# 执行不通过可以手动给权限
sudo chmod +x kubectl && mv kubectl /usr/local/bin/kubectl
# 查看版本
kubectl version --client
#yaml格式输出
kubectl version --client --output=yaml
```
## kubectl命令自动补全工具---bash-completion

因为我使用的是bash，没有安装zsh或其他的命令行工具，所以选择的bash-completion。其他的命令行工具可以去网上查找对应的资源。

```bash
#检查是否安装bash-completion
type _init_completion
#安装bash-completion
yum install bash-completion
#编辑 ~/.bashrc 或者 /etc/bashrc 文件 加入下面代码
# 内容
[[ $PS1 && -f /usr/share/bash-completion/bash_completion ]] && \
    . /usr/share/bash-completion/bash_completion

# 刷新配置
source ~/.bashrc # or source /etc/bashrc
#检查是否安装成功
type _init_completion
#启用kubectl自动补全
#只给当前用户设置
echo 'source <(kubectl completion bash)' >>~/.bashrc
#系统全局设置
kubectl completion bash | sudo tee /etc/bash_completion.d/kubectl > /dev/null
```
### 懒人设置 --- kubectl => k
```bash
# 关联别名:kubectl单词太长,用k代替，kubectl=>k
echo 'alias k=kubectl' >>~/.bashrc # echo 'alias k=kubectl' >>/etc/bashrc
echo 'complete -F __start_kubectl k' >>~/.bashrc #echo 'complete -F __start_kubectl k' >>/etc/bashrc
#刷新配置
source ~/.bashrc # source /etc/bashrc
```
