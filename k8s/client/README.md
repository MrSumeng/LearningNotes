#Kubernetes 学习笔记(三）--- 访问集群
> [官方文档](https://kubernetes.io/zh/docs/tasks/access-application-cluster/access-cluster/)

我这里分享的是通过编程的方式去访问k8s集群，之所以分享这种方式，是因为我们可能有需要对k8s进行扩展，或者二次开发的需求，而我在学习时，苦于没有简单明了的文档来帮助我学习,以至于走了很多弯路，所以我把我的学习历程记录了下来，分享给大家，有什么不对的地方，还请大家指正。

## 前提
+ 你需要有一个k8s集群
+ 将k8s集群的admin.conf文件拷贝到你访问的地方
+ 保护好你的admin.conf不要泄露-它拥有完全的k8s控制权限

如果你没有k8s集群，可以参考这篇文档，搭建一个自己的[集群](https://www.jianshu.com/p/aa684a76ee7f)

## 获取客户端
k8s官方给我们提供几种获取k8s客户端方式，他们都被封装 [client-go](https://github.com/kubernetes/client-go) 这个package里

### 1: RESTClient
RESTClint是k8s官方提供的最基础的客户端，它将常见的 Kubernetes API 约定封装在一起。
它提供的Kubernetes API有：
+ GetRateLimiter() flowcontrol.RateLimiter:用于获取限速器
+ Verb(verb string) *Request ：用于设置请求的操作
+ Post() *Request ：用于创建资源
+ Put() *Request ：用于更新资源
+ Patch(pt types.PatchType) *Request ：用于更新资源，与PUT有一定不同 
+ Get() *Request ：用于获取资源
+ Delete() *Request ：用于删除资源
+ APIVersion() schema.GroupVersion ：用于获取GroupVersion

```go
package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var nameSpace = "kube-system"
var podName = "coredns-64897985d-cmmk9"

func main() {
	// 获取k8s集群的admin.conf 并转换为*rest.Config
	config := getConfig()
	config.GroupVersion = &v1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	config.APIPath = "/api"
	// 获取RESTClient
	client, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}
	// 获取POD
	var pod v1.Pod
	// 下面这行代码相当于在拼接URL 
	req := client.Get().Namespace(nameSpace).Resource("pods").Name(podName)
	fmt.Println(req.URL()) // https://ip:prot/api/v1/namespaces/kube-system/pods/coredns-64897985d-cmmk9
	err = req.Do(context.TODO()).Into(&pod)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(pod.Name)
	}
}
func getConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", "./admin.conf")
	if err != nil {
		panic(err)
	}
	return config
}
```

### 2: DynamicCli

动态客户端 通常用于查询自定义CRD或者本身服务里没有的资源,他解析出来的对象是一个map[string]interface{}

```go
func dynamicCli() {
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	config := getConfig()
	cli, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	cluster, err := cli.Resource(resource).List(context.TODO(), v12.ListOptions{})
	if err != nil {
		fmt.Println(err)
	} else {
		for _, item := range cluster.Items {
			fmt.Println(item.GetName())
		}
	}
}
func getConfig() *rest.Config {
    config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhousong/admin.conf")
    if err != nil {
        panic(err)
    }
return config
}
```
### 3: ClientSet
ClientSet是k8s官方提供的官方资源的客户端，如，pods，nodes，role...等等
同时，我们自定义的CRD也可以生成client代码
ClientSet本质是在RESTClient上的一层针对特定资源的封装
```go
//这段代码就是 clientSet.CoreV1().Pods 的源码 ，我们可以看出pod的client，其实就是RESTClient
type pods struct {
	client rest.Interface
	ns     string
}

// newPods returns a Pods
func newPods(c *CoreV1Client, namespace string) *pods {
	return &pods{
		client: c.RESTClient(),
		ns:     namespace,
	}
}
```

```go
func clientCmd() {
	config := getConfig()
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	pod, err := clientSet.CoreV1().Pods(nameSpace).Get(context.TODO(), podName, v12.GetOptions{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(pod.Name)
	}
	pods, err := clientSet.CoreV1().Pods(nameSpace).List(context.TODO(), v12.ListOptions{})
	if err != nil {
		fmt.Println(err)
	} else {
		for _, item := range pods.Items {
			fmt.Println(item.Name)
		}
	}
	res, err := clientSet.ServerPreferredNamespacedResources()
	if err != nil {
		fmt.Println(err)
	} else {
		for _, re := range res {
			fmt.Println(re.GroupVersionKind().String(), re.GroupVersion)
		}
	}
}

func getConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/zhousong/admin.conf")
	if err != nil {
		panic(err)
	}
	return config
}
```

## Client 创建详解
在上一部分中，我们了解了如何创建Client，现在我们来看看Client的源码，看看它是如何创建出来的。
### RESTClient
通常情况下，我们一般调用 "k8s.io/client-go/rest" 包中的 RESTClientFor 方法来创建 RESTClient，接下我们就看这个方法中，是如何构建 RESTClient 的。

**RESTClientFor**
RESTClientFor 主要是用于校验一些配置错误和 调用**HTTPClientFor**获取http.client，然后将 config 和 http.client 传递给**RESTClientForConfigAndClient**创建 RESTClient.
```go
func RESTClientFor(config *Config) (*RESTClient, error) {
	//1.会首先校验我们传入进去的配置文件中的 **GroupVersion/NegotiatedSerializer** 是否存在，不存在就返回错误，说明这两个配置是我们创建 Client 所必须的。
	//  这也是在上一节RESTClient 创建的过程中，我们为什么要给Config单独赋值的原因。 (注：常规情况下，我们获取的Config中，这两个值是没有赋值的。)
	if config.GroupVersion == nil {
        return nil, fmt.Errorf("GroupVersion is required when initializing a RESTClient")
    }
    if config.NegotiatedSerializer == nil {
        return nil, fmt.Errorf("NegotiatedSerializer is required when initializing a RESTClient")
    }
	
	//2.快速验证我们传入的 config.HOST 是否能给正确解析，如果验证失败的就返回错误，避免我们在后续的使用过程中重复验证
    _, _, err := defaultServerUrlFor(config)
    if err != nil {
        return nil, err
    }
    //3.根据传入的 config 构建一个 http client.
    httpClient, err := HTTPClientFor(config)
    if err != nil {
        return nil, err
    }
    //4.调用 RESTClientForConfigAndClient 创建 RESTClient.
    return RESTClientForConfigAndClient(config, httpClient)
}
```

**HTTPClientFor**
HTTPClientFor 主要用于 根据config 创建 http.client
```go
func HTTPClientFor(config *Config) (*http.Client, error) {
    transport, err := TransportFor(config)
    if err != nil {
        return nil, err
    }
    var httpClient *http.Client
    if transport != http.DefaultTransport || config.Timeout > 0 {
        httpClient = &http.Client{
            Transport: transport,
            Timeout:   config.Timeout,
        }
    } else {
        httpClient = http.DefaultClient
    }
    return httpClient, nil
}
```

**RESTClientForConfigAndClient**
RESTClientForConfigAndClient 则是根据 config 和 httpClient 创建出 RESTClient
```go
func RESTClientForConfigAndClient(config *Config, httpClient *http.Client) (*RESTClient, error) {
    //同1
    if config.GroupVersion == nil {
        return nil, fmt.Errorf("GroupVersion is required when initializing a RESTClient")
    }
    if config.NegotiatedSerializer == nil {
        return nil, fmt.Errorf("NegotiatedSerializer is required when initializing a RESTClient")
    }
    // 同2 同时获取 baseURL 和 versionedAPIPath
    baseURL, versionedAPIPath, err := defaultServerUrlFor(config)
    if err != nil {
        return nil, err
    }
    
    //5. 从配置文件获取限速器，如果没有，就自己创建一个默认的限速器。
    rateLimiter := config.RateLimiter
    if rateLimiter == nil {
        qps := config.QPS
        if config.QPS == 0.0 {
            qps = DefaultQPS
        }
        burst := config.Burst
        if config.Burst == 0 {
            burst = DefaultBurst
        }
        if qps > 0 {
            rateLimiter = flowcontrol.NewTokenBucketRateLimiter(qps, burst)
        }
    }
    
    //6.利用已经获取参数构建一个ClientContentConfig
    var gv schema.GroupVersion
    if config.GroupVersion != nil {
        gv = *config.GroupVersion
    }
    clientContent := ClientContentConfig{
        AcceptContentTypes: config.AcceptContentTypes,
        ContentType:        config.ContentType,
        GroupVersion:       gv,
        Negotiator:         runtime.NewClientNegotiator(config.NegotiatedSerializer, gv),
    }
    
    //7.利用已经获取参数构建一个RESTClient
    restClient, err := NewRESTClient(baseURL, versionedAPIPath, clientContent, rateLimiter, httpClient)
    if err == nil && config.WarningHandler != nil {
        restClient.warningHandler = config.WarningHandler
    }
    return restClient, err
}
```

以上便是获取 RESTClient 的全过程了，可以总结为以下流程：
1. 会首先校验我们传入进去的 config 中的 **GroupVersion/NegotiatedSerializer** 是否存在。
2. 验证我们传入的 config.HOST 是否能够正确解析。
3. 根据配置获取 http.client。
4. 获取baseURL 和 versionedAPIPath。
5. 从配置文件获取限速器，如果没有，就自己创建一个默认的限速器。
6. 利用已经获取参数构建一个ClientContentConfig。
7. 利用已经获取参数构建一个RESTClient。

### ClientSet
通常情况下，我们一般调用 "k8s.io/client-go/kubernetes" 包中的 NewForConfig 方法来创建 ClientSet，接下我们就看这个方法中，是如何构建 ClientSet 的。

**NewForConfig**
NewForConfig 主要是用于 调用**HTTPClientFor**获取http.client，然后将 config 和 http.client 传递给**NewForConfigAndClient**创建 RESTClient.
````go
func NewForConfig(c *rest.Config) (*Clientset, error) {
    // 0.复制一份 config
    configShallowCopy := *c
    
    // 1.验证 UserAgent,不存在就自己构建一个 默认格式是  "os.Args[0]/GitVersion (GOOS/GOARCH) GitCommit"
    if configShallowCopy.UserAgent == "" {
        configShallowCopy.UserAgent = rest.DefaultKubernetesUserAgent()
    }
    
    //2.根据传入的 config 构建一个 http client.
    httpClient, err := rest.HTTPClientFor(&configShallowCopy)
    if err != nil {
        return nil, err
    }
    
    //3. 创建ClientSet
    return NewForConfigAndClient(&configShallowCopy, httpClient)
}
````
**NewForConfigAndClient**
NewForConfigAndClient 就比较粗暴了。主要是用于调用各个资源的Client创建方法，创建 Client， 毕竟**ClientSet**这个词还是很好理解的，就是客户端的集合。
```go
func NewForConfigAndClient(c *rest.Config, httpClient *http.Client) (*Clientset, error) {
	//同0
    configShallowCopy := *c
	//同1
    if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
        if configShallowCopy.Burst <= 0 {
            return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
        }
        configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
    }
    
    var cs Clientset
    var err error
    cs.admissionregistrationV1, err = admissionregistrationv1.NewForConfigAndClient(&configShallowCopy, httpClient)
    if err != nil {
        return nil, err
    }
    cs.admissionregistrationV1beta1, err = admissionregistrationv1beta1.NewForConfigAndClient(&configShallowCopy, httpClient)
    if err != nil {
        return nil, err
    }
	/*=====创建各种资源client的方法=====*/
	...
    return &cs, nil
}
```

看完这个，我们找一个例子，看看各种资源是怎么创建自己的客户端的。拿 核心组资源 CoreV1 看看。它是怎么创建的。

**NewForConfigAndClient**
NewForConfigAndClient 主要用于创建 核心组资源的客户端，
```go
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*CoreV1Client, error) {
    config := *c
	//4.设置默认配置
    if err := setConfigDefaults(&config); err != nil {
        return nil, err
    }
	//5.创建RESTClient, 同上一节RESTClient的 第4步
    client, err := rest.RESTClientForConfigAndClient(&config, h)
    if err != nil {
        return nil, err
    }
    return &CoreV1Client{client}, nil
}
```

**setConfigDefaults**
setConfigDefaults 主要用于为配置文件补充完成 GroupVersion，APIPath，NegotiatedSerializer，UserAgent，等信息。
```go
func setConfigDefaults(config *rest.Config) error {
    gv := v1.SchemeGroupVersion
    config.GroupVersion = &gv
    config.APIPath = "/api"
    config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
    
    if config.UserAgent == "" {
        config.UserAgent = rest.DefaultKubernetesUserAgent()
    }
    
    return nil
}
```

以上便是获取 ClientSet 的全过程了，可以总结为以下流程：
1. 拷贝一份 config，防止影响到其他使用的地方。
2. 补全 UserAgent。
3. 使用配置文件创建 http.Client。
4. 调用各种资源的 Client 创建方法创建 Client ，然后设置到 ClientSet。
5. 调用各种资源的 Client 创建过程中。拷贝一份 config。
6. 然后补全一些默认配置。如： GroupVersion，APIPath，NegotiatedSerializer，UserAgent等。
7. 最后调用 RESTClientForConfigAndClient 创建 RESTClint。

### DynamicCli 
DynamicClient 整体结构差不多，不同的是默认配置的设置，因为往往某一类资源Client，它们的配置是固定的，而DynamicClient则是动态的，需要适配不同的资源。 因此，它的核心反而是动态资源的适配。