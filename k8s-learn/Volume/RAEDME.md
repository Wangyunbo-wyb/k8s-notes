## 存储

### 卷：Volume
容器中的文件在磁盘上是临时存放的，这给在容器中运行较重要的应用带来一些问题。 当容器崩溃或停止时会出现一个问题。此时容器状态未保存， 因此在容器生命周期内创建或修改的所有文件都将丢失。 在崩溃期间，kubelet 会以干净的状态重新启动容器。

Kubernetes 支持很多类型的卷。 Pod可以同时使用任意数目的卷类型。 临时卷类型的生命周期与 Pod 相同， 但持久卷可以比 Pod 的存活期长。 当 Pod 不再存在时，Kubernetes 也会销毁临时卷；不过 Kubernetes 不会销毁持久卷。 对于给定 Pod 中任何类型的卷，在容器重启期间数据都不会丢失。

Kubernetes支持多种Volume类型，包括空目录（emptyDir）、主机路径（hostPath）、NFS、云存储提供商（如AWS EBS、Azure Disk）、持久卷（PersistentVolume）等。

### 持久卷：Persistent Volumes：
PersistentVolume（PV）是集群中由管理员配置的一块存储资源，它独立于Pod存在。PV是集群级别的资源，可以被多个Pod使用。

PV与实际存储后端（如AWS EBS卷、GCE PD卷、Azure Disk、NFS卷等）绑定，由管理员在集群中静态配置或通过存储类动态分配。

> NFS 是一种网络文件系统，英文 Network File System(NFS)，能使使用者访问网络上别处的文件就像在使用自己的计算机一样，NFS 可以独立于集群中的节点，其存储空间可以在集群之外，然后通过网络为多个节点共享文件。不同的节点可以通过远程存取、操作文件，同时文件的状态为所有节点共享，保证每个节点上访问到的文件内容、状态都是一致的、实时的。

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: nfs
spec:
  storageClassName: manual
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteMany
  nfs:
    server: 10.244.1.4
    path: "/"
```
### 持久卷声明：PersistentVolumeClaim：
PersistentVolumeClaim（PVC）是Pod对存储资源的申请，它请求PV以便Pod可以使用持久存储。PVC存在于命名空间内，并且与Pod的生命周期绑定。

Pod通过声明PVC来请求存储资源，而不必关心具体的PV细节。管理员可以设置存储类（StorageClass），使得PVC可以动态地分配PV。

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: nfs
spec:
  accessModes:
    - ReadWriteMany
  storageClassName: manual
  resources:
    requests:
      storage: 1Gi
```

在成功地将 PVC 和 PV 进行绑定之后，Pod 就能够像使用 hostPath 等常规类型的 Volume 一样，在自己的 YAML 文件里声明使用这个 PVC 了，如下所示：
```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    role: web-frontend
spec:
  containers:
  - name: web
    image: nginx
    ports:
      - name: web
        containerPort: 80
    volumeMounts:
        - name: nfs
          mountPath: "/usr/share/nginx/html"
  volumes:
  - name: nfs
    persistentVolumeClaim:
      claimName: nfs
```
Pod 需要做的，就是在 volumes 字段里声明自己要使用的 PVC 名字。接下来，等这个 Pod 创建之后，kubelet 就会把这个 PVC 所对应的 PV，也就是一个 NFS 类型的 Volume，挂载在这个 Pod 容器内的目录上。

### 存储类：StorageClass：
StorageClass定义了动态分配PV的策略，是管理员提供给用户的抽象层。用户通过指定StorageClass在创建PVC时，Kubernetes根据其规则动态分配PV。

存储类可以定义复制策略、卷大小、IOPS要求等参数，以及如何处理卷的动态回收。

### 临时卷：Ephemeral Volume：
有些应用程序需要额外的存储，但并不关心数据在重启后是否仍然可用。 例如，缓存服务经常受限于内存大小，而且可以将不常用的数据转移到比内存慢的存储中，对总体性能的影响并不大。

另有些应用程序需要以文件形式注入的只读数据，比如配置数据或密钥。

临时卷就是为此类用例设计的。因为卷会遵从 Pod 的生命周期，与 Pod 一起创建和删除， 所以停止和重新启动 Pod 时，不会受持久卷在何处可用的限制。

临时卷在 Pod 规约中以内联方式定义，这简化了应用程序的部署和管理。
