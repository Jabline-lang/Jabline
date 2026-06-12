package stdlib

import (
	"context"
	"fmt"
	"jabline/pkg/object"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClients   = make(map[string]*redis.Client)
	redisClientsMu sync.Mutex
)

var RedisBuiltins = []struct {
	Name   string
	Object object.Object
}{
	{"redis_connect", &object.Builtin{Fn: redisConnect}},
	{"redis_get", &object.Builtin{Fn: redisGet}},
	{"redis_set", &object.Builtin{Fn: redisSet}},
	{"redis_del", &object.Builtin{Fn: redisDel}},
	{"redis_exists", &object.Builtin{Fn: redisExists}},
	{"redis_expire", &object.Builtin{Fn: redisExpire}},
	{"redis_incr", &object.Builtin{Fn: redisIncr}},
	{"redis_hset", &object.Builtin{Fn: redisHSet}},
	{"redis_hget", &object.Builtin{Fn: redisHGet}},
	{"redis_hgetall", &object.Builtin{Fn: redisHGetAll}},
	{"redis_hdel", &object.Builtin{Fn: redisHDel}},
	{"redis_lpush", &object.Builtin{Fn: redisLPush}},
	{"redis_rpush", &object.Builtin{Fn: redisRPush}},
	{"redis_lpop", &object.Builtin{Fn: redisLPop}},
	{"redis_rpop", &object.Builtin{Fn: redisRPop}},
	{"redis_llen", &object.Builtin{Fn: redisLLen}},
	{"redis_lrange", &object.Builtin{Fn: redisLRange}},
	{"redis_sadd", &object.Builtin{Fn: redisSAdd}},
	{"redis_smembers", &object.Builtin{Fn: redisSMembers}},
	{"redis_srem", &object.Builtin{Fn: redisSRem}},
	{"redis_ping", &object.Builtin{Fn: redisPing}},
	{"redis_close", &object.Builtin{Fn: redisClose}},
}

func init() {
	NativeModuleRegistry["_redis"] = RedisBuiltins
	NativeModulePrefixes["_redis"] = "redis_"
}

func getRedisClient(name string) (*redis.Client, error) {
	redisClientsMu.Lock()
	defer redisClientsMu.Unlock()
	client, ok := redisClients[name]
	if !ok {
		return nil, fmt.Errorf("redis client '%s' not found. call redis_connect first", name)
	}
	return client, nil
}

// redisOpCtx returns a timed context for Redis operations (10s timeout).
// The caller must call the returned cancel function to avoid a context leak.
func redisOpCtx() context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	time.AfterFunc(10*time.Second, cancel)
	return ctx
}

func redisConnect(args ...object.Object) object.Object {
	if len(args) < 2 {
		return newError("redis_connect expects at least 2 args: (name, addr, [password, db])")
	}
	name, ok := args[0].(*object.String)
	if !ok {
		return newError("name must be STRING, got %s", args[0].Type())
	}
	addr, ok := args[1].(*object.String)
	if !ok {
		return newError("addr must be STRING, got %s", args[1].Type())
	}

	password := ""
	if len(args) >= 3 {
		if p, ok := args[2].(*object.String); ok {
			password = p.Value
		}
	}

	db := 0
	if len(args) >= 4 {
		if d, ok := args[3].(*object.Integer); ok {
			db = int(d.Value)
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr.Value,
		Password: password,
		DB:       db,
	})

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		return newError("redis_connect failed: %s", err)
	}

	redisClientsMu.Lock()
	redisClients[name.Value] = client
	redisClientsMu.Unlock()

	return &object.Boolean{Value: true}
}

func redisGet(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_get expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	getCtx, getCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer getCancel()
	val, err := client.Get(getCtx, key.Value).Result()
	if err == redis.Nil {
		return &object.Null{}
	}
	if err != nil {
		return newError("redis_get failed: %s", err)
	}
	return &object.String{Value: val}
}

