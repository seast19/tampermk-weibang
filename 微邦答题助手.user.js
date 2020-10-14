// ==UserScript==
// @name         微邦答题助手
// @namespace    https://greasyfork.org/zh-CN/users/563657-seast19
// @description  微邦自动匹配答案脚本，解放你的大脑
// @version      2.0.5
// @author       seast19
// @icon         https://s1.ax1x.com/2020/05/18/YWucdO.png
// @require      https://cdn.staticfile.org/layer/3.1.1/mobile/layer.min.js
// @match        http*://weibang.youth.cn/webpage_sapi/examination/detail/*/showDetail/0/phone/platform/*/orgId/httphost/httpport/token
// @nomatch        https://www.nnjjtgs.com/user/nnjexamExercises/paper.html?testactivityId=*
// @nomatch        https://www.nnjjtgs.com/user/nnjexam/paper.html?tpid=*
// @grant        GM_xmlhttpRequest

// ==/UserScript==
;(function () {
  'use strict'

  // 加载layer css
  let link = document.createElement('link')
  link.rel = 'stylesheet'
  link.type = 'text/css'
  link.href = 'https://cdn.staticfile.org/layer/3.1.1/mobile/need/layer.min.css'
  document.head.appendChild(link)

  // ****************全局变量****************************

  // 自定义css
  const cusCSS = `
  #alterDoubleButton {
    display:none !important;
  } 
  `
  // 云函数api
  const serverAPI =
    'https://service-3k85ekq2-1254302252.gz.apigw.tencentcs.com/release/wb_helper_v3'

  // 题目答案列表
  let quesionsList = [] //本次题目
  let answerList = [] //服务器返回的答案
  let allList = [] //全局已经答过的题目

  // let globalTimeout //全局计时器
  let startFlag = false // 运行标志
  let feedBackFlag = false //反馈标记

  // 微邦模型
  let objWB = {
    Title() {
      return document.querySelector('.title_pageDetail').innerText
    },
    // 题目div的父节点 应返回题目div（包含题目选项等信息）的列表
    NodeFather() {
      return document.querySelectorAll('.questionListWrapper_content')
    },
    // 题目
    TextQuestion(subNode) {
      let ques = subNode.querySelector('.titleText_content').childNodes[1]
        .nodeValue
      // 特殊字符处理
      ques = filteSpecialSymbol(ques)
      return ques
    },
    // 题号
    TextNum(subNode) {
      let num = subNode.querySelector('span').innerText
      return num
    },
    // 选项A内容
    TextOptionA(subNode) {
      let optText = subNode.querySelector('label.optionContent_content')
        .innerText
      // 特殊字符处理
      optText = filteSpecialSymbol(optText)
      return optText
    },
    // 选项div 的父节点
    // return []ansNode
    NodeAnsFather(subNode) {
      return subNode.querySelectorAll('label.optionContent_content')
    },
    // 任意选项内容
    TextOpt(ansNode) {
      let optText = ansNode.innerText
      optText = filteSpecialSymbol(optText)
      return optText
    },
    // 点击的节点
    ClickNode(ansNode) {
      return ansNode
    },
  }

  // 众学网模型
  let objZXW = {
    // 父节点
    NodeFather: function () {
      return document.querySelectorAll('li.topic')
    },
    // 问题题目
    TextQuestion: function (subNode) {
      let ques = subNode.querySelector('dt>div').innerText
      // 特殊字符处理
      ques = filteSpecialSymbol(ques)
      return ques
    },
    // 题号
    TextNum: function (subNode) {
      let num = subNode.getAttribute('index')
      return num
    },
    // 选项A
    TextOptionA: function (subNode) {
      let optText = subNode.querySelector('dd').innerText
      // 特殊字符处理
      optText = optText.substring(2) //过滤选项前面的A
      optText = filteSpecialSymbol(optText)
      return optText
    },
    // 选项 父节点
    NodeAnsFather: function (subNode) {
      return subNode.querySelectorAll('dd')
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
      return ansNode.querySelector('label')
    },
  }

  // *****************工具函数***************************** //

  //全局添加样式
  function addGlobalStyle(css) {
    let head, style
    head = document.getElementsByTagName('head')[0]
    if (!head) {
      return
    }
    style = document.createElement('style')
    css = css.replace('!important', '')
    style.innerHTML = css.replace(/;/g, ' !important;')
    head.appendChild(style)
  }

  // 过滤特殊字符串
  function filteSpecialSymbol(oldS) {
    let newS = oldS.replace(//g, '')
    newS = newS.replace(/ /g, '')
    newS = newS.replace(/[\n\r]/g, '')
    return newS
  }

  //错误日搜集
  function feedBack(e, obj) {
    if (feedBackFlag) {
      return
    }

    feedBackFlag = true

    GM_xmlhttpRequest({
      method: 'post',
      headers: {
        'X-LC-Id': 'hYVRtO7xCsS9k7ac4o9bfjKn-gzGzoHsz',
        'X-LC-Key': 'u8XIvYFinbdemgmcSeFrLf87',
        'Content-Type': 'application/json',
      },
      url: 'https://lc-api.seast.net/1.1/classes/wb_feedback',
      data: JSON.stringify({
        title: obj.Title() || '',
        src: window.location.href || '',
        msg: e.msg || '',
        question: e.q || '',
        ans: e.a || '',
      }),
    })
  }

  // ****************爬虫函数**************** //

  // 通用获取问题
  function getQuestions(obj) {
    return new Promise((resolve, reject) => {
      //清空之前的数据
      quesionsList = []
      answerList = []

      //获取所有题目
      for (const subNode of obj.NodeFather()) {
        // 获取题目和选项A
        let tempQues = obj.TextQuestion(subNode)
        let tempOptA = obj.TextOptionA(subNode)

        // 添加未答过的题目
        if (allList.indexOf(tempQues) === -1) {
          quesionsList.push({
            q: tempQues,
            opta: tempOptA,
          })
        }
      }

      // 检验是否有新题目
      if (quesionsList.length > 0) {
        resolve()
      } else {
        reject('当前无新题目')
      }
    })
  }

  // 通用从服务器获取答案
  function getAnswers(obj) {
    return new Promise((resolve, reject) => {
      //跨域请求
      GM_xmlhttpRequest({
        method: 'post',
        url: serverAPI,
        timeout: 20000,
        data: JSON.stringify({
          title: obj.Title(),
          host: window.location.host,
          url: window.location.href,
          data: quesionsList,
        }),
        onload: function (res) {
          answerList = JSON.parse(res.responseText)
          if (answerList.length == undefined) {
            reject('获取题目失败')
          }

          console.log(
            `后台共匹配 ${answerList.length}/${quesionsList.length} 条题目`
          )

          // 反馈
          if (answerList.length != quesionsList.length) {
            feedBack(
              {
                msg: `后台与前端(${answerList.length}/${quesionsList.length})题目数不匹配`,
              },
              obj
            )
          }

          resolve()
        },
        ontimeout: function (e) {
          console.log(e)
          reject('后台服务器连接超时')
        },
        onerror: function (e) {
          console.log(e)
          reject('后台服务器连接失败')
        },
      })
    })
  }

  // 通用页面上选择正确答案
  function chooseAns(obj) {
    // 错误题数
    let errorCount = 0

    //遍历网页题目
    for (const subNode of obj.NodeFather()) {
      //获取本题目
      let question = obj.TextQuestion(subNode)

      //全局题目中有记录，则代表此题已经匹配
      if (allList.indexOf(question) != -1) {
        continue
      }

      //遍历服务器题目，找到网页与服务器匹配题目
      for (const answer of answerList) {
        //获取该题目答案
        if (filteSpecialSymbol(answer.q) === question) {
          let match = false

          //   遍历网页中的答案选项
          for (const ansNode of obj.NodeAnsFather(subNode)) {
            //如果答案在服务器答案数组中
            if (answer.a.indexOf(obj.TextOpt(ansNode)) != -1) {
              match = true

              obj.ClickNode(ansNode).click()

              if (answer.a.length && answer.a.length === 1) {
                break
              }
            }
          }

          //  答案不匹配则控制台显示答案
          if (!match) {
            errorCount += 1
            console.log(`无法匹配题目：${question} 无法匹配`)
            console.log(`正确答案：`)
            console.table(answer.a)
            // 反馈
            feedBack(
              {
                msg: '前端题目不匹配',
                q: `${question}`,
                a: `${answer.a.join('|')}`,
              },
              obj
            )
            break
          }

          // 添加此题目到全局题库，下次生成答案时不再匹配此题目
          allList.push(question)

          break
        }
      }
    }

    // 显示答题结果
    if (errorCount > 0) {
      layer.open({
        content: `共匹配 ${answerList.length} / ${quesionsList.length} 题，其中错误 ${errorCount} 题，请按F2查看详细信息`,
        skin: 'msg',
        time: 3, //2秒后自动关闭
      })
    } else {
      layer.open({
        content: `共匹配 ${answerList.length} / ${quesionsList.length} 题`,
        skin: 'msg',
        time: 3, //2秒后自动关闭
      })
    }
  }

  // 开始函数
  async function start() {
    // 防止多次点击
    if (startFlag === true) {
      console.log('点击频繁')
      return
    }

    startFlag = true

    // 分选答题平台
    let obj = ''
    switch (window.location.host) {
      case 'weibang.youth.cn': // 微邦
        obj = objWB
        break

      case 'www.nnjjtgs.com': // 众学网
        obj = objZXW
        break

      default:
        alert('当前页面暂不支持自动答题')
        return
    }

    // 开始答题
    try {
      await getQuestions(obj)
      layer.open({
        content: '开始答题..',
        skin: 'msg',
        time: 30, //2秒后自动关闭
      })
      await getAnswers(obj)
      chooseAns(obj)
    } catch (e) {
      console.log(e)
      layer.open({
        content: e,
        skin: 'msg',
        time: 3, //2秒后自动关闭
      })
    }

    startFlag = false
  }

  // *********************初始化***********************

  // 初始化函数
  function init() {
    // 屏蔽下载弹窗
    try {
      addGlobalStyle(cusCSS)
    } catch (e) {
      console.log('[weibang_helper]无法加载自定义样式')
    }

    // 加载layer css
    let link = document.createElement('link')
    link.rel = 'stylesheet'
    link.type = 'text/css'
    link.href =
      'https://cdn.staticfile.org/layer/3.1.1/mobile/need/layer.min.css'

    //侧边按钮
    let btnBox = document.createElement('div')
    btnBox.id = 'wk_btn'
    btnBox.style =
      'display:block;z-index:999;position:fixed; right:10px;top:45%; width:70px; height:30px;line-height:30px; background-color:#f50;color:#fff;text-align:center;font-size:16px;font-family:"Microsoft YaHei","微软雅黑",STXihei,"华文细黑",Georgia,"Times New Roman",Arial,sans-serif;font-weight:bold;cursor:pointer'
    btnBox.innerHTML = '生成答案'

    // 事件注入
    document.head.appendChild(link)
    document.querySelector('body').append(btnBox)
    document.getElementById('wk_btn').addEventListener('click', async () => {
      await start()
    })
  }

  init()
})()
