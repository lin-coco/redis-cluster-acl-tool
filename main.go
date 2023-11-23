package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/redis/go-redis/v9"
)

type Options struct {
	Addr     string `short:"a" long:"addr" description:"redis地址，例如: 127.0.0.1:6379" default:""`
	Password string `short:"p" long:"password" description:"默认用户密码" default:""`
	Acl      string `short:"c" long:"acl" description:"acl命令，例如: acl list" default:""`
}

type Redis struct {
	Type     string `yaml:"Type"` // default is standalone. values[standalone,cluster]
	Addr     string `yaml:"Addr"`
	Password string `yaml:"Password"`
}

var opt Options

func init() {
	if _, err := flags.Parse(&opt); err != nil {
		log.Fatal(err)
	}
	if opt.Addr == "" {
		opt.Addr = os.Getenv("ADDR")
	}
	if opt.Password == "" {
		opt.Password = os.Getenv("PASSWORD")
	}
	if opt.Acl == "" {
		opt.Acl = os.Getenv("ACL")
	}
	if opt.Addr == "" || opt.Password == "" || opt.Acl == "" {
		log.Fatal("请确保执行参数addr、password、acl不为空（或环境变量ADDR、PASSWORD、ACL）")
	}
}

func main() {
	log.Println(opt)
	fmt.Printf("\n--------------------测试集群连接--------------------\n")
	client := NewRedis(Redis{
		Type:     "cluster",
		Addr:     opt.Addr,
		Password: opt.Password,
	})
	defer func() {
		_ = client.Close()
	}()
	result, err := client.ClusterNodes(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed opening connection to redis: %v\n", err)
	}
	log.Printf("redis 集群节点 result:\n%v\n", result)

	fmt.Printf("\n--------------------节点执行acl--------------------\n")

	nodes := strings.Split(result, "\n")
	addrs := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node == "" {
			continue
		}
		if split := strings.Split(node, " "); len(split) < 2 {
			log.Fatal("cluster nodes 不足两列")
		} else {
			addr, _, found := strings.Cut(split[1], "@")
			if !found {
				log.Fatal("cluster nodes 第二列没有@")
			}
			addrs = append(addrs, addr)
		}
	}
	for _, addr := range addrs {
		nodeClient := NewRedis(Redis{
			Type:     "standalone",
			Addr:     addr,
			Password: opt.Password,
		})
		log.Printf("node client:%v exec: '%v'", addr, opt.Acl)
		aclList := strings.Split(opt.Acl, " ")
		aclInterface := make([]interface{}, len(aclList))
		for i, a := range aclList {
			aclInterface[i] = a
		}
		result, err := nodeClient.Do(context.Background(), aclInterface...).Result()
		if err != nil {
			log.Println(err)
		}
		if result != nil {
			log.Println(result)
		}
		_ = nodeClient.Close()
	}

	fmt.Printf("\nSuccess...Please make sure that the redis.conf configuration is the same as that of the ACL\n")
}

// NewRedis .
// 如果指定了 MasterName 选项，则返回 FailoverClient 哨兵客户端。
// 如果 Addrs 是 2 个以上的地址，则返回 ClusterClient 集群客户端。
// 其他情况，返回 Client 单节点客户端。
func NewRedis(c Redis) redis.UniversalClient {
	var err error
	var addrs = []string{c.Addr}

	if c.Type == "" || c.Type == "standalone" {
		// 单机
		addrs = []string{c.Addr}
	} else if c.Type == "cluster" {
		// 集群
		addrs = []string{c.Addr, c.Addr}
	} else {
		panic("Config Redis.Type has invalid value, recognized value[standalone,cluster] and default standalone")
	}

	client := redis.NewUniversalClient(
		&redis.UniversalOptions{
			Addrs:    addrs,
			Password: c.Password,
		},
	)
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed opening connection to redis: %v\n", err)
	}
	return client
}
