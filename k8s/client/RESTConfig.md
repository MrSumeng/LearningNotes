# rest.Config 详解
> rest.Config 是我们获取k8s集群client的最基础的配置文件，它是创建client的基础。
 
通常情况下，我们创建一个 k8s Client,无论是最基础的 RESTClient ，还是基于 RESTClient 去封装 ClientSet / DynamicClient / DiscoverClient 都需要 rest.Config 这个配置文件去建立。

那么 rest.Config 为我们提供了哪些配置，它们分别有什么用，它是怎么创建的？我们接下来就看一看。

## admin.conf
在看 rest.Config 前，我们先来看看我们在创建k8s集群后，集群为我们提供的admin.conf 里面提供了哪些信息。这些信息对我们接下来了解rest.Config 非常有用。
> admin.conf 的路径通常情况下为: /etc/kubernetes/admin.conf

我们可以通过阅读 admin.conf 发现，admin.conf 为我提供了以下信息：
+ 集群信息列表(clusters): 里面提供了集群的名称，集群的服务地址(包含IP和端口),集群的证书颁发机构。(列表，支持多集群)
+ 用户信息列表(users): 里面提供了用户的名称，用户的证书和密钥。(列表，支持多用户)
+ 上下文列表(contexts): 里面提供了集群与用户的绑定信息，它维系了用户和集群的关联关系；名称的格式为: user@cluster。(列表，支持上下文切换)
+ 当前使用的上下文(current-context): 即当前使用的集群和用户。

```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: xxxx
    server: https://xxxx:6443
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: kubernetes-admin
  name: kubernetes-admin@kubernetes
current-context: kubernetes-admin@kubernetes
preferences: {}
users:
- name: kubernetes-admin
  user:
    client-certificate-data: xxx
    client-key-data: xxx
```
看完admin.conf 接下来我们就正是开始了解 rest.Config。

## rest.Config 结构
> Config 包含可以在初始化时传递给 Kubernetes 客户端的通用属性。

其中：
+ Host是指向 kubernetes ApiServer 的地址.
+ ContentConfig 则包含数据传输用的内容类型声明和编码解码器
+ 其中还有一些用于连接 ApiServer 的 认证：Username，Password，TLSClientConfig 等等。
+ 还有 HTTP 请求的一些配置如：Transport, WrapTransport, Timeout 等等。
```go
type Config struct {
    // Host must be a host string, a host:port pair, or a URL to the base of the apiserver.
    // If a URL is given then the (optional) Path of that URL represents a prefix that must
    // be appended to all request URIs used to access the apiserver. This allows a frontend
    // proxy to easily relocate all of the apiserver endpoints.
    Host string
    // APIPath is a sub-path that points to an API root.
    APIPath string
    
    // ContentConfig contains settings that affect how objects are transformed when
    // sent to the server.
    ContentConfig
    
    // Server requires Basic authentication
    Username string
    Password string `datapolicy:"password"`
    
    // Server requires Bearer authentication. This client will not attempt to use
    // refresh tokens for an OAuth2 flow.
    // TODO: demonstrate an OAuth2 compatible client.
    BearerToken string `datapolicy:"token"`
    
    // Path to a file containing a BearerToken.
    // If set, the contents are periodically read.
    // The last successfully read value takes precedence over BearerToken.
    BearerTokenFile string
    
    // Impersonate is the configuration that RESTClient will use for impersonation.
    Impersonate ImpersonationConfig
    
    // Server requires plugin-specified authentication.
    AuthProvider *clientcmdapi.AuthProviderConfig
    
    // Callback to persist config for AuthProvider.
    AuthConfigPersister AuthProviderConfigPersister
    
    // Exec-based authentication provider.
    ExecProvider *clientcmdapi.ExecConfig
    
    // TLSClientConfig contains settings to enable transport layer security
    TLSClientConfig
    
    // UserAgent is an optional field that specifies the caller of this request.
    UserAgent string
    
    // DisableCompression bypasses automatic GZip compression requests to the
    // server.
    DisableCompression bool
    
    // Transport may be used for custom HTTP behavior. This attribute may not
    // be specified with the TLS client certificate options. Use WrapTransport
    // to provide additional per-server middleware behavior.
    Transport http.RoundTripper
    // WrapTransport will be invoked for custom HTTP behavior after the underlying
    // transport is initialized (either the transport created from TLSClientConfig,
    // Transport, or http.DefaultTransport). The config may layer other RoundTrippers
    // on top of the returned RoundTripper.
    //
    // A future release will change this field to an array. Use config.Wrap()
    // instead of setting this value directly.
    WrapTransport transport.WrapperFunc
    
    // QPS indicates the maximum QPS to the master from this client.
    // If it's zero, the created RESTClient will use DefaultQPS: 5
    QPS float32
    
    // Maximum burst for throttle.
    // If it's zero, the created RESTClient will use DefaultBurst: 10.
    Burst int
    
    // Rate limiter for limiting connections to the master from this client. If present overwrites QPS/Burst
    RateLimiter flowcontrol.RateLimiter
    
    // WarningHandler handles warnings in server responses.
    // If not set, the default warning handler is used.
    // See documentation for SetDefaultWarningHandler() for details.
    WarningHandler WarningHandler
    
    // The maximum length of time to wait before giving up on a server request. A value of zero means no timeout.
    Timeout time.Duration
    
    // Dial specifies the dial function for creating unencrypted TCP connections.
    Dial func(ctx context.Context, network, address string) (net.Conn, error)
    
    // Proxy is the proxy func to be used for all requests made by this
    // transport. If Proxy is nil, http.ProxyFromEnvironment is used. If Proxy
    // returns a nil *URL, no proxy is used.
    //
    // socks5 proxying does not currently support spdy streaming endpoints.
    Proxy func(*http.Request) (*url.URL, error)
    
    // Version forces a specific version to be used (if registered)
    // Do we need this?
    // Version string
}
```

