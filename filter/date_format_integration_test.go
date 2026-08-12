package filter

// import (
// 	"testing"
// 	"time"

// 	"github.com/Shadowmaple/logflow/internal/event"
// 	"github.com/Shadowmaple/logflow/model"
// )

// // 本文件为 dateFormat 过滤器的集成测试，验证其在真实日志处理管道中的行为。
// //
// // 管道装配方式与 process.InputHandler.startOne 一致:
// //
// //	filters := model.BuildFilterProcessors(conf, BuildFilter)
// //	pipeline := model.NewProcessList(filters..., output)
// //	for { event := input.ReceiveOne(); ...; pipeline.Process(event) }
// //
// // 由于 filter 包不能导入 process 包（会形成循环依赖），这里用实现 model.Processor
// // 的 captureOutput 直接作为管道末端，等价于 process.OutputProcessor 的行为。

// // sliceInput 是测试用的 model.Input 实现，按顺序吐出预设事件，结束后返回 nil。
// type sliceInput struct {
// 	events []*event.Event
// 	idx    int
// }

// func (i *sliceInput) ReceiveOne() *event.Event {
// 	if i.idx >= len(i.events) {
// 		return nil
// 	}
// 	ev := i.events[i.idx]
// 	i.idx++
// 	return ev
// }

// func (i *sliceInput) Close() {}

// // captureOutput 是测试用的管道末端，捕获所有到达的事件。
// // 同时实现 model.Output 与 model.Processor，可直接挂到处理链末端。
// type captureOutput struct {
// 	events []*event.Event
// }

// func (o *captureOutput) Handle(ev *event.Event) error {
// 	o.events = append(o.events, ev)
// 	return nil
// }

// func (o *captureOutput) Close() {}

// // Process 使其可作为 model.Processor 挂载到处理链末端（模拟 process.OutputProcessor）
// func (o *captureOutput) Process(ev *event.Event) bool {
// 	if err := o.Handle(ev); err != nil {
// 		return false
// 	}
// 	return true
// }

// // dateParseFilter 是测试用桩 filter，用于模拟 logflow.yaml 中 date filter 的预期行为
// // （将字符串时间字段解析为 time.Time）。当前 DateFilter.Filter 实现为空（被注释），
// // 因此管道中需要一个解析阶段来为 dateFormat 提供 time.Time 输入。
// type dateParseFilter struct {
// 	src    string
// 	target string
// 	layout string // Go 时间格式
// }

// func (f *dateParseFilter) Filter(ev *event.Event) error {
// 	raw, ok := ev.Data[f.src]
// 	if !ok {
// 		return nil
// 	}
// 	s, ok := raw.(string)
// 	if !ok {
// 		return nil
// 	}
// 	t, err := time.Parse(f.layout, s)
// 	if err != nil {
// 		return nil
// 	}
// 	ev.Data[f.target] = t
// 	return nil
// }

// func newDateParseFilter(conf map[string]any) model.Filter {
// 	f := &dateParseFilter{layout: time.RFC3339}
// 	if v, ok := conf["src"]; ok {
// 		f.src = v.(string)
// 	}
// 	if v, ok := conf["target"]; ok {
// 		f.target = v.(string)
// 	}
// 	if v, ok := conf["layout"]; ok {
// 		f.layout = v.(string)
// 	}
// 	if f.src == "" {
// 		panic("dateparse: src is required")
// 	}
// 	if f.target == "" {
// 		f.target = f.src
// 	}
// 	return f
// }

// func init() {
// 	// 注册测试用桩 filter，使其可像真实 filter 一样通过配置构建
// 	register("dateparse", newDateParseFilter)
// }

// // runPipeline 用给定的 filter 配置构建处理链，将 input 中的事件依次送入管道，
// // 返回到达末端的捕获输出。模拟 process.InputHandler.startOne 的事件处理循环。
// func runPipeline(t *testing.T, filterConfs []map[string]any, in model.Input) *captureOutput {
// 	t.Helper()
// 	out := &captureOutput{}

// 	fps := model.BuildFilterProcessors(filterConfs, BuildFilter)
// 	procs := make([]model.Processor, 0, len(fps)+1)
// 	for _, fp := range fps {
// 		procs = append(procs, fp)
// 	}
// 	procs = append(procs, out)

// 	pipeline := model.NewProcessList(procs...)
// 	for {
// 		ev := in.ReceiveOne()
// 		if ev == nil {
// 			break
// 		}
// 		pipeline.Process(ev)
// 	}
// 	return out
// }

