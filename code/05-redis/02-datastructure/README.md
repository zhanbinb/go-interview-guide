# Redis 底层数据结构（超高频）

> 面试**超高频**章节（仅次于数据类型）
> 重点：每个类型的**底层实现** + **为什么这么设计** + **优化点**
> 必问：**SDS 是什么？为什么用 SDS？ZSet 为什么用跳表不用红黑树？**

---

## 🔥 必问核心（一张总图）

| 外部类型 | 底层数据结构（Redis 7.x）| 关键优化点 |
|---------|------------------------|----------|
| String | SDS（int / embstr / raw 三种编码）| O(1) 长度、二进制安全、杜绝缓冲区溢出 |
| List | quicklist（双向链表 + listpack 节点）| 折中 linkedlist 和 ziplist 的优缺点 |
| Hash | listpack（小）/ dict（大）| 字段少时连续内存省空间，多时 O(1) 查 |
| Set | intset（全整数小）/ dict（其它）| 全是整数时连续内存省空间 |
| ZSet | listpack（小）/ skiplist + dict（大）| 同时支持 O(log N) 排序 + O(1) 查分数 |

> 面试话术：**"Redis 用 C 写，但所有键值对的数据结构都是自己实现的，最核心是 SDS / dict / quicklist / skiplist / listpack 五种。"**

---

## 1. SDS：Simple Dynamic String（String 底层）

### 1.1 为什么不直接用 C 字符串？

C 字符串（`char*` 以 `\0` 结尾）有三个致命问题：

| 问题 | 说明 | 后果 |
|------|------|------|
| **O(N) 取长度** | 必须遍历到 `\0` | `STRLEN` 慢 |
| **二进制不安全** | 不能存 `\0`（被当成结束符）| 无法存图片/音频等二进制 |
| **缓冲区溢出** | `strcat` 不检查空间 | 修改前必须保证目标有足够空间 |

### 1.2 SDS 结构体（Redis 实际定义）

```c
struct sdshdr {
    uint8_t  len;       // 已使用长度
    uint8_t  alloc;     // 总分配长度（不含 header 和 \0）
    uint8_t  flags;     // 低 3 位表示类型（sdshdr5/8/16/32/64）
    char     buf[];     // 实际字节数组
};
```

- 通过 `flags` 区分 `sdshdr5/8/16/32/64` 五种 header 大小（按 len 大小动态选，省内存）
- `buf[]` 柔性数组，紧跟在 header 后，**一次 malloc 同时分配 header + 数据**

### 1.3 SDS 三大优势

| 优势 | 原理 |
|------|------|
| **O(1) 取长度** | 直接读 `len` 字段 |
| **二进制安全** | 用 `len` 判结束，不用 `\0` |
| **杜绝缓冲区溢出** | 修改前先 `sdsMakeRoomFor` 扩容 |

### 1.4 空间分配策略（减少内存重分配）

**空间预分配**（`sdsMakeRoomFor`）：
- 修改后 `len < 1MB` → 分配 `2 * len`（双倍预留）
- 修改后 `len >= 1MB` → 分配 `len + 1MB`（固定预留）
- 最多 `N` 次追加只 `O(log N)` 次 realloc

**惰性空间释放**（`sdstrim`）：
- 缩短字符串时，不立刻 `realloc` 缩容
- 仅把多余的字节写进 `free` 字段，留给下次用
- 真要释放时调 `sdsRemoveFreeSpace` 或 `sdsfree`

### 1.5 三种字符串编码（`OBJECT ENCODING` 验证）

```bash
SET s1 1234           # int
SET s2 "hello"        # embstr（≤ 44 字节）
SET s3 "long...(>44)" # raw
OBJECT ENCODING s1    # "int"
OBJECT ENCODING s2    # "embstr"
OBJECT ENCODING s3    # "raw"
```

| 编码 | 触发条件 | 特点 |
|------|---------|------|
| **int** | 值是整数且能用 `long` 表示 | 整数直接存指针，省 SDS 头 |
| **embstr** | 字符串 ≤ 44 字节 | header + buf **一次 malloc**（连续内存）|
| **raw** | 字符串 > 44 字节 | header 和 buf 两次 malloc（SDS 头 + 数据）|

