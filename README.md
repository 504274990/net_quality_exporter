# net_quality_exporter
用于探测网络质量并暴露 prometheus 指标

**一个公网质量探测工具，主要用于边缘端项目上传数据到公有云时的网络质量检测，具有如下功能：**
> 公网的丢包率检测
> 网络的响应时间
> 网络的抖动值
> 可同时探测两种不同的 icmp 大小的包
> 可同时监测 k8s 内网和公网两种 dns 解析状态

**配置的grafana面板如下**
![image](https://user-images.githubusercontent.com/13415530/198239144-697e2762-2558-4a04-a51e-e7e351bc62f8.png)

**使用方法**

```
/ # net_quality_exporter --help
usage: net_quality_exporter [<flags>]

Flags:
  -h, --help                     Show context-sensitive help (also try --help-long and --help-man).
      --resolve.domain=www.allsmartcloud.com... ...  
                                 Detect the domain name resolved by dns, It is recommended to add two domain names, one public domain name and one k8s service name
      --web.listen-address=":9331"  
                                 Address to listen on for web interface.
      --web.telemetry-path="/metrics"  
                                 Path under which to expose metrics.
      --ping.interval=60s        Each ping runner interval time.
      --ping.size=32             The size of each ping packet.
      --ping.largesize=1000      The size of each large ping packet.
      --ping.timeout=20s         The entire timeout period of each ping runner.
      --ping.count=15            The number of packets sent by each ping runner.
      --ping.pkg.interval=400ms  The interval of each ping packet.
      --ping.target="106.11.172.9"  
                                 The interval of each ping packet, must be domain name, default is aliyun.com's address.
      --log.level=info           Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt        Output format of log messages. One of: [logfmt, json]
      --version                  Show application version.
```
