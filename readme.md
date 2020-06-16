## 微邦答题助手 - tampermonkey 插件

### 微邦答题助手.user.js
1. 此为油猴脚本，在打开微邦答题页面时显示`生成答案`按钮，点击按钮可自动匹配答案。
2. bugs：在pc端因目标页面使用react-virtualized滑动加载列表，无法一次性获取所有题目，需要手动滑动页面至没有获取到的题目处，再次点击`生成答案`以获取答案。或者可以在开发者工具中切换到移动设备，移动设备下可以一次性获取所有题目。


### weibang_serverless.go
1. 此文件为部署在腾讯云函数的匹配答案后台，通过油猴发送题目至云函数，返回题目答案以供油猴解析。
2. 此文件应搭配`./lib20200519.xlsx`使用。
3. 云函数[文档](https://cloud.tencent.com/document/product/583/18032#.E7.BC.96.E8.AF.91.E6.89.93.E5.8C.85)，应将`mian`文件和`./lib20200519.xlsx`打包在一起上传.
4. 编译命令
    ```
    GOOS=linux GOARCH=amd64 go build -o main scf_weibang.go
    ```

### spider_*
1. 用于获取题目

*** 

### 更新

* 2020/05/30 优化代码结构，计划支持众学网
* 2020/05/27 添加题库，过滤题目中的特殊字符以提高匹配率