> 44 字节来自：`64 - 16(header) - 3(buf 头) - 1(\0) = 44`

---

## 2. dict：hashtable（Hash / Set 底层）

### 2.1 整体结构

```
dict
├── ht[2]            // 两个 hashtable（用于渐进式 rehash）
│   ├── ht[0]        // 正在用的表
│   │   ├── table    // dictEntry* 数组
│   │   ├── size     // 桶数（总是 2^n）
│   │   ├── sizemask // size - 1（位掩码算下标）
│   │   └── used     // 已用 entry 数
│   └── ht[1]        // rehash 时的目标表（平时 unused）
├── rehashidx        // rehash 进度（-1 表示没在 rehash）
└── iterators        // 正在迭代的迭代器数量
```

```
dictEntry
├── key        // sds（键）
├── val        // union { sds / int64 / double }
├── next       // 链地址法解决冲突（指向下一个 entry）
└── ...        // (Redis 6.x+) 还可以放 LRU/lfu 等元数据
```

### 2.2 哈希冲突：链地址法

两个 key 哈希到同一桶 → 用 `next` 指针连成链表（同桶是链表，**不同桶是数组**）。

### 2.3 哈希函数：SipHash（Redis 4.0+）

- 之前用 `MurmurHash2` → 容易被哈希洪水攻击（构造大量冲突的 key 把 O(1) 退化到 O(N)）
- 改用 SipHash-1-2（带密钥的伪随机函数）+ 引入 `dict-can-resize` 控制

### 2.4 渐进式 rehash（核心设计！）

**为什么需要 rehash**：负载因子过高（used/size > 1）或过低（< 0.1）→ 扩缩容。

**为什么"渐进式"**：如果一次性把全部 entry 从 ht[0] 搬到 ht[1] → 单次操作会卡顿（百万级 key 时几秒）。所以**分多次搬**。

**触发条件**：
| 条件 | 操作 |
|------|------|
| `used / size >= 1` 且没在 `BGSAVE`/`BGREWRITEAOF` | 扩容（`size *= 2`）|
| `used / size >= 5`（强制）| 扩容（不管有没有在持久化）|
| `used / size < 0.1` | 收缩（`size /= 2`）|

> 持久化时主动避免扩容：子进程用的是 fork 出来的 COW（Copy-On-Write），扩容会触发大量写时复制。

**渐进式 rehash 步骤**：
1. 给 ht[1] 分配新空间（`rehashidx = 0`）
2. **每次增/删/查**操作时，把 ht[0] 一个桶里的 entry 搬到 ht[1]，`rehashidx++`
3. 当 ht[0] 全搬完 → 释放 ht[0]，ht[0] = ht[1]，`rehashidx = -1`
4. 期间所有**查操作**要查 ht[0] 和 ht[1] 两个表

> 面试话术：**"Redis 用分摊思想，把大字典的 rehash 拆到 N 次操作里完成，避免单次卡顿。"**

### 2.5 负载因子（load factor）

- `load_factor = ht[0].used / ht[0].size`
- 越接近 1 冲突越多；越小越浪费
- Redis 默认扩容阈值是 1（紧凑但易冲突）

---

## 3. quicklist：双向链表 + listpack 节点（List 底层）

### 3.1 历史演变

- **Redis ≤ 3.0**：`linkedlist`（双向链表）→ 每个节点单独 `malloc`，内存碎片多
- **Redis 3.2**：引入 `ziplist`（连续内存，省内存但修改要 realloc）
- **Redis 3.2+**：`quicklist = linkedlist of ziplist/listpack`（折中）
- **Redis 7.0**：quicklist 内部节点用 `listpack`（替代 ziplist，更安全）

### 3.2 quicklist 结构

