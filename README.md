**StarRocks Buckets分桶修复工具**

功能：

1. 支持整改全量表、分区表的分桶数，排序键、分桶键。

2. 支持查看全量表、分区表每个BE占用的tablet数，存储大小，占用百分比。

3. 支持查看全量表、分区表每个tablet大小。

4. 支持查看当前分桶数、最大容量的分区、表类型、推荐分桶数保底指标、推荐分桶数建议指标、表总大小/容量、表总数量/行数、表创建时间、表最大的合并版本、表注释。

5. 支持分析排序键、分桶键是否合理。

6. 支持清理数据表中的空分区。

命令:

1.`setbuckets -s <app> -t <schema.tablename>`

查看内表明细

![](img/2025-04-30-23-10-22-image.png)

这里最前面有个绿色的【正常】，代表数据表正常，无倾斜。因为数据量也不大，17GB，按照1GB=1~5bucket原则, 它的范围在4~9之间，虽然属于正常，但内部肯定存在微倾斜，因为最大的分区是4.8GB了，四舍五入后5bucket更合理。

2.`setbuckets -s <app> -t <schema.tablename> -l`

查看排序键、分桶键明细

![](img/2025-04-30-23-15-21-image.png)

2.`setbuckets -s <app> -t <schema.tablename> -id all`

查看数据表在所有be中的tablet大小

![](img/2025-04-30-23-17-59-image.png)

这里很明显看到tablet中是存在部分倾斜的，导致的原因是因为分桶键不合理。

真正均衡的内表，tablet长下面这样，更好的保持各项性能

![](img/2025-04-30-23-21-44-image.png)

3.`setbuckets -s <app> -t <schema.tablename> -splitkey <splitkey>`

重置数据表分桶键

![](img/2025-04-30-23-24-28-image.png)

原始语句展示

![](img/2025-04-30-23-25-19-image.png)

开始构建新语句

![](img/2025-04-30-23-26-06-image.png)

数据重新写入

![](img/2025-04-30-23-26-38-image.png)

最后整理列出最新tablet分布情况

![](img/2025-04-30-23-27-25-image.png)

![](img/2025-04-30-23-29-25-image.png)

tablet最大从1.6GB降到400MB，保持最佳性能。

4.`setbuckets -s <app> -t <schema.tablename> -id <BackendId>`

查看单个be中的tablet分布情况

![](img/2025-04-30-23-32-03-image.png)

5.`setbuckets -s <app> -t <schema.tablename> -b <BucketNum>`

修改数据表全局分桶数，从上面的4bucket改到7bucket

![](img/2025-04-30-23-33-43-image.png)

重新构建语句，调整分桶数，写入数据，最后交换

![](img/2025-04-30-23-34-45-image.png)

40s，1.6亿，SR还是不错的选择！

![](img/2025-04-30-23-39-59-image.png)

分桶数已经调整过来了

再查看tablet

![](img/2025-04-30-23-41-31-image.png)

5.`setbuckets -s <app> -t <schema.tablename> -a `

自动调整分桶模式，不需要指定分桶数，内部根据1GB=1~5bucket进行随机调整。

![](img/2025-04-30-23-44-25-image.png)

老套路。

5.`setbuckets -s <app> -t <schema.tablename> -c`

清理空分区

![](img/2025-04-30-23-47-29-image.png)

例如这个表，有853个空分区，只有1个分区是有数据的，这个时候，可以进行清理空分区，减轻元数据压力。

![](img/2025-04-30-23-49-44-image.png)

仅保留有数据的分区

![](img/2025-04-30-23-50-15-image.png)

数据重载

清理前

![](img/2025-04-30-23-52-21-image.png)

清理后

![](img/2025-04-30-23-51-11-image.png)

5.`setbuckets -s <app> -t <schema.tablename> -b <BucketNum> -p [PartitionName] -pset`

（全量表）重置分区级别的分桶（自动识别模式）

![](img/2025-04-30-23-55-09-image.png)

例如上面倾斜的表，进行分区层面的调整

![](img/2025-04-30-23-56-27-image.png)

![](img/2025-04-30-23-57-20-image.png)

结果

![](img/2025-04-30-23-59-25-image.png)

6.`setbuckets -s <app> -t <schema.tablename> -b <BucketNum> -pset`

（分区表）重置分区级别的分桶（自动识别模式）

![](img/2025-05-01-00-08-35-image.png)

自动计算所有异常的分区，并把1214个空分区，分桶调小。

降配到1

![](img/2025-05-01-00-10-11-image.png)

而其他拥有存储的分区，则自动调配。

![](img/2025-05-01-00-20-07-image.png)

另外加上-c，亦能删除空分区。
