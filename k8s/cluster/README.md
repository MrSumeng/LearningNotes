# k8s 集群搭建
## 准备工作

### 配置host
```bash
cat >> /etc/hosts << EOF
192.168.116.57 k8s-master
192.168.116.58 k8s-worker01
192.168.116.59 k8s-worker02
EOF
```
### 禁用SELinux
```bash
# 临时
sudo setenforce 0
# 永久
sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
```
### 禁用Swap
```bash
# 临时关闭
swapoff -a 
# 永久关闭
sed -ri 's/.*swap.*/#&/' /etc/fstab
```
###允许 iptables 检查桥接流量
```bash
cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf
br_netfilter
EOF

cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-ip6tables = 1
net.bridge.bridge-nf-call-iptables = 1
EOF
sudo sysctl --system
```


## 安装kubeadm kubectl kubelet
> 注意 所有节点都需要安装 docker kubectl kubelet kubeadm


在这里，我选择使用kubeadm来安装k8s集群
[官方文档](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

### 安装运行时
我选择使用docker，原因是使用惯了 ^ _ ^
已经写过docker安装了，不在赘述，需要的参考下面这篇文章。
[docker 安装](https://www.jianshu.com/p/62c487307f1e)
补充一点 修改一下cgroup
```bash
cat <<EOF | sudo tee /etc/docker/daemon.json
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2",
  "registry-mirrors": ["https://wefdoe4x.mirror.aliyuncs.com"]
}
EOF
#重启docker
systemctl daemon-reload
systemctl restart docker
```
### 安装kubeadm kubelet kubectl

```bash
# 添加k8s源
cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo
[kubernetes]
name=Kubernetes
baseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\$basearch
enabled=1
gpgcheck=1
repo_gpgcheck=1
gpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg
exclude=kubelet kubeadm kubectl
EOF

# 官方的不一定能拉到 换成阿里的
cat > /etc/yum.repos.d/kubernetes.repo << EOF
[kubernetes]
name=Kubernetes
baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
enabled=1
gpgcheck=0
repo_gpgcheck=0
gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
EOF
# 安装kubelet kubeadm kubectl
sudo yum install -y kubelet kubeadm  kubectl --disableexcludes=kubernetes
sudo systemctl enable --now kubelet # systemctl enable kubelet + systemctl start kubelet
```

### 初始化集群
#### 方式一： 直接初始化（推荐）

[出现错误看这里](https://kubernetes.io/zh/docs/setup/production-environment/tools/kubeadm/troubleshooting-kubeadm/)

[网络插件看这里](https://kubernetes.io/zh/docs/concepts/cluster-administration/networking/#how-to-implement-the-kubernetes-networking-model)

**注意**：安装部署节点过程中，很有可以因为防火墙的原因，导致端口不通，所以为了方便测试安装，可以考虑关掉防火墙。


```bash
#初始化master 节点 只有master节点需要执行这个命令
kubeadm init
#漫长的等待 init完成后会有提示 跟着提示操作
export KUBECONFIG=/etc/kubernetes/admin.conf # 这样只是临时的 关闭session后再打开就不可以了
# 持久化
cat >> ~/.bash_profile <<EOF
export KUBECONFIG=/etc/kubernetes/admin.conf
EOF
source ~/.bash_profile

#查看节点状态
kubectl get nodes -A # 现在只有master一个节点 并且状态还是NotReady 
# 这个时候需要安装网络插件 网络插件有很多 我们这里选择 canal
# 我没有直接apply 而是选择先下载下来 方便以后使用
curl -O https://projectcalico.docs.tigera.io/manifests/canal.yaml
kubectl apply -f canal.yaml
# 安装完成后我们get node 就会发现状态从NotReady => Ready

# 查看 加入master节点的token
kubeadm token create --print-join-command
```
接下来的操作是在**worker**节点
> 安装docker kubeadm kubectl kubelet 不在赘述
```bash
# worker节点加入到master节点
kubeadm join 192.168.116.57:6443 --token xxxx   --discovery-token-ca-cert-hash sha256:xxx 
```

##### 修改worker节点的角色
```bash
# 修改角色  master执行
kubectl label node <节点名称> node-role.kubernetes.io/worker=worker
```
### 安装dashboard
```bash
#安装dashboard
curl https://raw.githubusercontent.com/kubernetes/dashboard/v2.5.1/aio/deploy/recommended.yaml -O
kubectl apply -y recommended.yaml
```
#### 访问安装dashboard
 [官方文档](https://github.com/kubernetes/dashboard/blob/master/docs/user/accessing-dashboard/README.md)
#####代理方式
```bash
kubectl proxy
# 访问
http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/
```
#####NodePort
编辑 **kubernetes-dashboard** service
```bash
kubectl -n kubernetes-dashboard edit service kubernetes-dashboard
```
修改IP类型， type：ClusterIP ==> type: NodePort 然后保存退出
```yaml
apiVersion: v1
...
  name: kubernetes-dashboard
  namespace: kubernetes-dashboard
  resourceVersion: "343478"
  selfLink: /api/v1/namespaces/kubernetes-dashboard/services/kubernetes-dashboard
  uid: 8e48f478-993d-11e7-87e0-901b0e532516
spec:
  clusterIP: 10.100.124.90
  externalTrafficPolicy: Cluster
  ports:
  - port: 443
    protocol: TCP
    targetPort: 8443
  selector:
    k8s-app: kubernetes-dashboard
  sessionAffinity: None
  type: ClusterIP # 修改这里 ClusterIP ==> NodePort
status:
  loadBalancer: {}
```
获取服务端口
```bash
kebuctl get service  kubernetes-dashboard -n  kubernetes-dashboard
```
输出信息
```text
NAME                   TYPE       CLUSTER-IP       EXTERNAL-IP   PORT(S)        AGE
kubernetes-dashboard   NodePort   10.100.124.90   <nodes>       443:31707/TCP   21h
```
你可以看到Type变成了NodePort，并且Ports多了一个**31707**端口，这就是我们的nodePort。

如果你是单节点的服务，就可以直接使用https://master-ip:nodePort访问dashboard。

如果你不是单节点服务，你可以需要使用https://node-ip:nodePort访问dashboard。

这里的node-ip指的是运行kubernetes-dashboard服务的节点的IP，可以通过以下方式查看。
```bash
# 查看pod运行在那个节点
kubectl get pods -n kubernetes-dashboard -o wide
# 查看节点IP
kubectl get nodes -A -o wid
```
输出信息
```text
# pod info:kubectl get pods -n kubernetes-dashboard -o wide
NAME                                         READY   STATUS    RESTARTS   AGE   IP          NODE           NOMINATED NODE   READINESS GATES
dashboard-metrics-scraper-799d786dbf-lgnx2   1/1     Running   0          73m   10.36.0.0   k8s-worker02   <none>           <none>
kubernetes-dashboard-fb8648fd9-wnpkt         1/1     Running   0          73m   10.32.0.3   k8s-worker01   <none>           <none>

# node info:kubectl get nodes -A -o wid
NAME           STATUS   ROLES                  AGE   VERSION   INTERNAL-IP      EXTERNAL-IP   OS-IMAGE                KERNEL-VERSION          CONTAINER-RUNTIME
k8s-master     Ready    control-plane,master   18h   v1.23.6   192.168.116.57   <none>        CentOS Linux 7 (Core)   3.10.0-957.el7.x86_64   docker://20.10.14
k8s-worker01   Ready    worker                 18h   v1.23.6   192.168.116.58   <none>        CentOS Linux 7 (Core)   3.10.0-957.el7.x86_64   docker://20.10.14
k8s-worker02   Ready    worker                 18h   v1.23.6   192.168.116.59   <none>        CentOS Linux 7 (Core)   3.10.0-957.el7.x86_64   docker://20.10.14
```
##### 获取token
```bash
# 创建用户
cat > xiaoming.sercrt.yaml << EOF
apiVersion: v1
kind: ServiceAccount
metadata:
  name: xiaoming
  namespace: kubernetes-dashboard
EOF
kubectl create -f xiaoming.sercrt.yaml

# 绑定角色
cat > dashboard.token.yaml << EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: admin-user
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: xiaoming
  namespace: kubernetes-dashboard
EOF
kubectl apply -f dashboard.token.yaml

# 查看创建好的secrets
kubectl -n kubernetes-dashboard get secrets
# 输出
xiaoming-token-rgvpr               kubernetes.io/service-account-token   3      109m
default-token-rjzg9                kubernetes.io/service-account-token   3      124m
kubernetes-dashboard-certs         Opaque                                0      124m
kubernetes-dashboard-csrf          Opaque                                1      124m
kubernetes-dashboard-key-holder    Opaque                                2      124m
# 查看token
kubectl describe secrets -n kubernetes-dashboard xiaoming-token-rgvpr
# 输出
Name:         zhousong-token-8z495
Namespace:    kubernetes-dashboard
Labels:       <none>
Annotations:  kubernetes.io/service-account.name: zhousong
              kubernetes.io/service-account.uid: ef844eb7-37bc-476e-9bd6-ac80eefa4a5b

Type:  kubernetes.io/service-account-token

Data
====
ca.crt:     1099 bytes
namespace:  20 bytes
token:      <这里是很长的token,登录的时候复制这个就可以了>
```


## 其他命令
```bash
# 查看命名空间
kubectl get namespace -A
# 查看节点信息
kubectl get nodes
# 查看pod信息
kubectl get pod -A
# 查看描述信息
kubectl describe <resource tpye> <resource name> -n <namespace>
# 查看日志
kubectl logs [-f] [-p] POD [-c CONTAINER]
# 查看kubelet运行日志
journalctl -xe kubelet
```