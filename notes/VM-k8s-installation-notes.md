# 集群安装小白笔记

# 安装Virtualbox

直接在官网上下载包装包安装，我的电脑是MacOS，安装之后需要改一些权限。

从System Preference→Security & Privacy，General页会有一个关于Oracle的提示，点击Allow按钮。重启电脑。

然后打开Vbox，打开File→Host Network Manager， 里面一般是空的，这时候点击添加按钮，会加入一行。这个就是加上host only的网络，什么都不要改，确认退出。

![Screen Shot 2021-08-19 at 9.20.48 PM.png](%E9%9B%86%E7%BE%A4%E5%AE%89%E8%A3%85%E5%B0%8F%E7%99%BD%E7%AC%94%E8%AE%B0%20c81551372c5c4c0da5cb8aafc982010b/Screen_Shot_2021-08-19_at_9.20.48_PM.png)

# 建立VM

首先在Ubuntu官网下载一个iso. (https://releases.ubuntu.com/18.04.5/ubuntu-18.04.5-live-server-amd64.iso)

在VBox里面点击“New”，起名字，选Linux，Ubuntu 64bit，选择内存大小，其他的都选默认选项。
结束之后选择setting，进一步配置盒子。

1. 在Storage里的虚拟光驱下加入虚拟光盘。
    
    ![Screen Shot 2021-08-19 at 9.25.45 PM.png](%E9%9B%86%E7%BE%A4%E5%AE%89%E8%A3%85%E5%B0%8F%E7%99%BD%E7%AC%94%E8%AE%B0%20c81551372c5c4c0da5cb8aafc982010b/Screen_Shot_2021-08-19_at_9.25.45_PM.png)
    
2. 设置network。第一个adapter是NAT，默认即可。点击Adapter2， enable之后选Host-only即可。Mac版本的Vbox没有在这里手动加ip的选项，暂时不用管。
3. 在system里面点击processor tab，保证CPU个数不小于2。

然后点击Start，启动虚拟机盒子。

启动之后，会直接从iso安装操作系统，一路默认就可以了，也不需要安装额外的软件（一切之后手动安装）。

在这里我遇到一个问题，安装之后系统开始升级，我选了“cancel update and reboot”之后就一直卡在canceling update那里。查看log之后发现系统已经成功安装完毕就直接拔了虚拟电源，在setting里面弹出iso再次开机。一般来说直接等系统自己重启比较好。

# 配置系统

## 安装openssh不然无法工作——界面太糟糕了。

```bash
sudo su
apt-get install openssh-server
```

安装完毕之后记得要启动这个服务：

```bash
systemctl enable ssh  # system will start this service automatically when machine starts
systemctl start ssh
```

systemctl enable 本质是加了一个symbolic link到system init文件里面。

## 配置静态地址

打开 /etc/netplan/00-installer-config.yaml 文件，把 enp0s8下的 dhcp4对应值从true改成no，然后手动加上静态ip地址避免以后使用时每次都要在系统查地址。地址的前三位要跟你在Vbox Host Network Manager里面加network里面给出的一样（我的192.168.56），第四位可以自己选一个数字。改好之后保存。（我是sudo vim打开）

```
network:
  ethernets:
    enp0s3:
      dhcp4: true
    enp0s8:
      dhcp4: no
      addresses:
        - 192.168.56.103/24
  version: 2
```

然后运行

```
netplan apply
```

之后就可以自己开一个terminal ssh进去了。（Mac我用的iTerm2，系统自带也可以。Windows可以装一个putty）

```
ssh username@192.168.56.103
```

最后注意，折腾完了关机用下面的命令，不用每次直接拔虚拟电源了……（直接叉掉运行窗口）

```
poweroff
```

# 安装Kubernettes

## 安装Docker

直接抄老师的命令

```bash
apt-get update
apt-get install \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg-agent \
    software-properties-common
```

### **Add Docker’s official GPG key:**

```bash
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
```

### **Add Docker repositry**

```bash
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io
```

## 安装Kubernetes

第一个cat eof我没看懂，跳过了。先跑了gpg

```bash
gpg --keyserver keyserver.ubuntu.com --recv-keys BA07F4FB
gpg --export --armor BA07F4FB | sudo apt-key add -
```

然后没有update就直接install，出错。在外网所以就直接按照官网[https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/)

上的步骤，直接从google下载而不是走镜像。

1. Update the `apt` package index and install packages needed to use the Kubernetes `apt` repository:
    
    ```
    sudo apt-get update
    sudo apt-get install -y apt-transport-https ca-certificates curl
    ```
    
2. Download the Google Cloud public signing key:
    
    ```
    sudo curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
    ```
    
3. Add the Kubernetes `apt` repository:
    
    ```
    echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
    ```
    
4. Update `apt` package index, install kubelet, kubeadm and kubectl, and pin their version:
    
    ```
    sudo apt-get update
    sudo apt-get install -y kubelet kubeadm kubectl
    
    ```
    
    版本都是1.22.1
    

### **kubeadm init**

因为docker和kubernetes用的cgroup不一样，这里要先改一改不然init会失败

I faced similar issue recently. The problem was cgroup driver. Kubernetes cgroup driver was set to systems but docker was set to systemd. So I created 

```bash
cd /etc/docker
vi daemon.json
```

写入：

```
{
"exec-opts": ["native.cgroupdriver=systemd"]
}
```

然后重启服务：

```bash
systemctl daemon-reload
systemctl restart docker
systemctl restart kubelet
//Run kubeadm init or kubeadm join again.et kubeadm kubectl
//sudo apt-mark hold kubelet kubeadm kubectl
```

参考地址：[https://stackoverflow.com/questions/52119985/kubeadm-init-shows-kubelet-isnt-running-or-healthy](https://stackoverflow.com/questions/52119985/kubeadm-init-shows-kubelet-isnt-running-or-healthy)

抄老师的命令，把版本号和ip改一下

```bash
kubeadm init \
 --image-repository registry.aliyuncs.com/google_containers \
 --kubernetes-version v1.22.1 \
 --apiserver-advertise-address=192.168.56.103
```

这一步折腾了一阵子，除了上面最大的cgroup问题之外， 还遇到以下错误：

1. 忘记开docker或者kubelet服务
2. VM的CPU至少是2不然会报错

如果之前有失败的init，重新跑的时候需要reset 之后再重新init

```
kubeadm reset
```

最后init结束，出来下面的提示

```
Your Kubernetes control-plane has initialized successfully!

To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config

Alternatively, if you are the root user, you can run:

  export KUBECONFIG=/etc/kubernetes/admin.conf

You should now deploy a pod network to the cluster.
Run "kubectl apply -f [podnetwork].yaml" with one of the options listed at:
  https://kubernetes.io/docs/concepts/cluster-administration/addons/

Then you can join any number of worker nodes by running the following on each as root:

kubeadm join 192.168.56.103:6443 --token ji1kq0.7oh2o559xby49ync \
	--discovery-token-ca-cert-hash sha256:2c5271d412dc783cf058dead990e0d623f562fa0a2feb901d716658c0129ae06
```

### **copy kubeconfig**

```
$ mkdir -p $HOME/.kube
$ sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
$ sudo chown $(id -u):$(id -g) $HOME/.kube/config
```

### **untaint master**

```
$ kubectl taint nodes --all node-role.kubernetes.io/master-
```

### **join other node**

```
kubeadm join 192.168.56.103:6443 --token ji1kq0.7oh2o559xby49ync \
	--discovery-token-ca-cert-hash sha256:2c5271d412dc783cf058dead990e0d623f562fa0a2feb901d716658c0129ae06
```

join 出错， 下面的步骤就都还没做

```
[preflight] Running pre-flight checks
error execution phase preflight: [preflight] Some fatal errors occurred:
	[ERROR DirAvailable--etc-kubernetes-manifests]: /etc/kubernetes/manifests is not empty
	[ERROR FileAvailable--etc-kubernetes-kubelet.conf]: /etc/kubernetes/kubelet.conf already exists
	[ERROR Port-10250]: Port 10250 is in use
	[ERROR FileAvailable--etc-kubernetes-pki-ca.crt]: /etc/kubernetes/pki/ca.crt already exists
[preflight] If you know what you are doing, you can make a check non-fatal with `--ignore-preflight-errors=...`
To see the stack trace of this error execute with --v=5 or higher
```

## **install cilium**

```
helm install cilium cilium/cilium --version 1.9.1 \
    --namespace kube-system \
    --set kubeProxyReplacement=strict \
    --set k8sServiceHost=192.168.34.2 \
    --set k8sServicePort=6443
```

## **install calico cni plugin**

[https://docs.projectcalico.org/getting-started/kubernetes/quickstart](https://docs.projectcalico.org/getting-started/kubernetes/quickstart)

`$ kubectl create -f https://docs.projectcalico.org/manifests/tigera-operator.yaml
$ kubectl create -f https://docs.projectcalico.org/manifests/custom-resources.yaml`

`for i in `kubectl api-resources | grep true | awk '{print \$1}'`; do echo $i;kubectl get $i -n rook-ceph; done`