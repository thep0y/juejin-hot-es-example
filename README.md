# 掘金 - Elasticsearch 示例项目
爬取掘金的热门推荐页面部分信息作为示例数据保存到 es 中进行查询。

本项目中对 es 基本的创建、查询和删除操作均有简单实现。

## 简介

查询时使用命令行进行，示例项目的命令如下：

```shell
juejin allows you to index and search hot-recommended article's titles

Usage:
  juejin [command]

Available Commands:
  delete      Delete item with id
  help        Help about any command
  index       Index juejin hot-recommended articles into Elasticsearch
  search      Search juejin hot recommended articles

Flags:
  -h, --help           help for juejin
  -i, --index string   Index name (default "juejin")

Use "juejin [command] --help" for more information about a command.
```

可选命令为`index`、`search`和`delete`。

其中`index`也有可选命令：

```shell
      --pages int   The count of pages you want to crawl (default 5)
      --setup       Create Elasticsearch index
```

本项目使用的是本地 es ，推荐用 docker 创建，es 中需要安装 [ik 中文分词插件](https://github.com/medcl/elasticsearch-analysis-ik)。

### 1 创建索引

```shell
go run main.go index --setup
```

默认会根据项目中指定的 [mapping](https://github.com/thep0y/juejin-hot-es-example/blob/b2e760c2565783fa3c9339a3000f73813ae8c158/commands/index.go#L273) 创建索引，并爬取存储 5 页、共 100 条信息。

结果如下所示：

```shell
8:10PM INF Creating index with mapping
8:10PM INF Starting the crawl with 0 workers at 0 offset
8:10PM INF Stored doc Article ID=6957974706943164447 title="算法篇01、排序算法"
8:10PM INF Stored doc Article ID=6953868764362309639 title="如何处理浏览器的断网情况？"
...
8:10PM INF Skipping existing doc ID=6957726578692341791
8:10PM INF Skipping existing doc ID=6957925118429364255
8:10PM INF Skipping existing doc ID=6953868764362309639
8:10PM INF Skipping existing doc ID=6957981912669519903
8:10PM INF Skipping existing doc ID=6953059119561441287
8:10PM INF Skipping existing doc ID=6955336007839383588
...
8:10PM INF Stored doc Article ID=6957930535574306847 title="Node系列-阻塞和非阻塞的理解"
8:10PM INF Stored doc Article ID=6956602138201948196 title="《前端领域的转译打包工具链》上篇"
8:10PM INF Stored doc Article ID=6957982556885090312 title="JS篇：事件流"
```

终端结果截图：

![截屏2021-05-03 20.14.12](https://z3.ax1x.com/2021/05/03/gm0zqJ.png)

因为每页有 20 条，共爬 5 页，所以理论上应存储 100 条信息，但其中可能会存在几条重复信息，所以最后保存时可能会小于 100 条。

### 2 爬取 10 页

```shell
go run main.go index --pages 10
```

运行这条命令时，不会再创建索引，而是直接开始爬虫，因为只是示例项目，所以没有增加起始页和最终页的选择，只提供最终页码作为可选参数。

运行结果与上小节基本相同：

![截屏2021-05-03 20.17.38](https://z3.ax1x.com/2021/05/03/gmBqwd.png)

### 3 查询

查询时，使用的是[词组查询](https://github.com/thep0y/juejin-hot-es-example/blob/3fdc55c4062fc575a5b9e977919800a42dd18a53/search/store.go#L230)，中文更适合使用词组查询，不然每个查询词被拆分成单字查询，结果一般不是我们想要的。

```shell
go run main.go search 前端
```

查询到的结果中会将查询词高亮显示：

![截屏2021-05-03 20.22.39](https://z3.ax1x.com/2021/05/03/gmDDAA.png)

### 4 删除

```shell
go run main.go delete [id]
```

如：

![截屏2021-05-04 10.21.27](https://z3.ax1x.com/2021/05/04/gnk44f.png)

对已删除的 id 再执行删除操作：

![截屏2021-05-04 10.22.04](https://z3.ax1x.com/2021/05/04/gnk7vQ.png)