# -*- coding: utf-8 -*-
"""
从 HowToCook-master.zip 的 dishes/aquatic 目录导入菜谱到 cookbook 数据库。

字段映射（依据 manifest/init.sql）：
  title         <- md 一级标题去掉「的做法」
  summary       <- 标题后首段正文（截断到 200 字）
  calories      <- 「预估卡路里：NNN 大卡」
  difficulty    <- 「预估烹饪难度：★…」5 星制 → 3 档（1 简单 / 2 中等 / 3 较难）
  cook_minutes  <- 正文中「约 N 分钟 / N 小时」推断，取不到则 0
  servings      <- 「计算」小节中的份数描述，默认 1
  ingredients   <- 「计算」小节的列表项 → [{"name","amount","optional"}]
  tools         <- 「必备原料和工具」中识别出的厨具/器具
  steps         <- 「操作」小节的编号步骤 → [{"content","media":[attachmentId]}]
  cover_attachment_id / steps[].media <- md 中 ![alt](./x.jpg) 的图片，以 BLOB 写入
                 attachments.content（不落盘），登记 attachments 表
  source        <- 'import'
  tags          <- 固定 3=水产；再按正文关键词补菜系标签

用法：
  python hack/import_howtocook.py           # 正式导入
  python hack/import_howtocook.py --dry-run # 只解析和打印，不落库
"""

import argparse
import hashlib
import json
import os
import random
import re
import sqlite3
import string
import sys
import time
import zipfile

ZIP_PATH = os.environ.get("HOWTOCOOK_ZIP", "HowToCook-master.zip")
PROJECT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DB_PATH = os.path.join(PROJECT, "manifest", "cookbook.db")
SRC_PREFIX = "HowToCook-master/dishes/aquatic/"

# tags 表预置 id
TAG_AQUATIC = 3   # 水产
TAG_HOME = 11     # 家常菜
TAG_SICHUAN = 12  # 川菜
TAG_CANTON = 13   # 粤菜
TAG_HUNAN = 14    # 湘菜
TAG_SHANDONG = 15  # 鲁菜

MIME = {
    "jpg": "image/jpeg", "jpeg": "image/jpeg", "png": "image/png",
    "webp": "image/webp", "gif": "image/gif", "svg": "image/svg+xml",
}

# 「必备原料和工具」里识别为工具的常见词
TOOL_WORDS = [
    "炒锅", "砂锅", "平底锅", "不粘锅", "汤锅", "蒸锅", "高压锅", "电饭煲", "烤箱",
    "微波炉", "蒸屉", "密封袋", "保鲜袋", "厨房纸", "锡纸", "烤盘", "烤架", "竹签",
    "筷子", "漏勺", "笊篱", "炒勺", "锅铲", "铲子", "菜刀", "刀", "砧板", "案板",
    "碗", "盘", "碟", "盆", "纱布", "料理机", "搅拌机", "破壁机", "打蛋器", "温度计",
]


def log(*a):
    print(*a, flush=True)


def rnd_name(ext):
    alphabet = string.ascii_lowercase + string.digits
    return "".join(random.choice(alphabet) for _ in range(17)) + "." + ext


# ---------------------------------------------------------------- 解析

DIFF_MAP = {1: 1, 2: 1, 3: 2, 4: 3, 5: 3}


def clean_md(s):
    """去掉 markdown 行内标记，保留纯文本。"""
    s = re.sub(r"!\[[^\]]*\]\([^)]+\)", "", s)          # 图片
    s = re.sub(r"\[([^\]]*)\]\([^)]+\)", r"\1", s)       # 链接 → 文字
    s = s.replace("**", "").replace("*", "").replace("`", "")
    return s.strip()


def parse_sections(text):
    """按 ## 标题切分正文，返回 {标题: 内容}（保持顺序）。"""
    sections = {}
    order = []
    cur = None
    buf = []
    for line in text.splitlines():
        m = re.match(r"^##\s+(.+?)\s*$", line)
        if m:
            if cur is not None:
                sections[cur] = "\n".join(buf)
            cur = m.group(1).strip()
            order.append(cur)
            buf = []
        elif cur is not None:
            buf.append(line)
    if cur is not None:
        sections[cur] = "\n".join(buf)
    return sections, order


