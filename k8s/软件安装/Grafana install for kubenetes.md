# 准备

> k8s 集群: version >= 1.19.7

# Yaml 编写

## Namespace

创建 grafana 运行所需的命名空间

```yaml
apiVersion: v1  
kind: Namespace  
metadata:  
  name: grafana-system
```

## 存储

创建 grafana 运行所需的存储

```yaml
---  
apiVersion: storage.k8s.io/v1  
kind: StorageClass  
metadata:  
  name: grafana-sc  
  namespace: grafana-system  
provisioner: kubernetes.io/no-provisioner  
volumeBindingMode: WaitForFirstConsumer  
---  
apiVersion: v1  
kind: PersistentVolume  
metadata:  
  name: grafana-pv  
  namespace: grafana-system  
spec:  
  storageClassName: grafana-sc  
  capacity:  
    storage: 1Gi  
  accessModes:  
    - ReadWriteOnce  
  hostPath:  
    path: /tmp/grafana  
---  
apiVersion: v1  
kind: PersistentVolumeClaim  
metadata:  
  name: grafana-pvc  
  namespace: grafana-system  
spec:  
  storageClassName: grafana-sc  
  accessModes:  
    - ReadWriteOnce  
  resources:  
    requests:  
      storage: 1Gi
```

## Deployment

创建 grafana 运行所需的工作负载

```yaml
apiVersion: apps/v1  
kind: Deployment  
metadata:  
  labels:  
    app: grafana  
  name: grafana  
  namespace: grafana-system  
spec:  
  selector:  
    matchLabels:  
      app: grafana  
  template:  
    metadata:  
      labels:  
        app: grafana  
    spec:  
      securityContext:  
        fsGroup: 472  
        supplementalGroups:  
          - 0  
      containers:  
        - name: grafana  
          image: grafana/grafana:9.1.0  
          imagePullPolicy: IfNotPresent  
          ports:  
            - containerPort: 3000  
              name: http-grafana  
              protocol: TCP  
          readinessProbe:  
            failureThreshold: 3  
            httpGet:  
              path: /robots.txt  
              port: 3000  
              scheme: HTTP  
            initialDelaySeconds: 10  
            periodSeconds: 30  
            successThreshold: 1  
            timeoutSeconds: 2  
          livenessProbe:  
            failureThreshold: 3  
            initialDelaySeconds: 30  
            periodSeconds: 10  
            successThreshold: 1  
            tcpSocket:  
              port: 3000  
            timeoutSeconds: 1  
          resources:  
            requests:  
              cpu: 250m  
              memory: 750Mi  
          volumeMounts:  
            - mountPath: /var/lib/grafana  
              name: grafana-pv  
      volumes:  
        - name: grafana-pv  
          persistentVolumeClaim:  
            claimName: grafana-pvc
```

## 修改挂载目录权限

如果就这样部署，会发生错误：

```shell
➜  ~ kubectl logs -f grafana-xxx -n grafana-system
GF_PATHS_DATA='/var/lib/grafana' is not writable.
You may have issues with file permissions, more information here: http://docs.grafana.org/installation/docker/#migration-from-a-previous-version-of-the-docker-container-to-5-1-or-later
mkdir: cannot create directory '/var/lib/grafana/plugins': Permission denied
```

可以看出是由于没有目录权限引起的，我们可以通过修改目录权限来解决这个问题，我们可以了利用一个 Job 来帮我们解决这个问题：

```yaml
apiVersion: batch/v1  
kind: Job  
metadata:  
  name: grafana-chown  
  namespace: grafana-system  
spec:  
  template:  
    spec:  
      restartPolicy: Never  
      containers:  
        - name: grafana-chown  
          command: ["chown", "-R", "472:472", "/var/lib/grafana"]  
          image: busybox  
          imagePullPolicy: IfNotPresent  
          volumeMounts:  
            - name: grafana-pvc  
              mountPath: /var/lib/grafana  
      volumes:  
        - name: grafana-pvc  
          persistentVolumeClaim:  
            claimName: grafana-pvc
```

## Service

创建 grafana 访问所需的 service, 可以根据自己需求自行修改，我这里因为我的集群部署了 Ingress 控制器，所以使用 ClusterIP。

```yaml
apiVersion: v1  
kind: Service  
metadata:  
  name: grafana  
  namespace: grafana-system  
spec:  
  ports:  
    - port: 3000  
      protocol: TCP  
      targetPort: http-grafana  
  selector:  
    app: grafana  
  sessionAffinity: None  
  type: ClusterIP
```

## Ingress

创建 grafana 访问所需的 ingress。

```yaml
apiVersion: networking.k8s.io/v1  
kind: Ingress  
metadata:  
  name: grafana  
  namespace: grafana-system  
  annotations:  
    nginx.ingress.kubernetes.io/proxy-buffer-size: "32k"  
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "600"  
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"  
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"  
    nginx.ingress.kubernetes.io/proxy-body-size: "512m"  
    nginx.ingress.kubernetes.io/ingress.class: 'nginx'  
spec:  
  ingressClassName: nginx  
  rules:  
    - host: grafana.test.com  
      http:  
        paths:  
          - path: /  
            pathType: Prefix  
            backend:  
              service:  
                name: grafana  
                port:  
                  number: 3000
```

## 完整 yaml

deploy.yaml