## 创建
了解完 rest.Config 的配置，我们再来看看 rest.Config 是怎么创建出来的。

通常情况下，我们会通过 k8s.io/client-go/tools/clientcmd.BuildConfigFromFlags() 或是 sigs.k8s.io/controller-runtime.GetConfigOrDie() 来创建rest.Config。

我们先来了解 BuildConfigFromFlags() 是怎么创建rest.Config 的。

### BuildConfigFromFlags
在 BuildConfigFromFlags 中 创建 Config 主要分为两种方式：
+ 一种是在配置文件的路径和 master url 都没有提供的情况下，去通过集群内部配置获取 Config。
+ 另一种是在提供了配置文件的路径或 master url 情况下 通过，非交互式的延迟加载客户端配置来获取 Config 。
```go
func BuildConfigFromFlags(masterUrl, kubeconfigPath string) (*restclient.Config, error) {
    if kubeconfigPath == "" && masterUrl == "" {
        klog.Warning("Neither --kubeconfig nor --master was specified.  Using the inClusterConfig.  This might not work.")
        kubeconfig, err := restclient.InClusterConfig()
        if err == nil {
            return kubeconfig, nil
        }
        klog.Warning("error creating inClusterConfig, falling back to default config: ", err)
    }
    return NewNonInteractiveDeferredLoadingClientConfig(
        &ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
        &ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: masterUrl}}).ClientConfig()
}
```

#### InClusterConfig
先来在配置文件的路径和 master url 都没有提供的情况下，是怎么通过集群内部配置获取 Config的。
1. 通过环境变量获取和集群通信的HOST和端口。(KUBERNETES_SERVICE_HOST,KUBERNETES_SERVICE_PORT)
2. 通过固定配置文件路径获取用来进行认证的token。(/var/run/secrets/kubernetes.io/serviceaccount/token)
3. 创建一个空的TSl认证配置
4. 通过固定配置文件路径获取用来进行认证的CA证书。(/var/run/secrets/kubernetes.io/serviceaccount/ca.crt)
5. 组装配置文件

```go
func InClusterConfig() (*Config, error) {
    const (
        tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
        rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
    )
    host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
    if len(host) == 0 || len(port) == 0 {
        return nil, ErrNotInCluster
    }
    
    token, err := ioutil.ReadFile(tokenFile)
    if err != nil {
        return nil, err
    }
    
    tlsClientConfig := TLSClientConfig{}
    
    if _, err := certutil.NewPool(rootCAFile); err != nil {
        klog.Errorf("Expected to load root CA config from %s, but got err: %v", rootCAFile, err)
    } else {
        tlsClientConfig.CAFile = rootCAFile
    }
    
    return &Config{
        Host:            "https://" + net.JoinHostPort(host, port),
        TLSClientConfig: tlsClientConfig,
        BearerToken:     string(token),
        BearerTokenFile: tokenFile,
    }, nil
}
```
通过阅读代码，我们不难发现，这样的方式仅仅只能对拥有k8s配置的pod内部使用，否则，我们是获取不到对应的配置的。因为这些配置，通常情况下k8s集群会在生成pod时帮我们自动创建。

