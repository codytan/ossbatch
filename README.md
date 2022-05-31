## 说明



## 使用说明

ossbatch 为针对主流对象存储系统（暂支持七牛云）进行批量操作的小工具，可按条件检索、批量更改存储类型、批量删除，以清理或者优化存储费用，免于一直增长的存储费用。

和各大厂商自身提供的工具相比，主要就是集成多家（暂未实现）和可以多协助并发处理，加快进度。



#### 配置AK/SK

```
ob config --ak as_string -dF2Lai5Q --sk sk_string 
```



#### 扫描符合要求文件

```
ob list --bucket bucket-name --end_date 20200101 --limit 10000
```



#### 批量操作

```
批量删除
batch --bucket bucket-name --type delete --csv 20200101_0.csv --worker 10
```





#### 批量下载

```
ob pull --bucket bucket-name --domain http://domain.com --worker 10 --csv 20200101_0.csv
```