def parse_intro(text):
    """标题后、「## 必备原料和工具」之前的所有段落作为简介素材。"""
    body = text.split("\n", 1)[1] if "\n" in text else ""
    body = re.split(r"^##\s+", body, maxsplit=1, flags=re.M)[0]
    return body


def parse_difficulty(text):
    m = re.search(r"预估烹饪难度[：:]\s*(★+)", text)
    if not m:
        return 0
    return DIFF_MAP.get(len(m.group(1)), 0)


def parse_calories(text):
    m = re.search(r"预估卡路里[：:]\s*(\d+)", text)
    return int(m.group(1)) if m else 0


CN_NUM = {"一": 1, "两": 2, "二": 2, "三": 3, "四": 4, "五": 5,
          "六": 6, "七": 7, "八": 8, "九": 9, "十": 10, "半": 0.5}


def cn_to_int(s):
    """中文数字转整数（支持 一/两/三…/十/二十/二十五）。"""
    s = s.strip()
    if not s:
        return None
    if s in CN_NUM:
        return CN_NUM[s]
    if "十" in s:
        parts = s.split("十")
        left = CN_NUM.get(parts[0], 1) if parts[0] else 1
        right = CN_NUM.get(parts[1], 0) if len(parts) > 1 and parts[1] else 0
        return left * 10 + right
    if all(ch in CN_NUM for ch in s):
        return sum(CN_NUM[ch] for ch in s)
    return None


def parse_cook_minutes(text):
    """从正文推断总耗时，取所有「约 N 分钟/N 小时」中最大的那个。支持中文数字。"""
    cands = []
    # 阿拉伯数字：约 2-2.5 小时 / 约 30 分钟
    for m in re.finditer(r"约\s*(?:需|需要)?\s*(\d+(?:\.\d+)?)\s*(?:[–~-]\s*(\d+(?:\.\d+)?)\s*)?小时", text):
        cands.append(float(m.group(2) or m.group(1)) * 60)
    m = re.search(r"(?:大约|预计|约|需|需要)\s*(?:只需|需要)?\s*(\d+)\s*分钟", text)
    if m:
        cands.append(float(m.group(1)))
    # 中文数字：大约只需三十分钟 / 约二小时 / 约半小时
    for m in re.finditer(r"(?:大约|预计|约|需|需要)\s*(?:只需|需要)?\s*([一二两三四五六七八九十半]+)\s*小时", text):
        v = cn_to_int(m.group(1))
        if v:
            cands.append(v * 60)
    for m in re.finditer(r"(?:大约|预计|约|需|需要)\s*(?:只需|需要)?\s*([一二两三四五六七八九十半]+)\s*分钟", text):
        v = cn_to_int(m.group(1))
        if v:
            cands.append(float(v))
    return int(max(cands)) if cands else 0


def parse_summary(intro):
    """简介：取首段真正的正文（跳过图片行、标题、难度/卡路里行），截断 200 字。"""
    for ln in intro.splitlines():
        s = ln.strip()
        if not s:
            continue
        if s.startswith("#"):
            continue
        if s.startswith("预估"):
            continue
        # 纯图片行（可能带列表符号或表头）
        if re.match(r"^[-*+\s]*!\[", s):
            continue
        if re.match(r"^[-*+\s]*$", s):
            continue
        # 表格行
        if s.startswith("|"):
            continue
        s = clean_md(s)
        s = re.sub(r"\s+", "", s)
        if s:
            return s[:200]
    return ""


def parse_servings(calc_text, text):
    """从「计算」或正文中提取份量。"""
    m = re.search(r"每\s*(\d+)\s*份", calc_text)
    if m:
        return int(m.group(1))
    m = re.search(r"一份正好够\s*(\d+)\s*个人", calc_text)
    if m:
        return int(m.group(1))
    m = re.search(r"够\s*(\d+)\s*人食用", text)
    if m:
        return int(m.group(1))
    m = re.search(r"每份", calc_text)
    if m:
        return 1
    return 1


