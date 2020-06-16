// ==UserScript==
// @name         微邦答题助手(CMD)
// @namespace    https://greasyfork.org/zh-CN/users/563657-seast19
// @version      1.2
// @description  微邦自动匹配答案脚本，解放你的大脑
// @icon         https://s1.ax1x.com/2020/05/18/YWucdO.png
// @author       seast19
// @match        https://weibang.youth.cn/webpage_sapi/examination/detail/*/showDetail/0/phone/platform/*/orgId/httphost/httpport/token
// @match        https://www.nnjjtgs.com/user/nnjexamExercises/paper.html?testactivityId=*
// @match        https://www.nnjjtgs.com/user/nnjexam/paper.html?tpid=*
// @grant        GM_xmlhttpRequest
// @note         2020/05/30 1.2 优化代码结构，添加对众学网支持
// @note         2020/05/28 1.0 修复多选题重复点击生成答案时会取消选中的情况
// @note         2020/05/26 0.7 修复多选题选项中有重复的答案选项导致无法全部选中的问题
// @note         2020/05/25 0.6 修复因空白字符引起的答案不匹配问题




// ==/UserScript==
(function () {
    'use strict';

    // ****************全局变量****************************

    // 云函数api
    const serverAPI = "https://service-l6svgo68-1254302252.gz.apigw.tencentcs.com/release/weibang_ans_v2"

    // 题目答案列表
    let quesionsList = [] //本次题目
    let answerList = [] //服务器返回的答案
    let allList = [] //全局已经答过的题目

    //全局计时器
    let globalTimeout

    // 运行标志
    let startFlag = false


    // *****************公共函数*****************************

    // 过滤特殊字符串
    function filteSpecialSymbol(oldS) {
        let newS = oldS.replace(//g, "")
        newS = newS.replace(/ /g, "")
        newS = newS.replace(/ /g, "")
        newS = newS.replace(/[\n\r]/g, "")
        return newS
    }

    // 去除重复数组
    function filteRepeatArray(arr) {
        for (let i = 0; i < arr.length - 1; i++) { //遍历获取“前一个元素”，最后一个数不用获取，它本身已经被前面所有元素给排除过了
            for (let j = i + 1; j < arr.length; j++) { //遍历获取剩下的元素，“后一个元素”的起始索引就是“前一个元素”的索引+1
                if (arr[i] == arr[j]) { //如果“前一个元素”与后面剩下的元素之一相同，那么就要删除后面的这个元素
                    arr.splice(j, 1);
                    j--; //如果删除了这个元素，那么后面的元素索引值就会发生改变，所以这里的j需要-1
                }
            }
        }
        return arr
    }

    // 显示信息框
    function showMsgBox(text) {
        //避免调用频繁导致上一次的隐藏消息框影响到本次显示
        clearTimeout(globalTimeout)

        document.getElementById("wk_msg").innerText = text
        document.getElementById("wk_msg").style.display = 'block'
        globalTimeout = setTimeout(() => {
            document.getElementById("wk_msg").style.display = 'none'
        }, 6000)
    }

    //错误日搜集
    function feedBack(e) {
        GM_xmlhttpRequest({
            method: "post",
            headers: {
                "X-LC-Id": "hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz",
                "X-LC-Key": "u8XIvYFinbdemgmcSeFrLf87",
                "Content-Type": "application/json",
            },
            url: "https://lc-api.seast.net" + "/1.1/classes/wb_feedback",
            data: JSON.stringify({
                msg: e.msg,
                question: e.q,
                ans: e.a
            }),
        });
    }


    // 通用获取问题
    function getQuestions(obj) {
        return new Promise((resolve, reject) => {
            showMsgBox('生成答案中..')

            //清空之前的数据
            quesionsList = []
            answerList = []

            //获取所有题目
            for (const subNode of obj.NodeFather()) {
                // 获取题目和选项A
                let tempQues = obj.TextQuestion(subNode)
                let tempOptA = obj.TextOptionA(subNode)

                // 若已经有答题记录，则不再重新获取答案
                if (allList.indexOf(tempQues) === -1) {
                    quesionsList.push({
                        q: tempQues,
                        opta: tempOptA
                    })
                }
            }

            // 检验是否有新题目
            if (quesionsList.length > 0) {
                resolve()
            } else {
                reject("当前无新题目")
            }

        })
    }

    // 通用从服务器获取答案
    function getAnswers() {
        return new Promise((resolve, reject) => {
            //跨域请求
            GM_xmlhttpRequest({
                method: "post",
                url: serverAPI,
                data: JSON.stringify(quesionsList),
                onload: function (res) {
                    answerList = JSON.parse(res.responseText)
                    console.log(`后台共匹配 ${answerList.length}/${quesionsList.length} 条题目`)

                    // 反馈
                    if (answerList.length != quesionsList.length) {
                        feedBack({
                            msg: "后台匹配失败",
                            q: '',
                            a: ""
                        })
                    }

                    resolve()
                },
                onerror: function (e) {
                    console.log(e)
                    reject("后台服务器连接失败")
                }
            });
        })

    }

    

    // 通用控制台显示正确答案
    function showAnsByCMD(obj) {
        return new Promise((resolve, reject) => {

            console.log("*******************");

            showMsgBox(`答案已生成，共[${answerList.length}/${quesionsList.length}]题，请按F12查看`)

            let fatherNode = obj.NodeFather()

            for (let i = 0; i < answerList.length; i++) {
                // 查找题号
                for (const subNode of fatherNode) {
                    let num = obj.TextNum(subNode)
                    let ques = obj.TextQuestion(subNode)

                    if (ques === filteSpecialSymbol(answerList[i].q)) {
                        console.log(`[${answerList[i].opt}] [第${num}题] (${answerList[i].a}) --> 题目：${answerList[i].q}`);
                        break
                    }
                }
                allList.push(answerList[i].q)
            }
            console.log("*******************");

            resolve()
        })

    }


    // 开始函数
    function start(obj) {
        // 防止多次点击
        if (startFlag === true) {
            console.log("点击频繁");
            return
        }

        startFlag = true


        // 获取题目
        getQuestions(obj)
            .then(() => {
                // 获取答案
                return getAnswers()
            })
            .then(() => {
                // 选择正确答案
                // return chooseAns(obj)

                // 控制台显示答案
                return showAnsByCMD(obj)
            })
            .then(() => {
                startFlag = false
            })
            .catch(e => {
                showMsgBox(e)
                startFlag = false
            })
    }

    // *********************微邦*********************

    let objWB = {
        // 题目块父节点 应返回题目块（包含题目选项等信息）的列表
        NodeFather: function () {
            return document.querySelectorAll(".questionListWrapper_content")
        },
        // 题目选择器
        TextQuestion: function (subNode) {
            let ques = subNode.querySelector(".titleText_content").childNodes[1].nodeValue
            // 特殊字符处理
            ques = filteSpecialSymbol(ques)
            return ques
        },
        // 题号
        TextNum: function (subNode) {
            let num = subNode.querySelector("span").innerText
            return num

        },
        // 选项A内容
        TextOptionA: function (subNode) {
            let optText = subNode.querySelector("label.optionContent_content").innerText
            // 特殊字符处理
            optText = filteSpecialSymbol(optText)
            return optText
        },
        // 选项块 父节点
        NodeAnsFather: function (subNode) {
            return subNode.querySelectorAll("label.optionContent_content")
        },
        // 任意选项内容
        TextOpt: function (ansNode) {
            let optText = ansNode.innerText
            optText = filteSpecialSymbol(optText)
            return optText
        },
        // 点击的节点
        ClickNode: function (ansNode) {
            return ansNode
        }
    }


    // ************众学网*********************

    let objZXW = {
        // 父节点
        NodeFather: function () {
            return document.querySelectorAll("li.topic")
        },
        // 问题题目
        TextQuestion: function (subNode) {
            let ques = subNode.querySelector("dt>div").innerText
            // 特殊字符处理
            ques = filteSpecialSymbol(ques)
            return ques
        },
        // 题号
        TextNum: function (subNode) {
            let num = subNode.getAttribute("index")
            return num
        },
        // 选项A
        TextOptionA: function (subNode) {
            let optText = subNode.querySelector("dd").innerText
            // 特殊字符处理
            optText = optText.substring(2) //过滤选项前面的A 
            optText = filteSpecialSymbol(optText)
            return optText
        },
        // 选项 父节点
        NodeAnsFather: function (subNode) {
            return subNode.querySelectorAll("dd")
        },
        // 任意选项
        TextOpt: function (ansNode) {
            let optText = ansNode.innerText
            optText = optText.substring(2)
            optText = filteSpecialSymbol(optText)
            return optText
        },
        //点击的节点
        ClickNode: function (ansNode) {
            return ansNode.querySelector("label")
        }

    }

    // *********************初始化***********************

    //侧边按钮
    let btnBox = document.createElement('div');
    btnBox.id = 'wk_btn';
    btnBox.style = 'display:block;z-index:999;position:fixed; right:10px;top:45%; width:70px; height:30px;line-height:30px; background-color:#f50;color:#fff;text-align:center;font-size:16px;font-family:"Microsoft YaHei","微软雅黑",STXihei,"华文细黑",Georgia,"Times New Roman",Arial,sans-serif;font-weight:bold;cursor:pointer';
    btnBox.innerHTML = '生成答案';

    // 消息提示框
    let msgBox = document.createElement('div');
    msgBox.id = 'wk_msg';
    msgBox.style = 'display:none;z-index:999;position:fixed; left:10%;bottom:10%;width:80%; border-radius: 3px;padding-left: 6px;padding-right: 6px; height:30px;line-height:30px; background-color:#6b6b6b;box-shadow: 6px 6px 4px #e0e0e0;color:#fff;text-align:center;font-size:16px;font-family:"Microsoft YaHei","微软雅黑",STXihei,"华文细黑",Georgia,"Times New Roman",Arial,sans-serif;font-weight:bold;cursor:pointer';
    msgBox.innerHTML = 'this is dafault msg';

    document.querySelector('body').append(btnBox)
    document.querySelector('body').append(msgBox)

    document.getElementById("wk_btn").addEventListener("click", () => {
        // 分选答题平台
        switch (window.location.host) {
            case "weibang.youth.cn": // 微邦
                start(objWB)
                break;

            case "www.nnjjtgs.com": // 众学网
                start(objZXW)
                break

            default:
                alert("当前页面暂不支持自动答题")
                break;
        }
    })




})();