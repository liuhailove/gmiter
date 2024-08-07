package redis

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"

	"github.com/liuhailove/gmiter/core/base"
	"github.com/liuhailove/gmiter/core/config"
	"github.com/liuhailove/gmiter/core/flow"
	"github.com/liuhailove/gmiter/logging"
)

// 定义Lua脚本
var luaScript = `
		local resourceName = KEYS[1]
		local globalThreshold = tonumber(ARGV[1])
		local acquireCount = tonumber(ARGV[2])
		local current = tonumber(redis.call('get', resourceName) or "0")
		local ttl = tonumber(redis.call('pttl', resourceName) or "-1")
		-- 如果TTL大于1000，则说明系统时钟出现了问题，此处重新设置Key的过期时间
		if ttl > 1000 then
			redis.call('pexpire', resourceName, 1000)
		end
		if current + acquireCount > globalThreshold then
			return 0
		else
			redis.call('incrby', resourceName, acquireCount)
			if ttl < 0 then
				redis.call('pexpire', resourceName, 1000)
			end
			return 1
		end
	`

var (
	jsonTraffic = jsoniter.ConfigCompatibleWithStandardLibrary
)

const (
	DefaultSeaPrefix = "{Sea}"
)

// RedisClusterTokenService redis集群Token服务
type RedisClusterTokenService struct {
	base.TokenService
	client    redis.UniversalClient
	scriptSha string
}

func NewRedisClient(conf *config.RedisClusterConfig) (*RedisClusterTokenService, error) {
	// redis客户端地址不可以为空
	if len(conf.Host) <= 0 {
		logging.Error(errors.New("redis cluster host cannot be empty"), "redis host  cannot be empty")
		return nil, errors.New("redis host  cannot be empty")
	}

	var redisAddrList []string = nil
	if strings.Contains(conf.Host, ":") {
		redisAddrList = strings.Split(conf.Host, ",")
	} else {
		addrs, err := net.LookupHost(conf.Host)
		if err != nil {
			logging.Error(errors.New("redis LookupHost fail"), "redis LookupHost fail")
			return nil, errors.New("redis LookupHost fail")
		}
		redisAddrList = make([]string, len(addrs))
		for index, ip := range addrs {
			redisAddrList[index] = fmt.Sprintf("%s:%d", ip, conf.Port)
		}
	}
	var redisClient redis.UniversalClient
	if conf.IsCluster {
		redisClient = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:        redisAddrList,
			Password:     conf.Password,
			DB:           0,
			PoolSize:     10,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		})
	} else {
		redisClient = redis.NewClient(&redis.Options{
			Addr:         redisAddrList[0],
			Password:     conf.Password,
			DB:           0,
			PoolSize:     10,
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 1 * time.Second,
		})
	}
	// 将Lua脚本加载到Redis中
	scriptSha, err := redis.NewScript(luaScript).Load(context.Background(), redisClient).Result()
	// 流控规则(无论什么时候对象都要存在，要不然引用时会存在空指针)
	tokenServiceClient := &RedisClusterTokenService{client: redisClient, scriptSha: scriptSha}
	if err != nil {
		logging.Error(err, "redis cluster load script error")
		//Lua 脚本上传失败 scriptSha置为空""，后续直接通过脚本调用
		tokenServiceClient.scriptSha = ""
	}
	return tokenServiceClient, nil
}

// RequestToken 从远程TokenServer请求tokens
//
// @param rule 规则信息
// @param acquireCount 请求token的数量
// @param prioritized 请求是否需要优先处理
// @return token请求处理结果
func (r *RedisClusterTokenService) RequestToken(rule string, acquireCount uint32, prioritized int32) *base.TokResult {
	var ru = new(flow.Rule)
	var err = jsonTraffic.Unmarshal([]byte(rule), ru)
	if err != nil {
		return base.BadResult
	}
	if r.notValidRequestSimple(ru.ID, acquireCount) {
		return base.BadResult
	}
	return r.acquireClusterToken(r.client, ru, acquireCount, prioritized)
}

// RequestParamToken 从远程TokenServer为一个具体的参数请求Token
//
// @param rule 规则信息
// @param acquireCount 请求token的数量
// @param params 参数列表
// @return token请求处理结果
func (r *RedisClusterTokenService) RequestParamToken(rule string, acquireCount uint32, params []interface{}) *base.TokResult {
	//TODO implement me
	panic("implement me")
}

// RequestConcurrentToken 从远程TokenServer获取的并发token数
//
// @param rule 规则信息
// @param acquireCount 并发获取的token数
// @return token请求处理结果
func (r *RedisClusterTokenService) RequestConcurrentToken(rule string, acquireCount uint32) *base.TokResult {
	//TODO implement me
	panic("implement me")
}

// ReleaseConcurrentToken 远程TokenServer异步释放Token数
//
// @param rule 规则信息
// @param tokenId 全局唯一tokenId
func (r *RedisClusterTokenService) ReleaseConcurrentToken(rule string, tokenId string) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClusterTokenService) notValidRequestSimple(id string, count uint32) bool {
	return id == "" || count <= 0
}

func (r *RedisClusterTokenService) notValidRequest(address string, id string, count uint32) bool {
	return address == "" || id == "" || count <= 0
}

func (r *RedisClusterTokenService) acquireClusterToken(client redis.UniversalClient, rule *flow.Rule, acquireCount uint32, prioritized int32) *base.TokResult {
	var result interface{}
	var err error
	if r.scriptSha == "" {
		result, err = client.Eval(context.Background(), luaScript, []string{DefaultSeaPrefix + "_" + config.AppName() + "_" + rule.ID + "_" + rule.Resource}, rule.ClusterConfig.GlobalThreshold, acquireCount).Result()
	} else {
		result, err = client.EvalSha(context.Background(), r.scriptSha, []string{DefaultSeaPrefix + "_" + config.AppName() + "_" + rule.ID + "_" + rule.Resource}, rule.ClusterConfig.GlobalThreshold, acquireCount).Result()

	}
	if err != nil {
		var e *net.OpError
		if errors.As(err, &e) && strings.Contains(e.Err.Error(), "connect") {
			// 重建，等待下次恢复
			_ = redisTokenServiceInst.ReInitial()
		} else if strings.HasPrefix(err.Error(), "NOSCRIPT ") {
			// 重建，等待下次恢复
			_ = redisTokenServiceInst.ReInitial()
		}
		logging.Error(err, "redis cluster load script error")
		return base.FailResult
	}
	if result.(int64) == 1 {
		return base.StatusOkResult
	}
	return base.BlockedResult
}

func (r *RedisClusterTokenService) Destroy() {
	if r.client != nil {
		var err = r.client.Close()
		if err != nil {
			logging.Error(err, "redis cluster close error")
		}
	}
}
