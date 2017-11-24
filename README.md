# Swagger generated server

CSE-Java Pojo Server


## Overview
这里的代码是根据用户提供的swagger.yaml文件自动生成。生成的代码总的来说分成以下几类：

1，入口main函数在com.service.service.ServiceApplication这个类中。

2，com.service.service.controller.ServiceImpl这个类插入通过定义RpcSchema这个annotation，拦截用户rpc请求，并将处理的handler代理给serviceDelegate类。

3，com.service.service.controller.ServiceDelegate类是用户实现具体handler实现的类，理论上只需要用户在这个类里面实现业务逻辑。

4，com.service.service.controller.Service这个类定义了由swagger.yaml里定义了的operation的函数footprint。它可以理解为是com.service.service.controller.ServiceImpl实现类的接口。

5，com.service.service.controller.model这个包下面是swagger.yaml里面定义的definitions，也就是model层的类。

6，resource目录下定义了一些配置文件，包括log4j的配置文件，微服务的配置信息，以及swagger协议文件等等。

7，特别注意两点：microservice.yaml里面配置信息要填写正确，生成的有可能有些偏差；swagger.yaml里面的x-java-interface配置项一定要指定为com.service.service.controller.Service。