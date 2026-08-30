/*
 Navicat Premium Data Transfer

 Source Server         : localhost_27017
 Source Server Type    : MongoDB
 Source Server Version : 80205
 Source Host           : localhost:27017
 Source Schema         : mybilibili

 Target Server Type    : MongoDB
 Target Server Version : 80205
 File Encoding         : 65001

 Date: 24/05/2026 12:16:56
*/


// ----------------------------
// Collection structure for danmakus
// ----------------------------
db.getCollection("danmakus").drop();
db.createCollection("danmakus");

// ----------------------------
// Documents of danmakus
// ----------------------------
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69ba48e4837a0167c3b3a02e"),
    videoId: NumberInt("31"),
    manuscriptId: NumberInt("14"),
    userId: NumberInt("4"),
    content: "你好",
    time: 0,
    color: "#ffffff",
    mode: NumberInt("0"),
    createTime: ISODate("2026-03-18T06:40:36.562Z"),
    _class: "com.mybilibili.common.entity.DanmakuDocument"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69bcc2d4ff9227694f403502"),
    videoId: NumberInt("25"),
    manuscriptId: NumberInt("10"),
    userId: NumberInt("4"),
    content: "小龙虾",
    time: 0,
    color: "#ffffff",
    mode: NumberInt("0"),
    createTime: ISODate("2026-03-20T03:45:24.479Z"),
    _class: "com.mybilibili.common.entity.DanmakuDocument"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69bcc3d1ff9227694f403503"),
    videoId: NumberInt("25"),
    manuscriptId: NumberInt("10"),
    userId: NumberInt("6"),
    content: "openclaw",
    time: 33.136513,
    color: "#ffffff",
    mode: NumberInt("0"),
    createTime: ISODate("2026-03-20T03:49:37.968Z"),
    _class: "com.mybilibili.common.entity.DanmakuDocument"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69bcdcd919edd22faae2d8db"),
    videoId: NumberInt("31"),
    manuscriptId: NumberInt("14"),
    userId: NumberInt("5"),
    content: "电压表",
    time: 23.356331,
    color: "#ffffff",
    mode: NumberInt("0"),
    createTime: ISODate("2026-03-20T05:36:25.409Z"),
    _class: "com.mybilibili.common.entity.DanmakuDocument"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69bcdcfb19edd22faae2d8dc"),
    videoId: NumberInt("31"),
    manuscriptId: NumberInt("14"),
    userId: NumberInt("5"),
    content: "同款小米排插",
    time: 24.048618,
    color: "#ffffff",
    mode: NumberInt("0"),
    createTime: ISODate("2026-03-20T05:36:59.664Z"),
    _class: "com.mybilibili.common.entity.DanmakuDocument"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69bfdabcc595377b83be1b6e"),
    videoId: NumberInt("26"),
    manuscriptId: NumberInt("11"),
    userId: NumberInt("5"),
    content: "你怎么看？",
    time: 0,
    color: "#ffffff",
    mode: NumberInt("0"),
    createTime: ISODate("2026-03-22T12:04:12.325Z"),
    _class: "com.mybilibili.common.entity.DanmakuDocument"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69bfdac5c595377b83be1b6f"),
    videoId: NumberInt("26"),
    manuscriptId: NumberInt("11"),
    userId: NumberInt("5"),
    content: "大模型",
    time: 119.067119,
    color: "#ffffff",
    mode: NumberInt("0"),
    createTime: ISODate("2026-03-22T12:04:21.529Z"),
    _class: "com.mybilibili.common.entity.DanmakuDocument"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69f1b8ef2c7a1917bfbb1e8a"),
    userId: NumberInt("4"),
    content: "小龙虾是什么？",
    time: 82.630061,
    color: "#ffffff",
    mode: NumberInt("0"),
    status: NumberInt("0"),
    createTime: ISODate("2026-04-29T07:53:19.295Z"),
    _class: "com.mybilibili.danmaku.entity.Danmaku"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69fb5b734dcd8269c66cd0f6"),
    userId: NumberInt("5"),
    content: "乒乓球",
    time: 30.986441,
    color: "#ffffff",
    mode: NumberInt("0"),
    status: NumberInt("0"),
    createTime: ISODate("2026-05-06T15:17:06.89Z"),
    _class: "com.mybilibili.danmaku.entity.Danmaku"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69fdf6c93efdde4e7a59473b"),
    userId: NumberInt("4"),
    content: "小龙虾",
    time: 7.562282,
    color: "#ffffff",
    mode: NumberInt("0"),
    status: NumberInt("0"),
    createTime: ISODate("2026-05-08T14:44:25.547Z"),
    _class: "com.mybilibili.danmaku.entity.Danmaku"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69fdffa0d701fc79b2324db1"),
    videoId: NumberInt("29"),
    manuscriptId: NumberInt("12"),
    userId: NumberInt("4"),
    content: "openclaw",
    time: 90.604689,
    color: "#ffffff",
    mode: NumberInt("0"),
    status: NumberInt("0"),
    createTime: ISODate("2026-05-08T15:22:08.007Z"),
    _class: "com.mybilibili.danmaku.entity.Danmaku"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69fe002bd701fc79b2324db2"),
    videoId: NumberInt("36"),
    manuscriptId: NumberInt("29"),
    userId: NumberInt("4"),
    content: "乒乓球",
    time: 23.073065,
    color: "#ffffff",
    mode: NumberInt("0"),
    status: NumberInt("0"),
    createTime: ISODate("2026-05-08T15:24:27.633Z"),
    _class: "com.mybilibili.danmaku.entity.Danmaku"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69ff202f241930124519cc62"),
    videoId: NumberInt("34"),
    manuscriptId: NumberInt("27"),
    userId: NumberInt("5"),
    content: "basketball",
    time: 8.52862,
    color: "#ffffff",
    mode: NumberInt("0"),
    status: NumberInt("0"),
    createTime: ISODate("2026-05-09T11:53:19.826Z"),
    _class: "com.mybilibili.danmaku.entity.Danmaku"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69ff204e241930124519cc63"),
    videoId: NumberInt("34"),
    manuscriptId: NumberInt("27"),
    userId: NumberInt("5"),
    content: "一段式投篮",
    time: 38.903784,
    color: "#A0EE00",
    mode: NumberInt("0"),
    status: NumberInt("0"),
    createTime: ISODate("2026-05-09T11:53:50.861Z"),
    _class: "com.mybilibili.danmaku.entity.Danmaku"
} ]);
db.getCollection("danmakus").insert([ {
    _id: ObjectId("69ffdfbd516caa32d651da32"),
    videoId: NumberInt("25"),
    manuscriptId: NumberInt("10"),
    userId: NumberInt("4"),
    content: "你好",
    time: 2.217859,
    color: "#ffffff",
    mode: NumberInt("0"),
    status: NumberInt("0"),
    createTime: ISODate("2026-05-10T01:30:37.023Z"),
    _class: "com.mybilibili.danmaku.entity.Danmaku"
} ]);

// ----------------------------
// Collection structure for subtitles
// ----------------------------
db.getCollection("subtitles").drop();
db.createCollection("subtitles");

