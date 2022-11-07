package main

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var nameSpace = "default"
var podName = "dnsutils"

func main() {
	restClient()
	clientCmd()
	dynamicCli()
}

func dynamicCli() {
	resource := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	config := getConfig()
	cli, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	pod, err := cli.Resource(resource).Namespace(nameSpace).Get(context.TODO(), podName, v12.GetOptions{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(pod.GetName())
	}
}

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
}

func restClient() {
	config := getConfig()
	config.GroupVersion = &v1.SchemeGroupVersion
	config.NegotiatedSerializer = scheme.Codecs
	config.APIPath = "/api"
	client, err := rest.RESTClientFor(config)
	if err != nil {
		panic(err)
	}
	var pod v1.Pod
	req := client.Get().Namespace("default").Resource("pods").Name("dnsutils")
	err = req.Do(context.TODO()).Into(&pod)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(pod.Name)
	}
}

func getConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", "xxx")
	if err != nil {
		panic(err)
	}
	return config
}
