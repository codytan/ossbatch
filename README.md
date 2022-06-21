## 说明

ossbatch 为针对主流对象存储系统（暂支持七牛云）进行批量操作的小工具，可按条件检索、批量更改存储类型、批量删除，以清理或者优化存储费用，免于一直增长的存储费用。

和各大厂商自身提供的工具相比，主要就是集成多家（暂未实现）和可以多协程并发处理，加快进度。



#### 配置AK/SK

```
ob config --ak as_string --sk sk_string 
```



#### 扫描符合要求文件

```
ob list --bucket bucket-name --end_date 20200101 --limit 1000000 -o xxx.csv
```
limit参数根据实际批次处理时间决定，测试10万一个批次为一个半天时间方便，也可以一次获取更多，本地切分csv后可多台机器同时进行处理。 
```
以10万行为界限拆分csv
split -d -100000 xxx.csv xxx_
```
如为本地备份数据，则可以管理起来对应的csv文件，用于本地查找的索引。  

#### 批量操作
```
批量删除
batch --bucket bucket-name --type delete --csv xxx.csv --worker 10

批量修改为深度归档存储
batch --bucket bucket-name --type chang3 --csv xxx.csv --worker 10

```


#### 批量下载

```
ob pull --bucket bucket-name --domain http://domain.com --worker 10 --csv xxx.csv
```



