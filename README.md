# logflow

数据消费管道

## 字段约定

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

#### drop

满足条件则抛弃消息

```YAML
drop:
  - if: ['EQ(level, "DEBUG")']
```

#### lowercase

```YAML
lowercase: ['domain', 'url']
```

#### Remove

```YAML
Remove:
  fields: ['domain', 'url']
```

#### Rename

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

#### file

#### console