// // TestDateFormatPipeline_EndToEnd 模拟 logflow.yaml 中的真实管道:
// // json -> dateparse -> dateFormat -> output。
// // JSON 消息中的 ISO8601 字符串被解析为 time.Time，再由 dateFormat 格式化。
// func TestDateFormatPipeline_EndToEnd(t *testing.T) {
// 	filterConfs := []map[string]any{
// 		{"json": map[string]any{"source": "message"}},
// 		{"dateparse": map[string]any{"src": "logtime", "target": "logtime"}},
// 		{"dateFormat": map[string]any{
// 			"source":   "logtime",
// 			"target":   "@timestamp",
// 			"format":   "yyyy-MM-dd HH:mm:ss",
// 			"location": "Asia/Shanghai",
// 		}},
// 	}
// 	raw := `{"logtime":"2024-01-15T10:30:45Z","level":"INFO"}`
// 	in := &sliceInput{events: []*event.Event{
// 		newEvent(map[string]any{"message": raw}),
// 	}}

// 	out := runPipeline(t, filterConfs, in)
// 	if len(out.events) != 1 {
// 		t.Fatalf("expected 1 event captured, got %d", len(out.events))
// 	}
// 	got := out.events[0]

// 	// UTC 10:30:45 -> Asia/Shanghai 18:30:45
// 	if ts := got.Data["@timestamp"]; ts != "2024-01-15 18:30:45" {
// 		t.Fatalf("@timestamp = %v, want %q", ts, "2024-01-15 18:30:45")
// 	}
// 	// dateparse 应将 logtime 替换为 time.Time，供 dateFormat 消费
// 	if _, ok := got.Data["logtime"].(time.Time); !ok {
// 		t.Fatalf("logtime should be time.Time after pipeline, got %T", got.Data["logtime"])
// 	}
// 	// json 解析出的其它字段应保留
// 	if level := got.Data["level"]; level != "INFO" {
// 		t.Fatalf("level = %v, want %q", level, "INFO")
// 	}
// }

// // TestDateFormatPipeline_ConditionGates 验证通过 if 条件控制 dateFormat 是否执行。
// // 条件不满足时 dateFormat 被跳过，事件仍继续流向 output。
// func TestDateFormatPipeline_ConditionGates(t *testing.T) {
// 	filterConfs := []map[string]any{
// 		{"dateFormat": map[string]any{
// 			"source": "logtime",
// 			"target": "@timestamp",
// 			"format": "yyyy-MM-dd",
// 			"if":     []string{"EXIST(logtime)"},
// 		}},
// 	}
// 	in := &sliceInput{events: []*event.Event{
// 		newEvent(map[string]any{"logtime": testTime}),
// 		newEvent(map[string]any{"message": "no time here"}),
// 	}}

// 	out := runPipeline(t, filterConfs, in)
// 	if len(out.events) != 2 {
// 		t.Fatalf("both events should reach output, got %d", len(out.events))
// 	}
// 	// 第一条事件含 logtime -> 条件成立 -> 格式化
// 	if ts := out.events[0].Data["@timestamp"]; ts != "2024-01-15" {
// 		t.Fatalf("first @timestamp = %v, want %q", ts, "2024-01-15")
// 	}
// 	// 第二条事件无 logtime -> 条件不成立 -> dateFormat 跳过，未写入 @timestamp
// 	if _, exists := out.events[1].Data["@timestamp"]; exists {
// 		t.Fatal("second event should not have @timestamp (condition skipped dateFormat)")
// 	}
// }

// // TestDateFormatPipeline_NeverDropsOrFailTag 验证 dateFormat 的关键管道级契约:
// // 即使源字段类型错误且配置了 fail_tag，dateFormat 也永不返回错误，
// // 因此事件不会被丢弃，也不会被打上 fail_tag。
// func TestDateFormatPipeline_NeverDropsOrFailTag(t *testing.T) {
// 	filterConfs := []map[string]any{
// 		{"dateFormat": map[string]any{
// 			"source":   "logtime",
// 			"target":   "@timestamp",
// 			"format":   "yyyy-MM-dd",
// 			"fail_tag": "date_failed",
// 		}},
// 	}
// 	in := &sliceInput{events: []*event.Event{
// 		newEvent(map[string]any{"logtime": "not-a-time"}),
// 	}}

// 	out := runPipeline(t, filterConfs, in)
// 	if len(out.events) != 1 {
// 		t.Fatalf("event should NOT be dropped by dateFormat, got %d captured", len(out.events))
// 	}
// 	if _, hasFail := out.events[0].Data["fail_tag"]; hasFail {
// 		t.Fatal("fail_tag should not be set because dateFormat never returns an error")
// 	}
// 	if _, hasTs := out.events[0].Data["@timestamp"]; hasTs {
// 		t.Fatal("@timestamp should not be set when source is not time.Time and no set_if_fail")
// 	}
// }

