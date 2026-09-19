# -*- coding: utf-8 -*-
"""
向 cookbook 数据库加入一批家常菜谱（含分类标签、用餐时段、食材与简单步骤）。

特点：
  - 数据以本文件内的 RECIPES 列表为准，可重复执行（同名菜谱默认跳过）
  - 封面留空（cover_attachment_id = 0）
  - source 记为 'import'，review_status=0（待人工核对）
  - 食材/工具/步骤按 recipes 表的 JSON 列约定序列化

用法：
  python hack/seed_home_recipes.py --dry-run   # 只打印，不落库
  python hack/seed_home_recipes.py             # 正式写入
  python hack/seed_home_recipes.py --replace   # 同名菜谱覆盖更新
"""

import argparse
import json
import os
import sqlite3
import sys
import time

PROJECT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DB_PATH = os.path.join(PROJECT, "manifest", "cookbook.db")

# 标签 id（与 manifest/init.sql 的预置标签一致）
T = {
    "荤菜": 1, "素菜": 2, "水产": 3, "主食": 4, "汤与粥": 5, "早餐": 6,
    "甜品": 7, "饮品": 8, "调味品": 9, "半成品加工": 10, "家常菜": 11,
    "川菜": 12, "粤菜": 13, "湘菜": 14, "鲁菜": 15, "其他": 16,
}

# 用餐时段位掩码：1=早餐 2=午餐 4=晚餐 8=加餐
B, L, D, S = 1, 2, 4, 8
L_D = L | D

