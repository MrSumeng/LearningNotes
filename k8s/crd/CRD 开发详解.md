# 前言

最近一直在从事云原生相关的开发，开发的时候难免要进行 CRD 的控制器和 webhook 的开发。从最开始的磕磕绊绊，到现在还算顺畅，所以打算将这个过程总结了一下，写成文档，让后来者少走些弯路。

> CRD 是什么，大家可以自行查阅，以后有时间精力，我也会写一篇文档总结一下。

这篇文档，主要分为以下几个部分：
+ kubebuilder 安装
+ 项目创建
+ 项目实现
+ 项目部署

在 **kubebuilder 安装**部分中，会和大家如何进行安装 kubebuilder。

在**项目创建**部分会和大家一起使用 kubebuilder 创建我们的项目。

在**项目实现**部分会和大家完成我们的项目代码。

在**项目部署**会部分和大家一起将我们的代码部署到k8s集群。

# kubebuilder 安装

## 介绍

[Kubebuilder 是一个使用自定义资源定义 (CRD)](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-custom-resource-definitions) 构建 Kubernetes API 的框架。它可以提升我们开发CRD的效率，降低我们的开发成本，使我们低成本的进行 k8s  Operator 开发。

和大多数开源项目一样，kubebuilder 的安装非常简单。 只需要下载好我们需要的对应的版本的安装包就可以安装了。

需要注意的是，版本兼容问题：不同 kubebuilder 版本创建出来 controller 可能与 k8s 集群存在一定的兼容性问题，开发前，需要先明确自己的集群版本，然后再去选择 kubebuilder 的版本。



## 安装

在这里我选择的是 v3.3.0 的 kubebuilder 进行安装。

