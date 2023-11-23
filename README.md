# redis-cluster-acl-tool

redis集群acl管理工具

由于官方不支持集群模式下acl规则传播，比较常见的做法是所有节点执行一下acl规则，或修改配置文件acl，重启所有节点。前者操作比较麻烦，后者会重置客户端与集群的连接

此工具简化了为所有节点配置acl规则



## 快速执行
```shell
go run main.go -a 10.233.127.68:6379 -p 123456 -c "acl list"
# make all
# ./acltool -a 10.233.127.68:6379 -p 123456 -c "acl list" 
```

1. -a为地址请改成自己的集群中一个node地址
2. -p为默认用户的密码
3. -c为acl规则命令


## 参数解释

参数：
```go
// 命令行参数
type Options struct {
	Addr     string `short:"a" long:"addr" description:"redis地址，例如: 127.0.0.1:6379" default:""`
	Password string `short:"p" long:"password" description:"默认用户密码" default:""`
	Acl      string `short:"c" long:"acl" description:"acl命令，例如: acl list" default:""`
}
// 或者环境变量
if opt.Addr == "" {
    opt.Addr = os.Getenv("ADDR")
}
if opt.Password == "" {
    opt.Password = os.Getenv("PASSWORD")
}
if opt.Acl == "" {
    opt.Acl = os.Getenv("ACL")
}
```

命令行参数优先，addr为redis集群其中的一个节点，密码是默认用户的，acl是执行的命令

## docker镜像
构建并上传到自己的镜像仓库
```shell
docker buildx build --platform linux/arm64,linux/amd64 -t addr/username/redis-cluster-acl-tool:tag . --push
# docker build -t addr/username/redis-cluster-acl-tool:tag . 
```

快速获取
```shell
docker pull lincocoxue/redis-cluster-acl-tool:dev-02
```

## kubernetes最佳实践
创建Kubernetes ConfigMap资源，指定这三个参数

```yaml
kind: ConfigMap
apiVersion: v1
metadata:
  name: redis-cluster-acl-tool-config
  namespace: syndra
  annotations:
    kubesphere.io/alias-name: ''
    kubesphere.io/creator: lincoco
    kubesphere.io/description: redis集群acl工具配置
data:
  acl: >-
    acl setuser account on
    >shohZoowiequoe2eeyahphasxxs3afio8eeshohtheeyie8nee0ushoo3Ii3hoh4he ~account:*
    +@all -ACL -FLUSHALL -FLUSHDB -SHUTDOWN -SAVE -BGSAVE -CONFIG -MODULE -DEBUG
  addr: '10.233.127.68:6379'
  password: '123456'
```
创建kubernetes Job资源，要执行acl就先修改ConfigMap的三个参数，然后重新运行Job

```yaml
kind: Job
apiVersion: batch/v1
metadata:
  name: redis-cluster-acl-tool
  namespace: default
  labels:
    app: redis-cluster-acl-tool
  annotations:
    kubesphere.io/creator: lincoco
    kubesphere.io/description: redis集群配置acl工具
    revisions: >-
      {"1":{"status":"completed","succeed":1,"desire":1,"uid":"57304390-5de4-4bb7-a6e5-b28a8f72ff18","start-time":"2023-11-23T16:50:42+08:00","completion-time":"2023-11-23T16:51:00+08:00"},"2":{"status":"completed","succeed":1,"desire":1,"uid":"b7c0504a-2350-4025-8b7b-c9069f58b133","start-time":"2023-11-23T16:57:14+08:00","completion-time":"2023-11-23T16:57:42+08:00"}}
spec:
  parallelism: 1
  completions: 1
  backoffLimit: 6
  selector:
    matchLabels:
      controller-uid: b7c0504a-2350-4025-8b7b-c9069f58b133
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: redis-cluster-acl-tool
        controller-uid: b7c0504a-2350-4025-8b7b-c9069f58b133
        job-name: redis-cluster-acl-tool
      annotations:
        kubesphere.io/creator: lincoco
        kubesphere.io/imagepullsecrets: '{}'
    spec:
      containers:
        - name: acl-tool
          image: 'addr/username/redis-cluster-acl-tool:tag'
          command:
            - acltool
          env:
            - name: ADDR
              valueFrom:
                configMapKeyRef:
                  name: redis-cluster-acl-tool-config
                  key: addr
            - name: PASSWORD
              valueFrom:
                configMapKeyRef:
                  name: redis-cluster-acl-tool-config
                  key: password
            - name: ACL
              valueFrom:
                configMapKeyRef:
                  name: redis-cluster-acl-tool-config
                  key: acl
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          imagePullPolicy: Always
      restartPolicy: OnFailure
      terminationGracePeriodSeconds: 30
      dnsPolicy: ClusterFirst
      serviceAccountName: default
      serviceAccount: default
      securityContext: {}
      schedulerName: default-scheduler
  completionMode: NonIndexed
  suspend: false
```