```
quicklist
├── head → quicklistNode
├── tail → quicklistNode
└── count           // 所有 listpack 元素总数
    ├── prev
    ├── next
    quicklistNode ─┤
    ├── zi          // listpack* 指向一段连续内存
    ├── zl_size     // 这个 listpack 占用字节
    └── count       // 这个 listpack 元素个数
    ├── prev
    ├── next
    quicklistNode ─┤
    ├── zi          // listpack*
    ├── zl_size
    └── count
```

**本质**：一个**双向链表**，每个节点是一个**紧凑 listpack**（连续内存）。

### 3.3 关键参数

| 参数 | 默认 | 含义 |
|------|------|------|
| `list-max-listpack-size` | `-2` | 每个 listpack 节点大小（负数=字节，正数=元素数；-2 = 8KB）|
| `list-compress-depth` | `0` | 两端不压缩的深度（0=全不压缩，1=首尾各 1 节点不压缩）|

### 3.4 折中优势

- **比 linkedlist 省内存**（节点内 listpack 连续）
- **比纯 ziplist 修改快**（只动某个节点，不用整块 realloc）
- 越大的 list 用越小 listpack（控制单个节点大小，避免单次搬迁代价大）

---

## 4. listpack：紧凑列表（替代 ziplist）

### 4.1 ziplist 的问题：级联更新

- ziplist 每个 entry 存 `prevrawlen`（前一个 entry 的长度）
- 当某个 entry 长度从 1 字节变 253+ 字节，`prevrawlen` 从 1 字节变 5 字节 → 后一个 entry 也要变 → 雪崩
- 最坏情况 O(N) 次连锁更新（虽然概率小但存在）

### 4.2 listpack 怎么解决

- 不存 `prevrawlen` 这种依赖关系
- 每个 entry 存 `encoding-len` + `len` + `value`
- 用 `numele` 总元素数 + `backlen` 定位前一个 entry
- **彻底干掉级联更新**

### 4.3 触发 listpack 编码的阈值（Redis 7.x）

| 类型 | 阈值参数 | 默认值 |
|------|---------|--------|
| Hash | `hash-max-listpack-entries` / `-value` | 128 / 64 |
| List | `list-max-listpack-size`（每个 quicklist 节点）| 8KB |
| ZSet | `zset-max-listpack-entries` / `-value` | 128 / 64 |
| Set | 全是整数 + 元素数 ≤ `set-max-intset-entries` | 512 |

**超过阈值会自动升级**：
- Hash: listpack → dict
- Set: intset → dict
- ZSet: listpack → skiplist + dict
- List: 不升级（quicklist 永远用 listpack 节点）

---

## 5. intset：整数集合（小 Set 底层）

```c
typedef struct intset {
    uint32_t encoding;  // INTSET_ENC_INT16/32/64
    uint32_t length;    // 元素个数
    int8_t  contents[]; // 有序、紧凑、连续的整数数组
} intset;
```

**特点**：
- 全部是有序整数 → 二分查找 O(log N)
- 升级（`upgrade`）：新加更大范围的整数 → 整片 realloc 到更大 encoding
- 一旦加非整数 / 元素数超阈值 → 升级为 `hashtable`
- 内存极省（全是连续 int8/16/32/64，无指针/header 开销）

**示例**：
```bash
SADD s 1 2 3 4 5    # 全是整数 → intset
OBJECT ENCODING s   # "intset"
SADD s "abc"        # 加字符串 → 升级为 hashtable
OBJECT ENCODING s   # "hashtable"
```

---

## 6. skiplist：跳表（ZSet 排序部分）

### 6.1 结构

```
level 4: head ─────────────────────────────────────────→ nil
level 3: head ──────────────────→  node3 ──────────────→ nil
level 2: head ─────────→  node2 ─→ node3 ──────→ node4 → nil
level 1: head ──→ n1 ─→ node2 ─→ node3 ─→ n4 ─→ node4 → nil
```

- 多层有序链表，底层含所有节点
- 每个节点的 level 数随机（概率 1/2/4/... 决定晋升）
- 查找从最高层开始，能跳就跳 → 平均 O(log N)

### 6.2 查找过程（找 score=30 的节点）