[kubebuilder 开源仓库](https://github.com/kubernetes-sigs/kubebuilder)  

[kubebuilder 发版地址](https://github.com/kubernetes-sigs/kubebuilder/releases)

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

这样，我们就完成了 kubebuilder 的安装，是不是非常简便。

接下来，我们就可以使用它进行项目创建了。

# 项目创建

项目创建中，我们会使用到一些 kubebuilder 常用的命令来帮助我们快速创建项目。

## 命令介绍

我将在实战中使用下面3个命令来创建我们的项目：

```bash
# 初始化项目
# 使用 --domain 可以指定<域>, 我们在这个项目中所创建的所有的API组都将是<group>.<domain>
# 使用 --projiect-name 可以指定我们项目的名称。
kubebuilder init --domain test.crd --project-name test.crd

# 创建API 这个命令可以帮助我们快速创建CRD资源以及CRD控制器
# 使用 --group 可以指定资源组。
# 使用 --version 可以指定资源的版本。
# 使用  --kind 可以指定资源的类型，这个类型就是你自定义的CRD的名字。
kubebuilder create api --group test --version v1 --kind Test

# 创建WebHook
# --group  --kind  --version 和上述一致
# 使用 --defaulting 会为我们生成一个 webhook 的 Deafult 的接口实现，需要我们自己完成，用来补全CRD的默认值
# 使用--programmatic-validation 会为我们生成一个 webhook 的 Validation 的接口实现，要我们自己完成，用来校验CRD资源在创建，更新，删除时数据是否正确。
kubebuilder create webhook --group test --version v1 --kind Test --defaulting --programmatic-validation
```

## 实战

在这里，我们创建一个CRD，这个CRD是用来送外卖订单：
+ kind: Order
+ group: demo
+ domain: sumeng.com
+ version: v1

### 初始化项目

#### 创建项目

```bash
# 创建项目文件夹
mkdir crd-demo
cd crd-demo
# 初始化项目
kubebuilder init --domain sumeng.com --project-name crd-demo
# 查看项目结构
tree
.
├── config # 配置文件目录
│   ├── default
│   │   ├── kustomization.yaml
│   │   ├── manager_auth_proxy_patch.yaml
│   │   └── manager_config_patch.yaml
│   ├── manager
│   │   ├── controller_manager_config.yaml
│   │   ├── kustomization.yaml
│   │   └── manager.yaml
│   ├── prometheus
│   │   ├── kustomization.yaml
│   │   └── monitor.yaml
│   └── rbac
│       ├── auth_proxy_client_clusterrole.yaml
│       ├── auth_proxy_role_binding.yaml
│       ├── auth_proxy_role.yaml
│       ├── auth_proxy_service.yaml
│       ├── kustomization.yaml
│       ├── leader_election_role_binding.yaml
│       ├── leader_election_role.yaml
│       ├── role_binding.yaml
│       └── service_account.yaml
├── Dockerfile # 构建镜像的 Dockerfile
├── go.mod
├── go.sum
├── hack
│   └── boilerplate.go.txt # 代码头文件
├── LICENSE
├── main.go 
├── Makefile
├── PROJECT
└── README.md
```

#### 修改头文件

```bash
# 修改头文件
vim hack/boilerplate.go.txt
```

boilerplate.go.txt

``` txt
/*
Copyright 2022 The CRD-Demo Authors. # 修改了这里

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
```

#### 修改 Makefile 
使用golang版本为1.18时 直接使用 kubebuilder create api命令会出现 *controller-gen:no such file or directory* 的错误。我们需要修改一下 Makefile 来避免错误:
> 修改 go-get-tool 使用go install 来替代 go get

```Makefile
# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\ # 修改这里 go get --> go install
rm -rf $$TMP_DIR ;\
}
endef
```

### 创建API
```bash
# 创建API
# kubebuilder 会询问你是否创建资源和控制器，我们都选择是就可以了。
kubebuilder create api --group demo --version v1 --kind Order

# 查看结构 我们会发现多出来了以下内容。
tree
.
├── api # 我们资源的资源结构存放在这里。
│   └── v1 # 我们定义的版本在结构中以目录的形式呈现出来。
│       ├── groupversion_info.go # 我们定义的CRD的 Group 和 version 信息
│       ├── order_types.go # CRD 结构体定义，我们需要根据自己的需求修改。
│       └── zz_generated.deepcopy.go
├── bin
│   └── controller-gen # 用来生成控制器的工具
├── config # 该目录下的大多数配置文件都不需要手动编写，可以使用工具生成
│   ├── crd # crd 相关配置
│   │   ├── kustomization.yaml
│   │   ├── kustomizeconfig.yaml
│   │   └── patches
│   │       ├── cainjection_in_orders.yaml
│   │       └── webhook_in_orders.yam
│   ├── rbac # 权限相关
│   │   ├── order_editor_role.yaml
│   │   ├── order_viewer_role.yaml
│   └── samples # 使用demo
│       └── demo_v1_order.yaml
├── controllers # 控制器代码目录
│   ├── order_controller.go
│   └── suite_test.go

```


### 创建WebHook

```bash
# 创建WebHook
kubebuilder create webhook --group demo --version v1 --kind Order --defaulting --programmatic-validation
# 查看结构
tree
.
├── api
│   └── v1
│       ├── order_webhook.go # webhook 代码文件
│       └── webhook_suite_test.go
├── config
│   ├── certmanager #认证密钥相关
│   │   ├── certificate.yaml
│   │   ├── kustomization.yaml
│   │   └── kustomizeconfig.yaml
│   ├── default
│   │   ├── manager_webhook_patch.yaml
│   │   └── webhookcainjection_patch.yaml
│   └── webhook
│       ├── kustomization.yaml
│       ├── kustomizeconfig.yaml
│       └── service.yaml

```



# 项目实现

经过上述步骤，我们已经完成了项目的创建，接下来我们就开始项目实现的过程，主要分为以下几步：
+ 项目介绍
+ CRD 编写
+ 控制器实现
+ WebHook实现

## 项目介绍

外卖订单，订单里有客户的点商品信息，商家信息，以及配送信息。

订单状态变更为：
+ "" ---> 未接单
+ 未接单 ---> 已接单
+ 已接单 ---> 制作中
+ 制作中 ---> 制作完成
+ 制作完成 ---> 待配送
+ 待配送 ---> 配送中
+ 配送中 ---> 订单完成

因为我们没有实际的商家和骑手来更新状态，所以我们随机一个时间，来推动订单进程。

接下来，我们就来实现这样一个简单的demo吧！

## CRD 编写

./api/v1/order_types.go

### Spec

Spec 部分主要用于描述我们的自定义资源信息。在这里，我们用来编写订单的信息，如商品信息，商家信息。

```go
// OrderSpec defines the desired state of Order  
type OrderSpec struct {  
   // the information for the Shop  
   // +kubebuilder:validation:Required  
   Shop *ShopInfo `json:"shop"`  
  
   // Commodities is a list of CommodityInfo  
   // +kubebuilder:validation:Required  
   Commodities []CommodityInfo `json:"commodity"`  
  
   // TotalPrice is the total price of the Order  
   // +kubebuilder:validation:Required  
   TotalPrice int64 `json:"totalPrice"`  
  
   // Remark of Order  
   // +optional  
   Remark string `json:"remark,omitempty"`  
}  
  
type ShopInfo struct {  
   // Name of the shop   
   Name string `json:"name"`  
}  
  
type CommodityInfo struct {  
   // Name of the commodity   
   Name string `json:"name"`  
  
   // Price of the commodity   
   Price int64 `json:"price"`  
  
   // Quantity of commodity   
   Quantity int64 `json:"quantity"`  
}
```

上述代码中，我们可以看到一些带 ‘+’ 的注释，这些是我们在生成 crd yaml 文件时的一些标识，他可以帮我做一些简单的功能，节约我们的开发时间。

+ +kubebuilder:validation:Required 用于标记该字段是必填项。
+ +optional  用于标记该字段是可选项

更多的标记我们可以查阅：[Kubebuilder Book](https://book.kubebuilder.io/reference/markers.html)

### Status

status 用来记录资源的状态，这样我们用来记录订单的状态。

```go
// OrderStatus defines the observed state of Order  
type OrderStatus struct {  
   // Conditions a list of conditions an order can have.   
   // +optional   
   Conditions []OrderCondition `json:"conditions,omitempty"`  
   // +optional  
   Phase OrderPhase `json:"phase,omitempty"`  
   // +optional  
   Message string `json:"message,omitempty"`  
}

type OrderCondition struct {  
   // Type of order condition.   Type OrderConditionType `json:"type"`  
   // Phase of the condition, one of True, False, Unknown.  
   Status corev1.ConditionStatus `json:"status"`  
   // The last time this condition was updated.  
   LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`  
   // Last time the condition transitioned from one status to another.  
   LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`  
   // The reason for the condition's last transition.  
   Reason string `json:"reason,omitempty"`  
   // A human-readable message indicating details about the transition.  
   Message string `json:"message,omitempty"`  
}  
  
type OrderConditionType string  
  
const (  
   ConditionShop     OrderConditionType = "Shop"  
   ConditionDelivery OrderConditionType = "Delivery"  
)

type OrderPhase string  
  
const (  
   OrderNotAccepted  OrderPhase = "未接单"  
   OrderAccepted     OrderPhase = "已接单"  
   OrderInMaking     OrderPhase = "制作中"  
   OrderMakeComplete OrderPhase = "制作完成"  
   OrderWaiting      OrderPhase = "待配送"  
   OrderDelivery     OrderPhase = "配送中"  
   OrderFinish       OrderPhase = "订单完成"  
)
```

###  Order

主体部分我们不需要修改，在此处，仅仅只是加上打印标识，方便我们后续观察。

```go
//+genclient  
//+kubebuilder:object:root=true  
//+kubebuilder:subresource:status  
//+kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.phase",description="The order status phase"  
//+kubebuilder:printcolumn:name="MESSAGE",type="string",JSONPath=".status.message",description="The order status message"  
  
// Order is the Schema for the orders APItype Order struct {  
   metav1.TypeMeta   `json:",inline"`  
   metav1.ObjectMeta `json:"metadata,omitempty"`  
  
   Spec   OrderSpec   `json:"spec,omitempty"`  
   Status OrderStatus `json:"status,omitempty"`  
}
```

+ kubebuilder:printcolumn  打印列

上述我们就完成了我们代码CRD的编写，接下来，我们就开始我们代码的控制器的实现。

## 控制器实现

./controllers/order_controller.go

```go 
const MaxSpeedTime int = 60
func (r *OrderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {  
   // Fetch the order instance  
   order := &demov1.Order{}  
   err := r.Get(context.TODO(), req.NamespacedName, order)  
   if err != nil {  
      if errors.IsNotFound(err) {  
         return ctrl.Result{}, nil  
      }  
      // Error reading the object - requeue the request.  
      return ctrl.Result{}, err  
   }  
     
   status := order.Status.DeepCopy()  
   defer func() {  
      err := r.updateScaleStatusInternal(order, *status)  
      if err != nil {  
         klog.Errorf("update order(%s/%s) status failed: %s",  
            order.Namespace, order.Name, err.Error())  
         return  
      }  
   }()  
  
   //Simulate the time spent in each phase  
   speedTime := rand.Int() % MaxSpeedTime  
   switch status.Phase {  
   case "":  
      status.Phase = demov1.OrderNotAccepted  
      status.Message = "Order not accepted"  
   case demov1.OrderNotAccepted:  
      status.Phase = demov1.OrderAccepted  
      status.Message = "Order accepted"  
      cond := NewOrderCondition(demov1.ConditionShop, corev1.ConditionFalse, status.Message, status.Message)  
      SetOrderCondition(status, *cond)  
   case demov1.OrderAccepted:  
      status.Phase = demov1.OrderInMaking  
      status.Message = "Order in making"  
   case demov1.OrderInMaking:  
      status.Phase = demov1.OrderMakeComplete  
      status.Message = "Order make complete"  
      cond := NewOrderCondition(demov1.ConditionShop, corev1.ConditionTrue, status.Message, status.Message)  
      SetOrderCondition(status, *cond)  
   case demov1.OrderMakeComplete:  
      status.Phase = demov1.OrderWaiting  
      status.Message = "Order wait delivery"  
      cond := NewOrderCondition(demov1.ConditionDelivery, corev1.ConditionFalse, status.Message, status.Message)  
      SetOrderCondition(status, *cond)  
   case demov1.OrderWaiting:  
      status.Phase = demov1.OrderDelivery  
      status.Message = "Order delivery"  
   case demov1.OrderDelivery:  
      status.Phase = demov1.OrderFinish  
      status.Message = "Order finished,customer has signed"  
   case demov1.OrderFinish:  
      cond := NewOrderCondition(demov1.ConditionDelivery, corev1.ConditionTrue, "Success", status.Message)  
      SetOrderCondition(status, *cond)  
      return ctrl.Result{}, nil  
   }  
  
   return ctrl.Result{RequeueAfter: time.Duration(speedTime) * time.Second}, nil  
}  
  
func (r *OrderReconciler) updateScaleStatusInternal(scale *demov1.Order, newStatus demov1.OrderStatus) error {  
   if reflect.DeepEqual(scale.Status, newStatus) {  
      return nil  
   }  
   clone := scale.DeepCopy()  
   if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {  
      if err := r.Client.Get(context.TODO(),  
         types.NamespacedName{Name: scale.Name, Namespace: scale.Namespace},  
         clone); err != nil {  
         klog.Errorf("error getting updated scale(%s/%s) from client",  
            scale.Namespace, scale.Name)  
         return err  
      }  
      clone.Status = newStatus  
      if err := r.Client.Status().Update(context.TODO(), clone); err != nil {  
         return err  
      }  
      return nil  
   }); err != nil {  
      return err  
   }  
   oldBy, _ := json.Marshal(scale.Status)  
   newBy, _ := json.Marshal(newStatus)  
   klog.V(5).Infof("order(%s/%s) status from(%s) -> to(%s)", scale.Namespace, scale.Name, string(oldBy), string(newBy))  
   return nil  
}  
  
// SetupWithManager sets up the controller with the Manager.
func (r *OrderReconciler) SetupWithManager(mgr ctrl.Manager) error {  
   //Order status changes do not trigger the reconcile process  
   predicates := builder.WithPredicates(predicate.Funcs{  
      UpdateFunc: func(e event.UpdateEvent) bool {  
         oldObject := e.ObjectOld.(*demov1.Order)  
         newObject := e.ObjectNew.(*demov1.Order)  
         if oldObject.Generation != newObject.Generation || newObject.DeletionTimestamp != nil {  
            klog.V(3).Infof("Observed updated for order: %s/%s", newObject.Namespace, newObject.Name)  
            return true  
         }  
         return false  
      },  
   })  
   return ctrl.NewControllerManagedBy(mgr).  
      For(&demov1.Order{}, predicates).  
      Complete(r)  
}
```

上面便是控制器的逻辑，其实很多时候也和crud没太多区别，我们就在 Reconcile 里实现我们的业务逻辑就可以了。

## WebHook实现

./api/v1/order_webhook.go

```go 
// Default implements webhook.Defaulter so a webhook will be registered for the type
func (in *Order) Default() {  
   // Set the default value.  
   // However, we have noting to do in this crd resources.
}  
  

var _ webhook.Validator = &Order{}  
  
// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (in *Order) ValidateCreate() error {  
   orderlog.Info("validate create", "name", in.Name)  
   return in.Spec.validate()  
}  
  
// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (in *Order) ValidateUpdate(old runtime.Object) error {  
   orderlog.Info("validate update", "name", in.Name)  
   return in.Spec.validate()  
}  
  
// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (in *Order) ValidateDelete() error {  
   return nil  
}  
  
func (in *OrderSpec) validate() error {  
   if in.TotalPrice <= 0 {  
      return fmt.Errorf("total price must be greater than 0")  
   }  
  
   var totalPrice int64 = 0  
   for i := range in.Commodities {  
      err := in.Commodities[i].validate()  
      if err != nil {  
         return err  
      }  
      totalPrice += in.Commodities[i].Price * in.Commodities[i].Quantity  
   }  
  
   if totalPrice != in.TotalPrice {  
      return fmt.Errorf("the total price of the item is incorrect")  
   }  
   return nil  
}  
  
func (in *CommodityInfo) validate() error {  
   if in.Quantity <= 0 {  
      return fmt.Errorf("commodity %s quantity must be greater than 0", in.Name)  
   }  
   if in.Price <= 0 {  
      return fmt.Errorf("commodity %s price must be greater than 0", in.Name)  
   }  
   return nil  
}
```

在上面的代码中，我们只是简单的校验的订单数据，具体的开发中，我们可以根据我们的需求去实现。

# 项目部署

经过上面的步骤，我们项目算是编码完成了，接下来，我们就开始部署我们的服务。

主要分为以下几步：
+ 修改 Makefile
+ 修改 Dockerfile
+ 生成部署文件
+ 部署服务

## 修改 Makefile

大多数情况下，kubebuilder 生产的 makefile 文件已经够我们使用了，但为了更加方便我们的开发与部署，我们可以进行以下改造：

```makfile
# ===remove===  
# Image URL to use all building/pushing image targets  
# IMG ?= controller:latest  

# ===add=== 
# your docker repositories  
REPO ?=  
  
# your project name  
PROJ ?=order  
  
# your project tag or version  
TAG ?=latest

.PHONY: deploy-yaml  
# Generate deploy yaml.  
deploy-yaml: kustomize ## Generate deploy yaml.  
   $(call gen-yamls)

define gen-yamls  
{\  
set -e ;\  
[ -f ${PROJECT_DIR}/_output/yamls/build ] || mkdir -p ${PROJECT_DIR}/_output/yamls/build; \  
rm -rf ${PROJECT_DIR}/_output/yamls/build/${PROJ}; \  
cp -rf ${PROJECT_DIR}/config/* ${PROJECT_DIR}/_output/yamls/build/; \  
cd ${PROJECT_DIR}/_output/yamls/build/manager; \  
${KUSTOMIZE} edit set image controller=${REPO}/${PROJ}:${TAG}; \  
set +x ;\  
echo "==== create deploy.yaml in ${PROJECT_DIR}/_output/yamls/ ====";\  
${KUSTOMIZE} build ${PROJECT_DIR}/_output/yamls/build/default > ${PROJECT_DIR}/_output/yamls/deploy.yaml;\  
}  
endef

# ===modify===
.PHONY: docker-build  
docker-build: ## Build docker image with the manager.  
   docker build --no-cache . -t ${REPO}/${PROJ}:${TAG}  
   docker rmi `docker images | grep none | awk '{print $$3}'` # clean the assets of docker images when build endings  
  
.PHONY: docker-push  
docker-push: ## Push docker image with the manager.  
   docker push ${REPO}/${PROJ}:${TAG}

## Location to install dependencies to  
LOCALBIN ?= $(shell pwd)/bin  
$(LOCALBIN):  
   mkdir -p $(LOCALBIN)  
  
KUSTOMIZE = $(LOCALBIN)/kustomize  
KUSTOMIZE_VERSION ?= v3.8.7  
KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  
.PHONY: kustomize  
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.  
$(KUSTOMIZE): $(LOCALBIN)  
ifeq ($(wildcard $(KUSTOMIZE)),)  
   curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)  
endif

```

上述文件中：
+ 将 docker 镜像名称拆分成了3部分：REPO/PROJ/TAG,方便我们构建镜像。
+ 新增了开始生成 deploy.yaml 的代码。
+ 修改了获取 kustomize 的方式

## 修改 Dockerfile

原来的 Dockerfile 中，每次构建都需要重新拉去依赖，我们可以使用 go vendor 的形式减少省下拉取依赖的时间。

但是这就要求我们在构建代码前，现将依赖拉取到本地。

```Dockerfile
# ===remove===
RUN go mod download
# ===add===
COPY vendor/ vendor/
```

##  生成部署文件

使用 makefile 可以帮我们快速生成相关配置文件

### 生成CRD & RBAC

```bash
# 生成 crd 和 rbac yaml 文件
make manifests
# 查看结构
tree
.
├── config
│   ├── crd
│   │   ├── bases
│   │   │   └── demo.sumeng.com_orders.yaml # crd 文件

```

### 生成部署文件
#### 修改文件
修改 ./config/crd/kustomization.yaml

```yaml
resources:  
- bases/demo.sumeng.com_orders.yaml  
#+kubebuilder:scaffold:crdkustomizeresource  
  
patchesStrategicMerge:  
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix.  
# patches here are for enabling the conversion webhook for each CRD  
- patches/webhook_in_orders.yaml  
#+kubebuilder:scaffold:crdkustomizewebhookpatch  
  
# [CERTMANAGER] To enable cert-manager, uncomment all the sections with [CERTMANAGER] prefix.  
# patches here are for enabling the CA injection for each CRD  
- patches/cainjection_in_orders.yaml  
#+kubebuilder:scaffold:crdkustomizecainjectionpatch  
  
# the following config is for teaching kustomize how to do kustomization for CRDs.  
configurations:  
- kustomizeconfig.yaml
```

修改 ./config/default/kustomization.yaml

```yaml
# Adds namespace to all resources.  
namespace: crd-demo-system  
  
# Value of this field is prepended to the  
# names of all resources, e.g. a deployment named  
# "wordpress" becomes "alices-wordpress".  
# Note that it should also match with the prefix (text before '-') of the namespace  
# field above.  
namePrefix: crd-demo-  
  
# Labels to add to all resources and selectors.  
#commonLabels:  
#  someName: someValue  
  
bases:  
- ../crd  
- ../rbac  
- ../manager  
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in  
# crd/kustomization.yaml  
- ../webhook  
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'. 'WEBHOOK' components are required.  
- ../certmanager  
# [PROMETHEUS] To enable prometheus monitor, uncomment all sections with 'PROMETHEUS'.  
#- ../prometheus  
  
patchesStrategicMerge:  
# Protect the /metrics endpoint by putting it behind auth.  
# If you want your controller-manager to expose the /metrics  
# endpoint w/o any authn/z, please comment the following line.  
- manager_auth_proxy_patch.yaml  
  
# Mount the controller config file for loading manager configurations  
# through a ComponentConfig type  
#- manager_config_patch.yaml  
  
# [WEBHOOK] To enable webhook, uncomment all the sections with [WEBHOOK] prefix including the one in  
# crd/kustomization.yaml  
- manager_webhook_patch.yaml  
  
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER'.  
# Uncomment 'CERTMANAGER' sections in crd/kustomization.yaml to enable the CA injection in the admission webhooks.  
# 'CERTMANAGER' needs to be enabled to use ca injection  
- webhookcainjection_patch.yaml  
  
# the following config is for teaching kustomize how to do var substitution  
vars:  
# [CERTMANAGER] To enable cert-manager, uncomment all sections with 'CERTMANAGER' prefix.  
- name: CERTIFICATE_NAMESPACE # namespace of the certificate CR  
  objref:  
    kind: Certificate  
    group: cert-manager.io  
    version: v1  
    name: serving-cert # this name should match the one in certificate.yaml  
  fieldref:  
    fieldpath: metadata.namespace  
- name: CERTIFICATE_NAME  
  objref:  
    kind: Certificate  
    group: cert-manager.io  
    version: v1  
    name: serving-cert # this name should match the one in certificate.yaml  
- name: SERVICE_NAMESPACE # namespace of the service  
  objref:  
    kind: Service  
    version: v1  
    name: webhook-service  
  fieldref:  
    fieldpath: metadata.namespace  
- name: SERVICE_NAME  
  objref:  
    kind: Service  
    version: v1  
    name: webhook-service
```

修改  ./config/default/manager_auth_proxy_patch.yaml

```yaml
apiVersion: apps/v1  
kind: Deployment  
metadata:  
  name: controller-manager  
  namespace: system  
spec:  
  template:  
    spec:  
      containers:  
      - name: manager  
        args:  
        - "--health-probe-bind-address=:8081"  
        - "--metrics-bind-address=127.0.0.1:8080"  
        - "--leader-elect"
```

修改  ./config/default/webhookcainjection_patch.yaml

```yaml
apiVersion: admissionregistration.k8s.io/v1  
kind: ValidatingWebhookConfiguration  
metadata:  
  name: validating-webhook-configuration  
  annotations:  
    cert-manager.io/inject-ca-from: $(CERTIFICATE_NAMESPACE)/$(CERTIFICATE_NAME)
```

修改  ./config/webhook/manifests.yaml

```yaml
---  
apiVersion: admissionregistration.k8s.io/v1  
kind: ValidatingWebhookConfiguration  
metadata:  
  creationTimestamp: null  
  name: validating-webhook-configuration  
webhooks:  
- admissionReviewVersions:  
  - v1  
  clientConfig:  
    service:  
      name: webhook-service  
      namespace: system  
      path: /validate-demo-sumeng-com-v1-order  
  failurePolicy: Fail  
  name: vorder.kb.io  
  rules:  
  - apiGroups:  
    - demo.sumeng.com  
    apiVersions:  
    - v1  
    operations:  
    - CREATE  
    - UPDATE  
    resources:  
    - orders  
  sideEffects: None
```




#### 生成文件

```bash
# 安装 kustomize
make kustomize
# 生成部署文件
make deploy-yaml REPO=<xxx> TAG=<xxx>
# 生成的部署文件为 ./_output/yamls/deploy.yaml, 我们拿着这个文件去k8s集群部署就可以了
```


## 部署服务

经历了这么多，终于要开始部署服务了。开始之前，我们要先构建我们的服务。

```bash
# 拉取依赖
go mod tidy
go mod vendor
# 生成 deepcopy 文件
make generate
# 制作镜像
make docker-build REPO=<xxx> TAG=<xxx>
# 上传镜像
make docker-push REPO=<xxx> TAG=<xxx>

# 进行制作完成后，为了防止我们的指定的镜像不对，我们可以重新生成部署文件
make deploy-yaml REPO=<xxx> TAG=<xxx>

# 部署服务
# 这个是时候，我们就可以拿着这个文件去我们的集群部署服务了。
kubectl apply -f ./_output/yamls/deploy.yaml

# 我们查看pod，会发现我们的控制器一致起不来，是因为我们添加了webhook服务，需要安装certmanger
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.1/cert-manager.yaml
```

```bash 
kubectl apply -f ./_output/yamls/deploy.yaml
namespace/crd-demo-system created
customresourcedefinition.apiextensions.k8s.io/orders.demo.sumeng.com created
serviceaccount/crd-demo-controller-manager created
role.rbac.authorization.k8s.io/crd-demo-leader-election-role created
clusterrole.rbac.authorization.k8s.io/crd-demo-manager-role created
clusterrole.rbac.authorization.k8s.io/crd-demo-metrics-reader created
clusterrole.rbac.authorization.k8s.io/crd-demo-proxy-role created
rolebinding.rbac.authorization.k8s.io/crd-demo-leader-election-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/crd-demo-manager-rolebinding created
clusterrolebinding.rbac.authorization.k8s.io/crd-demo-proxy-rolebinding created
configmap/crd-demo-manager-config created
service/crd-demo-controller-manager-metrics-service created
service/crd-demo-webhook-service created
deployment.apps/crd-demo-controller-manager created
certificate.cert-manager.io/crd-demo-serving-cert created
issuer.cert-manager.io/crd-demo-selfsigned-issuer created
validatingwebhookconfiguration.admissionregistration.k8s.io/crd-demo-validating-webhook-configuration created
➜  crd-demo k get pod -n crd-demo-system 
NAME                                           READY   STATUS    RESTARTS   AGE
crd-demo-controller-manager-68dc8c65bc-nlkwp   1/1     Running   0          4m5s

```
经过上述步骤，我们的服务就跑起来了。

我们可编写一个资源去验证一下。

./config/samples/demo_v1_order.yaml

``` yaml
apiVersion: demo.sumeng.com/v1  
kind: Order  
metadata:  
  name: order-sample  
spec:  
  shop:  
    name: 鲜果市场  
  commodity:  
    - name: 白菜  
      price: 1  
      quantity: 5  
    - name: 黄瓜  
      price: 2  
      quantity: 5  
  totalPrice: 15  
  remark: 送货上门
```

观察她的状态变化

```bash
 k get orders.demo.sumeng.com order-sample -w
NAME           STATUS   MESSAGE  
order-sample   未接单      Order not accepted
order-sample   已接单      Order accepted
order-sample   制作中      Order in making
order-sample   制作完成     Order make complete
order-sample   待配送      Order wait delivery
```

这样我们便完整的进行了一个 CRD 控制器和 webhook 开发与部署。

# 附录

[github 项目地址](https://github.com/MrSumeng/crd-demo)

[kububuilder book](https://book.kubebuilder.io/introduction.html)

[cretmanager](https://cert-manager.io/docs/installation/)