def split_name_amount(item):
    """
    把「五花肉 500g」「鲈鱼 一条」「盐 10g」拆成 (name, amount)。
    中文菜谱里名字与用量常以空格分隔，用量以数字/量词开头。
    """
    item = clean_md(item)
    item = item.strip("。；;，,")
    if not item:
        return None
    # 去掉行首的序号/项目符号残留
    item = re.sub(r"^\d+[.、)]\s*", "", item)
    if not item:
        return None
    parts = item.split()
    if len(parts) >= 2:
        # 从右往左找第一个「像用量」的片段
        for i in range(1, len(parts)):
            tail = " ".join(parts[i:])
            if re.match(r"^[0-9０-９]", tail) or re.match(
                    r"^(适量|少许|若干|一把|几|半|一|两|三|四|五|六|七|八|九|十)", tail):
                return " ".join(parts[:i]).strip("，,、"), tail.strip("，,、")
        # 没有明显用量：第一段当名字，其余当用量
        return parts[0].strip("，,、"), " ".join(parts[1:]).strip("，,、")
    # 无空格：尝试在「中文名字 + 中文数量词」之间切开，如「鱼头一个」「八角两个」
    m = re.match(r"^([^0-9０-９]{2,}?)([一二两三四五六七八九十半]+\s*(?:个|只|条|根|块|瓣|片|把|棵|颗|粒|斤|两|头|张|段|勺|杯|碗|份|束|支|朵|枚|大卡|克|毫升|升))$", item)
    if m:
        return m.group(1).strip("，,、"), m.group(2).strip()
    return item, ""


def looks_like_name_amount(piece):
    """判断一个逗号片段是否自带「名字+用量」结构。"""
    na = split_name_amount(piece)
    if not na:
        return False
    name, amount = na
    return bool(name) and bool(amount) and len(name) <= 24


def tidy_name_amount(name, amount):
    """
    规整名字/用量：
    - 名字里残留的括号说明（如「蔬菜（比如土豆片/豆芽）」）保留括号但要整理
    - 用量里夹带的说明文字（如「250g 份数（建议 1-2 人份）」）截到第一个括号/空白前
    """
    name = name.strip("，,、 ")
    amount = amount.strip("，,、 ")
    # 用量：只保留前面的数字+单位部分
    if amount:
        # 先按逗号切开，取第一段（「片，450g」→「片」；「250g 份数（…）」→「250g 份数（…）」）
        first = re.split(r"[，,]", amount)[0].strip()
        m = re.match(r"^([0-9０-９][0-9０-９./~\-]*\s*[A-Za-z\u4e00-\u9fff]{0,3})", first)
        if m:
            compact = m.group(1).strip()
            if len(compact) <= 12:
                amount = compact
        else:
            amount = first
        amount = re.sub(r"[（(].*?[）)]", "", amount).strip()
    # 名字以数字结尾且用量以量词开头：说明名字尾部混进了数量词，补进用量
    m = re.match(r"^(.+?)[，,]\s*([0-9０-９]+)$", name)
    if m and amount:
        name = m.group(1).strip()
        amount = f"{m.group(2)}{amount}".strip()
    # 名字里残留的逗号统一成空格，读起来更自然
    name = re.sub(r"[，,]\s*", " ", name).strip()
    return name, amount


