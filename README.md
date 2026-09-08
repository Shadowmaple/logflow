# logflow

数据消费管道

数据从 input 读取，经过 filter 链依次处理，最终写入 output：

```
input ──> filter ──> filter ──> ... ──> output
```

## Build

### Run

```
make
logflow -c config/logflow.yaml
```

`-c` 指定配置文件路径，必须为 `.yaml` 或 `.yml` 文件。

### Docker

```Shell
docker build -t logflow:v1 .
# 或者
make docker
```

### 交叉编译

```Shell
make all
```

构建 windows / linux / darwin 多平台产物，输出到 `build/` 目录。

### 测试

```Shell
make test
```

## 字段约定

### JSONPATH 格式

如果以 `$.` 开头, 认为是这种格式

给几个下面文中的例子

```text
$.store.book[0].title

$['store']['book'][0]['title']

$.store.book[(@.length-1)].title

$.store.book[?(@.price < 10)].title
```

具体的格式和例子参见 [https://goessner.net/articles/JsonPath/](https://goessner.net/articles/JsonPath/)

### 多层字段

不以 `$.` 开头的字段名, 按 `[name][last]` 这样的格式表示多层字段。

## 配置规则

### 整体结构

```YAML
system:
  # 全局配置
input:
  # 输入列表
filter:
  # 过滤器列表, 按顺序依次执行
output:
  # 输出列表
```

#### system

```YAML
system:
  worker: 1
  logger:
    level: 'info'
    format: 'json'
    output: 'stdout'
    file_path: ''
```

- `worker`：并发处理数，即每个 input 开启 n 个任务协程并发处理，每个 input 都有单独的 filter 和 output 队列，默认 1
- `logger.level`：日志级别，debug / info / warn / error / fatal，默认 info
- `logger.format`：日志输出格式，json / console，默认 json
- `logger.output`：日志输出位置，stdout / stderr / file，默认 stdout
- `logger.file_path`：输出到文件时的文件路径

### input

目前支持 `kafka`、`console`、`random`。

#### kafka

```YAML
input:
  - kafka:
      brokers: ['127.0.0.1:9092']     # 必填
      topics: ['logflow']             # 必填, topics / topic_pattern 二选一
      group_id: logflow-group         # 必填
      codec: 'json'
      worker: 1
      auto_offset_earliest: false
      enable_auto_commit: true
      auto_commit_interval: 30        # 秒
      decorate_events: true
      messages_queue_length: 8
      client_id: logflow-client
      version: '2.8.2'
      sasl_enabled: false
      # sasl_mechanism: SCRAM-SHA-256   # 支持 SCRAM-SHA-256 / SCRAM-SHA-512
      # sasl_username: ''
      # sasl_password: ''
      tls_enabled: false
      # tls_ca: ''
      discard_on_error: false
```

- `brokers`：kafka broker 地址列表，必填
- `topics`：订阅的 topic 列表；`topic_pattern`（正则匹配 topic）暂不支持
- `group_id`：消费者组 ID，必填
- `codec`：消息解码方式，plain / json，默认 plain
- `worker`：并发消费数，默认 1
- `auto_offset_earliest`：消费者组第一次消费时是否从最早位置开始，默认 false（从最新位置开始）
- `enable_auto_commit`：是否自动提交偏移量，默认 true
- `auto_commit_interval`：自动提交偏移量的时间间隔（秒），默认 15
- `decorate_events`：是否添加 kafka 元数据字段 `@metadata.kafka.{topic, partition, offset, timestamp}`，默认 true
- `messages_queue_length`：内部消息队列长度，默认 8
- `client_id`：kafka 客户端 ID，默认 logflow
- `version`：kafka 协议版本
- `sasl_enabled`：是否开启 SASL 认证，开启时 `sasl_username`、`sasl_password`、`sasl_mechanism` 必填
- `tls_enabled`：是否开启 TLS；`tls_ca` 为自定义 CA 证书路径，可选
- `discard_on_error`：消息解码失败时是否丢弃消息，默认 false

#### console

从标准输入逐行读取数据。

```YAML
input:
  - console:
      codec: 'plain'    # plain / json, 默认 plain
```

#### random

随机生成 [from, to] 区间内的整型数字作为消息，仅用于测试。

```YAML
input:
  - random:
      from: 1           # 必填
      to: 100           # 必填
      max_messages: 10  # 最大生成条数, 不配置则不限制
```

#### file

【TODO】

### filter

filter 列表中的每个过滤器按配置顺序依次执行。

#### 通用字段

##### if

满足全部条件，则返回 true.

列表内为 and 关系，每个条件内可做 and、or 判断。

例如：

```YAML
drop:
  - if:
    - 'EQ(level, "DEBUG")'
    - 'EXIST(msg)'
    - 'EQ(tag, "backup") || EQ(tag, "prod")'
```

函数：

- `Exist(value)`：存在 value 字段
- `EQ`：字段值匹配是否相等，支持nil、整型、字符串
  - `EQ(value, 1)`：存在且value == 1
  - `EQ(value, '1')`：存在且value == '1'
  - `EQ(value, nil)`：存在且value == nil
- `In`：字段在列表中，支持整型、字符串
  - `In(items, "test")`：items 为列表且存在 "test"
- `NotIn`：字段不在列表中，支持整型、字符串
  - `NotIn(items, "test")`：items 为列表且不存在 "test"
- `HasPrefix`：字符串前缀匹配【TODO】
- `HasSuffix`：字符串后缀匹配【TODO】
- `Match`：正则匹配【TODO】

##### add_fields

当Filter执行成功时，可以添加一些字段。如果Filter失败, 则忽略。

##### remove_fields

当Filter执行成功时, 可以删除一些字段. 如果Filter失败, 则忽略。

##### failTag

当Filter执行失败时, 可以添加内容到 `@tags` 字段。

#### filters

将多个过滤器组合成一个嵌套过滤链，配合 `if` 使用：条件满足时执行内部的 `modules` 列表，条件不满足则跳过整组过滤器。

```YAML
- filters:
    if: ['Exist(level)']
    modules:
      - strip:
          fields: ['level']
      - uppercase:
          fields: ['level']
      - drop:
          if: ['EQ(level, "DEBUG")']
```

#### add

添加字段

```YAML
add:
  overwrite: true
  fields:
      flag: test
      '[name][first]': 'John'
```

- `fields`：要添加的字段及值，字段名支持 `[name][first]` 多层格式，不存在的层级会自动创建
- `overwrite`：字段已存在时是否覆盖，默认 true

#### json

将目标字段按照 json 格式解析

```YAML
json:
  source: message
  target: res
  include: ["client", "host"]
  exclude: ["request", "user_agent"]
```

##### source

源字段

##### target

目标字段, 如果不设置, 则将生成的所有字段写入到根一层.

##### include

只使用 include 中配置的字段，其他字段丢弃。优先级高于 exclude

##### exclude

exclude 中的字段不要。优先级低于 include

#### date

把一个字符串类型的字段, 转成一个 Time 类型的字段, 存到 target 里面去.

如果源字段不存在, 返回 false. 如果所有 formats 都匹配失败, 返回 false

```YAML
date:
  source: 'logtime'
  target: '@timestamp'
  location: 'Asia/Shanghai'
  add_year: false
  overwrite: true
  formats:
    - 'RFC3339'
    - 'ISO8601'
    - '2006-01-02T15:04:05'
    - '2006-01-02T15:04:05Z07:00'
    - '2006-01-02T15:04:05Z0700'
    - '2006-01-02'
    - 'UNIX'
    - 'UNIX_MS'
```

- `source`：源字段，必填
- `target`：目标字段，默认 `@timestamp`
- `location`：时区，源字段不带时区信息时使用
- `add_year`：解析失败时是否自动补充当前年份，默认 false
- `overwrite`：是否覆盖目标字段已有值，默认 true
- `formats`：时间格式列表，逐个尝试直到匹配成功；支持 Go 时间格式，以及内置格式别名：
  - `UNIX`：秒级时间戳
  - `UNIX_MS`：毫秒级时间戳
  - `RFC3339`：如 `2006-01-02T15:04:05Z07:00`
  - `ISO8601`：ISO8601 格式
  - 常见写法（如 `yyyy-MM-dd HH:mm:ss.SSS`）会自动转换为对应的 Go 时间格式

#### dateFormat

将 `time`类型的字段格式化到另一个字段中

```YAML
dateFormat:
  source: datatime
  location: 'Asia/Shanghai'
  format: 'yyyy.MM.dd'
  target: "@localtime"
  set_if_fail: '1970.01.01'
  set_if_nil: '1970.01.01'
```

- `source`、`target`、`format`：必填
- `format`：支持 `yyyy-MM-dd` 等常见写法，自动转换为 Go 时间格式

##### remove_if_fail

如果转换失败刚删除这个字段, 默认 false

##### set_if_fail: XX

如果转换失败, 刚将此字段的值设置为 XX . 优先级比 remove_if_fail 低.  如果 remove_if_fail 设置为 true, 则 set_if_fail 无效.

##### set_if_nil: XX

如果没有这个字段, 刚将此字段的值设置为 XX . 优先级最高

#### drop

满足条件则抛弃消息

```YAML
drop:
  - if: ['EQ(level, "DEBUG")']
```

#### lowercase

将字段的值转为小写

```YAML
lowercase:
  fields: ['domain', 'url']
```

#### remove

删除字段【TODO】

```YAML
Remove:
  fields: ['domain', 'url']
```

#### rename

字段改名【TODO】

```YAML
Rename:
  fields:
    host: hostname
    serverIP: server_ip
    '[name][last]': '[name][first]'
```

将 host 字段改名为 hostname, 将 serverIP 字段改名为 server_ip, 将 [name][last] 字段改名为 [name][first]

如果没有这个字段，不做任何操作

#### uppercase

将字段的值转为大写

```YAML
uppercase:
  fields: ['loglevel']
```

#### replace

字符串替换

```YAML
- replace:
    fields:
      msg: ['wang', 'Wang', 1]
- replace:
    fields:
      msg: ['en', 'eng']
```

- `fields`：key 为字段名，value 为替换规则列表 `[old, new]` 或 `[old, new, count]`
- 最后面的 1 代表, 只替换一次. 如果不给这个值, 代表替换所有的.
- 字段不存在时跳过, 不做任何操作; 字段值不是字符串时该过滤器执行失败

#### strip

去除字符串前后的空白字符、换行符

```YAML
strip:
  fields: ["name"]
```

### output

output 列表中的每条数据会写入到所有配置的 output。

#### elasticsearch

```YAML
elasticsearch:
  hosts: ['https://127.0.0.1:6009']   # 必填
  index: 'kafka-%{[@metadata][kafka][topic]}-%{@localtime}'   # 必填
  user: elastic
  password: test
  api_key: ''
  ssl: true
  cacert: ''
  sniff: false
  sniff_interval: 30 # 秒
  version: 7
  skip_ssl_verification: false # 是否跳过 ES 的证书校验
  max_retries: 3
  bulk_count: 5000
  bulk_size: 15   # MB
  flush_interval: 30 # 秒
```

- `index`：索引名，支持 `%{[jsonpath]}` 字段插值，如 `%{[@metadata][kafka][topic]}`、`%{@localtime}`
- `hosts`：ES 地址列表，必填
- `user` / `password`：Basic 认证
- `api_key`：Base64 编码的 API Key（由 `id:api_key` 组成后 Base64），优先级高于 user/password
- `ssl`：是否使用 https，默认 false；为 true 时，`skip_ssl_verification` 为 false 则必须配置 `cacert`（CA 证书路径）
- `sniff`：是否开启 sniff，默认 false；`sniff_interval` 为 sniff 间隔（秒）
- `version`：ES 版本，默认 7
- `bulk_count`：触发批量提交的消息条数阈值，默认 5000
- `bulk_size`：触发批量提交的字节阈值（MB），默认 15
- `flush_interval`：定时提交间隔（秒），默认 30，避免数据量过少时长时间不提交
- `max_retries`：最大重试次数，默认 3

#### console

将数据输出到标准输出。

```YAML
output:
  - console:
      codec: 'json'   # plain / json, 默认 plain
```

#### file

【TODO】

## codec 说明

input 与 output 通过 codec 指定编解码方式：

- `plain`（默认）：整条消息作为字符串写入 `message` 字段；输出时直接输出原始字符串
- `json`：将消息按 JSON 解析为字段；解析失败时将原始内容写入 `message` 字段，并添加 `@errorTag: JSODecodeFailed`
