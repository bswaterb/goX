package braft

import (
	"github.com/bswaterb/goX/braft/log"
	"github.com/bswaterb/goX/braft/message"
	"sync"
	"time"
)

const (
	ROLE_FOLLOWER = iota
	ROLE_CANDIDATER
	ROLE_LEADER
)

type Raft struct {
	mu sync.Mutex

	// Node State
	state int64
	// 节点当前任期
	currentTerm uint64
	// 投票给了哪个候选者节点
	votedFor int64
	// 心跳间隔
	heartbeatInterval time.Time
	// 上次接收/发送心跳包的时间
	heartbeatTime time.Time
	// 已经提交的日志索引
	commitIdx int64

	// 状态机日志
	log []log.Entry
}

// ReplyVote follower 响应 candidate 的投票请求
func (rf *Raft) ReplyVote(args *message.RequestVoteArgs, reply *message.RequestVoteReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = rf.currentTerm
	reply.VoteGranted = false
	if args.Term < rf.currentTerm {
		return
	}

	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.state = ROLE_FOLLOWER
		rf.votedFor = -1
	}

	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		rf.votedFor = args.CandidateId
		reply.VoteGranted = true
		rf.heartbeatTime = time.Now()
	}
	return
}

// AppendEntries follower 追加日志
func (rf *Raft) AppendEntries(args *message.AppendEntriesArgs, reply *message.AppendEntriesReply) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	reply.Term = rf.currentTerm
	reply.Success = false
	if args.Term < rf.currentTerm {
		return
	}

	if args.Term > rf.currentTerm {
		reply.Term = args.Term
		rf.currentTerm = args.Term
	}

	rf.state = ROLE_FOLLOWER
	lastLogIdx := len(rf.log) - 1
	// 校验 leader 传来的日志条目索引与任期是否与本地日志记载相同
	if args.PrevLogIdx > rf.log[lastLogIdx].Idx ||
		rf.log[args.PrevLogIdx].Term != args.PrevLogTerm {
		return
	}

	reply.Success = true

	// 比较日志条目的任期，确认是否能够安全追加本次日志
	// (1) 如果两个节点的日志在相同索引位置上的任期号相同，则认为 <= 此索引的日志相同
	// (2) 如果给定索引的记录已提交，那么该索引前面的记录也已经提交

	idx := args.PrevLogIdx
	for i, entry := range args.Entries {
		idx++
		// 比较 leader 提供的日志与本地日志中的对应条目是否相同，截取掉本地不同的部分
		if idx < int64(len(rf.log)) {
			if rf.log[idx].Term == entry.Term {
				continue
			}
			rf.log = rf.log[:idx]
		}
		// 将 leader 的日志追加到本地
		rf.log = append(rf.log, args.Entries[i:]...)
		break
	}

	if rf.commitIdx < args.LeaderCommitIdx {
		lastLogIdx := rf.log[len(rf.log)-1].Idx
		// rf.commitIdx = min(lastLogIdx, args.LeaderCommitIdx)
		if args.LeaderCommitIdx > lastLogIdx {
			// ?
			rf.commitIdx = lastLogIdx
		} else {
			rf.commitIdx = args.LeaderCommitIdx
		}

		// 应用日志
		rf.apply()
	}

	rf.state = ROLE_FOLLOWER
	return

}

func (rf *Raft) apply() {
	// 应用 rf.log 日志中的命令
}
