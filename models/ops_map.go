package models

import (
	"strings"
)

type OpType string

const (
	GetOp     OpType = "GET"
	SetOp     OpType = "SET"
	DelOp     OpType = "DEL"
	EvalOp    OpType = "Eval"
	TotalOp   OpType = "Total"
	UnknownOp OpType = "Unknown"
)

// Map of Redis commands to their operation category
var redisCommandType = map[string]OpType{
	// GET-like — reads
	"GET":      GetOp,
	"MGET":     GetOp,
	"GETRANGE": GetOp,
	"STRLEN":   GetOp,
	"EXISTS":   GetOp,
	"TTL":      GetOp,
	"PTTL":     GetOp,
	"TYPE":     GetOp,
	"KEYS":     GetOp,
	"SCAN":     GetOp,
	"GETBIT":   GetOp,
	"BITCOUNT": GetOp,
	"BITPOS":   GetOp,
	// Hash reads
	"HGET":    GetOp,
	"HMGET":   GetOp,
	"HGETALL": GetOp,
	"HVALS":   GetOp,
	"HKEYS":   GetOp,
	"HLEN":    GetOp,
	"HEXISTS": GetOp,
	"HSCAN":   GetOp,
	// List reads
	"LRANGE": GetOp,
	"LLEN":   GetOp,
	"LINDEX": GetOp,
	"LPOS":   GetOp,
	// Set reads
	"SCARD":       GetOp,
	"SISMEMBER":   GetOp,
	"SMISMEMBER":  GetOp,
	"SMEMBERS":    GetOp,
	"SRANDMEMBER": GetOp,
	"SSCAN":       GetOp,
	// Sorted set reads
	"ZRANGE":           GetOp,
	"ZREVRANGE":        GetOp,
	"ZRANGEBYSCORE":    GetOp,
	"ZREVRANGEBYSCORE": GetOp,
	"ZRANGEBYLEX":      GetOp,
	"ZCARD":            GetOp,
	"ZRANK":            GetOp,
	"ZREVRANK":         GetOp,
	"ZSCORE":           GetOp,
	"ZMSCORE":          GetOp,
	"ZCOUNT":           GetOp,
	"ZLEXCOUNT":        GetOp,
	"ZSCAN":            GetOp,
	// Stream reads
	"XRANGE":    GetOp,
	"XREVRANGE": GetOp,
	"XREAD":     GetOp,
	"XLEN":      GetOp,
	"XINFO":     GetOp,
	"XPENDING":  GetOp,
	// HyperLogLog reads
	"PFCOUNT": GetOp,
	// Geo reads
	"GEOPOS":    GetOp,
	"GEODIST":   GetOp,
	"GEOSEARCH": GetOp,

	// SET-like — writes
	"SET":         SetOp,
	"MSET":        SetOp,
	"MSETNX":      SetOp,
	"SETNX":       SetOp,
	"SETEX":       SetOp,
	"PSETEX":      SetOp,
	"APPEND":      SetOp,
	"INCR":        SetOp,
	"DECR":        SetOp,
	"INCRBY":      SetOp,
	"DECRBY":      SetOp,
	"INCRBYFLOAT": SetOp,
	"SETBIT":      SetOp,
	"SETRANGE":    SetOp,
	"GETSET":      SetOp,
	"GETDEL":      SetOp,
	"GETEX":       SetOp,
	"EXPIRE":      SetOp,
	"PEXPIRE":     SetOp,
	"EXPIREAT":    SetOp,
	"PEXPIREAT":   SetOp,
	"PERSIST":     SetOp,
	"RENAME":      SetOp,
	"RENAMENX":    SetOp,
	"MOVE":        SetOp,
	"COPY":        SetOp,
	// Hash writes
	"HSET":         SetOp,
	"HMSET":        SetOp,
	"HSETNX":       SetOp,
	"HINCRBY":      SetOp,
	"HINCRBYFLOAT": SetOp,
	// List writes
	"LPUSH":   SetOp,
	"RPUSH":   SetOp,
	"LPUSHX":  SetOp,
	"RPUSHX":  SetOp,
	"LSET":    SetOp,
	"LINSERT": SetOp,
	// Set writes
	"SADD": SetOp,
	// Sorted set writes
	"ZADD":    SetOp,
	"ZINCRBY": SetOp,
	// Stream writes
	"XADD":   SetOp,
	"XACK":   SetOp,
	"XCLAIM": SetOp,
	// HyperLogLog writes
	"PFADD":   SetOp,
	"PFMERGE": SetOp,
	// Geo writes
	"GEOADD": SetOp,

	// DEL-like — deletes
	"DEL":      DelOp,
	"UNLINK":   DelOp,
	"FLUSHDB":  DelOp,
	"FLUSHALL": DelOp,
	// List deletes
	"LPOP":  DelOp,
	"RPOP":  DelOp,
	"LREM":  DelOp,
	"LTRIM": DelOp,
	// Set deletes
	"SPOP": DelOp,
	"SREM": DelOp,
	// Hash deletes
	"HDEL": DelOp,
	// Sorted set deletes
	"ZREM":             DelOp,
	"ZREMRANGEBYRANK":  DelOp,
	"ZREMRANGEBYSCORE": DelOp,
	"ZREMRANGEBYLEX":   DelOp,
	"ZPOPMIN":          DelOp,
	"ZPOPMAX":          DelOp,
	// Stream deletes
	"XDEL":  DelOp,
	"XTRIM": DelOp,

	// EVAL-like
	"EVAL":    EvalOp,
	"EVALSHA": EvalOp,
}

func GetOpType(command string) OpType {
	upper := strings.ToUpper(command)
	if op, ok := redisCommandType[upper]; ok {
		return op
	}
	return UnknownOp
}
