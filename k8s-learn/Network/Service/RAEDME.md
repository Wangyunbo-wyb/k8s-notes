## Service

Service是标准k8s资源，本质是 节点 上的iptable或ipvs规则，利用规则实现流量的转发到pod上，且也基于selector标签选择器选择、标识后端pod；

本质上，Service是一种四层负载均衡的抽象。
> 四层负载均衡示例：
假设有一个L4负载均衡器在TCP层进行负载均衡。客户端请求 http://example.com，负载均衡器根据源IP地址将请求分发到后端服务器1或服务器2。
客户端 -> 负载均衡器 (检查IP和端口) -> 服务器1或服务器2

如果你使用Deployment来运行你的应用，Deployment可以动态地创建和销毁Pod。 在任何时刻，你都不知道有多少个这样的 Pod 正在工作以及它们健康与否； 你可能甚至不知道如何辨别健康的 Pod。 Kubernetes Pod的创建和销毁是为了匹配集群的预期状态。 Pod是临时资源（你不应该期待单个Pod既可靠又耐用）。

每个 Pod 会获得属于自己的 IP 地址（Kubernetes 期待网络插件来保证这一点）。 对于集群中给定的某个 Deployment，这一刻运行的 Pod 集合可能不同于下一刻运行该应用的 Pod 集合。

这就带来了一个问题：如果某组 Pod（称为“后端”）为集群内的其他 Pod（称为“前端”） 集合提供功能，前端要如何发现并跟踪要连接的 IP 地址，以便其使用负载的后端组件呢？

这时候就要使用Service对象， 它是将运行在一个或一组Pod上的网络应用程序公开为网络服务的方法。

资源清单如下：

```YAML
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app.kubernetes.io/name: MyApp
  ports:
    - name: http
      protocol: TCP
      port: 80
      targetPort: 9376
```
应用上述清单时，系统将创建一个名为 "my-service" 的、 服务类型默认为 ClusterIP 的 Service。 该 Service 指向带有标签 app.kubernetes.io/name: MyApp 的所有 Pod 的 TCP 端口 9376。

### 类型
Service有几种不同的类型，每种类型提供了不同的访问策略和使用场景。以下是Kubernetes中Service的主要类型：
1. ClusterIP：
    + 这是默认的Service类型。
    + 它为Service提供一个仅在Kubernetes集群内部可访问的虚拟IP地址（ClusterIP），适合用于集群内部的服务发现和通信。
2. NodePort：
    + NodePort类型的Service在每个节点的特定端口（NodePort）上暴露Service，通过<nodeIP>:<NodePort>可以从集群外部访问Service。
    + 适合用于测试或当需要从外部网络访问内部服务时。
3. LoadBalancer：
    + 这种类型的Service使用云提供商的负载均衡器，为Service提供一个外部可访问的IP地址。
    + 适合于公开暴露服务到互联网，例如Web应用。
4. ExternalName：
    + ExternalName类型的Service不路由流量，也不维护任何后端Pod的列表。
    + 它将Service名称映射到一个CNAME记录或外部DNS名称，允许Kubernetes集群内的客户端通过Service名称访问外部服务。
5. Headless Service：
    + 这是一种特殊类型的ClusterIP Service，没有指定ClusterIP。
    + 它允许直接通过Service名称和DNS服务发现Pod的IP地址，而不需要通过负载均衡。
    + 适合于不需要负载均衡的简单服务发现场景。

### 使用
假设我们有一个简单的web应用，它由两个微服务组成：一个前端服务和一个后端服务。我们将使用Kubernetes Service来使这些服务对内部和外部都可访问。他们的deployments部署清单如下：
前端：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: frontend
        image: nginx
        ports:
        - containerPort: 80
```
后端：
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: my-custom-backend-image:v1
        ports:
        - containerPort: 8080
```
他们的service资源如下
前端：
```yaml
apiVersion: v1
kind: Service
metadata:
   name: frontend-service
spec:
   type: LoadBalancer
   selector:
      app: frontend
   ports:
      - port: 80
        targetPort: 80
```
当我们访问的时候，由于frontend-service的类型是LoadBalancer，云提供商会为这个Service分配一个外部可访问的IP地址或域名。你可以通过这个IP或域名在浏览器中访问前端服务。
后端：
```yaml
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  type: ClusterIP
  selector:
    app: backend
  ports:
    - port: 8080
      targetPort: 8080
```
backend-service的类型是ClusterIP，这意味着它只能在集群内部访问。前端服务可以通过backend-service的名称和端口（8080）来访问后端服务，例如，如果前端服务需要调用后端服务的API，它可以通过http://backend-service:8080/api来访问。

### 实现
Service的底层逻辑确实涉及到选择Pod并根据访问请求进行负载均衡，以下是Service的工作原理的详细解释：

1. **选择Pod**:Service定义了一个选择器（selector），这个选择器用于匹配一组具有特定标签（labels）的Pod。Service负责监控这些Pod的状态，如果Pod被添加或移除，Service会自动更新其内部的Pod列表。
2. **IP地址**:每个Service在创建时会被分配一个唯一的虚拟IP地址（ClusterIP），这个IP地址是内部的，只在Kubernetes集群内部可用。Pod通过这个IP地址和Service定义的端口来通信。
3. **DNS解析**:Kubernetes集群内部有一个DNS服务，它允许Pod通过Service的名称来解析对应的ClusterIP。当Pod尝试连接到Service时，集群的DNS服务会将Service名称解析为Service的ClusterIP。
4. **负载均衡**:Kubernetes使用一种称为kube-proxy的组件来实现Service的负载均衡。kube-proxy在每个节点上运行，负责将发往Service的流量转发到后端的Pod。
5. **kube-proxy工作原理**：
   + kube-proxy监控Service和Endpoints资源的变化。
   + 当Service或Endpoints（包含实际Pod IP和端口信息）发生变化时，kube-proxy会更新其内部的路由规则。
   + 当有流量到达Service的ClusterIP和端口时，kube-proxy会根据其内部的路由规则，选择一个Pod进行流量转发。
   + kube-proxy使用随机选择、轮询、最小连接数等策略来进行负载均衡。 
6. **访问Service**:无论是从集群内部还是外部访问Service，客户端都不需要知道后端Pod的具体IP地址。它们只需要知道Service的名称和端口，然后通过Kubernetes的DNS服务或外部负载均衡器来访问Service。

