# logflow

数据消费管道

## Build

### Run

```
make
logflow -c config/logflow.yaml
```

### Docker

```Shell
docker build -t logflow:v1 .
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

## 配置规则

### input

#### kafka

#### file

### filter

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

#### add

添加字段

```YAML
Add:
  overwrite: true
  fields:
      flag: test
```

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
  src: 'logtime'
  target: '@timestamp'
  location: 'Asia/Shanghai'
  add_year: false
  overwrite: true
  formats:
      - 'RFC3339'
      - '2006-01-02T15:04:05'
      - '2006-01-02T15:04:05Z07:00'
      - '2006-01-02T15:04:05Z0700'
      - '2006-01-02'
      - 'UNIX'
      - 'UNIX_MS'
```

#### dateFormat

将 `time`类型的字段格式化到另一个字段中

```YAML
dateFormat:
  source: datatime
  location: 'Asia/Shanghai'
  format: '2006.01.02'
  target: "@localtime"
  set_if_fail: '1970.01.01'
  set_if_nil: '1970.01.01'
```

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

```YAML
lowercase:
  fields: ['domain', 'url']
```

#### remove

```YAML
Remove:
  fields: ['domain', 'url']
```

#### rename

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

```YAML
uppercase: ['loglevel']
```

#### replace

```YAML
- replace:
  fields:
    msg: ['wang', 'Wang', 1]
- replace:
  fields:
    msg: ['en', 'eng']
```

最后面的 1 代表, 只替换一次. 如果不给这个值, 代表替换所有的.

#### strip

去除字符串前后的空白字符、换行符

```YAML
strip: ["name"]
```

### output

#### elasticsearch

```YAML
elasticsearch:
  hosts: ['https://127.0.0.1:6009']
  index: 'kafka-%{[@metadata][kafka][topic]}-%{@localtime}'
  user: elastic
  password: test
  ssl: true
  cacert: ''
  sniff: false
  sniff_interval: 30 # 秒
  version: 7
  skip_ssl_verification: false # 是否跳过 ES 的证书校验
  bulk_count: 5000
  bulk_size: 15
  flush_interval: 30
```

#### file

todo

#### console