1. 从 level 3 头节点开始
2. `next.score <= 30` 就跳过去；否则下降一层
3. 重复直到 level 1 找到或确认不存在

### 6.3 跳表 vs 红黑树（**面试必问**）

| 维度 | 跳表 (skiplist) | 红黑树 (RB-Tree) |
|------|----------------|------------------|
| 实现复杂度 | **简单**（指针操作）| 复杂（左旋/右旋/变色）|
| 范围查询 | **O(log N) 找到起点后顺序遍历** | O(log N) 找起点后**中序遍历**（要回溯）|
| 插入/删除 | 简单（改指针 + 随机 level）| 复杂（5 种 case 修复平衡）|
| 并发友好 | **局部锁**即可 | 全局锁（要锁整棵子树）|
| 内存占用 | 略多（每节点多级指针数组）| 略少（每个节点 3 指针 + 颜色）|
| 极端退化 | 概率 1/2^N，**可调节**| 不会退化（严格平衡）|

**面试答案模板**（直接背）：

> Redis 作者 antirez 在 Twitter 上说过：实现 skiplist 比 RB-Tree 简单得多，**范围操作更方便**（ZSet 经常按 score 区间查），并发场景下**局部锁粒度更细**，虽然有 1/2^N 的概率退化但可以接受。综合下来跳表更适合 Redis 的场景。

### 6.4 ZSet 为什么同时用 skiplist + dict？

- `skiplist`：按 score 排序 → `ZADD` / `ZRANGE` / `ZRANGEBYSCORE` O(log N)
- `dict`：key → score 映射 → `ZSCORE` O(1)
- 两个结构**同时维护**，单 `ZADD` 同时改两个 → 用一份内存换 O(1) 取分数

---

## 7. 编码转换总表（必背！）

| 外部类型 | 编码 1（小）| 编码 2（大）| 转换时机 |
|---------|-----------|-----------|---------|
| String（整数）| int | — | 写字符串就转 embstr/raw |
| String（≤44 字节）| embstr | raw | > 44 字节 |
| String（> 44 字节）| raw | — | — |
| Hash | listpack | dict | 字段或值超阈值 |
| List（quicklist 节点）| listpack | listpack | **永不升级**（quicklist 永远）|
| Set（全整数）| intset | dict | 加非整数 / 元素数超阈值 |
| Set | dict | — | — |
| ZSet | listpack | skiplist + dict | 元素数 / value 长度超阈值 |

**验证方法**：`OBJECT ENCODING key`（Redis 4.0+ 推荐用 `OBJECT ENCODING` 而不是 `DEBUG OBJECT`）

---

## 8. 面试高频追问

### Q1：String 类型底层？SDS 是什么？

答：Redis 自己实现的简单动态字符串（Simple Dynamic String）。结构含 `len`（已用长度）、`alloc`（总分配）、`flags`（类型标识）、`buf[]`（数据）。相比 C 字符串有 4 大优势：O(1) 取长度、二进制安全、杜绝缓冲区溢出、空间预分配/惰性释放减少 realloc。

### Q2：渐进式 rehash 怎么做的？为什么？

答：dict 用 ht[0] 和 ht[1] 两张哈希表，扩容时给 ht[1] 分配新空间（更大或更小），然后**每次增删改查操作**顺带把 ht[0] 一个桶的 entry 搬到 ht[1]，逐步推进。期间查操作要查两个表。
目的：避免一次性搬迁百万级 key 导致的几秒卡顿，把成本**分摊到 N 次请求**里。

### Q3：ZSet 为什么用跳表不用红黑树？

答：① 实现简单（指针 vs 左旋右旋变色），bug 少；② 范围查询方便（跳表顺序遍历 O(N)，红黑树要中序遍历回溯）；③ 并发场景下局部锁粒度更细；④ 退化概率 1/2^N 可接受。综合工程考虑跳表更合适。（antirez 原话：实现简单 + 范围操作方便）

### Q4：quicklist 是什么？为什么不用 linkedlist 或 ziplist？