// // TestDateFormatPipeline_SetIfFailFlowsThrough 验证 set_if_fail 的回退值能流经管道到达 output。
// func TestDateFormatPipeline_SetIfFailFlowsThrough(t *testing.T) {
// 	filterConfs := []map[string]any{
// 		{"dateFormat": map[string]any{
// 			"source":      "logtime",
// 			"target":      "@timestamp",
// 			"format":      "yyyy-MM-dd",
// 			"set_if_fail": "1970-01-01",
// 		}},
// 	}
// 	in := &sliceInput{events: []*event.Event{
// 		newEvent(map[string]any{"logtime": 12345}),
// 	}}

// 	out := runPipeline(t, filterConfs, in)
// 	if len(out.events) != 1 {
// 		t.Fatalf("expected 1 event, got %d", len(out.events))
// 	}
// 	if ts := out.events[0].Data["@timestamp"]; ts != "1970-01-01" {
// 		t.Fatalf("@timestamp = %v, want %q", ts, "1970-01-01")
// 	}
// }

// // TestDateFormatPipeline_SetIfNilFlowsThrough 验证源字段缺失时 set_if_nil 的值能流经管道到达 output。
// func TestDateFormatPipeline_SetIfNilFlowsThrough(t *testing.T) {
// 	filterConfs := []map[string]any{
// 		{"dateFormat": map[string]any{
// 			"source":     "logtime",
// 			"target":     "@timestamp",
// 			"format":     "yyyy-MM-dd",
// 			"set_if_nil": "1970-01-01",
// 		}},
// 	}
// 	in := &sliceInput{events: []*event.Event{
// 		newEvent(map[string]any{"message": "no logtime field"}),
// 	}}

// 	out := runPipeline(t, filterConfs, in)
// 	if len(out.events) != 1 {
// 		t.Fatalf("expected 1 event, got %d", len(out.events))
// 	}
// 	if ts := out.events[0].Data["@timestamp"]; ts != "1970-01-01" {
// 		t.Fatalf("@timestamp = %v, want %q", ts, "1970-01-01")
// 	}
// }

// // TestDateFormatPipeline_PrecedingFilterDropsEvent 是对照测试:
// // 当前置 filter 返回错误并配置 fail_tag 时，事件在管道中被丢弃，
// // dateFormat 与 output 都不会收到该事件。
// // 这与 dateFormat 自身“永不丢弃事件”的行为形成对比。
// func TestDateFormatPipeline_PrecedingFilterDropsEvent(t *testing.T) {
// 	filterConfs := []map[string]any{
// 		// json filter 在 source 字段缺失时返回错误
// 		{"json": map[string]any{
// 			"source":   "message",
// 			"fail_tag": "json_failed",
// 		}},
// 		{"dateFormat": map[string]any{
// 			"source": "logtime",
// 			"target": "@timestamp",
// 			"format": "yyyy-MM-dd",
// 		}},
// 	}
// 	// 事件中没有 message 字段 -> json filter 返回错误 -> 事件被丢弃
// 	in := &sliceInput{events: []*event.Event{
// 		newEvent(map[string]any{"logtime": testTime}),
// 	}}

// 	out := runPipeline(t, filterConfs, in)
// 	if len(out.events) != 0 {
// 		t.Fatalf("event should be dropped by failing json filter, got %d captured", len(out.events))
// 	}
// }

// // TestDateFormatPipeline_ChainedWithDownstreamFilter 验证 dateFormat 的输出可被
// // 后续 filter 消费：dateFormat 写入 @timestamp 后，lowercase 对其再处理。
// func TestDateFormatPipeline_ChainedWithDownstreamFilter(t *testing.T) {
// 	filterConfs := []map[string]any{
// 		{"dateFormat": map[string]any{
// 			"source": "logtime",
// 			"target": "ts",
// 			"format": "yyyy-MM-dd z", // z 为时区名，UTC 时间 -> "UTC"
// 		}},
// 		// 对 ts 字段做小写处理，验证 dateFormat 的输出能被下游读取
// 		{"lowercase": map[string]any{"fields": []string{"ts"}}},
// 	}
// 	in := &sliceInput{events: []*event.Event{
// 		newEvent(map[string]any{"logtime": testTime}),
// 	}}

// 	out := runPipeline(t, filterConfs, in)
// 	if len(out.events) != 1 {
// 		t.Fatalf("expected 1 event, got %d", len(out.events))
// 	}
// 	// "yyyy-MM-dd z" -> "2024-01-15 UTC" -> lowercase -> "2024-01-15 utc"
// 	if ts := out.events[0].Data["ts"]; ts != "2024-01-15 utc" {
// 		t.Fatalf("ts = %v, want %q", ts, "2024-01-15 utc")
// 	}
// }