def parse_ingredients(calc_text):
    """「计算」小节里的 - / * 列表项 → 食材。"""
    out = []
    seen = set()
    for line in calc_text.splitlines():
        s = line.strip()
        if not re.match(r"^[-*+]\s+", s):
            continue
        raw = re.sub(r"^[-*+]\s+", "", s)
        if not raw or raw.startswith("注："):
            continue
        # 只有当一个列表项里的每个逗号片段都自带「名+量」时才拆开，
        # 否则视为「名字 用量」整体（如「黑鳕鱼，带皮，2 片，450g」）
        pieces = re.split(r"[，,、]+\s*", raw)
        target = pieces if len(pieces) > 1 and all(looks_like_name_amount(p) for p in pieces) else [raw]
        for piece in target:
            na = split_name_amount(piece)
            if not na:
                continue
            name, amount = na
            # 末尾用量没拆出来时，再试一次「整体 名字+量词」模式
            if not amount:
                m = re.match(r"^(.+?)\s+([0-9０-９][0-9０-９./~\-]*\s*\S*)$", name)
                if m:
                    name, amount = m.group(1), m.group(2)
            # 仍未拆出：尝试最后一个逗号边界（如「青葱，葱白，25g」）
            if not amount and ("，" in name or "," in name):
                m = re.match(r"^(.+?)[，,]\s*([0-9０-９][^，,]*)$", name)
                if m:
                    nm, am = m.group(1), m.group(2)
                    # 名字里若还有逗号，取最后一段作为更贴切的食材名
                    tail = re.split(r"[，,]", nm)[-1].strip()
                    if tail:
                        nm = tail
                    name, amount = nm, am
            if not name or name in seen:
                continue
            name, amount = tidy_name_amount(name, amount)
            # 过滤纯说明性文字
            if len(name) > 24 or not name:
                continue
            seen.add(name)
            optional = 1 if ("可选" in piece or "自选" in piece) else 0
            out.append({"name": name, "amount": amount, "optional": optional})
    return out


def parse_tools(text):
    """从「必备原料和工具」中识别工具词。"""
    found = []
    tools_section = ""
    m = re.search(r"^##\s+必备原料和工具\s*$(.*?)(?=^##\s)", text, re.M | re.S)
    if m:
        tools_section = m.group(1)
    for w in TOOL_WORDS:
        if w in tools_section and w not in found:
            found.append(w)
    return found


def parse_steps(ops_text):
    """
    「操作」小节 → 步骤列表。
    返回 [{"content": str, "imgs": [相对图片路径], "sub": str|None}]
    子标题（###）作为上下文前缀拼进步骤内容。
    """
    steps = []
    sub = None
    cur = None
    for line in ops_text.splitlines():
        mh = re.match(r"^###\s+(.+?)\s*$", line)
        if mh:
            if cur:
                steps.append(cur)
                cur = None
            sub = mh.group(1).strip()
            continue
        s = line.strip()
        if not s:
            continue
        # 编号步骤
        mn = re.match(r"^(\d+)[.、)]\s*(.*)$", s)
        if mn:
            if cur:
                steps.append(cur)
            cur = {"content": mn.group(2).strip(), "imgs": [], "sub": sub}
            continue
        # 图片行 / 图片独立成行
        if re.match(r"^!\[", s):
            if cur is not None:
                cur["imgs"].append(s)
            continue
        # 纯图片列表项
        if re.match(r"^[-*+]\s*!\[", s):
            if cur is not None:
                cur["imgs"].append(re.sub(r"^[-*+]\s*", "", s))
            continue
        # 续行（缩进的注/子项）
        if cur is not None:
            if re.match(r"^[-*+]\s+", s) or s.startswith("*") or s.startswith("注"):
                extra = clean_md(re.sub(r"^[-*+\s]+", "", s))
                if extra:
                    cur["content"] += " " + extra
            elif not re.match(r"^[#>|]", s):
                cur["content"] += " " + clean_md(s)
    if cur:
        steps.append(cur)

    # 清理：内容为空的步骤丢弃（只含图片的保留图）
    result = []
    for st in steps:
        content = clean_md(st["content"])
        content = re.sub(r"\s+", " ", content).strip()
        imgs = st["imgs"]
        if not content and not imgs:
            continue
        if st["sub"] and content:
            content = f"[{st['sub']}] {content}"
        elif st["sub"] and not content:
            content = f"[{st['sub']}]"
        result.append({"content": content, "imgs": imgs})
    return result


# 正文里所有图片引用（含操作区之外的封面）
def all_img_refs(text):
    return re.findall(r"!\[([^\]]*)\]\(([^)]+)\)", text)


# ---------------------------------------------------------------- 标签