# 字段：title, summary, tags, meal_mask, calories, difficulty, servings, cook_minutes,
#       ingredients[(名, 量, 是否可选)], tools[], steps[]
RECIPES = [
    dict(
        title="葱烧豆腐", summary="葱香浓郁、豆腐滑嫩入味的家常下饭菜",
        tags=["素菜", "家常菜", "鲁菜"], meal_mask=L_D, calories=260, difficulty=1,
        servings=2, cook_minutes=20,
        ingredients=[("北豆腐", "400g", 0), ("大葱", "2 根", 0), ("生抽", "15ml", 0),
                     ("蚝油", "10g", 0), ("白糖", "3g", 0), ("淀粉", "5g", 0),
                     ("食用油", "30ml", 0), ("清水", "150ml", 0)],
        tools=["平底锅"],
        steps=["大葱切斜段，葱白葱绿分开放；豆腐切 1.5cm 厚片，用厨房纸吸干表面水分。",
               "平底锅烧热放 20ml 油，下豆腐片中火煎至两面金黄，盛出备用。",
               "锅内留底油，下葱白段小火煸出香味至微黄，再放葱绿略炒。",
               "加生抽、蚝油、白糖和清水烧开，放回豆腐，中小火烧 3 分钟让豆腐入味。",
               "淀粉加水调成水淀粉，沿锅边淋入勾薄芡，汤汁浓稠后关火装盘。"],
    ),
    dict(
        title="鱼香肉丝", summary="酸甜微辣、肉丝滑嫩的川菜经典",
        tags=["荤菜", "川菜", "家常菜"], meal_mask=L_D, calories=420, difficulty=2,
        servings=2, cook_minutes=30,
        ingredients=[("猪里脊", "250g", 0), ("木耳", "30g", 0), ("胡萝卜", "半根", 0),
                     ("青椒", "1 个", 0), ("泡椒", "20g", 0), ("豆瓣酱", "15g", 0),
                     ("蒜", "3 瓣", 0), ("姜", "1 小块", 0), ("生抽", "15ml", 0),
                     ("醋", "15ml", 0), ("白糖", "15g", 0), ("淀粉", "10g", 0),
                     ("食用油", "40ml", 0)],
        tools=["炒锅"],
        steps=["里脊切细丝，加 5g 淀粉、5ml 生抽抓匀，腌制 10 分钟。",
               "木耳泡发切丝，胡萝卜、青椒切丝，蒜姜切末。",
               "调鱼香汁：生抽 10ml、醋 15ml、白糖 15g、淀粉 5g 加 2 勺清水拌匀。",
               "热锅放油，下肉丝大火滑散至变色盛出。",
               "锅中留油，下泡椒、豆瓣酱炒出红油，放蒜姜末爆香。",
               "倒入木耳、胡萝卜、青椒丝大火翻炒 1 分钟，回锅肉丝。",
               "淋入鱼香汁，快速翻炒至汤汁裹匀收浓即可出锅。"],
    ),
    dict(
        title="丝瓜炒虾仁", summary="虾仁弹嫩、丝瓜清甜的清爽小炒",
        tags=["水产", "家常菜", "粤菜"], meal_mask=L_D, calories=200, difficulty=1,
        servings=2, cook_minutes=15,
        ingredients=[("丝瓜", "2 根", 0), ("虾仁", "200g", 0), ("蒜", "3 瓣", 0),
                     ("盐", "3g", 0), ("料酒", "5ml", 0), ("淀粉", "3g", 0),
                     ("食用油", "20ml", 0)],
        tools=["炒锅"],
        steps=["虾仁去虾线，加料酒、淀粉和少许盐抓匀，腌 5 分钟。",
               "丝瓜去皮切滚刀块，蒜切末。",
               "热锅放油，下虾仁大火滑炒至变粉红立刻盛出。",
               "锅内留油，爆香蒜末，下丝瓜中火翻炒至微微变软出水。",
               "回锅虾仁，加盐调味，翻炒均匀后即可出锅。"],
    ),
    dict(
        title="丝瓜炒鸡蛋", summary="清甜软嫩的快手家常菜",
        tags=["素菜", "家常菜"], meal_mask=B | L_D, calories=180, difficulty=1,
        servings=2, cook_minutes=12,
        ingredients=[("丝瓜", "2 根", 0), ("鸡蛋", "3 个", 0), ("蒜", "2 瓣", 0),
                     ("盐", "3g", 0), ("食用油", "25ml", 0)],
        tools=["炒锅"],
        steps=["鸡蛋打散加少许盐拌匀；丝瓜去皮切滚刀块，蒜切末。",
               "热锅放 15ml 油，倒入蛋液炒至凝固成块，盛出备用。",
               "锅内补油，爆香蒜末，下丝瓜翻炒至变软出水。",
               "回锅鸡蛋，加盐调味，翻炒均匀出锅。"],
    ),
    dict(
        title="芹菜炒肉", summary="芹菜爽脆、肉片咸香的经典家常小炒",
        tags=["荤菜", "家常菜"], meal_mask=L_D, calories=320, difficulty=1,
        servings=2, cook_minutes=20,
        ingredients=[("芹菜", "300g", 0), ("猪瘦肉", "150g", 0), ("蒜", "3 瓣", 0),
                     ("生抽", "10ml", 0), ("淀粉", "5g", 0), ("盐", "3g", 0),
                     ("食用油", "25ml", 0)],
        tools=["炒锅"],
        steps=["猪瘦肉切薄片，加生抽、淀粉抓匀腌 10 分钟。",
               "芹菜去叶切段，较粗的茎剖开；蒜切片。",
               "热锅放油，下肉片中火滑炒至变色盛出。",
               "锅中留油爆香蒜片，下芹菜大火翻炒 2 分钟至断生。",
               "回锅肉片，加盐调味，炒匀出锅。"],
    ),
    dict(
        title="素炒杏鲍菇", summary="口感似鲍鱼、鲜香下饭的素菜",
        tags=["素菜", "家常菜"], meal_mask=L_D, calories=150, difficulty=1,
        servings=2, cook_minutes=15,
        ingredients=[("杏鲍菇", "3 个", 0), ("蒜", "3 瓣", 0), ("小葱", "1 根", 0),
                     ("生抽", "15ml", 0), ("蚝油", "10g", 0), ("盐", "2g", 0),
                     ("食用油", "20ml", 0)],
        tools=["炒锅"],
        steps=["杏鲍菇切 3mm 厚片，再改刀划十字花刀便于入味；蒜切末。",
               "热锅放油，下杏鲍菇中火煎至两面微黄、边缘卷起。",
               "加蒜末爆香，放生抽、蚝油和盐翻炒均匀。",
               "大火收汁至汤汁裹匀，撒葱花出锅。"],
    ),
    dict(
        title="番茄西兰花炒蛋", summary="色彩明快、酸甜营养的家常搭配",
        tags=["素菜", "家常菜"], meal_mask=B | L_D, calories=230, difficulty=1,
        servings=2, cook_minutes=18,
        ingredients=[("番茄", "2 个", 0), ("西兰花", "半颗", 0), ("鸡蛋", "3 个", 0),
                     ("蒜", "2 瓣", 0), ("盐", "3g", 0), ("白糖", "3g", 0),
                     ("食用油", "25ml", 0)],
        tools=["炒锅"],
        steps=["西兰花掰小朵，沸水加盐焯 1 分钟捞出过凉水。",
               "番茄去皮切块，鸡蛋打散加少许盐，蒜切末。",
               "热锅放油，倒蛋液炒成块盛出。",
               "锅内留油爆香蒜末，下番茄中火炒出汁，加白糖和盐。",
               "放入西兰花和鸡蛋，大火翻炒均匀出锅。"],
    ),
    dict(
        title="西葫芦炒牛肉", summary="牛肉滑嫩、西葫芦清鲜的快手下饭菜",
        tags=["荤菜", "家常菜"], meal_mask=L_D, calories=330, difficulty=2,
        servings=2, cook_minutes=25,
        ingredients=[("牛肉", "200g", 0), ("西葫芦", "1 根", 0), ("蒜", "3 瓣", 0),
                     ("生抽", "15ml", 0), ("蚝油", "10g", 0), ("淀粉", "5g", 0),
                     ("料酒", "5ml", 0), ("食用油", "30ml", 0), ("盐", "2g", 0)],
        tools=["炒锅"],
        steps=["牛肉逆纹切薄片，加生抽、料酒、淀粉抓匀，再拌 5ml 油腌 15 分钟。",
               "西葫芦对半切开切片，蒜切末。",
               "热锅放油烧至微微冒烟，下牛肉大火快速滑散至变色，盛出。",
               "锅中留油爆香蒜末，下西葫芦大火翻炒 1 分钟。",
               "回锅牛肉，加蚝油和盐，快速翻炒均匀出锅。"],
    ),
    dict(
        title="青炒菠菜", summary="清淡爽口、几分钟搞定的绿叶菜",
        tags=["素菜", "家常菜"], meal_mask=L_D, calories=90, difficulty=1,
        servings=2, cook_minutes=8,
        ingredients=[("菠菜", "400g", 0), ("蒜", "3 瓣", 0), ("盐", "3g", 0),
                     ("食用油", "15ml", 0)],
        tools=["炒锅"],
        steps=["菠菜择洗干净，沸水焯 30 秒去除草酸，捞出沥干切段。",
               "蒜切末。热锅放油，爆香蒜末。",
               "下菠菜大火快速翻炒 1 分钟，加盐调味即可出锅。"],
    ),
    dict(
        title="清蒸鲈鱼", summary="鱼肉细嫩、原汁原味的粤式蒸菜",
        tags=["水产", "粤菜", "家常菜"], meal_mask=L_D, calories=385, difficulty=2,
        servings=2, cook_minutes=30,
        ingredients=[("鲈鱼", "1 条（约 600g）", 0), ("香葱", "3 根", 0), ("姜", "1 块", 0),
                     ("蒸鱼豉油", "15ml", 0), ("料酒", "10ml", 0), ("食用盐", "5g", 0),
                     ("食用油", "15ml", 0)],
        tools=["蒸锅"],
        steps=["鲈鱼处理干净，两面各划几刀，用盐和料酒抹匀内外，腌 10 分钟。",
               "姜切片和丝，葱白切段、葱绿切丝，葱丝泡冷水备用。",
               "蒸盘底铺葱段和姜片，用筷子将鱼架起（避免浸在汤汁里）。",
               "水开后上锅，大火蒸 8-10 分钟，取出倒掉盘中腥水，去掉姜葱。",
               "鱼身淋上蒸鱼豉油，铺上葱丝和姜丝。",
               "另起锅烧热 15ml 油至冒烟，淋在葱姜丝上激香即可上桌。"],
    ),
    dict(
        title="小米红枣粥", summary="温润养胃、米香枣甜的早餐粥",
        tags=["汤与粥", "早餐"], meal_mask=B | S, calories=180, difficulty=1,
        servings=2, cook_minutes=40,
        ingredients=[("小米", "80g", 0), ("红枣", "8 颗", 0), ("清水", "1200ml", 0)],
        tools=["汤锅"],
        steps=["小米淘洗两遍，红枣去核撕成小块。",
               "锅中放水和红枣，大火烧开。",
               "倒入小米，再次烧开后转小火，盖盖留缝煮 25 分钟。",
               "期间搅拌两三次防粘底，煮至米粒开花、粥体粘稠即可。"],
    ),
    dict(
        title="紫菜蛋花汤", summary="鲜香清淡、三分钟出锅的快手汤",
        tags=["汤与粥", "家常菜"], meal_mask=L_D, calories=80, difficulty=1,
        servings=2, cook_minutes=10,
        ingredients=[("紫菜", "10g", 0), ("鸡蛋", "2 个", 0), ("小葱", "1 根", 0),
                     ("盐", "3g", 0), ("香油", "3ml", 0), ("清水", "700ml", 0)],
        tools=["汤锅"],
        steps=["紫菜撕小块放碗中；鸡蛋打散；葱切花。",
               "锅中加水烧开，加盐调味。",
               "转小火，将蛋液沿筷子缓缓淋入锅中形成蛋花。",
               "关火，把汤冲入紫菜碗中，撒葱花、淋香油即可。"],
    ),
    dict(
        title="山药瘦肉粥", summary="健脾养胃、口感绵软的咸粥",
        tags=["汤与粥", "早餐"], meal_mask=B | S, calories=260, difficulty=1,
        servings=2, cook_minutes=50,
        ingredients=[("大米", "80g", 0), ("山药", "150g", 0), ("猪瘦肉", "100g", 0),
                     ("姜", "2 片", 0), ("盐", "4g", 0), ("清水", "1300ml", 0)],
        tools=["汤锅"],
        steps=["大米淘净，山药去皮切小丁（处理时戴手套防痒），瘦肉切末。",
               "锅中加水和大米，大火烧开转小火煮 20 分钟。",
               "加入山药丁和姜片，继续煮 15 分钟至山药软烂。",
               "放入肉末搅散煮 5 分钟，加盐调味，挑出姜片即可。"],
    ),
    dict(
        title="清炒油麦菜", summary="带着独特清香、脆嫩爽口的绿叶菜",
        tags=["素菜", "家常菜"], meal_mask=L_D, calories=85, difficulty=1,
        servings=2, cook_minutes=8,
        ingredients=[("油麦菜", "400g", 0), ("蒜", "4 瓣", 0), ("盐", "3g", 0),
                     ("食用油", "15ml", 0)],
        tools=["炒锅"],
        steps=["油麦菜洗净沥干，切成长段，菜梗菜叶分开。",
               "蒜切片。热锅放油，爆香蒜片。",
               "先下菜梗大火炒 30 秒，再下菜叶快速翻炒至变软。",
               "加盐调味，立即出锅保持脆嫩。"],
    ),
    dict(
        title="清炖牛腩", summary="汤色清亮、牛腩软烂的原味炖菜",
        tags=["荤菜", "家常菜", "粤菜"], meal_mask=L_D, calories=520, difficulty=2,
        servings=3, cook_minutes=120,
        ingredients=[("牛腩", "800g", 0), ("白萝卜", "1 根", 0), ("姜", "1 块", 0),
                     ("大葱", "1 根", 0), ("八角", "2 个", 0), ("料酒", "20ml", 0),
                     ("盐", "8g", 0), ("清水", "2000ml", 0)],
        tools=["砂锅"],
        steps=["牛腩切 3cm 块，冷水下锅加料酒焯水，撇净浮沫后捞出洗净。",
               "牛腩放入砂锅，加清水、姜片、葱段、八角，大火烧开转小火炖 60 分钟。",
               "白萝卜去皮切滚刀块。",
               "将萝卜块放入锅中，继续炖 30 分钟至牛腩软烂、萝卜透明。",
               "出锅前加盐调味，撒葱花即可。"],
    ),
    dict(
        title="红薯焖饭", summary="薯香四溢、微甜软糯的一锅饭",
        tags=["主食", "家常菜"], meal_mask=L_D, calories=380, difficulty=1,
        servings=2, cook_minutes=40,
        ingredients=[("大米", "200g", 0), ("红薯", "250g", 0), ("清水", "260ml", 0)],
        tools=["电饭煲"],
        steps=["大米淘洗干净，按正常水量浸泡 15 分钟。",
               "红薯去皮切 2cm 见方的块。",
               "把红薯块铺在米上，按下煮饭键。",
               "跳闸后焖 10 分钟，开盖用饭勺翻拌均匀即可。"],
    ),
    dict(
        title="冬瓜排骨汤", summary="清甜解腻、汤鲜味美的家常汤",
        tags=["汤与粥", "荤菜", "家常菜"], meal_mask=L_D, calories=320, difficulty=1,
        servings=3, cook_minutes=70,
        ingredients=[("排骨", "500g", 0), ("冬瓜", "400g", 0), ("姜", "1 块", 0),
                     ("料酒", "15ml", 0), ("盐", "6g", 0), ("清水", "1500ml", 0),
                     ("小葱", "1 根", 0)],
        tools=["汤锅"],
        steps=["排骨冷水下锅，加料酒焯水去血沫，捞出冲净。",
               "排骨和姜片放入锅中，加清水大火烧开转小火炖 40 分钟。",
               "冬瓜去皮去瓤切厚片。",
               "放入冬瓜继续煮 15 分钟至透明软烂。",
               "加盐调味，撒葱花出锅。"],
    ),
    dict(
        title="鲫鱼豆腐汤", summary="汤白鲜浓、鱼肉嫩豆腐滑的滋补汤",
        tags=["水产", "汤与粥", "家常菜"], meal_mask=L_D, calories=290, difficulty=2,
        servings=3, cook_minutes=40,
        ingredients=[("鲫鱼", "1 条（约 400g）", 0), ("嫩豆腐", "300g", 0), ("姜", "1 块", 0),
                     ("小葱", "2 根", 0), ("料酒", "10ml", 0), ("盐", "5g", 0),
                     ("食用油", "25ml", 0), ("清水", "1200ml", 0)],
        tools=["炒锅", "汤锅"],
        steps=["鲫鱼处理干净擦干水分，两面各划两刀；豆腐切块；姜切片。",
               "热锅放油，下鲫鱼中火煎至两面金黄（不要频繁翻动）。",
               "加姜片、料酒，一次性倒入开水，大火滚煮 10 分钟至汤色奶白。",
               "转中火，放入豆腐块煮 10 分钟。",
               "加盐调味，撒葱花出锅。全程不要加冷水，否则汤色不白。"],
    ),
    dict(
        title="清炒荷兰豆", summary="翠绿脆嫩、清甜爽口的小炒",
        tags=["素菜", "家常菜"], meal_mask=L_D, calories=110, difficulty=1,
        servings=2, cook_minutes=12,
        ingredients=[("荷兰豆", "300g", 0), ("蒜", "3 瓣", 0), ("盐", "3g", 0),
                     ("食用油", "15ml", 0)],
        tools=["炒锅"],
        steps=["荷兰豆撕去两侧老筋，洗净沥干。",
               "沸水加少许盐和油，焯 40 秒后捞出过冷水，保持翠绿。",
               "热锅放油爆香蒜片，下荷兰豆大火翻炒 1 分钟。",
               "加盐调味，炒匀出锅。"],
    ),
    dict(
        title="番茄炒鸡蛋", summary="酸甜开胃、国民级的家常第一菜",
        tags=["素菜", "家常菜"], meal_mask=B | L_D, calories=210, difficulty=1,
        servings=2, cook_minutes=12,
        ingredients=[("番茄", "2 个", 0), ("鸡蛋", "3 个", 0), ("小葱", "1 根", 0),
                     ("盐", "3g", 0), ("白糖", "5g", 0), ("食用油", "25ml", 0)],
        tools=["炒锅"],
        steps=["番茄去皮切块（顶部划十字用开水烫一下更好去皮）；鸡蛋打散加少许盐。",
               "热锅放油，倒入蛋液，待边缘凝固后炒成大块盛出。",
               "锅中留油，下番茄中火翻炒出汁，加白糖和盐。",
               "回锅鸡蛋，翻炒让蛋块裹上番茄汁，撒葱花出锅。"],
    ),
    dict(
        title="清蒸鳕鱼", summary="肉质雪白细嫩、无刺易食的清淡蒸鱼",
        tags=["水产", "家常菜"], meal_mask=L_D, calories=230, difficulty=1,
        servings=2, cook_minutes=20,
        ingredients=[("鳕鱼", "300g", 0), ("姜", "1 块", 0), ("小葱", "2 根", 0),
                     ("蒸鱼豉油", "15ml", 0), ("料酒", "10ml", 0), ("盐", "2g", 0),
                     ("食用油", "15ml", 0)],
        tools=["蒸锅"],
        steps=["鳕鱼解冻后擦干，加盐和料酒腌 10 分钟。",
               "盘底铺姜片和葱段，放上鳕鱼，鱼身再铺几片姜。",
               "水开后大火蒸 8 分钟，取出倒掉盘中水分，去掉姜葱。",
               "淋上蒸鱼豉油，铺葱丝。",
               "烧热 15ml 油至冒烟，淋在葱丝上即可。"],
    ),
    dict(
        title="山药乌鸡汤", summary="滋补温润、汤清味鲜的养生汤",
        tags=["汤与粥", "荤菜"], meal_mask=L_D, calories=340, difficulty=2,
        servings=3, cook_minutes=100,
        ingredients=[("乌鸡", "半只", 0), ("山药", "250g", 0), ("红枣", "6 颗", 0),
                     ("枸杞", "10g", 0), ("姜", "1 块", 0), ("料酒", "15ml", 0),
                     ("盐", "6g", 0), ("清水", "1800ml", 0)],
        tools=["砂锅"],
        steps=["乌鸡剁块，冷水下锅加料酒焯水，撇沫后捞出冲净。",
               "鸡块、姜片、红枣放入砂锅，加清水大火烧开转小火炖 60 分钟。",
               "山药去皮切段（戴手套操作），放入锅中继续炖 20 分钟。",
               "加枸杞再煮 5 分钟，出锅前加盐调味。"],
    ),
    dict(
        title="紫米饭", summary="颗粒分明、带天然米香的粗粮主食",
        tags=["主食"], meal_mask=B | L_D, calories=340, difficulty=1,
        servings=2, cook_minutes=45,
        ingredients=[("紫米", "100g", 0), ("大米", "100g", 0), ("清水", "300ml", 0)],
        tools=["电饭煲"],
        steps=["紫米提前用清水浸泡 2 小时（紫米较硬，泡过才易软）。",
               "紫米与大米混合淘洗，放入电饭煲。",
               "加 300ml 水（比纯白米饭略多），按下煮饭键。",
               "跳闸后焖 10 分钟再开盖，用饭勺翻松即可。"],
    ),
    dict(
        title="清炒茼蒿", summary="带着独特蒿香、清口解腻的绿叶菜",
        tags=["素菜", "家常菜"], meal_mask=L_D, calories=80, difficulty=1,
        servings=2, cook_minutes=8,
        ingredients=[("茼蒿", "400g", 0), ("蒜", "4 瓣", 0), ("盐", "3g", 0),
                     ("食用油", "15ml", 0)],
        tools=["炒锅"],
        steps=["茼蒿择洗干净，切成 5cm 长段，茎叶分开。",
               "蒜切片。热锅放油爆香蒜片。",
               "下茼蒿茎部大火炒 30 秒，再下叶子快速翻炒。",
               "加盐调味，叶子一变软立刻出锅，避免出水过多。"],
    ),
    dict(
        title="蒜蓉娃娃菜", summary="蒜香浓郁、清甜软嫩的蒸/炒皆宜菜",
        tags=["素菜", "家常菜", "粤菜"], meal_mask=L_D, calories=120, difficulty=1,
        servings=2, cook_minutes=15,
        ingredients=[("娃娃菜", "2 颗", 0), ("蒜", "1 整头", 0), ("生抽", "15ml", 0),
                     ("蚝油", "10g", 0), ("盐", "2g", 0), ("食用油", "25ml", 0)],
        tools=["炒锅", "蒸锅"],
        steps=["娃娃菜纵向切成 4 瓣，洗净沥干，摆入盘中。",
               "蒜剁成蓉，一半用油小火煸至微黄成金蒜，与另一半生蒜蓉混合。",
               "加生抽、蚝油、盐拌匀成蒜蓉酱，铺在娃娃菜上。",
               "水开后上锅大火蒸 8 分钟，取出即可（也可直接下锅炒软后淋酱）。"],
    ),
    dict(
        title="红烧黄花鱼", summary="色泽红亮、鱼肉鲜嫩入味的下饭菜",
        tags=["水产", "家常菜", "鲁菜"], meal_mask=L_D, calories=430, difficulty=2,
        servings=2, cook_minutes=35,
        ingredients=[("黄花鱼", "1 条（约 500g）", 0), ("姜", "1 块", 0), ("蒜", "4 瓣", 0),
                     ("大葱", "1 根", 0), ("生抽", "20ml", 0), ("老抽", "5ml", 0),
                     ("料酒", "15ml", 0), ("白糖", "10g", 0), ("醋", "5ml", 0),
                     ("淀粉", "5g", 0), ("食用油", "40ml", 0), ("清水", "200ml", 0)],
        tools=["炒锅"],
        steps=["黄花鱼处理干净，两面划刀，用厨房纸擦干，薄薄拍一层淀粉。",
               "热锅放油烧至微冒烟，下鱼中火煎至两面金黄定型后盛出。",
               "锅中留油，下姜片、蒜瓣、葱段爆香。",
               "加生抽、老抽、料酒、白糖、醋和清水烧开。",
               "放回鱼，中小火烧 8 分钟，期间用勺舀汤汁浇在鱼身上。",
               "大火收汁至浓稠，把汤汁淋在鱼上即可。"],
    ),
]


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--replace", action="store_true", help="同名菜谱覆盖更新，而非跳过")
    args = ap.parse_args()

    if not os.path.exists(DB_PATH):
        print("数据库不存在:", DB_PATH)
        return 1

    # 自检：标签合法、字段齐全
    for r in RECIPES:
        assert r["tags"], f"{r['title']} 缺标签"
        for t in r["tags"]:
            assert t in T, f"{r['title']} 标签非法: {t}"
        assert r["ingredients"], f"{r['title']} 缺食材"
        assert r["steps"], f"{r['title']} 缺步骤"
        for ing in r["ingredients"]:
            assert len(ing) == 3 and ing[0], f"{r['title']} 食材格式错误: {ing}"

    con = sqlite3.connect(DB_PATH)
    con.execute("PRAGMA foreign_keys=OFF")
    cur = con.cursor()
    existing = {r[0]: r[1] for r in cur.execute("select title, id from recipes").fetchall()}

    log = []
    n_new = n_upd = n_skip = 0
    now = int(time.time())

    for r in RECIPES:
        title = r["title"]
        ingredients = [{"name": n, "amount": a, "optional": o} for n, a, o in r["ingredients"]]
        steps = [{"content": s, "media": []} for s in r["steps"]]
        payload = dict(
            summary=r["summary"], calories=r["calories"], difficulty=r["difficulty"],
            servings=r["servings"], cook_minutes=r["cook_minutes"], meal_mask=r["meal_mask"],
            ingredients=json.dumps(ingredients, ensure_ascii=False),
            tools=json.dumps(r["tools"], ensure_ascii=False),
            steps=json.dumps(steps, ensure_ascii=False),
        )

        if title in existing:
            if not args.replace:
                n_skip += 1
                log.append((title, f"SKIP 已存在(id={existing[title]})", len(ingredients), len(steps)))
                continue
            if not args.dry_run:
                cur.execute(
                    "update recipes set summary=?, calories=?, difficulty=?, servings=?,"
                    " cook_minutes=?, meal_mask=?, ingredients=?, tools=?, steps=?,"
                    " updated_at=? where title=?",
                    (*payload.values(), now, title))
                rid = existing[title]
                cur.execute("delete from recipe_tags where recipe_id=?", (rid,))
            else:
                rid = existing[title]
            n_upd += 1
            log.append((title, f"REPLACE id={rid}", len(ingredients), len(steps)))
        else:
            if args.dry_run:
                n_new += 1
                log.append((title, "DRY NEW", len(ingredients), len(steps)))
                continue
            cur.execute(
                "insert into recipes (user_id,title,summary,cover_attachment_id,tips,"
                "ingredients,tools,steps,meal_mask,calories,difficulty,servings,cook_minutes,"
                "source,ai_model,review_status,is_deleted,created_at,updated_at)"
                " values (0,?,?,0,'',?,?,?,?,?,?,?,?,'import','',0,0,?,?)",
                (title, r["summary"], payload["ingredients"], payload["tools"], payload["steps"],
                 r["meal_mask"], r["calories"], r["difficulty"], r["servings"],
                 r["cook_minutes"], now, now))
            rid = cur.lastrowid
            n_new += 1
            log.append((title, f"NEW id={rid}", len(ingredients), len(steps)))

        if not args.dry_run:
            for t in r["tags"]:
                cur.execute("insert or ignore into recipe_tags (recipe_id, tag_id) values (?,?)",
                            (rid, T[t]))

    if not args.dry_run:
        con.commit()

    print(f"{'菜名':<14} {'结果':<24} {'食材':>4} {'步骤':>4}  标签")
    print("-" * 92)
    for title, res, ni, ns in log:
        tags = next(r["tags"] for r in RECIPES if r["title"] == title)
        print(f"{title:<14} {res:<24} {ni:>4} {ns:>4}  {'/'.join(tags)}")
    print("-" * 92)
    print(f"新增 {n_new} · 覆盖 {n_upd} · 跳过 {n_skip} | 库中 recipes 总数: "
          f"{cur.execute('select count(*) from recipes').fetchone()[0]}")
    con.close()
    return 0


if __name__ == "__main__":
    sys.exit(main())
