package message

import "github.com/bswaterb/goX/braft/log"

type RequestVoteArgs struct {
	// 任期
	Term uint64
	// 候选人id
	CandidateId int64
}

type RequestVoteReply struct {
	Term uint64
	// 是否同意本次投票
	VoteGranted bool
}

type AppendEntriesArgs struct {
	// leader 所在任期号
	Term uint64
	// 发送该请求的 leader id
	LeaderId int64

	// 前一个日志条目的索引和任期号，用于接受者进行本地信息校验
	PrevLogIdx  int64
	PrevLogTerm uint64

	// leader 已提交的最大的日志索引，用于跟随者提交
	LeaderCommitIdx int64

	// 本次需要进行复制的日志内容，用于心跳消息时无需携带
	Entries []log.Entry
}

type AppendEntriesReply struct {
	Term    uint64
	Success bool
}