我们可以尝试验证一下，在这里，我准备了一个集群和一个pod：
```bash
kubectl get pod
# NAME                                  READY   STATUS             RESTARTS          AGE
# dnsutils                              1/1     Running            0                 1h

# 进入到pod内
# kubectl exec -it podName -n namespace -- /bin/sh
kubectl exec -it dnsutils -- /bin/sh
#依次查看这些配置
# echo $KUBERNETES_SERVICE_HOST
10.96.0.1
# echo $KUBERNETES_SERVICE_PORT
443
# cat /var/run/secrets/kubernetes.io/serviceaccount/token
eyxxxx...
# cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
-----BEGIN CERTIFICATE-----
xxxxxx
+eU=...
-----END CERTIFICATE-----
```
通过以上操作，我们可以发现，k8s集群确实在创建pod的时候会为我们准备这些配置。

#### NewNonInteractiveDeferredLoadingClientConfig(xx,xx).ClientConfig()
> 注:NewNonInteractiveDeferredLoadingClientConfig(xx,xx).ClientConfig() 名字太长，下面就用 ClientConfig 代替。

在 ClientConfig 中主要是调用以下调用链路去生成 Config:
1. 调用 createClientConfig 生成Config。
2. 在 createClientConfig 调用 Load 去加载 Config。
3. 在 Load 通过我们提供的 配置文件路径 获取客户端 Config，然后合并客户端 Config。
4. 再通过合并客户端 Config 创建 DirectClientConfig 并返回。
5. 然后 createClientConfig 中调用 DirectClientConfig 的 ClientConfig 生成 Config
```go
// step 1
func (config *DeferredLoadingClientConfig) ClientConfig() (*restclient.Config, error) {
	mergedClientConfig, err := config.createClientConfig()
	...
	mergedConfig, err := mergedClientConfig.ClientConfig()
	switch {
	...
	case mergedConfig != nil:
		if !config.loader.IsDefaultConfig(mergedConfig) {
			return mergedConfig, nil
		}
	}
	if config.icc.Possible() {
		return config.icc.ClientConfig()
	}
	return mergedConfig, err
}
// step 2
func (config *DeferredLoadingClientConfig) createClientConfig() (ClientConfig, error) {

    if config.clientConfig != nil {
        return config.clientConfig, nil
    }
    ...
    var currentContext string
    if config.overrides != nil {
        currentContext = config.overrides.CurrentContext
    }
    if config.fallbackReader != nil {
        config.clientConfig = NewInteractiveClientConfig(*mergedConfig, currentContext, config.overrides, config.fallbackReader, config.loader)
    } else {
        config.clientConfig = NewNonInteractiveClientConfig(*mergedConfig, currentContext, config.overrides, config.loader)
    }
    return config.clientConfig, nil
}

// step 3
func (rules *ClientConfigLoadingRules) Load() (*clientcmdapi.Config, error) {
    ...
    kubeConfigFiles := []string{}
    if len(rules.ExplicitPath) > 0 {
        if _, err := os.Stat(rules.ExplicitPath); os.IsNotExist(err) {
            return nil, err
        }
        kubeConfigFiles = append(kubeConfigFiles, rules.ExplicitPath)
    } else {
        kubeConfigFiles = append(kubeConfigFiles, rules.Precedence...)
    }
    kubeconfigs := []*clientcmdapi.Config{}
    for _, filename := range kubeConfigFiles {
		...
		config, err := LoadFromFile(filename)
		...
        kubeconfigs = append(kubeconfigs, config)
    }
    ...
    mapConfig := clientcmdapi.NewConfig()
    for _, kubeconfig := range kubeconfigs {
        mergo.Merge(mapConfig, kubeconfig, mergo.WithOverride)
    }
    ...
    nonMapConfig := clientcmdapi.NewConfig()
    for i := len(kubeconfigs) - 1; i >= 0; i-- {
        kubeconfig := kubeconfigs[i]
        mergo.Merge(nonMapConfig, kubeconfig, mergo.WithOverride)
    }
    config := clientcmdapi.NewConfig()
    mergo.Merge(config, mapConfig, mergo.WithOverride)
    mergo.Merge(config, nonMapConfig, mergo.WithOverride)
    
    if rules.ResolvePaths() {
        if err := ResolveLocalPaths(config); err != nil {
            errlist = append(errlist, err)
        }
    }
    return config, utilerrors.NewAggregate(errlist)
}

//step 4
func (config *DirectClientConfig) ClientConfig() (*restclient.Config, error) {
    configAuthInfo, err := config.getAuthInfo()
    ...
    configClusterInfo, err := config.getCluster()
    ...
    clientConfig := &restclient.Config{}
    clientConfig.Host = configClusterInfo.Server
    if configClusterInfo.ProxyURL != "" {
        u, err := parseProxyURL(configClusterInfo.ProxyURL)
        ...
        clientConfig.Proxy = http.ProxyURL(u)
    }
    
    if config.overrides != nil && len(config.overrides.Timeout) > 0 {
        timeout, err := ParseTimeout(config.overrides.Timeout)
        ...
        clientConfig.Timeout = timeout
    }
    
    if u, err := url.ParseRequestURI(clientConfig.Host); err == nil && u.Opaque == "" && len(u.Path) > 1 {
        u.RawQuery = ""
        u.Fragment = ""
        clientConfig.Host = u.String()
    }
    if len(configAuthInfo.Impersonate) > 0 {
        clientConfig.Impersonate = restclient.ImpersonationConfig{
            UserName: configAuthInfo.Impersonate,
            UID:      configAuthInfo.ImpersonateUID,
            Groups:   configAuthInfo.ImpersonateGroups,
            Extra:    configAuthInfo.ImpersonateUserExtra,
        }
    }
    
    // only try to read the auth information if we are secure
    if restclient.IsConfigTransportTLS(*clientConfig) {
        var err error
        var persister restclient.AuthProviderConfigPersister
        if config.configAccess != nil {
        authInfoName, _ := config.getAuthInfoName()
        persister = PersisterForUser(config.configAccess, authInfoName)
        }
        userAuthPartialConfig, err := config.getUserIdentificationPartialConfig(configAuthInfo, config.fallbackReader, persister, configClusterInfo)
        if err != nil {
        return nil, err
        }
        mergo.Merge(clientConfig, userAuthPartialConfig, mergo.WithOverride)
        
        serverAuthPartialConfig, err := getServerIdentificationPartialConfig(configAuthInfo, configClusterInfo)
        if err != nil {
        return nil, err
        }
        mergo.Merge(clientConfig, serverAuthPartialConfig, mergo.WithOverride)
    }
    
    return clientConfig, nil
}
```