def detect_tags(text, title):
    tags = {TAG_AQUATIC}
    if re.search(r"川菜|四川|麻辣|水煮鱼|辣子", text):
        tags.add(TAG_SICHUAN)
    if re.search(r"粤式|粤菜|广东|白灼|清蒸鲈鱼|顺德", text):
        tags.add(TAG_CANTON)
    if re.search(r"鲁菜|山东|葱烧海参", text):
        tags.add(TAG_SHANDONG)
    if re.search(r"湘菜|湖南|剁椒", text):
        tags.add(TAG_HUNAN)
    if re.search(r"家常", text):
        tags.add(TAG_HOME)
    return sorted(tags)


# ---------------------------------------------------------------- 主流程

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--dry-run", action="store_true")
    ap.add_argument("--reset", action="store_true",
                    help="先清除此前 source='import' 的菜谱及其附件记录与文件，再重新导入")
    args = ap.parse_args()

    if not os.path.exists(DB_PATH):
        log("数据库不存在:", DB_PATH)
        return 1

    con = sqlite3.connect(DB_PATH)
    con.execute("PRAGMA foreign_keys=OFF")
    cur = con.cursor()

    if args.reset and not args.dry_run:
        old = [r[0] for r in cur.execute(
            "select id from recipes where source='import'").fetchall()]
        # 只清理本脚本写入的附件：storage_path 形如 <yyyyMM>/<随机名>（不含 demo/ 与 .svg）
        paths = [r[0] for r in cur.execute(
            "select storage_path from attachments where storage_path not like 'demo/%'"
            " and storage_path not like '%.svg'").fetchall()]
        cur.execute("delete from recipe_tags where recipe_id in"
                    " (select id from recipes where source='import')")
        cur.execute("delete from recipes where source='import'")
        cur.execute("delete from attachments where storage_path not like 'demo/%'"
                    " and storage_path not like '%.svg'")
        con.commit()
        # blob 化后附件内容随记录删除，无磁盘文件可清
        log(f"reset: 删除菜谱 {len(old)} 条、附件记录 {len(paths)} 条")

    existing_titles = {r[0] for r in cur.execute("select title from recipes").fetchall()}
    log("库中现有菜谱:", len(existing_titles), sorted(existing_titles))

    with zipfile.ZipFile(ZIP_PATH) as z:
        names = z.namelist()
        mds = sorted(n for n in names
                     if n.startswith(SRC_PREFIX) and n.endswith(".md"))

        stats = {"parsed": 0, "inserted": 0, "skipped": 0, "img": 0}
        report = []

        for md_path in mds:
            raw_title = md_path[len(SRC_PREFIX):]
            text = z.read(md_path).decode("utf-8")
            dirname = md_path.rsplit("/", 1)[0]

            m = re.match(r"^#\s+(.+?)(?:的做法)?\s*$", text.splitlines()[0])
            if m:
                title = m.group(1).strip()
            else:
                title = os.path.basename(md_path)[:-3]  # 去掉 .md，取文件名
            title = re.sub(r"的做法$", "", title).strip()

            sections, order = parse_sections(text)
            intro = parse_intro(text)
            calc = sections.get("计算", "")
            ops = sections.get("操作", "")

            summary = parse_summary(intro)
            difficulty = parse_difficulty(text)
            calories = parse_calories(text)
            cook_minutes = parse_cook_minutes(intro)
            servings = parse_servings(calc, text)
            ingredients = parse_ingredients(calc)
            tools = parse_tools(text)
            steps = parse_steps(ops)
            tags = detect_tags(text, title)

            # 该菜谱引用的所有图片（相对路径 → zip 内绝对路径）
            img_map = {}  # 相对路径 -> (alt, zip_path, ext)
            for alt, rel in all_img_refs(text):
                if rel.startswith("http"):
                    continue
                clean = rel.lstrip("./")
                zp = f"{dirname}/{clean}"
                if zp in names:
                    ext = clean.rsplit(".", 1)[-1].lower()
                    img_map[rel] = (alt, zp, ext)

            stats["parsed"] += 1

            if title in existing_titles:
                stats["skipped"] += 1
                report.append((title, "SKIP 已存在", len(ingredients), len(steps), len(img_map), tags))
                continue

            if args.dry_run:
                report.append((title, "DRY", len(ingredients), len(steps), len(img_map), tags))
                continue

            # ---- 附件以 BLOB 直接入库（attachments.content），不落盘 ----
            def attach_for(rel):
                if rel not in img_map:
                    return 0
                alt, zp, ext = img_map[rel]
                data = z.read(zp)
                fname = rnd_name(ext)
                month = time.strftime("%Y%m")
                # storage_path 仅为唯一标识（blob 化后不指向磁盘文件）
                storage = f"{month}/{fname}"
                sha = hashlib.sha256(data).hexdigest()
                cur.execute(
                    "insert into attachments (owner_user_id, kind, file_name, storage_path,"
                    " mime_type, size_bytes, width, height, duration, sha256, is_deleted,"
                    " created_at, updated_at, content) values (0,'image',?,?,?,?,0,0,0,?,0,?,?,?)",
                    (os.path.basename(zp), storage, MIME.get(ext, "application/octet-stream"),
                     len(data), sha, int(time.time()), int(time.time()), data))
                stats["img"] += 1
                return cur.lastrowid

            # 封面：优先「预览图 / 成品」类图片，否则第一张
            cover_rel = None
            for rel, (alt, zp, ext) in img_map.items():
                if re.search(r"预览|成品|摆盘|参考", alt or "") or re.search(r"预览|成品|摆盘|参考", os.path.basename(zp)):
                    cover_rel = rel
                    break
            if cover_rel is None and img_map:
                cover_rel = list(img_map.keys())[0]
            cover_id = attach_for(cover_rel) if cover_rel else 0

            # 步骤：把图片引用替换成 attachment id
            step_rows = []
            for st in steps:
                media = []
                # 步骤内容里残留的图片 markdown → 提取
                for alt, rel in all_img_refs(st["content"]):
                    aid = attach_for(rel)
                    if aid:
                        media.append(aid)
                for img_md in st["imgs"]:
                    for alt, rel in all_img_refs(img_md):
                        aid = attach_for(rel)
                        if aid:
                            media.append(aid)
                content = clean_md(st["content"])
                step_rows.append({"content": content, "media": media})

            now = int(time.time())
            cur.execute(
                "insert into recipes (user_id,title,summary,cover_attachment_id,tips,"
                "ingredients,tools,steps,meal_mask,calories,difficulty,servings,cook_minutes,"
                "source,ai_model,review_status,is_deleted,created_at,updated_at)"
                " values (0,?,?,?,?,?,?,?,?,?,?,?,?,'import','',0,0,?,?)",
                (title, summary, cover_id, "",
                 json.dumps(ingredients, ensure_ascii=False),
                 json.dumps(tools, ensure_ascii=False),
                 json.dumps(step_rows, ensure_ascii=False),
                 0, calories, difficulty, servings, cook_minutes,
                 now, now))
            rid = cur.lastrowid
            for t in tags:
                cur.execute("insert or ignore into recipe_tags (recipe_id, tag_id) values (?,?)", (rid, t))
            existing_titles.add(title)
            stats["inserted"] += 1
            report.append((title, f"OK id={rid} cover={cover_id}", len(ingredients), len(step_rows), len(img_map), tags))

        if not args.dry_run:
            con.commit()

        log("\n{:<28} {:<22} {:>4} {:>4} {:>4}  tags".format("菜名", "结果", "食材", "步骤", "图"))
        log("-" * 100)
        for r in report:
            log("{:<28} {:<22} {:>4} {:>4} {:>4}  {}".format(r[0], r[1], r[2], r[3], r[4], r[5]))
        log("-" * 100)
        log("统计:", stats)
        log("recipes 总数:", cur.execute("select count(*) from recipes").fetchone()[0])
        log("attachments 总数:", cur.execute("select count(*) from attachments").fetchone()[0])
        log("recipe_tags 总数:", cur.execute("select count(*) from recipe_tags").fetchone()[0])

    con.close()
    return 0


if __name__ == "__main__":
    sys.exit(main())