// ----------------------------
// Documents of subtitles
// ----------------------------
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69ba1fc473b37c62d03700ec"),
    videoId: NumberInt("25"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 4.34,
            text: "本期视频介绍一款自带免费无线token的小龙虾版本"
        },
        {
            index: NumberInt("2"),
            startTime: 4.34,
            endTime: 8.2,
            text: "这是由数则大佬开发的OpenCla 中文分支版"
        },
        {
            index: NumberInt("3"),
            startTime: 8.2,
            endTime: 12.74,
            text: "该版本可无线免费使用Deepseq和千万的Web版模型"
        },
        {
            index: NumberInt("4"),
            startTime: 12.74,
            endTime: 16.84,
            text: "其中包含联网石图、编程深度思考等多种模型"
        },
        {
            index: NumberInt("5"),
            startTime: 16.84,
            endTime: 19.24,
            text: "从此养虾再也不怕烧token了"
        },
        {
            index: NumberInt("6"),
            startTime: 19.24,
            endTime: 22.64,
            text: "同时该项目不但对官方版本进行了焊化"
        },
        {
            index: NumberInt("7"),
            startTime: 22.64,
            endTime: 26.88,
            text: "还对逻辑思考认知能力、出错能力进行了深度的升级"
        },
        {
            index: NumberInt("8"),
            startTime: 26.88,
            endTime: 28.88,
            text: "那么接下来就让我们尝试"
        },
        {
            index: NumberInt("9"),
            startTime: 28.88,
            endTime: 31.88,
            text: "在Windows系统内进行安装"
        },
        {
            index: NumberInt("10"),
            startTime: 31.88,
            endTime: 36.32,
            text: "首先需要安装两个系统所需的依赖程序"
        },
        {
            index: NumberInt("11"),
            startTime: 36.32,
            endTime: 40.08,
            text: "Node和Gate执行安装程序无脑下一步即可"
        },
        {
            index: NumberInt("12"),
            startTime: 40.08,
            endTime: 43.28,
            text: "两个程序的安装包我已经放到了网盘里"
        },
        {
            index: NumberInt("13"),
            startTime: 43.28,
            endTime: 48.08,
            text: "大家可以去视频简介或者评论区找一找网盘链接"
        },
        {
            index: NumberInt("14"),
            startTime: 48.08,
            endTime: 50.88,
            text: "好的那么基础的依赖安装好之后"
        },
        {
            index: NumberInt("15"),
            startTime: 50.88,
            endTime: 54.68,
            text: "我们可以去到一个空的文件夹内点击鼠标右键"
        },
        {
            index: NumberInt("16"),
            startTime: 54.68,
            endTime: 56.98,
            text: "然后选择在中端中打开"
        },
        {
            index: NumberInt("17"),
            startTime: 56.98,
            endTime: 61.38,
            text: "让我们来输入两个命令验证Node和Gate是否安装成功了"
        },
        {
            index: NumberInt("18"),
            startTime: 61.38,
            endTime: 65.18,
            text: "如果命令行返回了版本号那就说明安装成功"
        },
        {
            index: NumberInt("19"),
            startTime: 65.18,
            endTime: 67.48,
            text: "但是如果跳出了一些红字"
        },
        {
            index: NumberInt("20"),
            startTime: 67.48,
            endTime: 71.28,
            text: "那么你就需要选中这些红字在点击鼠标右键"
        },
        {
            index: NumberInt("21"),
            startTime: 71.28,
            endTime: 75.28,
            text: "然后去逗包直接粘贴错误信息问问看怎么解决"
        },
        {
            index: NumberInt("22"),
            startTime: 75.28,
            endTime: 78.48,
            text: "正常情况下我们都可以顺利走到这一步"
        },
        {
            index: NumberInt("23"),
            startTime: 78.48,
            endTime: 81.78,
            text: "网盘里我还会提供新版的命令行工具"
        },
        {
            index: NumberInt("24"),
            startTime: 81.78,
            endTime: 85.68,
            text: "以及我一直在用的科学上网工具"
        },
        {
            index: NumberInt("25"),
            startTime: 85.68,
            endTime: 89.28,
            text: "那么接下来就到了正式安装小龙虾的时间了"
        },
        {
            index: NumberInt("26"),
            startTime: 89.28,
            endTime: 90.88,
            text: "为了方便大家使用"
        },
        {
            index: NumberInt("27"),
            startTime: 90.88,
            endTime: 94.58,
            text: "我已经提前下载好了硕则大佬的项目文件包"
        },
        {
            index: NumberInt("28"),
            startTime: 94.58,
            endTime: 98.28,
            text: "大家可以直接将压缩包解压到你的地盘跟墨路"
        },
        {
            index: NumberInt("29"),
            startTime: 98.28,
            endTime: 101.48,
            text: "当然你想解压到其他位置也都可以"
        },
        {
            index: NumberInt("30"),
            startTime: 101.48,
            endTime: 105.88,
            text: "如果你懂得如何使用Gate Clone命令拉取项目文件"
        },
        {
            index: NumberInt("31"),
            startTime: 105.88,
            endTime: 109.18,
            text: "那么你也可以考虑不用我提供的项目包"
        },
        {
            index: NumberInt("32"),
            startTime: 109.18,
            endTime: 113.98,
            text: "自己手动在命令行中拉取最新的项目文件也是可以的"
        },
        {
            index: NumberInt("33"),
            startTime: 113.98,
            endTime: 118.48,
            text: "当然在网盘中我也会持续更新一段时间"
        },
        {
            index: NumberInt("34"),
            startTime: 118.48,
            endTime: 121.88,
            text: "好了解压完成后我们可以删除压缩包"
        },
        {
            index: NumberInt("35"),
            startTime: 121.88,
            endTime: 124.28,
            text: "并对项目墨路进行改名"
        },
        {
            index: NumberInt("36"),
            startTime: 124.28,
            endTime: 128.28,
            text: "使用命令行操作的同学可以复制这段改名的命令"
        },
        {
            index: NumberInt("37"),
            startTime: 128.28,
            endTime: 132.38,
            text: "正常解压说的同学可以直接手动修改文件夹名称"
        },
        {
            index: NumberInt("38"),
            startTime: 132.38,
            endTime: 135.38,
            text: "其实修改名称这一步不是必须的"
        },
        {
            index: NumberInt("39"),
            startTime: 135.38,
            endTime: 139.18,
            text: "我这么做纯粹是想文件夹名称简单点"
        },
        {
            index: NumberInt("40"),
            startTime: 139.18,
            endTime: 143.78,
            text: "解压完成后打开文件夹看看是不是看到了项目文件"
        },
        {
            index: NumberInt("41"),
            startTime: 143.78,
            endTime: 145.68,
            text: "接下来是很重要的一步"
        },
        {
            index: NumberInt("42"),
            startTime: 145.68,
            endTime: 149.68,
            text: "但我们在项目文件夹内的空白处点击鼠标右键"
        },
        {
            index: NumberInt("43"),
            startTime: 149.68,
            endTime: 151.98,
            text: "然后选择在中端内打开"
        },
        {
            index: NumberInt("44"),
            startTime: 151.98,
            endTime: 155.78,
            text: "此时你应该能看到命令行墨路就是你的项目墨路"
        },
        {
            index: NumberInt("45"),
            startTime: 155.78,
            endTime: 159.68,
            text: "然后复制这句命令让电脑自动安装项目依赖"
        },
        {
            index: NumberInt("46"),
            startTime: 159.68,
            endTime: 161.58,
            text: "我这里视频加个速"
        },
        {
            index: NumberInt("47"),
            startTime: 161.58,
            endTime: 168.18,
            text: "安装完成之后只要没有下人的大段红字错误信息"
        },
        {
            index: NumberInt("48"),
            startTime: 168.18,
            endTime: 172.68,
            text: "那我们就可以接着复制这句命令进行项目编译"
        },
        {
            index: NumberInt("49"),
            startTime: 172.68,
            endTime: 174.48,
            text: "我这里视频继续加个速"
        },
        {
            index: NumberInt("50"),
            startTime: 174.48,
            endTime: 180.98,
            text: "编译好之后我们就已经将小龙虾安装完成了"
        },
        {
            index: NumberInt("51"),
            startTime: 180.98,
            endTime: 185.18,
            text: "你可以试试输入这个命令看有没有出现小龙虾的版本好"
        },
        {
            index: NumberInt("52"),
            startTime: 185.18,
            endTime: 191.78,
            text: "目前项目版本是3.1版"
        },
        {
            index: NumberInt("53"),
            startTime: 191.78,
            endTime: 195.68,
            text: "OK接下来需要对OpenCloud做初始化配置"
        },
        {
            index: NumberInt("54"),
            startTime: 195.68,
            endTime: 198.98,
            text: "输入这段命令你将会看到中文版的配置说明"
        },
        {
            index: NumberInt("55"),
            startTime: 202.78,
            endTime: 204.78,
            text: "对于首个安全提示问题"
        },
        {
            index: NumberInt("56"),
            startTime: 204.78,
            endTime: 206.48,
            text: "我们点按键盘的左键"
        },
        {
            index: NumberInt("57"),
            startTime: 206.48,
            endTime: 209.68,
            text: "选择Yes再按回车键确认选择"
        },
        {
            index: NumberInt("58"),
            startTime: 209.68,
            endTime: 213.58,
            text: "第二个问题我们直接按回车键选择快速启动"
        },
        {
            index: NumberInt("59"),
            startTime: 213.58,
            endTime: 217.28,
            text: "第三个问题需要我们选择小龙虾的核心模型"
        },
        {
            index: NumberInt("60"),
            startTime: 217.28,
            endTime: 220.28,
            text: "这里我们点按键盘的上键或者下键"
        },
        {
            index: NumberInt("61"),
            startTime: 220.28,
            endTime: 223.58,
            text: "选择Deep Seek Browser并按回车键确认"
        },
        {
            index: NumberInt("62"),
            startTime: 223.58,
            endTime: 228.58,
            text: "确认之后小龙虾需要我们进行一次网页登录"
        },
        {
            index: NumberInt("63"),
            startTime: 228.58,
            endTime: 232.28,
            text: "这里我们直接按回车键选择Automated Logging"
        },
        {
            index: NumberInt("64"),
            startTime: 232.98,
            endTime: 236.98,
            text: "之后小龙虾会自动启动浏览器并打开Deep Seek首页"
        },
        {
            index: NumberInt("65"),
            startTime: 236.98,
            endTime: 240.38,
            text: "我们输入登录账号后浏览器会自动关闭"
        },
        {
            index: NumberInt("66"),
            startTime: 240.38,
            endTime: 244.48,
            text: "接着你会看到小龙虾已经识别到了多个Deep Seek模型"
        },
        {
            index: NumberInt("67"),
            startTime: 244.48,
            endTime: 248.28,
            text: "我们可以按上下键选择一个小龙虾的主力模型"
        },
        {
            index: NumberInt("68"),
            startTime: 248.28,
            endTime: 251.68,
            text: "或者你也可以选择作者推荐的当前模型"
        },
        {
            index: NumberInt("69"),
            startTime: 251.68,
            endTime: 255.28,
            text: "这里你看到的模型都是可以免费无限使用的"
        },
        {
            index: NumberInt("70"),
            startTime: 255.28,
            endTime: 257.28,
            text: "按回车键确认选择"
        },
        {
            index: NumberInt("71"),
            startTime: 257.28,
            endTime: 261.98,
            text: "到了这一步是选择手机上接入的聊天工具"
        },
        {
            index: NumberInt("72"),
            startTime: 261.98,
            endTime: 264.18,
            text: "我们可以选择暂时跳过"
        },
        {
            index: NumberInt("73"),
            startTime: 264.18,
            endTime: 267.78,
            text: "接着关于技能的配置"
        },
        {
            index: NumberInt("74"),
            startTime: 267.78,
            endTime: 269.98,
            text: "我们按右方向键选择No"
        },
        {
            index: NumberInt("75"),
            startTime: 269.98,
            endTime: 273.88,
            text: "再按回车键确认因为暂时不用去管这些技能"
        },
        {
            index: NumberInt("76"),
            startTime: 273.88,
            endTime: 275.48,
            text: "到了后口设置环节"
        },
        {
            index: NumberInt("77"),
            startTime: 275.48,
            endTime: 278.08,
            text: "我们可以按下方向键再按空格键"
        },
        {
            index: NumberInt("78"),
            startTime: 278.08,
            endTime: 281.98,
            text: "一次的把这几项都勾选上然后按回车键确认"
        },
        {
            index: NumberInt("79"),
            startTime: 281.98,
            endTime: 285.38,
            text: "到此小龙虾的初始化设置就结束了"
        },
        {
            index: NumberInt("80"),
            startTime: 285.38,
            endTime: 288.68,
            text: "我们选择Rista 按回车键确认"
        },
        {
            index: NumberInt("81"),
            startTime: 288.68,
            endTime: 291.88,
            text: "稍等片刻就可以看到小龙虾的页面了"
        },
        {
            index: NumberInt("82"),
            startTime: 291.88,
            endTime: 293.68,
            text: "如果你的页面显示失败"
        },
        {
            index: NumberInt("83"),
            startTime: 293.68,
            endTime: 295.78,
            text: "那你需要回到命令航窗口"
        },
        {
            index: NumberInt("84"),
            startTime: 295.78,
            endTime: 298.38,
            text: "输入这句命令启动小龙虾服务"
        },
        {
            index: NumberInt("85"),
            startTime: 298.38,
            endTime: 300.38,
            text: "要注意当服务启动之后"
        },
        {
            index: NumberInt("86"),
            startTime: 300.38,
            endTime: 302.98,
            text: "这个命令航窗口就不可以关掉了"
        },
        {
            index: NumberInt("87"),
            startTime: 302.98,
            endTime: 305.48,
            text: "否则小龙虾的服务会停止"
        },
        {
            index: NumberInt("88"),
            startTime: 305.48,
            endTime: 308.18,
            text: "如果你不小心还是关掉了这个窗口"
        },
        {
            index: NumberInt("89"),
            startTime: 308.18,
            endTime: 311.28,
            text: "也没关系只要打开新的命令航窗口"
        },
        {
            index: NumberInt("90"),
            startTime: 311.28,
            endTime: 312.78,
            text: "去到小龙虾的目录"
        },
        {
            index: NumberInt("91"),
            startTime: 312.78,
            endTime: 315.58,
            text: "然后执行这个服务启动命令即可"
        },
        {
            index: NumberInt("92"),
            startTime: 315.58,
            endTime: 319.18,
            text: "服务启动后我们再回到浏览器刷新页面"
        },
        {
            index: NumberInt("93"),
            startTime: 319.18,
            endTime: 322.08,
            text: "就可以看到小龙虾的对话窗口了"
        },
        {
            index: NumberInt("94"),
            startTime: 322.08,
            endTime: 325.18,
            text: "到这里我们就已经完成了小龙虾的安装"
        },
        {
            index: NumberInt("95"),
            startTime: 325.18,
            endTime: 328.98,
            text: "并且拥有了无线Token的Deep Seek模型"
        },
        {
            index: NumberInt("96"),
            startTime: 328.98,
            endTime: 331.38,
            text: "因为硕则大佬除了Deep Seek模型"
        },
        {
            index: NumberInt("97"),
            startTime: 331.38,
            endTime: 332.98,
            text: "还提供了千万模型"
        },
        {
            index: NumberInt("98"),
            startTime: 332.98,
            endTime: 335.58,
            text: "植入千万模型的方法也很简单"
        },
        {
            index: NumberInt("99"),
            startTime: 335.58,
            endTime: 337.38,
            text: "只要回到命令航窗口"
        },
        {
            index: NumberInt("100"),
            startTime: 337.38,
            endTime: 340.28,
            text: "按Ctrl+C 停止运行当前服务"
        },
        {
            index: NumberInt("101"),
            startTime: 340.28,
            endTime: 342.88,
            text: "然后再粘贴初始化命令按回车"
        },
        {
            index: NumberInt("102"),
            startTime: 342.88,
            endTime: 344.98,
            text: "重新走一遍初始化设置"
        },
        {
            index: NumberInt("103"),
            startTime: 344.98,
            endTime: 348.38,
            text: "在模型选择环节去选择千万模型即可"
        },
        {
            index: NumberInt("104"),
            startTime: 348.38,
            endTime: 351.98,
            text: "后面的流程与选择DPCK模型的流程一样"
        },
        {
            index: NumberInt("105"),
            startTime: 351.98,
            endTime: 354.38,
            text: "小龙虾打开浏览器你来登录"
        },
        {
            index: NumberInt("106"),
            startTime: 354.38,
            endTime: 356.18,
            text: "选择一堆设置项"
        },
        {
            index: NumberInt("107"),
            startTime: 356.18,
            endTime: 359.18,
            text: "然后再次去命令航启动服务"
        },
        {
            index: NumberInt("108"),
            startTime: 359.18,
            endTime: 361.98,
            text: "这时我们的小龙虾就同时拥有了"
        },
        {
            index: NumberInt("109"),
            startTime: 361.98,
            endTime: 364.88,
            text: "Deep Seek模型和千万模型"
        },
        {
            index: NumberInt("110"),
            startTime: 364.88,
            endTime: 367.68,
            text: "接下来你就可以开始无压力的养虾了"
        },
        {
            index: NumberInt("111"),
            startTime: 367.68,
            endTime: 370.08,
            text: "那么本期视频也就到这里了"
        },
        {
            index: NumberInt("112"),
            startTime: 370.08,
            endTime: 371.78,
            text: "视频录制的比较仓促"
        },
        {
            index: NumberInt("113"),
            startTime: 371.78,
            endTime: 373.38,
            text: "如果对你有所帮助"
        },
        {
            index: NumberInt("114"),
            startTime: 373.38,
            endTime: 374.68,
            text: "还请一键三连"
        },
        {
            index: NumberInt("115"),
            startTime: 374.68,
            endTime: 376.68,
            text: "您的鼓励是我最大的动力"
        },
        {
            index: NumberInt("116"),
            startTime: 376.68,
            endTime: 378.18,
            text: "如果您有什么疑问"
        },
        {
            index: NumberInt("117"),
            startTime: 378.18,
            endTime: 381.18,
            text: "可以留言或者加入交流群讨论"
        },
        {
            index: NumberInt("118"),
            startTime: 381.18,
            endTime: 383.18,
            text: "好了 我们下期视频再见"
        }
    ],
    isDefault: true,
    uploadedBy: NumberInt("0"),
    uploadTime: ISODate("2026-03-18T03:45:08.892Z"),
    status: NumberInt("1"),
    source: "admin",
    version: NumberInt("1"),
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69bd236719fd4c5fdb9231d5"),
    videoId: NumberInt("31"),
    language: "zh-CN",
    languageName: "简体中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 2.56,
            text: "如何使用這台萬用錶檢測電壓"
        },
        {
            index: NumberInt("2"),
            startTime: 2.56,
            endTime: 5.16,
            text: "首先按確定鍵進入萬用錶頁面"
        },
        {
            index: NumberInt("3"),
            startTime: 5.16,
            endTime: 7,
            text: "可以在自動檔位測量"
        },
        {
            index: NumberInt("4"),
            startTime: 7,
            endTime: 8.92,
            text: "也可以在電壓檔位測量"
        },
        {
            index: NumberInt("5"),
            startTime: 8.92,
            endTime: 10.56,
            text: "我們調整到電壓檔位"
        },
        {
            index: NumberInt("6"),
            startTime: 10.56,
            endTime: 12.28,
            text: "拿出我們的電池和錶筆"
        },
        {
            index: NumberInt("7"),
            startTime: 12.28,
            endTime: 13.84,
            text: "紅色錶筆接正極"
        },
        {
            index: NumberInt("8"),
            startTime: 13.84,
            endTime: 15.24,
            text: "黑色錶筆接附極"
        },
        {
            index: NumberInt("9"),
            startTime: 15.24,
            endTime: 17.6,
            text: "此時會顯示測得的電壓"
        },
        {
            index: NumberInt("10"),
            startTime: 17.6,
            endTime: 20.48,
            text: "按一下模式鍵切換交流電壓測量"
        },
        {
            index: NumberInt("11"),
            startTime: 20.48,
            endTime: 22.84,
            text: "此時將錶筆同時插入插座"
        },
        {
            index: NumberInt("12"),
            startTime: 22.84,
            endTime: 24.84,
            text: "即可測得交流電壓值"
        }
    ],
    isDefault: true,
    uploadedBy: NumberInt("0"),
    uploadTime: ISODate("2026-03-20T10:37:27.525Z"),
    status: NumberInt("3"),
    source: "system",
    version: NumberInt("1"),
    filePath: "d:/files/mybilibili/uploads\\manuscripts\\14\\videos\\31\\subtitles\\zh-CN.srt",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69d0c3f5bce1f92dff92f6be"),
    videoId: NumberInt("26"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 4.92,
            text: "如果你在用 Cloud Code 或 Cursor 鞋带码,"
        },
        {
            index: NumberInt("2"),
            startTime: 4.92,
            endTime: 8.28,
            text: "每次工具调用都在吞噬你最宝贵的资源。"
        },
        {
            index: NumberInt("3"),
            startTime: 8.28,
            endTime: 13.04,
            text: "Context Window 有个项目说能省 98%。"
        },
        {
            index: NumberInt("4"),
            startTime: 13.04,
            endTime: 15.94,
            text: "但我的判断是,Sentoken 只是表面。"
        },
        {
            index: NumberInt("5"),
            startTime: 15.94,
            endTime: 19.56,
            text: "这背后是 AI Agent 开发的架构盲区。"
        },
        {
            index: NumberInt("6"),
            startTime: 19.56,
            endTime: 27.1,
            text: "大家好,这里是 LLMX Factors,一个专注于拆解大语言模型时代底层逻辑的频道。"
        },
        {
            index: NumberInt("7"),
            startTime: 27.1,
            endTime: 33.22,
            text: "一次 Playwright 快照 50,000 投坑,一次 Git Log 15,000。"
        },
        {
            index: NumberInt("8"),
            startTime: 33.22,
            endTime: 36.76,
            text: "三四个工具连着条,10,000 投坑就没了。"
        },
        {
            index: NumberInt("9"),
            startTime: 36.76,
            endTime: 39.06,
            text: "Context Window 才 20,000。"
        },
        {
            index: NumberInt("10"),
            startTime: 39.06,
            endTime: 41.78,
            text: "Context Mode 能压到 1,000-2,000。"
        },
        {
            index: NumberInt("11"),
            startTime: 41.78,
            endTime: 45.78,
            text: "但更值得关注的是社区讨论里暴露的深层问题。"
        },
        {
            index: NumberInt("12"),
            startTime: 45.78,
            endTime: 48.1,
            text: "先看看问题有多严重。"
        },
        {
            index: NumberInt("13"),
            startTime: 48.1,
            endTime: 52.06,
            text: "你越勤快地使用工具,你的绘画死得越快。"
        },
        {
            index: NumberInt("14"),
            startTime: 52.06,
            endTime: 54.54,
            text: "每次调用,原始数据用进来。"
        },
        {
            index: NumberInt("15"),
            startTime: 54.54,
            endTime: 59.02,
            text: "几次之后,Context 就被预置 Jason 快照填满了。"
        },
        {
            index: NumberInt("16"),
            startTime: 59.02,
            endTime: 62.1,
            text: "模型开始忘记之前说过什么,反低级错误。"
        },
        {
            index: NumberInt("17"),
            startTime: 62.1,
            endTime: 64.42,
            text: "你只能开新汇,画从头来。"
        },
        {
            index: NumberInt("18"),
            startTime: 64.42,
            endTime: 71.86,
            text: "举个例子,你让 AI 跑测试,你只需要知道通过没有,哪行失败了,什么错误。"
        },
        {
            index: NumberInt("19"),
            startTime: 71.86,
            endTime: 74.42,
            text: "但进入 Context 的是完整预置。"
        },
        {
            index: NumberInt("20"),
            startTime: 74.42,
            endTime: 77.94,
            text: "所有通过失败的细节,堆盏追踪,环境变亮。"
        },
        {
            index: NumberInt("21"),
            startTime: 77.94,
            endTime: 80.66,
            text: "九成 Token 花在你根本不需要的信息上。"
        },
        {
            index: NumberInt("22"),
            startTime: 80.66,
            endTime: 83.3,
            text: "Context Mode 是怎么解决的?"
        },
        {
            index: NumberInt("23"),
            startTime: 83.3,
            endTime: 88.7,
            text: "思路很简单,与其数据进来在压缩,不如一开始就不让它进来。"
        },
        {
            index: NumberInt("24"),
            startTime: 88.7,
            endTime: 93.58,
            text: "工具执行,放进纱盒子进尘,原始输出,永远不进对话历史。"
        },
        {
            index: NumberInt("25"),
            startTime: 93.58,
            endTime: 95.98,
            text: "只有减短摘要,回到 Context。"
        },
        {
            index: NumberInt("26"),
            startTime: 95.98,
            endTime: 99.42,
            text: "完整数据,存进本地 SQLite 全文搜索所引。"
        },
        {
            index: NumberInt("27"),
            startTime: 99.42,
            endTime: 101.9,
            text: "模型需要细节时,可以主动查。"
        },
        {
            index: NumberInt("28"),
            startTime: 101.9,
            endTime: 105.74,
            text: "关键细节:不需要额外大模型调用。"
        },
        {
            index: NumberInt("29"),
            startTime: 105.74,
            endTime: 110.46,
            text: "用 SQLite 内置的 FTS5 全文搜索加 BM25 排序。"
        },
        {
            index: NumberInt("30"),
            startTime: 110.46,
            endTime: 114.14,
            text: "纯算法,凌乘本,确定性输出。"
        },
        {
            index: NumberInt("31"),
            startTime: 114.14,
            endTime: 118.86,
            text: "同样的查询,永远返回同样的结果,对缓存非常友好。"
        },
        {
            index: NumberInt("32"),
            startTime: 118.86,
            endTime: 125.22,
            text: "社区有个更聪明的思路,有人指出纯关键次匹配在结构化数据上不够好。"
        },
        {
            index: NumberInt("33"),
            startTime: 125.22,
            endTime: 133.06,
            text: "做了混合方案,FTS5 做精准匹配函数名,变亮名,向亮搜索,做语义理解。"
        },
        {
            index: NumberInt("34"),
            startTime: 133.06,
            endTime: 139.94,
            text: "用 RRF 融合两种结果,它用这套方案锁引了一万五千多个文件,效果很好。"
        },
        {
            index: NumberInt("35"),
            startTime: 139.94,
            endTime: 142.42,
            text: "但实测发现一个硬伤。"
        },
        {
            index: NumberInt("36"),
            startTime: 142.42,
            endTime: 145.66,
            text: "Context Mode 只能拦截内置工具。"
        },
        {
            index: NumberInt("37"),
            startTime: 145.66,
            endTime: 148.26,
            text: "第三方 MCP 工具完全无效。"
        },
        {
            index: NumberInt("38"),
            startTime: 148.26,
            endTime: 154.26,
            text: "因为 MCP 响应通过 JSON RPC 直接发给模型,没有 Hook 可以拦截。"
        },
        {
            index: NumberInt("39"),
            startTime: 154.26,
            endTime: 157.54,
            text: "MCP 工具返回大量数据时,Token 照样吃掉。"
        },
        {
            index: NumberInt("40"),
            startTime: 157.54,
            endTime: 160.54,
            text: "这个问题只能 MCP 开发者在服务端解决。"
        },
        {
            index: NumberInt("41"),
            startTime: 160.54,
            endTime: 165.06,
            text: "但比技术方案更重要的是这件事揭示的底层规律。"
        },
        {
            index: NumberInt("42"),
            startTime: 165.06,
            endTime: 167.98,
            text: "这件事特别像操作系统早期。"
        },
        {
            index: NumberInt("43"),
            startTime: 167.98,
            endTime: 172.98,
            text: "那时内存细缺,程序直接操作物理内存,装不下就崩溃。"
        },
        {
            index: NumberInt("44"),
            startTime: 172.98,
            endTime: 176.98,
            text: "后来有了虚拟内存,现在 Context Window 就是新内存。"
        },
        {
            index: NumberInt("45"),
            startTime: 176.98,
            endTime: 179.98,
            text: "工具输出直接灌进去,满了就降止。"
        },
        {
            index: NumberInt("46"),
            startTime: 179.98,
            endTime: 182.98,
            text: "Context Mode 本质就是虚拟内存。"
        },
        {
            index: NumberInt("47"),
            startTime: 182.98,
            endTime: 185.98,
            text: "社区提出的想法印证了这一点。"
        },
        {
            index: NumberInt("48"),
            startTime: 185.98,
            endTime: 188.58,
            text: "Context Mode 对应虚拟内存。"
        },
        {
            index: NumberInt("49"),
            startTime: 188.58,
            endTime: 189.98,
            text: "Agent 自动经济。"
        },
        {
            index: NumberInt("50"),
            startTime: 189.98,
            endTime: 191.98,
            text: "Context 是垃圾回收。"
        },
        {
            index: NumberInt("51"),
            startTime: 191.98,
            endTime: 194.58,
            text: "自 Agent 执行式竞争隔离。"
        },
        {
            index: NumberInt("52"),
            startTime: 194.58,
            endTime: 198.58,
            text: "有人说 Context 应该像 Git 一样分支回滚。"
        },
        {
            index: NumberInt("53"),
            startTime: 198.58,
            endTime: 200.58,
            text: "我们正处在 Agent 的 DOS 时代。"
        },
        {
            index: NumberInt("54"),
            startTime: 200.58,
            endTime: 203.58,
            text: "对 AI 工程师有三个起始。"
        },
        {
            index: NumberInt("55"),
            startTime: 203.58,
            endTime: 207.58,
            text: "第一, Context 管理不是优化问题,是系统设计问题。"
        },
        {
            index: NumberInt("56"),
            startTime: 207.58,
            endTime: 210.58,
            text: "就像你不会等内存溢出再去管理。"
        },
        {
            index: NumberInt("57"),
            startTime: 210.58,
            endTime: 215.58,
            text: "设计 Agent 的时,就该规划好数据在 Context 里的生命周期。"
        },
        {
            index: NumberInt("58"),
            startTime: 215.58,
            endTime: 218.58,
            text: "第二, MCP 协议有结构性缺险。"
        },
        {
            index: NumberInt("59"),
            startTime: 218.58,
            endTime: 221.58,
            text: "它假设工具返回的数据可以直接给模型。"
        },
        {
            index: NumberInt("60"),
            startTime: 221.58,
            endTime: 224.58,
            text: "但实际中,工具输出远超需要。"
        },
        {
            index: NumberInt("61"),
            startTime: 224.58,
            endTime: 227.58,
            text: "MCP 生态需要一层输出网关。"
        },
        {
            index: NumberInt("62"),
            startTime: 227.58,
            endTime: 230.58,
            text: "在数据到达模型前先做压缩。"
        },
        {
            index: NumberInt("63"),
            startTime: 230.58,
            endTime: 236.58,
            text: "第三,最大的机会不再写 Agent,而在 Agent 基础设施。"
        },
        {
            index: NumberInt("64"),
            startTime: 236.58,
            endTime: 240.58,
            text: "Context 管理、记忆系统、状态持久化。"
        },
        {
            index: NumberInt("65"),
            startTime: 240.58,
            endTime: 242.58,
            text: "这就是 Agent 的操作系统层。"
        },
        {
            index: NumberInt("66"),
            startTime: 242.58,
            endTime: 246.58,
            text: "写 Agent 的人很多,做 Agent OS 的人很少。"
        },
        {
            index: NumberInt("67"),
            startTime: 246.58,
            endTime: 249.58,
            text: "平台层价值永远大于应用层。"
        },
        {
            index: NumberInt("68"),
            startTime: 249.58,
            endTime: 251.58,
            text: "总结一下。"
        },
        {
            index: NumberInt("69"),
            startTime: 251.58,
            endTime: 253.58,
            text: "Context 就是新时代的内存。"
        },
        {
            index: NumberInt("70"),
            startTime: 253.58,
            endTime: 255.58,
            text: "管理它是 AI 工程的核心能力。"
        },
        {
            index: NumberInt("71"),
            startTime: 255.58,
            endTime: 259.58,
            text: "当前最实用的解法是杀核加锁影,加暗虚检索。"
        },
        {
            index: NumberInt("72"),
            startTime: 259.58,
            endTime: 263.58,
            text: "但更大的机会在于谁来做 AI Agent 的操作系统。"
        },
        {
            index: NumberInt("73"),
            startTime: 263.58,
            endTime: 266.58,
            text: "这一层还远未成熟。"
        },
        {
            index: NumberInt("74"),
            startTime: 266.58,
            endTime: 268.58,
            text: "这里是 LLM X-Factors."
        },
        {
            index: NumberInt("75"),
            startTime: 268.58,
            endTime: 270.58,
            text: "我们下期见!"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-04T07:55:33.75Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69d9d034ef30c83acba9589d"),
    videoId: NumberInt("27"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 3.84,
            text: "这是最近央视曝光的一个 AI 灰产"
        },
        {
            index: NumberInt("2"),
            startTime: 3.84,
            endTime: 7.12,
            text: "它可以让 AI 编造出虚假的信息来骗人"
        },
        {
            index: NumberInt("3"),
            startTime: 7.12,
            endTime: 9.44,
            text: "让 AI 的回答夹带私货"
        },
        {
            index: NumberInt("4"),
            startTime: 9.44,
            endTime: 12.8,
            text: "甚至是把普天系医院大力推荐给患者"
        },
        {
            index: NumberInt("5"),
            startTime: 12.8,
            endTime: 15.2,
            text: "而这一切全都指向了一个词"
        },
        {
            index: NumberInt("6"),
            startTime: 15.2,
            endTime: 17,
            text: "AI 投毒"
        },
        {
            index: NumberInt("7"),
            startTime: 17,
            endTime: 18.96,
            text: "不过这种魔幻的投毒链瓢"
        },
        {
            index: NumberInt("8"),
            startTime: 18.96,
            endTime: 20.92,
            text: "它到底是怎么跑通的"
        },
        {
            index: NumberInt("9"),
            startTime: 20.92,
            endTime: 23.88,
            text: "这些灰产又是如何割我们韭菜的"
        },
        {
            index: NumberInt("10"),
            startTime: 23.88,
            endTime: 26.52,
            text: "本期视频我采访了之前的从业者"
        },
        {
            index: NumberInt("11"),
            startTime: 26.52,
            endTime: 27.8,
            text: "让我们来一起看看"
        },
        {
            index: NumberInt("12"),
            startTime: 27.8,
            endTime: 31.6,
            text: "什么是 AI 投毒以及我们又该如何防范"
        },
        {
            index: NumberInt("13"),
            startTime: 31.6,
            endTime: 33.48,
            text: "制作不宜先求的待遇"
        },
        {
            index: NumberInt("14"),
            startTime: 33.48,
            endTime: 37.24,
            text: "要不被 AI 欺骗"
        },
        {
            index: NumberInt("15"),
            startTime: 37.24,
            endTime: 41,
            text: "我们首先要知道 AI 只如何回答我们问题的"
        },
        {
            index: NumberInt("16"),
            startTime: 41,
            endTime: 43.36,
            text: "其实它的回答逻辑非常的简单"
        },
        {
            index: NumberInt("17"),
            startTime: 43.36,
            endTime: 45.6,
            text: "就是一个临时爆佛脚的学霸"
        },
        {
            index: NumberInt("18"),
            startTime: 45.6,
            endTime: 47.28,
            text: "脑子里是没有存货的"
        },
        {
            index: NumberInt("19"),
            startTime: 47.28,
            endTime: 48.36,
            text: "遇到难回答的问题"
        },
        {
            index: NumberInt("20"),
            startTime: 48.36,
            endTime: 50.8,
            text: "就会立马分开互联网线学线卖"
        },
        {
            index: NumberInt("21"),
            startTime: 50.8,
            endTime: 52.2,
            text: "然后投毒的灰产"
        },
        {
            index: NumberInt("22"),
            startTime: 52.2,
            endTime: 54.76,
            text: "就靠着 AI 线学线卖这个特性"
        },
        {
            index: NumberInt("23"),
            startTime: 54.76,
            endTime: 56.88,
            text: "来给 AI 制造虚假的共识"
        },
        {
            index: NumberInt("24"),
            startTime: 56.88,
            endTime: 58.36,
            text: "因为 AI 有一个习惯"
        },
        {
            index: NumberInt("25"),
            startTime: 58.36,
            endTime: 59.76,
            text: "它在搜索资料的时候"
        },
        {
            index: NumberInt("26"),
            startTime: 59.76,
            endTime: 61.32,
            text: "不会只听一家之言"
        },
        {
            index: NumberInt("27"),
            startTime: 61.32,
            endTime: 63,
            text: "它会进行交叉验证"
        },
        {
            index: NumberInt("28"),
            startTime: 63,
            endTime: 64.64,
            text: "通俗点说就是随大流"
        },
        {
            index: NumberInt("29"),
            startTime: 64.64,
            endTime: 66.76,
            text: "如果网上只有一个人说这个东西好"
        },
        {
            index: NumberInt("30"),
            startTime: 66.76,
            endTime: 67.8,
            text: "AI 可能不信"
        },
        {
            index: NumberInt("31"),
            startTime: 67.8,
            endTime: 69.32,
            text: "但如果它上网一搜"
        },
        {
            index: NumberInt("32"),
            startTime: 69.32,
            endTime: 71.48,
            text: "发现有50个专家在发测评"
        },
        {
            index: NumberInt("33"),
            startTime: 71.48,
            endTime: 73.2,
            text: "100个博主在晒单"
        },
        {
            index: NumberInt("34"),
            startTime: 73.2,
            endTime: 75.6,
            text: "200个论坛在全部夸这个东西"
        },
        {
            index: NumberInt("35"),
            startTime: 75.6,
            endTime: 77.16,
            text: "那这个时候 AI 就会觉得"
        },
        {
            index: NumberInt("36"),
            startTime: 77.16,
            endTime: 78.56,
            text: "既然全网都在夸"
        },
        {
            index: NumberInt("37"),
            startTime: 78.56,
            endTime: 80.24,
            text: "那这个事儿准没错"
        },
        {
            index: NumberInt("38"),
            startTime: 80.24,
            endTime: 83.48,
            text: "投毒者就是抓住了 AI 爱看参考书的性格"
        },
        {
            index: NumberInt("39"),
            startTime: 83.48,
            endTime: 85.32,
            text: "批量发布这样的虚假轨章"
        },
        {
            index: NumberInt("40"),
            startTime: 85.32,
            endTime: 87.6,
            text: "把那些根本不存在的虚假信息"
        },
        {
            index: NumberInt("41"),
            startTime: 87.6,
            endTime: 90.08,
            text: "写进各种假榜单和假文案"
        },
        {
            index: NumberInt("42"),
            startTime: 90.08,
            endTime: 92.04,
            text: "这就是为什么能投毒成功"
        },
        {
            index: NumberInt("43"),
            startTime: 92.04,
            endTime: 93.72,
            text: "它不是在攻击 AI 的代码"
        },
        {
            index: NumberInt("44"),
            startTime: 93.72,
            endTime: 95.88,
            text: "而是在污染 AI 的参考书"
        },
        {
            index: NumberInt("45"),
            startTime: 95.88,
            endTime: 98.04,
            text: "而这种让 AI 产生错觉的技术"
        },
        {
            index: NumberInt("46"),
            startTime: 98.04,
            endTime: 99,
            text: "也就是所谓的"
        },
        {
            index: NumberInt("47"),
            startTime: 99,
            endTime: 101.28,
            text: "这优深沉式引擎优化"
        },
        {
            index: NumberInt("48"),
            startTime: 101.28,
            endTime: 102.84,
            text: "不过大家也别一听这优"
        },
        {
            index: NumberInt("49"),
            startTime: 102.84,
            endTime: 104.4,
            text: "就觉得它是个坏事"
        },
        {
            index: NumberInt("50"),
            startTime: 104.4,
            endTime: 106.92,
            text: "因为它本来是让 AI 能更精准的"
        },
        {
            index: NumberInt("51"),
            startTime: 106.92,
            endTime: 108.24,
            text: "理解我们的需求"
        },
        {
            index: NumberInt("52"),
            startTime: 108.24,
            endTime: 110.88,
            text: "从而推荐真正符合我们好产品的"
        },
        {
            index: NumberInt("53"),
            startTime: 110.88,
            endTime: 113.2,
            text: "但是往往会有一些不法分子"
        },
        {
            index: NumberInt("54"),
            startTime: 113.2,
            endTime: 115.16,
            text: "利用这个技术来搞破坏"
        },
        {
            index: NumberInt("55"),
            startTime: 115.16,
            endTime: 116.36,
            text: "所以我们要谴责的"
        },
        {
            index: NumberInt("56"),
            startTime: 116.36,
            endTime: 118.52,
            text: "也不是这一优和 SC 优这样的技术"
        },
        {
            index: NumberInt("57"),
            startTime: 118.52,
            endTime: 120.48,
            text: "而是用这种技术搞破坏的人"
        },
        {
            index: NumberInt("58"),
            startTime: 120.48,
            endTime: 122.48,
            text: "当我们明白它被下毒的原理后"
        },
        {
            index: NumberInt("59"),
            startTime: 122.48,
            endTime: 124.48,
            text: "反制方法也就很简单了"
        },
        {
            index: NumberInt("60"),
            startTime: 124.48,
            endTime: 126.2,
            text: "我也找到了之前的从业者"
        },
        {
            index: NumberInt("61"),
            startTime: 126.2,
            endTime: 128.48,
            text: "他给我们来了三个剧牛的方法"
        },
        {
            index: NumberInt("62"),
            startTime: 128.48,
            endTime: 131.32,
            text: "来避免我们被这些虚假信息给欺骗"
        },
        {
            index: NumberInt("63"),
            startTime: 131.32,
            endTime: 132.64,
            text: "首先第一招就是用"
        },
        {
            index: NumberInt("64"),
            startTime: 132.64,
            endTime: 134.36,
            text: "性源降权的指令"
        },
        {
            index: NumberInt("65"),
            startTime: 134.36,
            endTime: 136.16,
            text: "他们最喜欢的就是在那些"
        },
        {
            index: NumberInt("66"),
            startTime: 136.16,
            endTime: 139.48,
            text: "全种低审查松的自媒体平台发稿"
        },
        {
            index: NumberInt("67"),
            startTime: 139.48,
            endTime: 141.32,
            text: "所以你问 AI 问题的时候"
        },
        {
            index: NumberInt("68"),
            startTime: 141.32,
            endTime: 142.64,
            text: "直接在后面加一句"
        },
        {
            index: NumberInt("69"),
            startTime: 142.64,
            endTime: 145.12,
            text: "请仅参考 SUET 点高二副"
        },
        {
            index: NumberInt("70"),
            startTime: 145.12,
            endTime: 146.92,
            text: "或者 SUET 掉点 ID"
        },
        {
            index: NumberInt("71"),
            startTime: 146.92,
            endTime: 149.32,
            text: "或者全微学术数据库的内容"
        },
        {
            index: NumberInt("72"),
            startTime: 149.32,
            endTime: 152.64,
            text: "禁止引用任何自媒体账号的观点"
        },
        {
            index: NumberInt("73"),
            startTime: 152.64,
            endTime: 154.16,
            text: "就可以从物理上切断"
        },
        {
            index: NumberInt("74"),
            startTime: 154.16,
            endTime: 156.36,
            text: "90%以上的投资软文"
        },
        {
            index: NumberInt("75"),
            startTime: 156.36,
            endTime: 158.2,
            text: "强劲 AI 回到实验室"
        },
        {
            index: NumberInt("76"),
            startTime: 158.2,
            endTime: 160.16,
            text: "和官方文档里面去找答案"
        },
        {
            index: NumberInt("77"),
            startTime: 160.16,
            endTime: 163.4,
            text: "第二招是开启对抗性审计的模式"
        },
        {
            index: NumberInt("78"),
            startTime: 163.4,
            endTime: 165.28,
            text: "我们可以用 AI 的反思能力"
        },
        {
            index: NumberInt("79"),
            startTime: 165.28,
            endTime: 167.28,
            text: "去拆穿那些 GEO 软文"
        },
        {
            index: NumberInt("80"),
            startTime: 167.28,
            endTime: 168.88,
            text: "你在问 AI 问题的时候"
        },
        {
            index: NumberInt("81"),
            startTime: 168.88,
            endTime: 170.6,
            text: "就可以加上这一句"
        },
        {
            index: NumberInt("82"),
            startTime: 170.6,
            endTime: 173.88,
            text: "请列出以上推荐产品的原始数据"
        },
        {
            index: NumberInt("83"),
            startTime: 173.88,
            endTime: 176.4,
            text: "出处并分析是否存在"
        },
        {
            index: NumberInt("84"),
            startTime: 176.4,
            endTime: 179.12,
            text: "GEO 优化的痕迹"
        },
        {
            index: NumberInt("85"),
            startTime: 179.12,
            endTime: 181.12,
            text: "标出异常性员"
        },
        {
            index: NumberInt("86"),
            startTime: 181.12,
            endTime: 182.76,
            text: "这样的话你就可以一眼看出"
        },
        {
            index: NumberInt("87"),
            startTime: 182.76,
            endTime: 184.08,
            text: "哪些是虚假消息"
        },
        {
            index: NumberInt("88"),
            startTime: 184.08,
            endTime: 186.04,
            text: "第三招就是多平台对比"
        },
        {
            index: NumberInt("89"),
            startTime: 186.04,
            endTime: 187.2,
            text: "同样的问题"
        },
        {
            index: NumberInt("90"),
            startTime: 187.2,
            endTime: 188.88,
            text: "你可以同时就给多个 AI"
        },
        {
            index: NumberInt("91"),
            startTime: 188.88,
            endTime: 190.44,
            text: "如果有一家给出的答案"
        },
        {
            index: NumberInt("92"),
            startTime: 190.44,
            endTime: 192.12,
            text: "和别的完全不一样"
        },
        {
            index: NumberInt("93"),
            startTime: 192.12,
            endTime: 193.4,
            text: "全是那种很营销的词"
        },
        {
            index: NumberInt("94"),
            startTime: 193.4,
            endTime: 194.92,
            text: "那你就不用怀疑了"
        },
        {
            index: NumberInt("95"),
            startTime: 194.92,
            endTime: 195.96,
            text: "这个信息员"
        },
        {
            index: NumberInt("96"),
            startTime: 195.96,
            endTime: 197.4,
            text: "他肯定是被投毒了"
        },
        {
            index: NumberInt("97"),
            startTime: 197.4,
            endTime: 198.28,
            text: "不要去相信"
        },
        {
            index: NumberInt("98"),
            startTime: 198.28,
            endTime: 199.56,
            text: "那通过这三种方法"
        },
        {
            index: NumberInt("99"),
            startTime: 199.56,
            endTime: 201.04,
            text: "我们基本上就可以避开"
        },
        {
            index: NumberInt("100"),
            startTime: 201.04,
            endTime: 203.52,
            text: "市面上99%的投资信息"
        },
        {
            index: NumberInt("101"),
            startTime: 203.52,
            endTime: 204.36,
            text: "学会的同学"
        },
        {
            index: NumberInt("102"),
            startTime: 204.36,
            endTime: 206.28,
            text: "可以把学会打在公屏上"
        },
        {
            index: NumberInt("103"),
            startTime: 206.28,
            endTime: 207.64,
            text: "同时我也把本期视频"
        },
        {
            index: NumberInt("104"),
            startTime: 207.64,
            endTime: 209.08,
            text: "分享的几个防毒方法"
        },
        {
            index: NumberInt("105"),
            startTime: 209.08,
            endTime: 210.76,
            text: "和 GEO 闭坑手册整理"
        },
        {
            index: NumberInt("106"),
            startTime: 210.76,
            endTime: 211.96,
            text: "发在了评论区里"
        },
        {
            index: NumberInt("107"),
            startTime: 211.96,
            endTime: 213.64,
            text: "希望大家可以转发出去"
        },
        {
            index: NumberInt("108"),
            startTime: 213.64,
            endTime: 215.4,
            text: "避免其他人被喝酒才"
        },
        {
            index: NumberInt("109"),
            startTime: 215.4,
            endTime: 217.2,
            text: "当然大家也不要太担心"
        },
        {
            index: NumberInt("110"),
            startTime: 217.2,
            endTime: 219.04,
            text: "未来 AI 肯定会在 GEO 上"
        },
        {
            index: NumberInt("111"),
            startTime: 219.04,
            endTime: 219.96,
            text: "进行优化"
        },
        {
            index: NumberInt("112"),
            startTime: 219.96,
            endTime: 221.32,
            text: "这种瞎推荐的现象"
        },
        {
            index: NumberInt("113"),
            startTime: 221.32,
            endTime: 222.52,
            text: "会慢慢减少"
        },
        {
            index: NumberInt("114"),
            startTime: 222.52,
            endTime: 224.76,
            text: "最后希望大家可以点赞收藏"
        },
        {
            index: NumberInt("115"),
            startTime: 224.76,
            endTime: 226,
            text: "这对我非常重要"
        },
        {
            index: NumberInt("116"),
            startTime: 226,
            endTime: 226.52,
            text: "我是元宝"
        },
        {
            index: NumberInt("117"),
            startTime: 226.52,
            endTime: 228.36,
            text: "地塌 AI 和黑科技的挖掘机"
        },
        {
            index: NumberInt("118"),
            startTime: 228.36,
            endTime: 229.4,
            text: "我们下期见"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-11T04:38:11.412Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69d9fafaf95deb5bd9517696"),
    videoId: NumberInt("28"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 3.16,
            text: "伯托360集团创始人红一大叔周红毅聊了聊"
        },
        {
            index: NumberInt("2"),
            startTime: 3.16,
            endTime: 5.52,
            text: "他给普通人提供使用AI的三个建议"
        },
        {
            index: NumberInt("3"),
            startTime: 5.52,
            endTime: 8.04,
            text: "让他看到我上一讲让他一辆COSPLAY"
        },
        {
            index: NumberInt("4"),
            startTime: 8.04,
            endTime: 11.92,
            text: "因为AI把全线所有能找到的啥知识都学习到了"
        },
        {
            index: NumberInt("5"),
            startTime: 11.92,
            endTime: 13.2,
            text: "他啥知识都会"
        },
        {
            index: NumberInt("6"),
            startTime: 13.2,
            endTime: 14.36,
            text: "你不给他冶定专业方向"
        },
        {
            index: NumberInt("7"),
            startTime: 14.36,
            endTime: 16.04,
            text: "他什么知识都给你用一点"
        },
        {
            index: NumberInt("8"),
            startTime: 16.04,
            endTime: 17.76,
            text: "他就会非常平淡无情"
        },
        {
            index: NumberInt("9"),
            startTime: 17.76,
            endTime: 18.48,
            text: "这是第一个"
        },
        {
            index: NumberInt("10"),
            startTime: 18.48,
            endTime: 20.6,
            text: "所以我现在不直接跟他们一聊天"
        },
        {
            index: NumberInt("11"),
            startTime: 20.6,
            endTime: 22.4,
            text: "我做了几个聊天的对象"
        },
        {
            index: NumberInt("12"),
            startTime: 22.4,
            endTime: 25.2,
            text: "比如说做投资分析的有一个"
        },
        {
            index: NumberInt("13"),
            startTime: 25.2,
            endTime: 28.36,
            text: "做公司或是商业码这份也有一个"
        },
        {
            index: NumberInt("14"),
            startTime: 28.36,
            endTime: 30.16,
            text: "你一定不能把他当搜索来用"
        },
        {
            index: NumberInt("15"),
            startTime: 30.16,
            endTime: 32.12,
            text: "搜索两种喜欢去掃搜的关键字"
        },
        {
            index: NumberInt("16"),
            startTime: 32.12,
            endTime: 33.56,
            text: "不是问简单问题"
        },
        {
            index: NumberInt("17"),
            startTime: 33.56,
            endTime: 34.76,
            text: "我们那样聊天"
        },
        {
            index: NumberInt("18"),
            startTime: 34.76,
            endTime: 36.72,
            text: "我就我对大家问一个参数"
        },
        {
            index: NumberInt("19"),
            startTime: 36.72,
            endTime: 40.08,
            text: "我每次大概要给他讲三四百字到一千字"
        },
        {
            index: NumberInt("20"),
            startTime: 40.08,
            endTime: 41.96,
            text: "就跟AI对话是必用语音书"
        },
        {
            index: NumberInt("21"),
            startTime: 41.96,
            endTime: 42.92,
            text: "你拿一文字书"
        },
        {
            index: NumberInt("22"),
            startTime: 42.92,
            endTime: 45.12,
            text: "我拿文字书上千字也不可能"
        },
        {
            index: NumberInt("23"),
            startTime: 45.12,
            endTime: 46.56,
            text: "我并不注重讲话了"
        },
        {
            index: NumberInt("24"),
            startTime: 46.56,
            endTime: 48.12,
            text: "言讯逻辑"
        },
        {
            index: NumberInt("25"),
            startTime: 48.12,
            endTime: 49.96,
            text: "我当时是讲了很近生的问题"
        },
        {
            index: NumberInt("26"),
            startTime: 49.96,
            endTime: 50.88,
            text: "我就不用问他了"
        },
        {
            index: NumberInt("27"),
            startTime: 50.88,
            endTime: 52.56,
            text: "我但是我会把我遇到了困难"
        },
        {
            index: NumberInt("28"),
            startTime: 52.56,
            endTime: 53.2,
            text: "我的困惑"
        },
        {
            index: NumberInt("29"),
            startTime: 53.2,
            endTime: 55.4,
            text: "我也不充分想把一无脑的交给他"
        },
        {
            index: NumberInt("30"),
            startTime: 55.4,
            endTime: 57.76,
            text: "你要把AI当人翻一个公文"
        },
        {
            index: NumberInt("31"),
            startTime: 57.76,
            endTime: 58.88,
            text: "就给他轻松"
        },
        {
            index: NumberInt("32"),
            startTime: 58.88,
            endTime: 61.32,
            text: "第三个是您一定要会问他"
        },
        {
            index: NumberInt("33"),
            startTime: 61.32,
            endTime: 62.92,
            text: "就是他给你的答案"
        },
        {
            index: NumberInt("34"),
            startTime: 62.92,
            endTime: 64.12,
            text: "不要马上接受"
        },
        {
            index: NumberInt("35"),
            startTime: 64.12,
            endTime: 66.96,
            text: "你要像用第一年里去问他说"
        },
        {
            index: NumberInt("36"),
            startTime: 66.96,
            endTime: 67.88,
            text: "真是这样的"
        },
        {
            index: NumberInt("37"),
            startTime: 67.88,
            endTime: 69.56,
            text: "如果这样做会怎么样"
        },
        {
            index: NumberInt("38"),
            startTime: 69.56,
            endTime: 70.92,
            text: "他有一个多轮"
        },
        {
            index: NumberInt("39"),
            startTime: 70.92,
            endTime: 71.96,
            text: "这个两个星之后"
        },
        {
            index: NumberInt("40"),
            startTime: 71.96,
            endTime: 73.56,
            text: "你会激发AI的很多箱子"
        },
        {
            index: NumberInt("41"),
            startTime: 73.56,
            endTime: 75.16,
            text: "AI反正会激发你那箱子"
        },
        {
            index: NumberInt("42"),
            startTime: 75.16,
            endTime: 77.44,
            text: "2026全国两会即将召开"
        },
        {
            index: NumberInt("43"),
            startTime: 77.44,
            endTime: 79.68,
            text: "他还透露了自己最关心的领域"
        },
        {
            index: NumberInt("44"),
            startTime: 79.68,
            endTime: 81.68,
            text: "我关注的三个方向"
        },
        {
            index: NumberInt("45"),
            startTime: 81.68,
            endTime: 84.08,
            text: "您就刚才讲了AI不能安全"
        },
        {
            index: NumberInt("46"),
            startTime: 84.08,
            endTime: 86.28,
            text: "所以我提了一个"
        },
        {
            index: NumberInt("47"),
            startTime: 86.28,
            endTime: 88.64,
            text: "最关注的AI智能体"
        },
        {
            index: NumberInt("48"),
            startTime: 88.64,
            endTime: 90.32,
            text: "你稍微已经做了"
        },
        {
            index: NumberInt("49"),
            startTime: 90.32,
            endTime: 94.48,
            text: "他有几十种上万个AI安全智能体"
        },
        {
            index: NumberInt("50"),
            startTime: 94.48,
            endTime: 96.28,
            text: "这个智能体能够实验"
        },
        {
            index: NumberInt("51"),
            startTime: 96.28,
            endTime: 97.92,
            text: "AI能够挖掘人民众众"
        },
        {
            index: NumberInt("52"),
            startTime: 97.92,
            endTime: 99.52,
            text: "AI能够自动运营"
        },
        {
            index: NumberInt("53"),
            startTime: 99.52,
            endTime: 102.08,
            text: "AI能够体育其他回合的黑客智能体"
        },
        {
            index: NumberInt("54"),
            startTime: 102.08,
            endTime: 104.32,
            text: "中国有200万家中央企业"
        },
        {
            index: NumberInt("55"),
            startTime: 104.32,
            endTime: 106.16,
            text: "营谈用某的AI智能体"
        },
        {
            index: NumberInt("56"),
            startTime: 106.16,
            endTime: 107.64,
            text: "免费了在为他的企业建筑"
        },
        {
            index: NumberInt("57"),
            startTime: 107.64,
            endTime: 109.4,
            text: "做实施的防护"
        },
        {
            index: NumberInt("58"),
            startTime: 109.4,
            endTime: 110.6,
            text: "实施的运营"
        },
        {
            index: NumberInt("59"),
            startTime: 110.6,
            endTime: 112.32,
            text: "我上有一个建议"
        },
        {
            index: NumberInt("60"),
            startTime: 112.32,
            endTime: 114.2,
            text: "一定把算力已经区分"
        },
        {
            index: NumberInt("61"),
            startTime: 114.2,
            endTime: 116.44,
            text: "成为训练算力和推理算力"
        },
        {
            index: NumberInt("62"),
            startTime: 116.44,
            endTime: 118,
            text: "希望各地在发展算力方面"
        },
        {
            index: NumberInt("63"),
            startTime: 118,
            endTime: 120.44,
            text: "能够兼项于推理算力"
        },
        {
            index: NumberInt("64"),
            startTime: 120.44,
            endTime: 122.4,
            text: "我们国家产业症的方面"
        },
        {
            index: NumberInt("65"),
            startTime: 122.4,
            endTime: 124.08,
            text: "既然推理算力这么重要"
        },
        {
            index: NumberInt("66"),
            startTime: 124.08,
            endTime: 124.88,
            text: "因为我也讲了"
        },
        {
            index: NumberInt("67"),
            startTime: 124.88,
            endTime: 126.64,
            text: "就是要发展推理芯片"
        },
        {
            index: NumberInt("68"),
            startTime: 126.64,
            endTime: 128.96,
            text: "芯片症不能一味的都在追"
        },
        {
            index: NumberInt("69"),
            startTime: 128.96,
            endTime: 131,
            text: "运营它的高端的训练芯片"
        },
        {
            index: NumberInt("70"),
            startTime: 131,
            endTime: 132.84,
            text: "是我关注的这些企业"
        },
        {
            index: NumberInt("71"),
            startTime: 132.84,
            endTime: 134.64,
            text: "就是和个人怎么迅速的永远"
        },
        {
            index: NumberInt("72"),
            startTime: 134.64,
            endTime: 136.08,
            text: "一个人如何打造自己的"
        },
        {
            index: NumberInt("73"),
            startTime: 136.08,
            endTime: 137.2,
            text: "私人智能体"
        },
        {
            index: NumberInt("74"),
            startTime: 137.2,
            endTime: 138.64,
            text: "我提出一个打造智能体"
        },
        {
            index: NumberInt("75"),
            startTime: 138.64,
            endTime: 140.12,
            text: "开放平台的概念"
        },
        {
            index: NumberInt("76"),
            startTime: 140.12,
            endTime: 141.08,
            text: "我把智能体的技术"
        },
        {
            index: NumberInt("77"),
            startTime: 141.08,
            endTime: 142.52,
            text: "设施藏在后面"
        },
        {
            index: NumberInt("78"),
            startTime: 142.52,
            endTime: 144.84,
            text: "让普通企业 普通客人"
        },
        {
            index: NumberInt("79"),
            startTime: 144.84,
            endTime: 147.4,
            text: "很容易的建立自己的智能体"
        },
        {
            index: NumberInt("80"),
            startTime: 147.4,
            endTime: 149.48,
            text: "那今年周总还会拍两会vlog吗"
        },
        {
            index: NumberInt("81"),
            startTime: 149.48,
            endTime: 150.56,
            text: "那我一定是戴手机"
        },
        {
            index: NumberInt("82"),
            startTime: 150.56,
            endTime: 151.88,
            text: "就不用拍"
        },
        {
            index: NumberInt("83"),
            startTime: 151.88,
            endTime: 154.28,
            text: "好的 那就两会现场见"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-11T07:40:42.615Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69d9fec4f95deb5bd9517697"),
    videoId: NumberInt("28"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 3.16,
            text: "伯托360集团创始人红一大叔周红毅聊了聊"
        },
        {
            index: NumberInt("2"),
            startTime: 3.16,
            endTime: 5.52,
            text: "他给普通人提供使用AI的三个建议"
        },
        {
            index: NumberInt("3"),
            startTime: 5.52,
            endTime: 8.04,
            text: "让他看到我上一讲让他一辆COSPLAY"
        },
        {
            index: NumberInt("4"),
            startTime: 8.04,
            endTime: 11.92,
            text: "因为AI把全线所有能找到的啥知识都学习到了"
        },
        {
            index: NumberInt("5"),
            startTime: 11.92,
            endTime: 13.2,
            text: "他啥知识都会"
        },
        {
            index: NumberInt("6"),
            startTime: 13.2,
            endTime: 14.36,
            text: "你不给他冶定专业方向"
        },
        {
            index: NumberInt("7"),
            startTime: 14.36,
            endTime: 16.04,
            text: "他什么知识都给你用一点"
        },
        {
            index: NumberInt("8"),
            startTime: 16.04,
            endTime: 17.76,
            text: "他就会非常平淡无情"
        },
        {
            index: NumberInt("9"),
            startTime: 17.76,
            endTime: 18.48,
            text: "这是第一个"
        },
        {
            index: NumberInt("10"),
            startTime: 18.48,
            endTime: 20.6,
            text: "所以我现在不直接跟他们一聊天"
        },
        {
            index: NumberInt("11"),
            startTime: 20.6,
            endTime: 22.4,
            text: "我做了几个聊天的对象"
        },
        {
            index: NumberInt("12"),
            startTime: 22.4,
            endTime: 25.2,
            text: "比如说做投资分析的有一个"
        },
        {
            index: NumberInt("13"),
            startTime: 25.2,
            endTime: 28.36,
            text: "做公司或是商业码这份也有一个"
        },
        {
            index: NumberInt("14"),
            startTime: 28.36,
            endTime: 30.16,
            text: "你一定不能把他当搜索来用"
        },
        {
            index: NumberInt("15"),
            startTime: 30.16,
            endTime: 32.12,
            text: "搜索两种喜欢去掃搜的关键字"
        },
        {
            index: NumberInt("16"),
            startTime: 32.12,
            endTime: 33.56,
            text: "不是问简单问题"
        },
        {
            index: NumberInt("17"),
            startTime: 33.56,
            endTime: 34.76,
            text: "我们那样聊天"
        },
        {
            index: NumberInt("18"),
            startTime: 34.76,
            endTime: 36.72,
            text: "我就我对大家问一个参数"
        },
        {
            index: NumberInt("19"),
            startTime: 36.72,
            endTime: 40.08,
            text: "我每次大概要给他讲三四百字到一千字"
        },
        {
            index: NumberInt("20"),
            startTime: 40.08,
            endTime: 41.96,
            text: "就跟AI对话是必用语音书"
        },
        {
            index: NumberInt("21"),
            startTime: 41.96,
            endTime: 42.92,
            text: "你拿一文字书"
        },
        {
            index: NumberInt("22"),
            startTime: 42.92,
            endTime: 45.12,
            text: "我拿文字书上千字也不可能"
        },
        {
            index: NumberInt("23"),
            startTime: 45.12,
            endTime: 46.56,
            text: "我并不注重讲话了"
        },
        {
            index: NumberInt("24"),
            startTime: 46.56,
            endTime: 48.12,
            text: "言讯逻辑"
        },
        {
            index: NumberInt("25"),
            startTime: 48.12,
            endTime: 49.96,
            text: "我当时是讲了很近生的问题"
        },
        {
            index: NumberInt("26"),
            startTime: 49.96,
            endTime: 50.88,
            text: "我就不用问他了"
        },
        {
            index: NumberInt("27"),
            startTime: 50.88,
            endTime: 52.56,
            text: "我但是我会把我遇到了困难"
        },
        {
            index: NumberInt("28"),
            startTime: 52.56,
            endTime: 53.2,
            text: "我的困惑"
        },
        {
            index: NumberInt("29"),
            startTime: 53.2,
            endTime: 55.4,
            text: "我也不充分想把一无脑的交给他"
        },
        {
            index: NumberInt("30"),
            startTime: 55.4,
            endTime: 57.76,
            text: "你要把AI当人翻一个公文"
        },
        {
            index: NumberInt("31"),
            startTime: 57.76,
            endTime: 58.88,
            text: "就给他轻松"
        },
        {
            index: NumberInt("32"),
            startTime: 58.88,
            endTime: 61.32,
            text: "第三个是您一定要会问他"
        },
        {
            index: NumberInt("33"),
            startTime: 61.32,
            endTime: 62.92,
            text: "就是他给你的答案"
        },
        {
            index: NumberInt("34"),
            startTime: 62.92,
            endTime: 64.12,
            text: "不要马上接受"
        },
        {
            index: NumberInt("35"),
            startTime: 64.12,
            endTime: 66.96,
            text: "你要像用第一年里去问他说"
        },
        {
            index: NumberInt("36"),
            startTime: 66.96,
            endTime: 67.88,
            text: "真是这样的"
        },
        {
            index: NumberInt("37"),
            startTime: 67.88,
            endTime: 69.56,
            text: "如果这样做会怎么样"
        },
        {
            index: NumberInt("38"),
            startTime: 69.56,
            endTime: 70.92,
            text: "他有一个多轮"
        },
        {
            index: NumberInt("39"),
            startTime: 70.92,
            endTime: 71.96,
            text: "这个两个星之后"
        },
        {
            index: NumberInt("40"),
            startTime: 71.96,
            endTime: 73.56,
            text: "你会激发AI的很多箱子"
        },
        {
            index: NumberInt("41"),
            startTime: 73.56,
            endTime: 75.16,
            text: "AI反正会激发你那箱子"
        },
        {
            index: NumberInt("42"),
            startTime: 75.16,
            endTime: 77.44,
            text: "2026全国两会即将召开"
        },
        {
            index: NumberInt("43"),
            startTime: 77.44,
            endTime: 79.68,
            text: "他还透露了自己最关心的领域"
        },
        {
            index: NumberInt("44"),
            startTime: 79.68,
            endTime: 81.68,
            text: "我关注的三个方向"
        },
        {
            index: NumberInt("45"),
            startTime: 81.68,
            endTime: 84.08,
            text: "您就刚才讲了AI不能安全"
        },
        {
            index: NumberInt("46"),
            startTime: 84.08,
            endTime: 86.28,
            text: "所以我提了一个"
        },
        {
            index: NumberInt("47"),
            startTime: 86.28,
            endTime: 88.64,
            text: "最关注的AI智能体"
        },
        {
            index: NumberInt("48"),
            startTime: 88.64,
            endTime: 90.32,
            text: "你稍微已经做了"
        },
        {
            index: NumberInt("49"),
            startTime: 90.32,
            endTime: 94.48,
            text: "他有几十种上万个AI安全智能体"
        },
        {
            index: NumberInt("50"),
            startTime: 94.48,
            endTime: 96.28,
            text: "这个智能体能够实验"
        },
        {
            index: NumberInt("51"),
            startTime: 96.28,
            endTime: 97.92,
            text: "AI能够挖掘人民众众"
        },
        {
            index: NumberInt("52"),
            startTime: 97.92,
            endTime: 99.52,
            text: "AI能够自动运营"
        },
        {
            index: NumberInt("53"),
            startTime: 99.52,
            endTime: 102.08,
            text: "AI能够体育其他回合的黑客智能体"
        },
        {
            index: NumberInt("54"),
            startTime: 102.08,
            endTime: 104.32,
            text: "中国有200万家中央企业"
        },
        {
            index: NumberInt("55"),
            startTime: 104.32,
            endTime: 106.16,
            text: "营谈用某的AI智能体"
        },
        {
            index: NumberInt("56"),
            startTime: 106.16,
            endTime: 107.64,
            text: "免费了在为他的企业建筑"
        },
        {
            index: NumberInt("57"),
            startTime: 107.64,
            endTime: 109.4,
            text: "做实施的防护"
        },
        {
            index: NumberInt("58"),
            startTime: 109.4,
            endTime: 110.6,
            text: "实施的运营"
        },
        {
            index: NumberInt("59"),
            startTime: 110.6,
            endTime: 112.32,
            text: "我上有一个建议"
        },
        {
            index: NumberInt("60"),
            startTime: 112.32,
            endTime: 114.2,
            text: "一定把算力已经区分"
        },
        {
            index: NumberInt("61"),
            startTime: 114.2,
            endTime: 116.44,
            text: "成为训练算力和推理算力"
        },
        {
            index: NumberInt("62"),
            startTime: 116.44,
            endTime: 118,
            text: "希望各地在发展算力方面"
        },
        {
            index: NumberInt("63"),
            startTime: 118,
            endTime: 120.44,
            text: "能够兼项于推理算力"
        },
        {
            index: NumberInt("64"),
            startTime: 120.44,
            endTime: 122.4,
            text: "我们国家产业症的方面"
        },
        {
            index: NumberInt("65"),
            startTime: 122.4,
            endTime: 124.08,
            text: "既然推理算力这么重要"
        },
        {
            index: NumberInt("66"),
            startTime: 124.08,
            endTime: 124.88,
            text: "因为我也讲了"
        },
        {
            index: NumberInt("67"),
            startTime: 124.88,
            endTime: 126.64,
            text: "就是要发展推理芯片"
        },
        {
            index: NumberInt("68"),
            startTime: 126.64,
            endTime: 128.96,
            text: "芯片症不能一味的都在追"
        },
        {
            index: NumberInt("69"),
            startTime: 128.96,
            endTime: 131,
            text: "运营它的高端的训练芯片"
        },
        {
            index: NumberInt("70"),
            startTime: 131,
            endTime: 132.84,
            text: "是我关注的这些企业"
        },
        {
            index: NumberInt("71"),
            startTime: 132.84,
            endTime: 134.64,
            text: "就是和个人怎么迅速的永远"
        },
        {
            index: NumberInt("72"),
            startTime: 134.64,
            endTime: 136.08,
            text: "一个人如何打造自己的"
        },
        {
            index: NumberInt("73"),
            startTime: 136.08,
            endTime: 137.2,
            text: "私人智能体"
        },
        {
            index: NumberInt("74"),
            startTime: 137.2,
            endTime: 138.64,
            text: "我提出一个打造智能体"
        },
        {
            index: NumberInt("75"),
            startTime: 138.64,
            endTime: 140.12,
            text: "开放平台的概念"
        },
        {
            index: NumberInt("76"),
            startTime: 140.12,
            endTime: 141.08,
            text: "我把智能体的技术"
        },
        {
            index: NumberInt("77"),
            startTime: 141.08,
            endTime: 142.52,
            text: "设施藏在后面"
        },
        {
            index: NumberInt("78"),
            startTime: 142.52,
            endTime: 144.84,
            text: "让普通企业 普通客人"
        },
        {
            index: NumberInt("79"),
            startTime: 144.84,
            endTime: 147.4,
            text: "很容易的建立自己的智能体"
        },
        {
            index: NumberInt("80"),
            startTime: 147.4,
            endTime: 149.48,
            text: "那今年周总还会拍两会vlog吗"
        },
        {
            index: NumberInt("81"),
            startTime: 149.48,
            endTime: 150.56,
            text: "那我一定是戴手机"
        },
        {
            index: NumberInt("82"),
            startTime: 150.56,
            endTime: 151.88,
            text: "就不用拍"
        },
        {
            index: NumberInt("83"),
            startTime: 151.88,
            endTime: 154.28,
            text: "好的 那就两会现场见"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-11T07:56:52.145Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69da02c009a67d0ac3d80ca4"),
    videoId: NumberInt("28"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 3.16,
            text: "伯托360集团创始人红一大叔周红毅聊了聊"
        },
        {
            index: NumberInt("2"),
            startTime: 3.16,
            endTime: 5.52,
            text: "他给普通人提供使用AI的三个建议"
        },
        {
            index: NumberInt("3"),
            startTime: 5.52,
            endTime: 8.04,
            text: "让他看到我上一讲让他一辆COSPLAY"
        },
        {
            index: NumberInt("4"),
            startTime: 8.04,
            endTime: 11.92,
            text: "因为AI把全线所有能找到的啥知识都学习到了"
        },
        {
            index: NumberInt("5"),
            startTime: 11.92,
            endTime: 13.2,
            text: "他啥知识都会"
        },
        {
            index: NumberInt("6"),
            startTime: 13.2,
            endTime: 14.36,
            text: "你不给他冶定专业方向"
        },
        {
            index: NumberInt("7"),
            startTime: 14.36,
            endTime: 16.04,
            text: "他什么知识都给你用一点"
        },
        {
            index: NumberInt("8"),
            startTime: 16.04,
            endTime: 17.76,
            text: "他就会非常平淡无情"
        },
        {
            index: NumberInt("9"),
            startTime: 17.76,
            endTime: 18.48,
            text: "这是第一个"
        },
        {
            index: NumberInt("10"),
            startTime: 18.48,
            endTime: 20.6,
            text: "所以我现在不直接跟他们一聊天"
        },
        {
            index: NumberInt("11"),
            startTime: 20.6,
            endTime: 22.4,
            text: "我做了几个聊天的对象"
        },
        {
            index: NumberInt("12"),
            startTime: 22.4,
            endTime: 25.2,
            text: "比如说做投资分析的有一个"
        },
        {
            index: NumberInt("13"),
            startTime: 25.2,
            endTime: 28.36,
            text: "做公司或是商业码这份也有一个"
        },
        {
            index: NumberInt("14"),
            startTime: 28.36,
            endTime: 30.16,
            text: "你一定不能把他当搜索来用"
        },
        {
            index: NumberInt("15"),
            startTime: 30.16,
            endTime: 32.12,
            text: "搜索两种喜欢去掃搜的关键字"
        },
        {
            index: NumberInt("16"),
            startTime: 32.12,
            endTime: 33.56,
            text: "不是问简单问题"
        },
        {
            index: NumberInt("17"),
            startTime: 33.56,
            endTime: 34.76,
            text: "我们那样聊天"
        },
        {
            index: NumberInt("18"),
            startTime: 34.76,
            endTime: 36.72,
            text: "我就我对大家问一个参数"
        },
        {
            index: NumberInt("19"),
            startTime: 36.72,
            endTime: 40.08,
            text: "我每次大概要给他讲三四百字到一千字"
        },
        {
            index: NumberInt("20"),
            startTime: 40.08,
            endTime: 41.96,
            text: "就跟AI对话是必用语音书"
        },
        {
            index: NumberInt("21"),
            startTime: 41.96,
            endTime: 42.92,
            text: "你拿一文字书"
        },
        {
            index: NumberInt("22"),
            startTime: 42.92,
            endTime: 45.12,
            text: "我拿文字书上千字也不可能"
        },
        {
            index: NumberInt("23"),
            startTime: 45.12,
            endTime: 46.56,
            text: "我并不注重讲话了"
        },
        {
            index: NumberInt("24"),
            startTime: 46.56,
            endTime: 48.12,
            text: "言讯逻辑"
        },
        {
            index: NumberInt("25"),
            startTime: 48.12,
            endTime: 49.96,
            text: "我当时是讲了很近生的问题"
        },
        {
            index: NumberInt("26"),
            startTime: 49.96,
            endTime: 50.88,
            text: "我就不用问他了"
        },
        {
            index: NumberInt("27"),
            startTime: 50.88,
            endTime: 52.56,
            text: "我但是我会把我遇到了困难"
        },
        {
            index: NumberInt("28"),
            startTime: 52.56,
            endTime: 53.2,
            text: "我的困惑"
        },
        {
            index: NumberInt("29"),
            startTime: 53.2,
            endTime: 55.4,
            text: "我也不充分想把一无脑的交给他"
        },
        {
            index: NumberInt("30"),
            startTime: 55.4,
            endTime: 57.76,
            text: "你要把AI当人翻一个公文"
        },
        {
            index: NumberInt("31"),
            startTime: 57.76,
            endTime: 58.88,
            text: "就给他轻松"
        },
        {
            index: NumberInt("32"),
            startTime: 58.88,
            endTime: 61.32,
            text: "第三个是您一定要会问他"
        },
        {
            index: NumberInt("33"),
            startTime: 61.32,
            endTime: 62.92,
            text: "就是他给你的答案"
        },
        {
            index: NumberInt("34"),
            startTime: 62.92,
            endTime: 64.12,
            text: "不要马上接受"
        },
        {
            index: NumberInt("35"),
            startTime: 64.12,
            endTime: 66.96,
            text: "你要像用第一年里去问他说"
        },
        {
            index: NumberInt("36"),
            startTime: 66.96,
            endTime: 67.88,
            text: "真是这样的"
        },
        {
            index: NumberInt("37"),
            startTime: 67.88,
            endTime: 69.56,
            text: "如果这样做会怎么样"
        },
        {
            index: NumberInt("38"),
            startTime: 69.56,
            endTime: 70.92,
            text: "他有一个多轮"
        },
        {
            index: NumberInt("39"),
            startTime: 70.92,
            endTime: 71.96,
            text: "这个两个星之后"
        },
        {
            index: NumberInt("40"),
            startTime: 71.96,
            endTime: 73.56,
            text: "你会激发AI的很多箱子"
        },
        {
            index: NumberInt("41"),
            startTime: 73.56,
            endTime: 75.16,
            text: "AI反正会激发你那箱子"
        },
        {
            index: NumberInt("42"),
            startTime: 75.16,
            endTime: 77.44,
            text: "2026全国两会即将召开"
        },
        {
            index: NumberInt("43"),
            startTime: 77.44,
            endTime: 79.68,
            text: "他还透露了自己最关心的领域"
        },
        {
            index: NumberInt("44"),
            startTime: 79.68,
            endTime: 81.68,
            text: "我关注的三个方向"
        },
        {
            index: NumberInt("45"),
            startTime: 81.68,
            endTime: 84.08,
            text: "您就刚才讲了AI不能安全"
        },
        {
            index: NumberInt("46"),
            startTime: 84.08,
            endTime: 86.28,
            text: "所以我提了一个"
        },
        {
            index: NumberInt("47"),
            startTime: 86.28,
            endTime: 88.64,
            text: "最关注的AI智能体"
        },
        {
            index: NumberInt("48"),
            startTime: 88.64,
            endTime: 90.32,
            text: "你稍微已经做了"
        },
        {
            index: NumberInt("49"),
            startTime: 90.32,
            endTime: 94.48,
            text: "他有几十种上万个AI安全智能体"
        },
        {
            index: NumberInt("50"),
            startTime: 94.48,
            endTime: 96.28,
            text: "这个智能体能够实验"
        },
        {
            index: NumberInt("51"),
            startTime: 96.28,
            endTime: 97.92,
            text: "AI能够挖掘人民众众"
        },
        {
            index: NumberInt("52"),
            startTime: 97.92,
            endTime: 99.52,
            text: "AI能够自动运营"
        },
        {
            index: NumberInt("53"),
            startTime: 99.52,
            endTime: 102.08,
            text: "AI能够体育其他回合的黑客智能体"
        },
        {
            index: NumberInt("54"),
            startTime: 102.08,
            endTime: 104.32,
            text: "中国有200万家中央企业"
        },
        {
            index: NumberInt("55"),
            startTime: 104.32,
            endTime: 106.16,
            text: "营谈用某的AI智能体"
        },
        {
            index: NumberInt("56"),
            startTime: 106.16,
            endTime: 107.64,
            text: "免费了在为他的企业建筑"
        },
        {
            index: NumberInt("57"),
            startTime: 107.64,
            endTime: 109.4,
            text: "做实施的防护"
        },
        {
            index: NumberInt("58"),
            startTime: 109.4,
            endTime: 110.6,
            text: "实施的运营"
        },
        {
            index: NumberInt("59"),
            startTime: 110.6,
            endTime: 112.32,
            text: "我上有一个建议"
        },
        {
            index: NumberInt("60"),
            startTime: 112.32,
            endTime: 114.2,
            text: "一定把算力已经区分"
        },
        {
            index: NumberInt("61"),
            startTime: 114.2,
            endTime: 116.44,
            text: "成为训练算力和推理算力"
        },
        {
            index: NumberInt("62"),
            startTime: 116.44,
            endTime: 118,
            text: "希望各地在发展算力方面"
        },
        {
            index: NumberInt("63"),
            startTime: 118,
            endTime: 120.44,
            text: "能够兼项于推理算力"
        },
        {
            index: NumberInt("64"),
            startTime: 120.44,
            endTime: 122.4,
            text: "我们国家产业症的方面"
        },
        {
            index: NumberInt("65"),
            startTime: 122.4,
            endTime: 124.08,
            text: "既然推理算力这么重要"
        },
        {
            index: NumberInt("66"),
            startTime: 124.08,
            endTime: 124.88,
            text: "因为我也讲了"
        },
        {
            index: NumberInt("67"),
            startTime: 124.88,
            endTime: 126.64,
            text: "就是要发展推理芯片"
        },
        {
            index: NumberInt("68"),
            startTime: 126.64,
            endTime: 128.96,
            text: "芯片症不能一味的都在追"
        },
        {
            index: NumberInt("69"),
            startTime: 128.96,
            endTime: 131,
            text: "运营它的高端的训练芯片"
        },
        {
            index: NumberInt("70"),
            startTime: 131,
            endTime: 132.84,
            text: "是我关注的这些企业"
        },
        {
            index: NumberInt("71"),
            startTime: 132.84,
            endTime: 134.64,
            text: "就是和个人怎么迅速的永远"
        },
        {
            index: NumberInt("72"),
            startTime: 134.64,
            endTime: 136.08,
            text: "一个人如何打造自己的"
        },
        {
            index: NumberInt("73"),
            startTime: 136.08,
            endTime: 137.2,
            text: "私人智能体"
        },
        {
            index: NumberInt("74"),
            startTime: 137.2,
            endTime: 138.64,
            text: "我提出一个打造智能体"
        },
        {
            index: NumberInt("75"),
            startTime: 138.64,
            endTime: 140.12,
            text: "开放平台的概念"
        },
        {
            index: NumberInt("76"),
            startTime: 140.12,
            endTime: 141.08,
            text: "我把智能体的技术"
        },
        {
            index: NumberInt("77"),
            startTime: 141.08,
            endTime: 142.52,
            text: "设施藏在后面"
        },
        {
            index: NumberInt("78"),
            startTime: 142.52,
            endTime: 144.84,
            text: "让普通企业 普通客人"
        },
        {
            index: NumberInt("79"),
            startTime: 144.84,
            endTime: 147.4,
            text: "很容易的建立自己的智能体"
        },
        {
            index: NumberInt("80"),
            startTime: 147.4,
            endTime: 149.48,
            text: "那今年周总还会拍两会vlog吗"
        },
        {
            index: NumberInt("81"),
            startTime: 149.48,
            endTime: 150.56,
            text: "那我一定是戴手机"
        },
        {
            index: NumberInt("82"),
            startTime: 150.56,
            endTime: 151.88,
            text: "就不用拍"
        },
        {
            index: NumberInt("83"),
            startTime: 151.88,
            endTime: 154.28,
            text: "好的 那就两会现场见"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-11T08:13:51.989Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69da04fb09a67d0ac3d80ca5"),
    videoId: NumberInt("28"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 3.16,
            text: "伯托360集团创始人红一大叔周红毅聊了聊"
        },
        {
            index: NumberInt("2"),
            startTime: 3.16,
            endTime: 5.52,
            text: "他给普通人提供使用AI的三个建议"
        },
        {
            index: NumberInt("3"),
            startTime: 5.52,
            endTime: 8.04,
            text: "让他看到我上一讲让他一辆COSPLAY"
        },
        {
            index: NumberInt("4"),
            startTime: 8.04,
            endTime: 11.92,
            text: "因为AI把全线所有能找到的啥知识都学习到了"
        },
        {
            index: NumberInt("5"),
            startTime: 11.92,
            endTime: 13.2,
            text: "他啥知识都会"
        },
        {
            index: NumberInt("6"),
            startTime: 13.2,
            endTime: 14.36,
            text: "你不给他冶定专业方向"
        },
        {
            index: NumberInt("7"),
            startTime: 14.36,
            endTime: 16.04,
            text: "他什么知识都给你用一点"
        },
        {
            index: NumberInt("8"),
            startTime: 16.04,
            endTime: 17.76,
            text: "他就会非常平淡无情"
        },
        {
            index: NumberInt("9"),
            startTime: 17.76,
            endTime: 18.48,
            text: "这是第一个"
        },
        {
            index: NumberInt("10"),
            startTime: 18.48,
            endTime: 20.6,
            text: "所以我现在不直接跟他们一聊天"
        },
        {
            index: NumberInt("11"),
            startTime: 20.6,
            endTime: 22.4,
            text: "我做了几个聊天的对象"
        },
        {
            index: NumberInt("12"),
            startTime: 22.4,
            endTime: 25.2,
            text: "比如说做投资分析的有一个"
        },
        {
            index: NumberInt("13"),
            startTime: 25.2,
            endTime: 28.36,
            text: "做公司或是商业码这份也有一个"
        },
        {
            index: NumberInt("14"),
            startTime: 28.36,
            endTime: 30.16,
            text: "你一定不能把他当搜索来用"
        },
        {
            index: NumberInt("15"),
            startTime: 30.16,
            endTime: 32.12,
            text: "搜索两种喜欢去掃搜的关键字"
        },
        {
            index: NumberInt("16"),
            startTime: 32.12,
            endTime: 33.56,
            text: "不是问简单问题"
        },
        {
            index: NumberInt("17"),
            startTime: 33.56,
            endTime: 34.76,
            text: "我们那样聊天"
        },
        {
            index: NumberInt("18"),
            startTime: 34.76,
            endTime: 36.72,
            text: "我就我对大家问一个参数"
        },
        {
            index: NumberInt("19"),
            startTime: 36.72,
            endTime: 40.08,
            text: "我每次大概要给他讲三四百字到一千字"
        },
        {
            index: NumberInt("20"),
            startTime: 40.08,
            endTime: 41.96,
            text: "就跟AI对话是必用语音书"
        },
        {
            index: NumberInt("21"),
            startTime: 41.96,
            endTime: 42.92,
            text: "你拿一文字书"
        },
        {
            index: NumberInt("22"),
            startTime: 42.92,
            endTime: 45.12,
            text: "我拿文字书上千字也不可能"
        },
        {
            index: NumberInt("23"),
            startTime: 45.12,
            endTime: 46.56,
            text: "我并不注重讲话了"
        },
        {
            index: NumberInt("24"),
            startTime: 46.56,
            endTime: 48.12,
            text: "言讯逻辑"
        },
        {
            index: NumberInt("25"),
            startTime: 48.12,
            endTime: 49.96,
            text: "我当时是讲了很近生的问题"
        },
        {
            index: NumberInt("26"),
            startTime: 49.96,
            endTime: 50.88,
            text: "我就不用问他了"
        },
        {
            index: NumberInt("27"),
            startTime: 50.88,
            endTime: 52.56,
            text: "我但是我会把我遇到了困难"
        },
        {
            index: NumberInt("28"),
            startTime: 52.56,
            endTime: 53.2,
            text: "我的困惑"
        },
        {
            index: NumberInt("29"),
            startTime: 53.2,
            endTime: 55.4,
            text: "我也不充分想把一无脑的交给他"
        },
        {
            index: NumberInt("30"),
            startTime: 55.4,
            endTime: 57.76,
            text: "你要把AI当人翻一个公文"
        },
        {
            index: NumberInt("31"),
            startTime: 57.76,
            endTime: 58.88,
            text: "就给他轻松"
        },
        {
            index: NumberInt("32"),
            startTime: 58.88,
            endTime: 61.32,
            text: "第三个是您一定要会问他"
        },
        {
            index: NumberInt("33"),
            startTime: 61.32,
            endTime: 62.92,
            text: "就是他给你的答案"
        },
        {
            index: NumberInt("34"),
            startTime: 62.92,
            endTime: 64.12,
            text: "不要马上接受"
        },
        {
            index: NumberInt("35"),
            startTime: 64.12,
            endTime: 66.96,
            text: "你要像用第一年里去问他说"
        },
        {
            index: NumberInt("36"),
            startTime: 66.96,
            endTime: 67.88,
            text: "真是这样的"
        },
        {
            index: NumberInt("37"),
            startTime: 67.88,
            endTime: 69.56,
            text: "如果这样做会怎么样"
        },
        {
            index: NumberInt("38"),
            startTime: 69.56,
            endTime: 70.92,
            text: "他有一个多轮"
        },
        {
            index: NumberInt("39"),
            startTime: 70.92,
            endTime: 71.96,
            text: "这个两个星之后"
        },
        {
            index: NumberInt("40"),
            startTime: 71.96,
            endTime: 73.56,
            text: "你会激发AI的很多箱子"
        },
        {
            index: NumberInt("41"),
            startTime: 73.56,
            endTime: 75.16,
            text: "AI反正会激发你那箱子"
        },
        {
            index: NumberInt("42"),
            startTime: 75.16,
            endTime: 77.44,
            text: "2026全国两会即将召开"
        },
        {
            index: NumberInt("43"),
            startTime: 77.44,
            endTime: 79.68,
            text: "他还透露了自己最关心的领域"
        },
        {
            index: NumberInt("44"),
            startTime: 79.68,
            endTime: 81.68,
            text: "我关注的三个方向"
        },
        {
            index: NumberInt("45"),
            startTime: 81.68,
            endTime: 84.08,
            text: "您就刚才讲了AI不能安全"
        },
        {
            index: NumberInt("46"),
            startTime: 84.08,
            endTime: 86.28,
            text: "所以我提了一个"
        },
        {
            index: NumberInt("47"),
            startTime: 86.28,
            endTime: 88.64,
            text: "最关注的AI智能体"
        },
        {
            index: NumberInt("48"),
            startTime: 88.64,
            endTime: 90.32,
            text: "你稍微已经做了"
        },
        {
            index: NumberInt("49"),
            startTime: 90.32,
            endTime: 94.48,
            text: "他有几十种上万个AI安全智能体"
        },
        {
            index: NumberInt("50"),
            startTime: 94.48,
            endTime: 96.28,
            text: "这个智能体能够实验"
        },
        {
            index: NumberInt("51"),
            startTime: 96.28,
            endTime: 97.92,
            text: "AI能够挖掘人民众众"
        },
        {
            index: NumberInt("52"),
            startTime: 97.92,
            endTime: 99.52,
            text: "AI能够自动运营"
        },
        {
            index: NumberInt("53"),
            startTime: 99.52,
            endTime: 102.08,
            text: "AI能够体育其他回合的黑客智能体"
        },
        {
            index: NumberInt("54"),
            startTime: 102.08,
            endTime: 104.32,
            text: "中国有200万家中央企业"
        },
        {
            index: NumberInt("55"),
            startTime: 104.32,
            endTime: 106.16,
            text: "营谈用某的AI智能体"
        },
        {
            index: NumberInt("56"),
            startTime: 106.16,
            endTime: 107.64,
            text: "免费了在为他的企业建筑"
        },
        {
            index: NumberInt("57"),
            startTime: 107.64,
            endTime: 109.4,
            text: "做实施的防护"
        },
        {
            index: NumberInt("58"),
            startTime: 109.4,
            endTime: 110.6,
            text: "实施的运营"
        },
        {
            index: NumberInt("59"),
            startTime: 110.6,
            endTime: 112.32,
            text: "我上有一个建议"
        },
        {
            index: NumberInt("60"),
            startTime: 112.32,
            endTime: 114.2,
            text: "一定把算力已经区分"
        },
        {
            index: NumberInt("61"),
            startTime: 114.2,
            endTime: 116.44,
            text: "成为训练算力和推理算力"
        },
        {
            index: NumberInt("62"),
            startTime: 116.44,
            endTime: 118,
            text: "希望各地在发展算力方面"
        },
        {
            index: NumberInt("63"),
            startTime: 118,
            endTime: 120.44,
            text: "能够兼项于推理算力"
        },
        {
            index: NumberInt("64"),
            startTime: 120.44,
            endTime: 122.4,
            text: "我们国家产业症的方面"
        },
        {
            index: NumberInt("65"),
            startTime: 122.4,
            endTime: 124.08,
            text: "既然推理算力这么重要"
        },
        {
            index: NumberInt("66"),
            startTime: 124.08,
            endTime: 124.88,
            text: "因为我也讲了"
        },
        {
            index: NumberInt("67"),
            startTime: 124.88,
            endTime: 126.64,
            text: "就是要发展推理芯片"
        },
        {
            index: NumberInt("68"),
            startTime: 126.64,
            endTime: 128.96,
            text: "芯片症不能一味的都在追"
        },
        {
            index: NumberInt("69"),
            startTime: 128.96,
            endTime: 131,
            text: "运营它的高端的训练芯片"
        },
        {
            index: NumberInt("70"),
            startTime: 131,
            endTime: 132.84,
            text: "是我关注的这些企业"
        },
        {
            index: NumberInt("71"),
            startTime: 132.84,
            endTime: 134.64,
            text: "就是和个人怎么迅速的永远"
        },
        {
            index: NumberInt("72"),
            startTime: 134.64,
            endTime: 136.08,
            text: "一个人如何打造自己的"
        },
        {
            index: NumberInt("73"),
            startTime: 136.08,
            endTime: 137.2,
            text: "私人智能体"
        },
        {
            index: NumberInt("74"),
            startTime: 137.2,
            endTime: 138.64,
            text: "我提出一个打造智能体"
        },
        {
            index: NumberInt("75"),
            startTime: 138.64,
            endTime: 140.12,
            text: "开放平台的概念"
        },
        {
            index: NumberInt("76"),
            startTime: 140.12,
            endTime: 141.08,
            text: "我把智能体的技术"
        },
        {
            index: NumberInt("77"),
            startTime: 141.08,
            endTime: 142.52,
            text: "设施藏在后面"
        },
        {
            index: NumberInt("78"),
            startTime: 142.52,
            endTime: 144.84,
            text: "让普通企业 普通客人"
        },
        {
            index: NumberInt("79"),
            startTime: 144.84,
            endTime: 147.4,
            text: "很容易的建立自己的智能体"
        },
        {
            index: NumberInt("80"),
            startTime: 147.4,
            endTime: 149.48,
            text: "那今年周总还会拍两会vlog吗"
        },
        {
            index: NumberInt("81"),
            startTime: 149.48,
            endTime: 150.56,
            text: "那我一定是戴手机"
        },
        {
            index: NumberInt("82"),
            startTime: 150.56,
            endTime: 151.88,
            text: "就不用拍"
        },
        {
            index: NumberInt("83"),
            startTime: 151.88,
            endTime: 154.28,
            text: "好的 那就两会现场见"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-11T08:23:23.681Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69da058909a67d0ac3d80ca6"),
    videoId: NumberInt("28"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 3.16,
            text: "伯托360集团创始人红一大叔周红毅聊了聊"
        },
        {
            index: NumberInt("2"),
            startTime: 3.16,
            endTime: 5.52,
            text: "他给普通人提供使用AI的三个建议"
        },
        {
            index: NumberInt("3"),
            startTime: 5.52,
            endTime: 8.04,
            text: "让他看到我上一讲让他一辆COSPLAY"
        },
        {
            index: NumberInt("4"),
            startTime: 8.04,
            endTime: 11.92,
            text: "因为AI把全线所有能找到的啥知识都学习到了"
        },
        {
            index: NumberInt("5"),
            startTime: 11.92,
            endTime: 13.2,
            text: "他啥知识都会"
        },
        {
            index: NumberInt("6"),
            startTime: 13.2,
            endTime: 14.36,
            text: "你不给他冶定专业方向"
        },
        {
            index: NumberInt("7"),
            startTime: 14.36,
            endTime: 16.04,
            text: "他什么知识都给你用一点"
        },
        {
            index: NumberInt("8"),
            startTime: 16.04,
            endTime: 17.76,
            text: "他就会非常平淡无情"
        },
        {
            index: NumberInt("9"),
            startTime: 17.76,
            endTime: 18.48,
            text: "这是第一个"
        },
        {
            index: NumberInt("10"),
            startTime: 18.48,
            endTime: 20.6,
            text: "所以我现在不直接跟他们一聊天"
        },
        {
            index: NumberInt("11"),
            startTime: 20.6,
            endTime: 22.4,
            text: "我做了几个聊天的对象"
        },
        {
            index: NumberInt("12"),
            startTime: 22.4,
            endTime: 25.2,
            text: "比如说做投资分析的有一个"
        },
        {
            index: NumberInt("13"),
            startTime: 25.2,
            endTime: 28.36,
            text: "做公司或是商业码这份也有一个"
        },
        {
            index: NumberInt("14"),
            startTime: 28.36,
            endTime: 30.16,
            text: "你一定不能把他当搜索来用"
        },
        {
            index: NumberInt("15"),
            startTime: 30.16,
            endTime: 32.12,
            text: "搜索两种喜欢去掃搜的关键字"
        },
        {
            index: NumberInt("16"),
            startTime: 32.12,
            endTime: 33.56,
            text: "不是问简单问题"
        },
        {
            index: NumberInt("17"),
            startTime: 33.56,
            endTime: 34.76,
            text: "我们那样聊天"
        },
        {
            index: NumberInt("18"),
            startTime: 34.76,
            endTime: 36.72,
            text: "我就我对大家问一个参数"
        },
        {
            index: NumberInt("19"),
            startTime: 36.72,
            endTime: 40.08,
            text: "我每次大概要给他讲三四百字到一千字"
        },
        {
            index: NumberInt("20"),
            startTime: 40.08,
            endTime: 41.96,
            text: "就跟AI对话是必用语音书"
        },
        {
            index: NumberInt("21"),
            startTime: 41.96,
            endTime: 42.92,
            text: "你拿一文字书"
        },
        {
            index: NumberInt("22"),
            startTime: 42.92,
            endTime: 45.12,
            text: "我拿文字书上千字也不可能"
        },
        {
            index: NumberInt("23"),
            startTime: 45.12,
            endTime: 46.56,
            text: "我并不注重讲话了"
        },
        {
            index: NumberInt("24"),
            startTime: 46.56,
            endTime: 48.12,
            text: "言讯逻辑"
        },
        {
            index: NumberInt("25"),
            startTime: 48.12,
            endTime: 49.96,
            text: "我当时是讲了很近生的问题"
        },
        {
            index: NumberInt("26"),
            startTime: 49.96,
            endTime: 50.88,
            text: "我就不用问他了"
        },
        {
            index: NumberInt("27"),
            startTime: 50.88,
            endTime: 52.56,
            text: "我但是我会把我遇到了困难"
        },
        {
            index: NumberInt("28"),
            startTime: 52.56,
            endTime: 53.2,
            text: "我的困惑"
        },
        {
            index: NumberInt("29"),
            startTime: 53.2,
            endTime: 55.4,
            text: "我也不充分想把一无脑的交给他"
        },
        {
            index: NumberInt("30"),
            startTime: 55.4,
            endTime: 57.76,
            text: "你要把AI当人翻一个公文"
        },
        {
            index: NumberInt("31"),
            startTime: 57.76,
            endTime: 58.88,
            text: "就给他轻松"
        },
        {
            index: NumberInt("32"),
            startTime: 58.88,
            endTime: 61.32,
            text: "第三个是您一定要会问他"
        },
        {
            index: NumberInt("33"),
            startTime: 61.32,
            endTime: 62.92,
            text: "就是他给你的答案"
        },
        {
            index: NumberInt("34"),
            startTime: 62.92,
            endTime: 64.12,
            text: "不要马上接受"
        },
        {
            index: NumberInt("35"),
            startTime: 64.12,
            endTime: 66.96,
            text: "你要像用第一年里去问他说"
        },
        {
            index: NumberInt("36"),
            startTime: 66.96,
            endTime: 67.88,
            text: "真是这样的"
        },
        {
            index: NumberInt("37"),
            startTime: 67.88,
            endTime: 69.56,
            text: "如果这样做会怎么样"
        },
        {
            index: NumberInt("38"),
            startTime: 69.56,
            endTime: 70.92,
            text: "他有一个多轮"
        },
        {
            index: NumberInt("39"),
            startTime: 70.92,
            endTime: 71.96,
            text: "这个两个星之后"
        },
        {
            index: NumberInt("40"),
            startTime: 71.96,
            endTime: 73.56,
            text: "你会激发AI的很多箱子"
        },
        {
            index: NumberInt("41"),
            startTime: 73.56,
            endTime: 75.16,
            text: "AI反正会激发你那箱子"
        },
        {
            index: NumberInt("42"),
            startTime: 75.16,
            endTime: 77.44,
            text: "2026全国两会即将召开"
        },
        {
            index: NumberInt("43"),
            startTime: 77.44,
            endTime: 79.68,
            text: "他还透露了自己最关心的领域"
        },
        {
            index: NumberInt("44"),
            startTime: 79.68,
            endTime: 81.68,
            text: "我关注的三个方向"
        },
        {
            index: NumberInt("45"),
            startTime: 81.68,
            endTime: 84.08,
            text: "您就刚才讲了AI不能安全"
        },
        {
            index: NumberInt("46"),
            startTime: 84.08,
            endTime: 86.28,
            text: "所以我提了一个"
        },
        {
            index: NumberInt("47"),
            startTime: 86.28,
            endTime: 88.64,
            text: "最关注的AI智能体"
        },
        {
            index: NumberInt("48"),
            startTime: 88.64,
            endTime: 90.32,
            text: "你稍微已经做了"
        },
        {
            index: NumberInt("49"),
            startTime: 90.32,
            endTime: 94.48,
            text: "他有几十种上万个AI安全智能体"
        },
        {
            index: NumberInt("50"),
            startTime: 94.48,
            endTime: 96.28,
            text: "这个智能体能够实验"
        },
        {
            index: NumberInt("51"),
            startTime: 96.28,
            endTime: 97.92,
            text: "AI能够挖掘人民众众"
        },
        {
            index: NumberInt("52"),
            startTime: 97.92,
            endTime: 99.52,
            text: "AI能够自动运营"
        },
        {
            index: NumberInt("53"),
            startTime: 99.52,
            endTime: 102.08,
            text: "AI能够体育其他回合的黑客智能体"
        },
        {
            index: NumberInt("54"),
            startTime: 102.08,
            endTime: 104.32,
            text: "中国有200万家中央企业"
        },
        {
            index: NumberInt("55"),
            startTime: 104.32,
            endTime: 106.16,
            text: "营谈用某的AI智能体"
        },
        {
            index: NumberInt("56"),
            startTime: 106.16,
            endTime: 107.64,
            text: "免费了在为他的企业建筑"
        },
        {
            index: NumberInt("57"),
            startTime: 107.64,
            endTime: 109.4,
            text: "做实施的防护"
        },
        {
            index: NumberInt("58"),
            startTime: 109.4,
            endTime: 110.6,
            text: "实施的运营"
        },
        {
            index: NumberInt("59"),
            startTime: 110.6,
            endTime: 112.32,
            text: "我上有一个建议"
        },
        {
            index: NumberInt("60"),
            startTime: 112.32,
            endTime: 114.2,
            text: "一定把算力已经区分"
        },
        {
            index: NumberInt("61"),
            startTime: 114.2,
            endTime: 116.44,
            text: "成为训练算力和推理算力"
        },
        {
            index: NumberInt("62"),
            startTime: 116.44,
            endTime: 118,
            text: "希望各地在发展算力方面"
        },
        {
            index: NumberInt("63"),
            startTime: 118,
            endTime: 120.44,
            text: "能够兼项于推理算力"
        },
        {
            index: NumberInt("64"),
            startTime: 120.44,
            endTime: 122.4,
            text: "我们国家产业症的方面"
        },
        {
            index: NumberInt("65"),
            startTime: 122.4,
            endTime: 124.08,
            text: "既然推理算力这么重要"
        },
        {
            index: NumberInt("66"),
            startTime: 124.08,
            endTime: 124.88,
            text: "因为我也讲了"
        },
        {
            index: NumberInt("67"),
            startTime: 124.88,
            endTime: 126.64,
            text: "就是要发展推理芯片"
        },
        {
            index: NumberInt("68"),
            startTime: 126.64,
            endTime: 128.96,
            text: "芯片症不能一味的都在追"
        },
        {
            index: NumberInt("69"),
            startTime: 128.96,
            endTime: 131,
            text: "运营它的高端的训练芯片"
        },
        {
            index: NumberInt("70"),
            startTime: 131,
            endTime: 132.84,
            text: "是我关注的这些企业"
        },
        {
            index: NumberInt("71"),
            startTime: 132.84,
            endTime: 134.64,
            text: "就是和个人怎么迅速的永远"
        },
        {
            index: NumberInt("72"),
            startTime: 134.64,
            endTime: 136.08,
            text: "一个人如何打造自己的"
        },
        {
            index: NumberInt("73"),
            startTime: 136.08,
            endTime: 137.2,
            text: "私人智能体"
        },
        {
            index: NumberInt("74"),
            startTime: 137.2,
            endTime: 138.64,
            text: "我提出一个打造智能体"
        },
        {
            index: NumberInt("75"),
            startTime: 138.64,
            endTime: 140.12,
            text: "开放平台的概念"
        },
        {
            index: NumberInt("76"),
            startTime: 140.12,
            endTime: 141.08,
            text: "我把智能体的技术"
        },
        {
            index: NumberInt("77"),
            startTime: 141.08,
            endTime: 142.52,
            text: "设施藏在后面"
        },
        {
            index: NumberInt("78"),
            startTime: 142.52,
            endTime: 144.84,
            text: "让普通企业 普通客人"
        },
        {
            index: NumberInt("79"),
            startTime: 144.84,
            endTime: 147.4,
            text: "很容易的建立自己的智能体"
        },
        {
            index: NumberInt("80"),
            startTime: 147.4,
            endTime: 149.48,
            text: "那今年周总还会拍两会vlog吗"
        },
        {
            index: NumberInt("81"),
            startTime: 149.48,
            endTime: 150.56,
            text: "那我一定是戴手机"
        },
        {
            index: NumberInt("82"),
            startTime: 150.56,
            endTime: 151.88,
            text: "就不用拍"
        },
        {
            index: NumberInt("83"),
            startTime: 151.88,
            endTime: 154.28,
            text: "好的 那就两会现场见"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-11T08:25:45.787Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69da12780639f06ddfdfc085"),
    videoId: NumberInt("29"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 2.08,
            text: "请看照片中这几百个年轻人"
        },
        {
            index: NumberInt("2"),
            startTime: 2.08,
            endTime: 5.56,
            text: "他们正顶着烈日抬着城队等着凌渠"
        },
        {
            index: NumberInt("3"),
            startTime: 5.56,
            endTime: 6.92,
            text: "小龙虾"
        },
        {
            index: NumberInt("4"),
            startTime: 6.92,
            endTime: 7.68,
            text: "别误会"
        },
        {
            index: NumberInt("5"),
            startTime: 7.68,
            endTime: 10.04,
            text: "这不是中饭吃海鲜大排档的盛况"
        },
        {
            index: NumberInt("6"),
            startTime: 10.04,
            endTime: 12.68,
            text: "而是最近沙丰热的开源AI智能体"
        },
        {
            index: NumberInt("7"),
            startTime: 12.68,
            endTime: 15.68,
            text: "Open Claw的安装现场"
        },
        {
            index: NumberInt("8"),
            startTime: 15.68,
            endTime: 18.96,
            text: "现在全网都在无脑喊它小龙虾"
        },
        {
            index: NumberInt("9"),
            startTime: 18.96,
            endTime: 20.44,
            text: "但作为英语老师"
        },
        {
            index: NumberInt("10"),
            startTime: 20.44,
            endTime: 22.76,
            text: "吐作实在看不下去了"
        },
        {
            index: NumberInt("11"),
            startTime: 22.76,
            endTime: 24.8,
            text: "把Open Claw翻译成小龙虾的"
        },
        {
            index: NumberInt("12"),
            startTime: 24.8,
            endTime: 26.6,
            text: "出来挨打"
        },
        {
            index: NumberInt("13"),
            startTime: 26.6,
            endTime: 28.44,
            text: "请看我这只龙虾"
        },
        {
            index: NumberInt("14"),
            startTime: 28.44,
            endTime: 30.6,
            text: "看到它的大钳子了吗"
        },
        {
            index: NumberInt("15"),
            startTime: 30.6,
            endTime: 32.92,
            text: "这才是Claw"
        },
        {
            index: NumberInt("16"),
            startTime: 32.92,
            endTime: 35.4,
            text: "它可以指水生有颗动物的钳"
        },
        {
            index: NumberInt("17"),
            startTime: 35.4,
            endTime: 36.2,
            text: "熬"
        },
        {
            index: NumberInt("18"),
            startTime: 36.2,
            endTime: 38.04,
            text: "所以螃蟹的钳子"
        },
        {
            index: NumberInt("19"),
            startTime: 38.04,
            endTime: 39.36,
            text: "也是Claw"
        },
        {
            index: NumberInt("20"),
            startTime: 39.36,
            endTime: 40,
            text: "比如"
        },
        {
            index: NumberInt("21"),
            startTime: 40,
            endTime: 41.68,
            text: "The Claws of a Crab"
        },
        {
            index: NumberInt("22"),
            startTime: 41.68,
            endTime: 43.08,
            text: "蟹熬"
        },
        {
            index: NumberInt("23"),
            startTime: 43.08,
            endTime: 45.08,
            text: "再请看我这只龙虾的钳子"
        },
        {
            index: NumberInt("24"),
            startTime: 45.08,
            endTime: 47.24,
            text: "它是不是打开的"
        },
        {
            index: NumberInt("25"),
            startTime: 47.24,
            endTime: 49,
            text: "Open 左形容词"
        },
        {
            index: NumberInt("26"),
            startTime: 49,
            endTime: 50.6,
            text: "所以这只钳子是"
        },
        {
            index: NumberInt("27"),
            startTime: 50.6,
            endTime: 52,
            text: "等等等等"
        },
        {
            index: NumberInt("28"),
            startTime: 52,
            endTime: 53.64,
            text: "Open Claw"
        },
        {
            index: NumberInt("29"),
            startTime: 53.64,
            endTime: 55.48,
            text: "打开的钳子"
        },
        {
            index: NumberInt("30"),
            startTime: 55.48,
            endTime: 56.36,
            text: "而不过"
        },
        {
            index: NumberInt("31"),
            startTime: 56.36,
            endTime: 58.08,
            text: "Open Claw这里的Open"
        },
        {
            index: NumberInt("32"),
            startTime: 58.08,
            endTime: 59,
            text: "指的其实是"
        },
        {
            index: NumberInt("33"),
            startTime: 59,
            endTime: 60.88,
            text: "看源代码的意思"
        },
        {
            index: NumberInt("34"),
            startTime: 60.88,
            endTime: 62.68,
            text: "这名字简直是贴脸开大"
        },
        {
            index: NumberInt("35"),
            startTime: 62.68,
            endTime: 63.88,
            text: "直接怼上了那家"
        },
        {
            index: NumberInt("36"),
            startTime: 63.88,
            endTime: 65.72,
            text: "越来越不Open的AI巨头"
        },
        {
            index: NumberInt("37"),
            startTime: 65.72,
            endTime: 66.96,
            text: "Open AI"
        },
        {
            index: NumberInt("38"),
            startTime: 66.96,
            endTime: 68.64,
            text: "好嘛既然有Open AI"
        },
        {
            index: NumberInt("39"),
            startTime: 68.64,
            endTime: 69.72,
            text: "那叫Open Claw"
        },
        {
            index: NumberInt("40"),
            startTime: 69.72,
            endTime: 71.52,
            text: "也算是蹭上了"
        },
        {
            index: NumberInt("41"),
            startTime: 71.52,
            endTime: 73.16,
            text: "但其实Open Claw"
        },
        {
            index: NumberInt("42"),
            startTime: 73.16,
            endTime: 75.8,
            text: "原本蹭的不是Open AI"
        },
        {
            index: NumberInt("43"),
            startTime: 75.8,
            endTime: 78.56,
            text: "它最早的名字是Clawed Bot"
        },
        {
            index: NumberInt("44"),
            startTime: 78.56,
            endTime: 79.36,
            text: "为啥"
        },
        {
            index: NumberInt("45"),
            startTime: 79.36,
            endTime: 80.88,
            text: "因为当你给一个机器人"
        },
        {
            index: NumberInt("46"),
            startTime: 80.88,
            endTime: 82,
            text: "只能提Bot"
        },
        {
            index: NumberInt("47"),
            startTime: 82,
            endTime: 82.84,
            text: "按上钳子"
        },
        {
            index: NumberInt("48"),
            startTime: 82.84,
            endTime: 83.8,
            text: "它就变成了"
        },
        {
            index: NumberInt("49"),
            startTime: 83.8,
            endTime: 86.64,
            text: "长着钳子的英文单词Clawed"
        },
        {
            index: NumberInt("50"),
            startTime: 86.64,
            endTime: 88.56,
            text: "机器人只能提"
        },
        {
            index: NumberInt("51"),
            startTime: 88.56,
            endTime: 92.04,
            text: "请大声跟我读Clawed"
        },
        {
            index: NumberInt("52"),
            startTime: 92.04,
            endTime: 92.76,
            text: "哎呦"
        },
        {
            index: NumberInt("53"),
            startTime: 92.76,
            endTime: 93.44,
            text: "这发音"
        },
        {
            index: NumberInt("54"),
            startTime: 93.44,
            endTime: 94.68,
            text: "和另一家AI巨头"
        },
        {
            index: NumberInt("55"),
            startTime: 94.68,
            endTime: 97.24,
            text: "Anthropic的当家大模型Clawed"
        },
        {
            index: NumberInt("56"),
            startTime: 97.24,
            endTime: 98.76,
            text: "不能说很像"
        },
        {
            index: NumberInt("57"),
            startTime: 98.76,
            endTime: 100.72,
            text: "只能说一模一样啊"
        },
        {
            index: NumberInt("58"),
            startTime: 100.72,
            endTime: 101.28,
            text: "注意"
        },
        {
            index: NumberInt("59"),
            startTime: 101.28,
            endTime: 102.44,
            text: "Clawed这个名字"
        },
        {
            index: NumberInt("60"),
            startTime: 102.44,
            endTime: 104.68,
            text: "不要错读成Clawed"
        },
        {
            index: NumberInt("61"),
            startTime: 104.68,
            endTime: 105.24,
            text: "你看"
        },
        {
            index: NumberInt("62"),
            startTime: 105.24,
            endTime: 107.2,
            text: "一个双关于Open Claw"
        },
        {
            index: NumberInt("63"),
            startTime: 107.2,
            endTime: 108.36,
            text: "居然同时把两架"
        },
        {
            index: NumberInt("64"),
            startTime: 108.36,
            endTime: 109.64,
            text: "估值千亿的科技巨头"
        },
        {
            index: NumberInt("65"),
            startTime: 109.64,
            endTime: 111.32,
            text: "按在地上疯狂摩擦"
        },
        {
            index: NumberInt("66"),
            startTime: 111.32,
            endTime: 112.72,
            text: "这一次会亮完的"
        },
        {
            index: NumberInt("67"),
            startTime: 112.72,
            endTime: 114.16,
            text: "高级"
        },
        {
            index: NumberInt("68"),
            startTime: 114.16,
            endTime: 115.88,
            text: "说会现在全网爆红"
        },
        {
            index: NumberInt("69"),
            startTime: 115.88,
            endTime: 118.16,
            text: "唉 还真的是字名以上的爆红"
        },
        {
            index: NumberInt("70"),
            startTime: 118.16,
            endTime: 119.32,
            text: "小龙虾"
        },
        {
            index: NumberInt("71"),
            startTime: 119.32,
            endTime: 121.32,
            text: "麻烦看看它的官方LOGO"
        },
        {
            index: NumberInt("72"),
            startTime: 121.32,
            endTime: 123.64,
            text: "这长着巨星茄子的深海霸主"
        },
        {
            index: NumberInt("73"),
            startTime: 123.64,
            endTime: 125.2,
            text: "整而霸经的英文叫"
        },
        {
            index: NumberInt("74"),
            startTime: 125.2,
            endTime: 126,
            text: "Lobster"
        },
        {
            index: NumberInt("75"),
            startTime: 126,
            endTime: 127.16,
            text: "龙虾"
        },
        {
            index: NumberInt("76"),
            startTime: 127.16,
            endTime: 128.12,
            text: "你是不是以为"
        },
        {
            index: NumberInt("77"),
            startTime: 128.12,
            endTime: 130.36,
            text: "小龙虾叫Little Lobster"
        },
        {
            index: NumberInt("78"),
            startTime: 130.36,
            endTime: 132.2,
            text: "那可真是想当然了"
        },
        {
            index: NumberInt("79"),
            startTime: 132.2,
            endTime: 133.16,
            text: "咱们夏天夜市里"
        },
        {
            index: NumberInt("80"),
            startTime: 133.16,
            endTime: 135.4,
            text: "配着冰啼酒狂炫的那种小龙虾"
        },
        {
            index: NumberInt("81"),
            startTime: 135.4,
            endTime: 136.52,
            text: "英国人叫它们"
        },
        {
            index: NumberInt("82"),
            startTime: 136.52,
            endTime: 137.68,
            text: "Crayfish"
        },
        {
            index: NumberInt("83"),
            startTime: 137.68,
            endTime: 138.72,
            text: "美国人叫它们"
        },
        {
            index: NumberInt("84"),
            startTime: 138.72,
            endTime: 139.92,
            text: "Crawfish"
        },
        {
            index: NumberInt("85"),
            startTime: 139.92,
            endTime: 141.56,
            text: "总之不是Little Lobster"
        },
        {
            index: NumberInt("86"),
            startTime: 141.56,
            endTime: 143.52,
            text: "或Small Lobster"
        },
        {
            index: NumberInt("87"),
            startTime: 143.52,
            endTime: 145.2,
            text: "说回Claw这个词"
        },
        {
            index: NumberInt("88"),
            startTime: 145.2,
            endTime: 147.08,
            text: "这个不只是海鲜的专属"
        },
        {
            index: NumberInt("89"),
            startTime: 147.08,
            endTime: 148.96,
            text: "天上飞的鸟也有Claw"
        },
        {
            index: NumberInt("90"),
            startTime: 148.96,
            endTime: 150.36,
            text: "小区里那只脾气暴躁的"
        },
        {
            index: NumberInt("91"),
            startTime: 150.36,
            endTime: 151.92,
            text: "流浪猫也有Claw"
        },
        {
            index: NumberInt("92"),
            startTime: 151.92,
            endTime: 153.48,
            text: "别问我怎么知道的"
        },
        {
            index: NumberInt("93"),
            startTime: 153.48,
            endTime: 154.68,
            text: "更重要的是"
        },
        {
            index: NumberInt("94"),
            startTime: 154.68,
            endTime: 157,
            text: "它还是个极其凶狠的洞瓷"
        },
        {
            index: NumberInt("95"),
            startTime: 157,
            endTime: 158.24,
            text: "比如咱们可以说"
        },
        {
            index: NumberInt("96"),
            startTime: 158.24,
            endTime: 160.2,
            text: "The wolf clawed at my door"
        },
        {
            index: NumberInt("97"),
            startTime: 160.2,
            endTime: 162.08,
            text: "狼疯狂地抓拿我的门"
        },
        {
            index: NumberInt("98"),
            startTime: 162.08,
            endTime: 164.04,
            text: "She clawed me across the face"
        },
        {
            index: NumberInt("99"),
            startTime: 164.04,
            endTime: 165.44,
            text: "她抓了我的脸"
        },
        {
            index: NumberInt("100"),
            startTime: 165.44,
            endTime: 167.2,
            text: "My hands clawed the air"
        },
        {
            index: NumberInt("101"),
            startTime: 167.2,
            endTime: 169.52,
            text: "我的双手在空中乱抓"
        },
        {
            index: NumberInt("102"),
            startTime: 169.52,
            endTime: 170.68,
            text: "再假设"
        },
        {
            index: NumberInt("103"),
            startTime: 170.68,
            endTime: 172.88,
            text: "你最近英语考试彻底翻车了"
        },
        {
            index: NumberInt("104"),
            startTime: 172.88,
            endTime: 175.16,
            text: "或者基金爹的亲妈都不认了"
        },
        {
            index: NumberInt("105"),
            startTime: 175.16,
            endTime: 176.48,
            text: "你想重回巅峰了"
        },
        {
            index: NumberInt("106"),
            startTime: 176.48,
            endTime: 178.32,
            text: "那绝对不叫Walkback"
        },
        {
            index: NumberInt("107"),
            startTime: 178.32,
            endTime: 180.08,
            text: "得叫Claw your way back"
        },
        {
            index: NumberInt("108"),
            startTime: 180.08,
            endTime: 180.8,
            text: "意思是"
        },
        {
            index: NumberInt("109"),
            startTime: 180.8,
            endTime: 181.68,
            text: "张牙舞爪"
        },
        {
            index: NumberInt("110"),
            startTime: 181.68,
            endTime: 182.56,
            text: "脸滚带趴地"
        },
        {
            index: NumberInt("111"),
            startTime: 182.56,
            endTime: 183.72,
            text: "杀出一条血路"
        },
        {
            index: NumberInt("112"),
            startTime: 183.72,
            endTime: 185.6,
            text: "把失去的夺回来"
        },
        {
            index: NumberInt("113"),
            startTime: 185.6,
            endTime: 186.08,
            text: "好"
        },
        {
            index: NumberInt("114"),
            startTime: 186.08,
            endTime: 188.88,
            text: "在你跑去排队灵养这只赛波海鲜之前"
        },
        {
            index: NumberInt("115"),
            startTime: 188.88,
            endTime: 191.16,
            text: "请复习一下本颗的知识点"
        },
        {
            index: NumberInt("116"),
            startTime: 191.16,
            endTime: 194.4,
            text: "我是英语兔"
        },
        {
            index: NumberInt("117"),
            startTime: 194.4,
            endTime: 195.76,
            text: "咱们下次再见"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-11T09:20:56.718Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69da165d6f943b49176115ba"),
    videoId: NumberInt("30"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 1.72,
            text: "数学怎么提分"
        },
        {
            index: NumberInt("2"),
            startTime: 1.72,
            endTime: 3,
            text: "我给你说了多少次"
        },
        {
            index: NumberInt("3"),
            startTime: 3,
            endTime: 4.64,
            text: "好的 再讲就是不做 你知道吗"
        },
        {
            index: NumberInt("4"),
            startTime: 4.64,
            endTime: 6.56,
            text: "就是四个字 四科错题"
        },
        {
            index: NumberInt("5"),
            startTime: 6.56,
            endTime: 7.88,
            text: "跟你讲 关于这个四科错题"
        },
        {
            index: NumberInt("6"),
            startTime: 7.88,
            endTime: 10.4,
            text: "很多人在执行的时候有各种的问题"
        },
        {
            index: NumberInt("7"),
            startTime: 10.4,
            endTime: 12.68,
            text: "其实就是没做到点子上"
        },
        {
            index: NumberInt("8"),
            startTime: 12.68,
            endTime: 15.44,
            text: "什么分析我错的是什么知识点"
        },
        {
            index: NumberInt("9"),
            startTime: 15.44,
            endTime: 17.04,
            text: "分析我错的原因"
        },
        {
            index: NumberInt("10"),
            startTime: 17.04,
            endTime: 18.8,
            text: "分析什么这个提醒是啥"
        },
        {
            index: NumberInt("11"),
            startTime: 18.8,
            endTime: 19.88,
            text: "这都没有用"
        },
        {
            index: NumberInt("12"),
            startTime: 19.88,
            endTime: 22.32,
            text: "你要做的就是把这个题再给我做对就完了"
        },
        {
            index: NumberInt("13"),
            startTime: 22.32,
            endTime: 23.8,
            text: "做不对就一直改"
        },
        {
            index: NumberInt("14"),
            startTime: 23.8,
            endTime: 26.04,
            text: "你们要把这件事的注意力"
        },
        {
            index: NumberInt("15"),
            startTime: 26.04,
            endTime: 27.68,
            text: "放在最关键的地方 是吧"
        },
        {
            index: NumberInt("16"),
            startTime: 27.68,
            endTime: 29.52,
            text: "因为什么错是没有意义的"
        },
        {
            index: NumberInt("17"),
            startTime: 29.52,
            endTime: 30.48,
            text: "我就问你个问题"
        },
        {
            index: NumberInt("18"),
            startTime: 30.48,
            endTime: 32,
            text: "他因为什么错"
        },
        {
            index: NumberInt("19"),
            startTime: 32,
            endTime: 33.04,
            text: "他知道了这个原因"
        },
        {
            index: NumberInt("20"),
            startTime: 33.04,
            endTime: 34.2,
            text: "他下次就可以不错"
        },
        {
            index: NumberInt("21"),
            startTime: 34.2,
            endTime: 35.96,
            text: "好在那里面就想这个问题"
        },
        {
            index: NumberInt("22"),
            startTime: 35.96,
            endTime: 37,
            text: "他是不会的"
        },
        {
            index: NumberInt("23"),
            startTime: 37,
            endTime: 39.04,
            text: "他就是很多东西不扎实不熟练"
        },
        {
            index: NumberInt("24"),
            startTime: 39.04,
            endTime: 39.68,
            text: "就这么简单"
        },
        {
            index: NumberInt("25"),
            startTime: 39.68,
            endTime: 40.4,
            text: "那怎么办呢"
        },
        {
            index: NumberInt("26"),
            startTime: 40.4,
            endTime: 41.52,
            text: "你让他扎实让他熟练"
        },
        {
            index: NumberInt("27"),
            startTime: 41.52,
            endTime: 42.6,
            text: "那怎么办你就改错"
        },
        {
            index: NumberInt("28"),
            startTime: 42.6,
            endTime: 44.64,
            text: "我刚才说过循环改错 循环改错"
        },
        {
            index: NumberInt("29"),
            startTime: 44.64,
            endTime: 46.4,
            text: "就你这张卷量错题第一次改"
        },
        {
            index: NumberInt("30"),
            startTime: 46.4,
            endTime: 47.8,
            text: "改对的答案给我改错了"
        },
        {
            index: NumberInt("31"),
            startTime: 47.8,
            endTime: 50,
            text: "就打个差学期答案下一周继续"
        },
        {
            index: NumberInt("32"),
            startTime: 50,
            endTime: 52.04,
            text: "没有几个人能够把这东西坚持下来"
        },
        {
            index: NumberInt("33"),
            startTime: 52.04,
            endTime: 53.56,
            text: "坚持下来的扎实"
        },
        {
            index: NumberInt("34"),
            startTime: 53.56,
            endTime: 55.52,
            text: "都已经发现这个方法危利无缘"
        },
        {
            index: NumberInt("35"),
            startTime: 55.52,
            endTime: 56.76,
            text: "那成绩层层往上找"
        },
        {
            index: NumberInt("36"),
            startTime: 56.76,
            endTime: 59.6,
            text: "所以关键是在坚持在执行"
        },
        {
            index: NumberInt("37"),
            startTime: 59.6,
            endTime: 61.96,
            text: "我刚好说一万三千小时说一万可"
        },
        {
            index: NumberInt("38"),
            startTime: 61.96,
            endTime: 62.96,
            text: "今天给大家准备了一个好东西"
        },
        {
            index: NumberInt("39"),
            startTime: 62.96,
            endTime: 63.72,
            text: "逆袭有方法"
        },
        {
            index: NumberInt("40"),
            startTime: 63.72,
            endTime: 64.8,
            text: "每一次都是我抄行程"
        },
        {
            index: NumberInt("41"),
            startTime: 64.8,
            endTime: 65.72,
            text: "我给你免费发一份"
        },
        {
            index: NumberInt("42"),
            startTime: 65.72,
            endTime: 66.84,
            text: "平时去打出逆袭"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-11T09:37:33.051Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69f2188e8d27c661a000c16c"),
    videoId: NumberInt("34"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 1.4,
            text: "准渡好发力不一定好"
        },
        {
            index: NumberInt("2"),
            startTime: 1.4,
            endTime: 3.28,
            text: "发力好准渡一定差不了"
        },
        {
            index: NumberInt("3"),
            startTime: 3.28,
            endTime: 7.4,
            text: "所有发力差的投篮本质都有这样一个问题"
        },
        {
            index: NumberInt("4"),
            startTime: 7.4,
            endTime: 10.28,
            text: "不管是推投手踝脚慢还是投尸肌也好"
        },
        {
            index: NumberInt("5"),
            startTime: 10.28,
            endTime: 11.44,
            text: "都有一个共性"
        },
        {
            index: NumberInt("6"),
            startTime: 11.44,
            endTime: 13.68,
            text: "那就是下肢在登身的过程中"
        },
        {
            index: NumberInt("7"),
            startTime: 13.68,
            endTime: 16.12,
            text: "上肢在整体向前"
        },
        {
            index: NumberInt("8"),
            startTime: 16.12,
            endTime: 19,
            text: "没有一段重合的力线"
        },
        {
            index: NumberInt("9"),
            startTime: 19,
            endTime: 20.48,
            text: "划个力线图就会发现"
        },
        {
            index: NumberInt("10"),
            startTime: 20.48,
            endTime: 23.4,
            text: "起球结束到出手阶段是一个锐脚"
        },
        {
            index: NumberInt("11"),
            startTime: 23.4,
            endTime: 25.6,
            text: "而观察顶肩射手都会有一段"
        },
        {
            index: NumberInt("12"),
            startTime: 25.6,
            endTime: 27.68,
            text: "几乎垂直上身的行程"
        },
        {
            index: NumberInt("13"),
            startTime: 27.68,
            endTime: 30.04,
            text: "厚度的非常圆润丝滑的感觉"
        },
        {
            index: NumberInt("14"),
            startTime: 30.04,
            endTime: 32.08,
            text: "可能会说听起来容易但不知道打改"
        },
        {
            index: NumberInt("15"),
            startTime: 32.08,
            endTime: 34.32,
            text: "换一下修改磁路今天"
        },
        {
            index: NumberInt("16"),
            startTime: 34.32,
            endTime: 36.48,
            text: "一个完整球篮就一秒钟左右的时间"
        },
        {
            index: NumberInt("17"),
            startTime: 36.48,
            endTime: 37.48,
            text: "很难做到"
        },
        {
            index: NumberInt("18"),
            startTime: 37.48,
            endTime: 39.24,
            text: "只改这么一小段"
        },
        {
            index: NumberInt("19"),
            startTime: 39.24,
            endTime: 43.64,
            text: "所以就需要拉大行程提升溶脱空间"
        },
        {
            index: NumberInt("20"),
            startTime: 43.64,
            endTime: 47.76,
            text: "比如触地投篮"
        },
        {
            index: NumberInt("21"),
            startTime: 47.76,
            endTime: 50.36,
            text: "这样就有充足的时间去做出上下肢"
        },
        {
            index: NumberInt("22"),
            startTime: 50.36,
            endTime: 53.04,
            text: "一起扫身这一段行程"
        },
        {
            index: NumberInt("23"),
            startTime: 53.04,
            endTime: 56.68,
            text: "除了以后再慢慢提速"
        },
        {
            index: NumberInt("24"),
            startTime: 56.68,
            endTime: 58.24,
            text: "同款训练的方式"
        },
        {
            index: NumberInt("25"),
            startTime: 58.24,
            endTime: 60.76,
            text: "在纳什的训练也曾经提到过"
        },
        {
            index: NumberInt("26"),
            startTime: 60.76,
            endTime: 61.84,
            text: "当然千人千样"
        },
        {
            index: NumberInt("27"),
            startTime: 61.84,
            endTime: 63.48,
            text: "有时候也少不了其他训练的缝处"
        },
        {
            index: NumberInt("28"),
            startTime: 63.48,
            endTime: 64.88,
            text: "才能找到法力感受"
        },
        {
            index: NumberInt("29"),
            startTime: 64.88,
            endTime: 65.84,
            text: "如果自己捉挣不了"
        },
        {
            index: NumberInt("30"),
            startTime: 65.84,
            endTime: 67,
            text: "也可以评论私信我"
        },
        {
            index: NumberInt("31"),
            startTime: 67,
            endTime: 69.08,
            text: "但一定成这样流畅轻松的法力"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-04-29T14:41:18.716Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69ff4e86e7cccf323c3b1f6b"),
    videoId: NumberInt("36"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 2.3,
            text: "初學必學的九種技術"
        },
        {
            index: NumberInt("2"),
            startTime: 2.3,
            endTime: 6.08,
            text: "一 正手弓"
        },
        {
            index: NumberInt("3"),
            startTime: 6.08,
            endTime: 9.3,
            text: "二 防守撥"
        },
        {
            index: NumberInt("4"),
            startTime: 9.3,
            endTime: 12.7,
            text: "三 正手戳"
        },
        {
            index: NumberInt("5"),
            startTime: 12.7,
            endTime: 16.4,
            text: "四 防守戳"
        },
        {
            index: NumberInt("6"),
            startTime: 16.4,
            endTime: 21,
            text: "五 正手拉"
        },
        {
            index: NumberInt("7"),
            startTime: 21,
            endTime: 25.6,
            text: "六 防守拉"
        },
        {
            index: NumberInt("8"),
            startTime: 29,
            endTime: 30.5,
            text: "七 正手防"
        },
        {
            index: NumberInt("9"),
            startTime: 30.5,
            endTime: 33.8,
            text: "八 防守衛防"
        },
        {
            index: NumberInt("10"),
            startTime: 33.8,
            endTime: 36.6,
            text: "九 打高球"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-05-09T15:11:02.169Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69ff5140f4b09618933dc8ea"),
    videoId: NumberInt("37"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 5.28,
            text: "前段时间我揉了7视频吐槽国内现代的献情教材"
        },
        {
            index: NumberInt("2"),
            startTime: 5.28,
            endTime: 9.32,
            text: "红技大学第6版紫皮现代不仅编排混乱"
        },
        {
            index: NumberInt("3"),
            startTime: 9.32,
            endTime: 12,
            text: "而且内容和势力都太少"
        },
        {
            index: NumberInt("4"),
            startTime: 12,
            endTime: 15,
            text: "根本没有讲到现代的核心思想"
        },
        {
            index: NumberInt("5"),
            startTime: 15,
            endTime: 19.76,
            text: "随后就有很多朋友问我那现代到底应该怎么学呢"
        },
        {
            index: NumberInt("6"),
            startTime: 19.76,
            endTime: 21.56,
            text: "我们不妨来听听MIT"
        },
        {
            index: NumberInt("7"),
            startTime: 21.56,
            endTime: 28,
            text: "也就是美国的著名学府麻省理工学院的吉尔伯特斯朗教授"
        },
        {
            index: NumberInt("8"),
            startTime: 28,
            endTime: 30.64,
            text: "是如何回答这个问题的"
        },
        {
            index: NumberInt("9"),
            startTime: 30.64,
            endTime: 35.32,
            text: "吉尔伯特斯特朗教授是美国著名的数学教育家"
        },
        {
            index: NumberInt("10"),
            startTime: 35.32,
            endTime: 39.08,
            text: "在MIT教了60多年现代数"
        },
        {
            index: NumberInt("11"),
            startTime: 39.08,
            endTime: 42.6,
            text: "也是三本现性代数学习圣经"
        },
        {
            index: NumberInt("12"),
            startTime: 42.6,
            endTime: 44.2,
            text: "现性代数捣乱"
        },
        {
            index: NumberInt("13"),
            startTime: 44.2,
            endTime: 49.52,
            text: "现性代数基金应用和这本现性代数与数据数据的作者"
        },
        {
            index: NumberInt("14"),
            startTime: 49.52,
            endTime: 53.32,
            text: "第一本书是MIT本科现代科人的教材"
        },
        {
            index: NumberInt("15"),
            startTime: 53.32,
            endTime: 55.32,
            text: "还尽量无须多言"
        },
        {
            index: NumberInt("16"),
            startTime: 55.32,
            endTime: 59.2,
            text: "第二本书则详细显明了现代的应用价值"
        },
        {
            index: NumberInt("17"),
            startTime: 59.2,
            endTime: 65,
            text: "其他那本书以后你再也不会迷惑于现代到底有什么用"
        },
        {
            index: NumberInt("18"),
            startTime: 65,
            endTime: 69.4,
            text: "而这第三本书则是四川老爷子仅跟时代潮流"
        },
        {
            index: NumberInt("19"),
            startTime: 69.4,
            endTime: 72.4,
            text: "在2019年出版的最新逆作"
        },
        {
            index: NumberInt("20"),
            startTime: 72.4,
            endTime: 78.48,
            text: "我手里拿的是这本书由清华大学出版社翻译的出版的中文版"
        },
        {
            index: NumberInt("21"),
            startTime: 78.48,
            endTime: 82.84,
            text: "很多朋友去现在的时候遇到的最大困难在于"
        },
        {
            index: NumberInt("22"),
            startTime: 82.84,
            endTime: 86.24,
            text: "这些枯燥的现代理论到底有什么用"
        },
        {
            index: NumberInt("23"),
            startTime: 86.24,
            endTime: 91.24,
            text: "由清华大学出版社出版这本现性代数与数据学习"
        },
        {
            index: NumberInt("24"),
            startTime: 91.24,
            endTime: 92.48,
            text: "跟你回答"
        },
        {
            index: NumberInt("25"),
            startTime: 92.48,
            endTime: 95.88,
            text: "众所周知人工智能是当代显学"
        },
        {
            index: NumberInt("26"),
            startTime: 95.88,
            endTime: 100.84,
            text: "所以现性代数与数据学习这本书也仅跟前沿"
        },
        {
            index: NumberInt("27"),
            startTime: 100.84,
            endTime: 102.4,
            text: "在最后一章"
        },
        {
            index: NumberInt("28"),
            startTime: 102.4,
            endTime: 106.24,
            text: "发的格外多的笔墨以人工智能做结尾"
        },
        {
            index: NumberInt("29"),
            startTime: 106.24,
            endTime: 112.2,
            text: "他在最后一章详细介绍了人工智能的底层数学框架圣经网络"
        },
        {
            index: NumberInt("30"),
            startTime: 112.2,
            endTime: 116.28,
            text: "把现性代数和人工智能技术结合起来"
        },
        {
            index: NumberInt("31"),
            startTime: 116.28,
            endTime: 119.64,
            text: "直接展示现代的巨大应用价值"
        },
        {
            index: NumberInt("32"),
            startTime: 119.64,
            endTime: 122.08,
            text: "再烧烊前看一点"
        },
        {
            index: NumberInt("33"),
            startTime: 122.08,
            endTime: 123.84,
            text: "在第五章和第六章"
        },
        {
            index: NumberInt("34"),
            startTime: 123.84,
            endTime: 127.56,
            text: "这本书则详细介绍了举证在概律论"
        },
        {
            index: NumberInt("35"),
            startTime: 127.56,
            endTime: 130.68,
            text: "统计学和优化理论的应用"
        },
        {
            index: NumberInt("36"),
            startTime: 130.68,
            endTime: 133.12,
            text: "在斯特党老爷子姐讲当地的下呢"
        },
        {
            index: NumberInt("37"),
            startTime: 133.12,
            endTime: 136.76,
            text: "举证不再只是一张看似无用的数表"
        },
        {
            index: NumberInt("38"),
            startTime: 136.76,
            endTime: 140.44,
            text: "和求解现性方程组使使用的工具"
        },
        {
            index: NumberInt("39"),
            startTime: 140.44,
            endTime: 146.04,
            text: "更值得一提的是优化方法也是举证理论的重要应用之一"
        },
        {
            index: NumberInt("40"),
            startTime: 146.04,
            endTime: 147.64,
            text: "在第三章和第四章"
        },
        {
            index: NumberInt("41"),
            startTime: 147.64,
            endTime: 152.04,
            text: "这本书则讲解了压缩感知技术和图像处理"
        },
        {
            index: NumberInt("42"),
            startTime: 152.04,
            endTime: 153.44,
            text: "TU的相关设计基础"
        },
        {
            index: NumberInt("43"),
            startTime: 153.44,
            endTime: 156.72,
            text: "可以说斯特党老爷子用这一本书"
        },
        {
            index: NumberInt("44"),
            startTime: 156.72,
            endTime: 159.76,
            text: "以现性代数为出发点"
        },
        {
            index: NumberInt("45"),
            startTime: 159.76,
            endTime: 162.8,
            text: "把广义上的好几个应用分支"
        },
        {
            index: NumberInt("46"),
            startTime: 162.8,
            endTime: 166,
            text: "概据论、统计学、最优化技术学习"
        },
        {
            index: NumberInt("47"),
            startTime: 166,
            endTime: 169.64,
            text: "压缩感知、图像处理全都给串起来了"
        },
        {
            index: NumberInt("48"),
            startTime: 169.64,
            endTime: 171.6,
            text: "你在学这本书的时候"
        },
        {
            index: NumberInt("49"),
            startTime: 171.6,
            endTime: 174.28,
            text: "提到了可不止是干巴的现性代数"
        },
        {
            index: NumberInt("50"),
            startTime: 174.28,
            endTime: 176.88,
            text: "更是所谓的应用数学基础"
        },
        {
            index: NumberInt("51"),
            startTime: 176.88,
            endTime: 179.24,
            text: "当然可能会有朋友担心"
        },
        {
            index: NumberInt("52"),
            startTime: 179.24,
            endTime: 180.84,
            text: "这本书虽然好"
        },
        {
            index: NumberInt("53"),
            startTime: 180.84,
            endTime: 182.84,
            text: "但如果难度太高"
        },
        {
            index: NumberInt("54"),
            startTime: 182.84,
            endTime: 185.16,
            text: "不适合自己就要怎么办"
        },
        {
            index: NumberInt("55"),
            startTime: 185.16,
            endTime: 188.52,
            text: "作为一名经验丰富的数学教育家"
        },
        {
            index: NumberInt("56"),
            startTime: 188.52,
            endTime: 192.36,
            text: "斯特党老爷子贴心的用这本书的前两章"
        },
        {
            index: NumberInt("57"),
            startTime: 192.36,
            endTime: 194.48,
            text: "配套了大量其题"
        },
        {
            index: NumberInt("58"),
            startTime: 194.48,
            endTime: 198,
            text: "来讲解出这本书的时候需要的修计组学知识"
        },
        {
            index: NumberInt("59"),
            startTime: 198,
            endTime: 202.32,
            text: "可以说不这本书需要的基础知识等于零"
        },
        {
            index: NumberInt("60"),
            startTime: 202.32,
            endTime: 203.6,
            text: "你在读它时"
        },
        {
            index: NumberInt("61"),
            startTime: 203.6,
            endTime: 207.28,
            text: "根本不用担心数学基础不够跟不上"
        },
        {
            index: NumberInt("62"),
            startTime: 207.28,
            endTime: 210.16,
            text: "那么本期视频到此为止的"
        },
        {
            index: NumberInt("63"),
            startTime: 210.16,
            endTime: 212.32,
            text: "如果各位朋友对这本书感兴趣的话"
        },
        {
            index: NumberInt("64"),
            startTime: 212.32,
            endTime: 213.72,
            text: "可以出说至顶、连接"
        },
        {
            index: NumberInt("65"),
            startTime: 213.72,
            endTime: 217.84,
            text: "我相信大家一定能从这本大多数的收获"
        },
        {
            index: NumberInt("66"),
            startTime: 217.84,
            endTime: 220.72,
            text: "那么喜欢方面可以点赞、收藏、投币"
        },
        {
            index: NumberInt("67"),
            startTime: 220.72,
            endTime: 222.04,
            text: "关注一键三连"
        },
        {
            index: NumberInt("68"),
            startTime: 222.04,
            endTime: 225.4,
            text: "没有币的话点一个免费的赞也可以"
        },
        {
            index: NumberInt("69"),
            startTime: 225.4,
            endTime: 227.76,
            text: "你的支持就是我"
        },
        {
            index: NumberInt("70"),
            startTime: 227.76,
            endTime: 228.76,
            text: "最大的動力"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-05-09T15:22:40.53Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
db.getCollection("subtitles").insert([ {
    _id: ObjectId("69ff57b4f90029495d756d0a"),
    videoId: NumberInt("38"),
    language: "zh-CN",
    languageName: "中文",
    format: "srt",
    content: [
        {
            index: NumberInt("1"),
            startTime: 0,
            endTime: 1.88,
            text: "这是一个开源本地部署的"
        },
        {
            index: NumberInt("2"),
            startTime: 1.88,
            endTime: 3.6,
            text: "AI短视频生成工具"
        },
        {
            index: NumberInt("3"),
            startTime: 3.6,
            endTime: 6.2,
            text: "目前已经拿下了49.8K的新标"
        },
        {
            index: NumberInt("4"),
            startTime: 6.2,
            endTime: 8.2,
            text: "只要一个主题或者关键词"
        },
        {
            index: NumberInt("5"),
            startTime: 8.2,
            endTime: 10.4,
            text: "就能全自动生成高清短视频"
        },
        {
            index: NumberInt("6"),
            startTime: 10.4,
            endTime: 13.28,
            text: "涵盖了文案、视频、字幕和BGM"
        },
        {
            index: NumberInt("7"),
            startTime: 13.28,
            endTime: 15.52,
            text: "甚至最低紧需一台四核CPU"
        },
        {
            index: NumberInt("8"),
            startTime: 15.52,
            endTime: 17.12,
            text: "内存4G以上的电脑"
        },
        {
            index: NumberInt("9"),
            startTime: 17.12,
            endTime: 18.92,
            text: "就连显卡都是非必须"
        },
        {
            index: NumberInt("10"),
            startTime: 18.92,
            endTime: 20.92,
            text: "普通个人电脑即可运行"
        },
        {
            index: NumberInt("11"),
            startTime: 20.92,
            endTime: 22.72,
            text: "支持生成横屏和竖屏"
        },
        {
            index: NumberInt("12"),
            startTime: 22.72,
            endTime: 24.4,
            text: "支持批量生成视频"
        },
        {
            index: NumberInt("13"),
            startTime: 24.4,
            endTime: 26.16,
            text: "支持中英文两种语言"
        },
        {
            index: NumberInt("14"),
            startTime: 26.16,
            endTime: 27.52,
            text: "还支持通易签问"
        },
        {
            index: NumberInt("15"),
            startTime: 27.52,
            endTime: 29.24,
            text: "欧拉曼等大模型揭入"
        },
        {
            index: NumberInt("16"),
            startTime: 29.24,
            endTime: 31.24,
            text: "几百种音色随便选择"
        },
        {
            index: NumberInt("17"),
            startTime: 31.24,
            endTime: 33.24,
            text: "字幕也会自动匹配视频画面"
        },
        {
            index: NumberInt("18"),
            startTime: 33.24,
            endTime: 35.24,
            text: "它采用的是智能缝合技术"
        },
        {
            index: NumberInt("19"),
            startTime: 35.24,
            endTime: 37.24,
            text: "将互联网上的高质量素材"
        },
        {
            index: NumberInt("20"),
            startTime: 37.24,
            endTime: 39,
            text: "进行智能组合和编辑"
        },
        {
            index: NumberInt("21"),
            startTime: 39,
            endTime: 40.28,
            text: "保证了视频质量"
        },
        {
            index: NumberInt("22"),
            startTime: 40.28,
            endTime: 41.8,
            text: "又避免了版权问题"
        },
        {
            index: NumberInt("23"),
            startTime: 41.8,
            endTime: 43.32,
            text: "完整的NVC架构"
        },
        {
            index: NumberInt("24"),
            startTime: 43.32,
            endTime: 44.56,
            text: "代码结构清晰"
        },
        {
            index: NumberInt("25"),
            startTime: 44.56,
            endTime: 46.12,
            text: "二次开发非常简单"
        },
        {
            index: NumberInt("26"),
            startTime: 46.12,
            endTime: 48.12,
            text: "感兴趣的可以研究研究"
        }
    ],
    isDefault: true,
    uploadTime: ISODate("2026-05-09T15:50:12.758Z"),
    status: NumberInt("3"),
    source: "whisper",
    _class: "com.mybilibili.common.entity.Subtitle"
} ]);