### sigs.k8s.io/controller-runtime.GetConfigOrDie()
GetConfigOrDie 获取配置的方式与BuildConfigFromFlags相似；
它会先去读环境变量中的配置文件路径，如果不存在就调用InClusterConfig获取配置；
获取失败，就自己构建配置文件路径然后调用NewNonInteractiveDeferredLoadingClientConfig(xx,xx).ClientConfig()去获取配置文件,获取失败就退出。


```go
GetConfigOrDie = config.GetConfigOrDie

func GetConfigOrDie() *rest.Config {
    config, err := GetConfig()
    if err != nil {
        os.Exit(1)
    }
    return config
}
func GetConfig() (*rest.Config, error) {
    return GetConfigWithContext("")
}
func GetConfigWithContext(context string) (*rest.Config, error) {
    cfg, err := loadConfig(context)
	if err != nil {
        return nil, err
    }
    ...
    return cfg, nil
}

func loadConfig(context string) (*rest.Config, error) {
	// 从启动参数中获取配置文件路径
    if len(kubeconfig) > 0 {
        return loadConfigWithContext("", &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}, context)
    }
    
    kubeconfigPath := os.Getenv(clientcmd.RecommendedConfigPathEnvVar)
    if len(kubeconfigPath) == 0 {
		// var loadInClusterConfig = rest.InClusterConfig
        if c, err := loadInClusterConfig(); err == nil {
            return c, nil
        }
    }
	//自己构建获取配置文件的规则
    loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
    if _, ok := os.LookupEnv("HOME"); !ok {
        u, err := user.Current()
        if err != nil {
            return nil, fmt.Errorf("could not get current user: %w", err)
        }
        loadingRules.Precedence = append(loadingRules.Precedence, filepath.Join(u.HomeDir, clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	}
    
    return loadConfigWithContext("", loadingRules, context)
}

func loadConfigWithContext(apiServerURL string, loader clientcmd.ClientConfigLoader, context string) (*rest.Config, error) {
    return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
        loader,
        &clientcmd.ConfigOverrides{
            ClusterInfo: clientcmdapi.Cluster{
				Server: apiServerURL,
            },
            CurrentContext: context,
        }).ClientConfig()
}
```