func redisSet(args ...object.Object) object.Object {
	if len(args) < 3 {
		return newError("redis_set expects at least 3 args: (name, key, value, [ttl_seconds])")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	val, ok := args[2].(*object.String)
	if !ok {
		return newError("value must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}

	ttl := time.Duration(0)
	if len(args) >= 4 {
		if t, ok := args[3].(*object.Integer); ok {
			ttl = time.Second * time.Duration(t.Value)
		}
	}

	setCtx, setCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer setCancel()
	if ttl > 0 {
		err = client.Set(setCtx, key.Value, val.Value, ttl).Err()
	} else {
		err = client.Set(setCtx, key.Value, val.Value, 0).Err()
	}
	if err != nil {
		return newError("redis_set failed: %s", err)
	}
	return &object.Boolean{Value: true}
}

func redisDel(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_del expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	delCtx, delCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer delCancel()
	n, err := client.Del(delCtx, key.Value).Result()
	if err != nil {
		return newError("redis_del failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisExists(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_exists expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	n, err := client.Exists(ctx2, key.Value).Result()
	if err != nil {
		return newError("redis_exists failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisExpire(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("redis_expire expects 3 args: (name, key, ttl_seconds)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	ttl, ok := args[2].(*object.Integer)
	if !ok {
		return newError("ttl must be INTEGER")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	ok, err = client.Expire(ctx2, key.Value, time.Second*time.Duration(ttl.Value)).Result()
	if err != nil {
		return newError("redis_expire failed: %s", err)
	}
	return &object.Boolean{Value: ok}
}

func redisIncr(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_incr expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	n, err := client.Incr(redisOpCtx(), key.Value).Result()
	if err != nil {
		return newError("redis_incr failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisHSet(args ...object.Object) object.Object {
	if len(args) != 4 {
		return newError("redis_hset expects 4 args: (name, key, field, value)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	field, ok := args[2].(*object.String)
	if !ok {
		return newError("field must be STRING")
	}
	val, ok := args[3].(*object.String)
	if !ok {
		return newError("value must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	n, err := client.HSet(redisOpCtx(), key.Value, field.Value, val.Value).Result()
	if err != nil {
		return newError("redis_hset failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisHGet(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("redis_hget expects 3 args: (name, key, field)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	field, ok := args[2].(*object.String)
	if !ok {
		return newError("field must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	val, err := client.HGet(redisOpCtx(), key.Value, field.Value).Result()
	if err == redis.Nil {
		return &object.Null{}
	}
	if err != nil {
		return newError("redis_hget failed: %s", err)
	}
	return &object.String{Value: val}
}

func redisHGetAll(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_hgetall expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	result, err := client.HGetAll(redisOpCtx(), key.Value).Result()
	if err != nil {
		return newError("redis_hgetall failed: %s", err)
	}
	pairs := make(map[object.HashKey]object.HashPair)
	for k, v := range result {
		ks := &object.String{Value: k}
		pairs[ks.HashKey()] = object.HashPair{Key: ks, Value: &object.String{Value: v}}
	}
	return &object.Hash{Pairs: pairs}
}

func redisHDel(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("redis_hdel expects 3 args: (name, key, field)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	field, ok := args[2].(*object.String)
	if !ok {
		return newError("field must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	n, err := client.HDel(redisOpCtx(), key.Value, field.Value).Result()
	if err != nil {
		return newError("redis_hdel failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisLPush(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("redis_lpush expects 3 args: (name, key, value)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	val, ok := args[2].(*object.String)
	if !ok {
		return newError("value must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	n, err := client.LPush(redisOpCtx(), key.Value, val.Value).Result()
	if err != nil {
		return newError("redis_lpush failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisRPush(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("redis_rpush expects 3 args: (name, key, value)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	val, ok := args[2].(*object.String)
	if !ok {
		return newError("value must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	n, err := client.RPush(redisOpCtx(), key.Value, val.Value).Result()
	if err != nil {
		return newError("redis_rpush failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisLPop(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_lpop expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	val, err := client.LPop(redisOpCtx(), key.Value).Result()
	if err == redis.Nil {
		return &object.Null{}
	}
	if err != nil {
		return newError("redis_lpop failed: %s", err)
	}
	return &object.String{Value: val}
}

func redisRPop(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_rpop expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	val, err := client.RPop(redisOpCtx(), key.Value).Result()
	if err == redis.Nil {
		return &object.Null{}
	}
	if err != nil {
		return newError("redis_rpop failed: %s", err)
	}
	return &object.String{Value: val}
}

func redisLLen(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_llen expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	n, err := client.LLen(redisOpCtx(), key.Value).Result()
	if err != nil {
		return newError("redis_llen failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisLRange(args ...object.Object) object.Object {
	if len(args) != 4 {
		return newError("redis_lrange expects 4 args: (name, key, start, stop)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	start, ok := args[2].(*object.Integer)
	if !ok {
		return newError("start must be INTEGER")
	}
	stop, ok := args[3].(*object.Integer)
	if !ok {
		return newError("stop must be INTEGER")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	vals, err := client.LRange(redisOpCtx(), key.Value, start.Value, stop.Value).Result()
	if err != nil {
		return newError("redis_lrange failed: %s", err)
	}
	elements := make([]object.Object, len(vals))
	for i, v := range vals {
		elements[i] = &object.String{Value: v}
	}
	return &object.Array{Elements: elements}
}

func redisSAdd(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("redis_sadd expects 3 args: (name, key, member)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	member, ok := args[2].(*object.String)
	if !ok {
		return newError("member must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	n, err := client.SAdd(redisOpCtx(), key.Value, member.Value).Result()
	if err != nil {
		return newError("redis_sadd failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisSMembers(args ...object.Object) object.Object {
	if len(args) != 2 {
		return newError("redis_smembers expects 2 args: (name, key)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	members, err := client.SMembers(redisOpCtx(), key.Value).Result()
	if err != nil {
		return newError("redis_smembers failed: %s", err)
	}
	elements := make([]object.Object, len(members))
	for i, m := range members {
		elements[i] = &object.String{Value: m}
	}
	return &object.Array{Elements: elements}
}

func redisSRem(args ...object.Object) object.Object {
	if len(args) != 3 {
		return newError("redis_srem expects 3 args: (name, key, member)")
	}
	name, _ := args[0].(*object.String)
	key, ok := args[1].(*object.String)
	if !ok {
		return newError("key must be STRING")
	}
	member, ok := args[2].(*object.String)
	if !ok {
		return newError("member must be STRING")
	}
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	n, err := client.SRem(redisOpCtx(), key.Value, member.Value).Result()
	if err != nil {
		return newError("redis_srem failed: %s", err)
	}
	return &object.Integer{Value: n}
}

func redisPing(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("redis_ping expects 1 arg: (name)")
	}
	name, _ := args[0].(*object.String)
	client, err := getRedisClient(name.Value)
	if err != nil {
		return newError("%s", err.Error())
	}
	err = client.Ping(redisOpCtx()).Err()
	if err != nil {
		return &object.Boolean{Value: false}
	}
	return &object.Boolean{Value: true}
}

func redisClose(args ...object.Object) object.Object {
	if len(args) != 1 {
		return newError("redis_close expects 1 arg: (name)")
	}
	name, _ := args[0].(*object.String)

	redisClientsMu.Lock()
	defer redisClientsMu.Unlock()

	client, ok := redisClients[name.Value]
	if !ok {
		return &object.Boolean{Value: false}
	}
	client.Close()
	delete(redisClients, name.Value)
	return &object.Boolean{Value: true}
}

// closeAllRedis closes all open Redis connections.
func closeAllRedis() {
	redisClientsMu.Lock()
	defer redisClientsMu.Unlock()
	for name, client := range redisClients {
		client.Close()
		delete(redisClients, name)
	}
}