```yaml
---  
apiVersion: v1  
kind: Namespace  
metadata:  
  name: grafana-system  
---  
apiVersion: storage.k8s.io/v1  
kind: StorageClass  
metadata:  
  name: grafana-sc  
  namespace: grafana-system  
provisioner: kubernetes.io/no-provisioner  
volumeBindingMode: WaitForFirstConsumer  
---  
apiVersion: v1  
kind: PersistentVolume  
metadata:  
  name: grafana-pv  
  namespace: grafana-system  
spec:  
  storageClassName: grafana-sc  
  capacity:  
    storage: 1Gi  
  accessModes:  
    - ReadWriteOnce  
  hostPath:  
    path: /tmp/grafana  
---  
apiVersion: v1  
kind: PersistentVolumeClaim  
metadata:  
  name: grafana-pvc  
  namespace: grafana-system  
spec:  
  storageClassName: grafana-sc  
  accessModes:  
    - ReadWriteOnce  
  resources:  
    requests:  
      storage: 1Gi  
---  
apiVersion: batch/v1  
kind: Job  
metadata:  
  name: grafana-chown  
  namespace: grafana-system  
spec:  
  template:  
    spec:  
      restartPolicy: Never  
      containers:  
        - name: grafana-chown  
          command: ["chown", "-R", "472:472", "/var/lib/grafana"]  
          image: busybox  
          imagePullPolicy: IfNotPresent  
          volumeMounts:  
            - name: grafana-pvc  
              mountPath: /var/lib/grafana  
      volumes:  
        - name: grafana-pvc  
          persistentVolumeClaim:  
            claimName: grafana-pvc  
---  
apiVersion: apps/v1  
kind: Deployment  
metadata:  
  labels:  
    app: grafana  
  name: grafana  
  namespace: grafana-system  
spec:  
  selector:  
    matchLabels:  
      app: grafana  
  template:  
    metadata:  
      labels:  
        app: grafana  
    spec:  
      securityContext:  
        fsGroup: 472  
        supplementalGroups:  
          - 0  
      containers:  
        - name: grafana  
          image: grafana/grafana:9.1.0  
          imagePullPolicy: IfNotPresent  
          ports:  
            - containerPort: 3000  
              name: http-grafana  
              protocol: TCP  
          readinessProbe:  
            failureThreshold: 3  
            httpGet:  
              path: /robots.txt  
              port: 3000  
              scheme: HTTP  
            initialDelaySeconds: 10  
            periodSeconds: 30  
            successThreshold: 1  
            timeoutSeconds: 2  
          livenessProbe:  
            failureThreshold: 3  
            initialDelaySeconds: 30  
            periodSeconds: 10  
            successThreshold: 1  
            tcpSocket:  
              port: 3000  
            timeoutSeconds: 1  
          resources:  
            requests:  
              cpu: 250m  
              memory: 750Mi  
          volumeMounts:  
            - mountPath: /var/lib/grafana  
              name: grafana-pv  
      volumes:  
        - name: grafana-pv  
          persistentVolumeClaim:  
            claimName: grafana-pvc  
---  
apiVersion: v1  
kind: Service  
metadata:  
  name: grafana  
  namespace: grafana-system  
spec:  
  ports:  
    - port: 3000  
      protocol: TCP  
      targetPort: http-grafana  
  selector:  
    app: grafana  
  sessionAffinity: None  
  type: ClusterIP  
---  
apiVersion: networking.k8s.io/v1  
kind: Ingress  
metadata:  
  name: grafana  
  namespace: grafana-system  
  annotations:  
    nginx.ingress.kubernetes.io/proxy-buffer-size: "32k"  
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "600"  
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"  
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"  
    nginx.ingress.kubernetes.io/proxy-body-size: "512m"  
    nginx.ingress.kubernetes.io/ingress.class: 'nginx'  
spec:  
  ingressClassName: nginx  
  rules:  
    - host: grafana.test.com  
      http:  
        paths:  
          - path: /  
            pathType: Prefix  
            backend:  
              service:  
                name: grafana  
                port:  
                  number: 3000
```

# 部署

```bash
# 部署 grafana
➜  ~ kubectl apply -f grafana-deploy.yaml
```

## 观察服务

### 查看存储绑定状态

```bash
➜  ~ kubectl get -n grafana-system sc
NAME         PROVISIONER                    RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
grafana-sc   kubernetes.io/no-provisioner   Delete          WaitForFirstConsumer   false                  3h21m
➜  ~ kubectl get -n grafana-system pv
NAME         CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM                        STORAGECLASS   REASON   AGE
grafana-pv   1Gi        RWO            Retain           Bound    grafana-system/grafana-pvc   grafana-sc              3h22m
➜  ~ kubectl get -n grafana-system pvc
NAME          STATUS   VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS   AGE
grafana-pvc   Bound    grafana-pv   1Gi        RWO            grafana-sc     3h22m
```

## 查看POD运行状态

```bash
➜  ~ kubectl get -n grafana-system pod
NAME                       READY   STATUS      RESTARTS   AGE
grafana-58445b6986-wx8tn   1/1     Running     1          3h23m
grafana-chown-x99bh        0/1     Completed   0          3h23m
```

## 查看 Serveice + Ingress

```bash
➜  ~ kubectl get -n grafana-system svc
NAME      TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
grafana   ClusterIP   10.98.141.112   <none>        3000/TCP   3h24m
➜  ~ kubectl get -n grafana-system ing
NAME      CLASS   HOSTS              ADDRESS       PORTS   AGE
grafana   nginx   grafana.test.com   10.99.6.241   80      3h20m
```

# 访问 grafana

## 配置host

### windows
文件路径：C:\Windows\System32\drivers\etc

```
# 修改 hosts 文件
master ip grafana.test.com
```

linux & mac

```bash
sudo echo "master ip grafana.test.com" >> /etc/hosts
```

浏览器访问： http://grafana.test.com