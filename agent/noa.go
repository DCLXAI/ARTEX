package agent

import (
	"log"
	"path/filepath"

	"github.com/Autumn-27/norma/agentcore"
	"github.com/Autumn-27/norma/noaadapter"
)

// noaWarn returns a diagnostics sink tagging non-fatal noa messages with the
// session, routed through the package logger (agents have no per-instance one).
func noaWarn(session string) func(string) {
	return func(msg string) { log.Printf("[noa] %s: %s", session, msg) }
}

// noa 是 norma v0.4.0 引入的「模型驱动上下文压缩」机制,作为平台实验功能由用户在
// 系统设置中开关。它与内置 compaction 互斥:noaadapter.Enable 是唯一入口,一次挂上
// 上下文接管器(Compactor)、Compress 工具与三段常驻提示词,不调用 Enable 即为关闭
// (内置 compaction 照常工作)。开关由每个 agent 注入的 noaEnabledFn 解析,每 run 读
// 一次,故切换只影响之后启动的 run,无需重建 agent。

// enableNoa 在解析器报告开启时把 noa 接入 opts。archiveRoot 是压缩原文的持久化基目录
// (取全局 workDir,各 agent 统一落在 <workDir>/noa 下,不随任务/意图目录分散),sessionID
// 命名其下的归档子目录(全局唯一,故同一基目录内不冲突)。
//
// noa 是实验功能:接入失败不得中断真实任务。发生错误时经 onWarn 上报并回退内置压缩。
// 启用成功时清掉 opts.Compaction,避免 agentcore 因「两个上下文管理器同时设置」告警。
// 返回接入的 noa 会话(可查看其解析后的配置);未开启或接入失败时返回 nil。
func enableNoa(opts *agentcore.Options, enabled func() bool, archiveRoot, sessionID string, onWarn func(string)) *noaadapter.Session {
	if enabled == nil || !enabled() {
		return nil
	}
	if opts.OnWarn == nil {
		opts.OnWarn = onWarn
	}
	// 接管前先把 profile 已解析的上下文窗口(如 250K、1M)交给 noa,否则其窗口会被悄悄换成
	// SDK 默认的 200K。MaxTokens 是单独的输出上限,不受影响;窗口缺失时保留 SDK 的 200K 兜底。
	window := 0
	if opts.Compaction != nil && opts.Compaction.ContextWindow > 0 {
		window = opts.Compaction.ContextWindow
	}
	session, err := noaadapter.EnableWithSession(opts, noaadapter.Options{
		ArchiveBaseDir:    filepath.Join(archiveRoot, "noa"),
		SessionID:         sessionID,
		OnWarn:            onWarn,
		ModelContextLimit: window,
	})
	if err != nil {
		if onWarn != nil {
			onWarn("noa 压缩启用失败,回退内置压缩:" + err.Error())
		}
		return nil
	}
	// Compactor 覆盖 Compaction,但两者并存时 agentcore 每次会告警;明确清掉。
	opts.Compaction = nil
	return session
}