答：quicklist = 双向链表 + listpack 节点。**纯 linkedlist** 每个节点单独 malloc，内存碎片多、cache 不友好；**纯 ziplist** 修改要整块 realloc，大列表很慢。quicklist 折中：节点内 listpack 连续内存省空间，节点间链表结构改局部不动整体。Redis 3.2+ 用 quicklist，7.0+ 节点从 ziplist 升级为 listpack（防级联更新）。

### Q5：ziplist 和 listpack 的区别？

答：ziplist 每个 entry 存 `prevrawlen`（前一个 entry 长度），某 entry 长度变化会触发**级联更新**（每个后续 entry 都要改 prevrawlen）。listpack 不存这种依赖，用 `numele` + entry 自己的 backlen 定位前一个，**彻底消灭级联更新**。Redis 7.0+ 全面用 listpack。

### Q6：Hash 类型字段数多 vs 少，底层有什么区别？

答：少时用 **listpack**（连续内存，省内存，O(N) 查）；多时（默认 128 字段或 value 总长超 64 字节）转 **dict**（hashtable，O(1) 查，多耗内存）。所以**字段少用 Hash 划算，字段多（>128）建议拆 key 或用 String + JSON**。

### Q7：intset 是什么？什么时候升级？

答：小 Set 全是整数时的紧凑编码（连续 int16/32/64 数组）。新增更大范围的整数会触发**升级**（整片 realloc），但**不会降级**。加非整数或元素数超 `set-max-intset-entries`（默认 512）就升级为 hashtable。

### Q8：怎么验证一个 key 底层用的什么编码？

答：`OBJECT ENCODING key`（推荐）。也可以 `OBJECT HELP` / `DEBUG OBJECT key`（不推荐，会触发回收）。

### Q9：哈希冲突怎么解决？为什么用链地址法？

答：用链地址法（每个桶是链表）。简单、不依赖探测序列、删除方便、扩容方便（rehash 时按桶搬迁）。Redis 没选开放地址法，因为**链地址法对哈希函数要求低**（即使哈希不均也只是某桶链表长，整体仍可用）。

### Q10：Redis 的哈希函数为什么用 SipHash？

答：之前用 MurmurHash2，但**可被哈希洪水攻击**（攻击者构造大量哈希冲突的 key，把 O(1) 操作退化到 O(N)，耗光 CPU）。SipHash 是带密钥的伪随机函数，**抗哈希洪水攻击**。Redis 4.0+ 改用 SipHash-1-2。

---

## 9. 时间复杂度速查

| 操作 | 复杂度 | 数据结构 |
|------|--------|----------|
| `GET/SET/INCR` (String) | O(1) | SDS |
| `HSET/HGET/HMSET/HGETALL` (Hash) | O(1) 单字段 / O(N) 全字段 | dict 或 listpack |
| `HGETALL` (Hash 大) | O(N) | dict |
| `LPUSH/RPUSH/LPOP/RPOP` (List) | O(1) | quicklist 头尾 |
| `LINDEX list 1000` (List 中间) | O(N) | 遍历 quicklist |
| `SADD/SMEMBERS/SISMEMBER` (Set) | O(1) / O(N) | intset / dict |
| `SINTER/SUNION/SDIFF` (Set) | O(N*M) | dict 遍历 |
| `ZADD/ZREM/ZSCORE` (ZSet) | O(log N) / O(log N) / O(1) | skiplist / dict |
| `ZRANGE/ZRANGEBYSCORE` (ZSet) | O(log N + M) | skiplist |
| `OBJECT ENCODING` | O(1) | 直接读 robj 字段 |

---

## 10. 一句话速记（面试答完收尾用）

> **"Redis 7.x 五大底层：SDS（O(1) 长度 + 二进制安全）、dict（hashtable + 渐进式 rehash）、quicklist（双向链表 + listpack 节点）、skiplist（ZSet 排序用，跳表比红黑树实现简单且范围查询友好）、listpack（替代 ziplist，彻底消除级联更新）。小数据用紧凑结构（intset / listpack / embstr）省内存，大数据升级到 hashtable / skiplist 保性能。"**
