package dataSource

import (
	"baiwan/models"
	"baiwan/sysconfig"
	"encoding/json"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/cast"
	"log"
	"sync"
	"time"
)

var (
	pool *redis.Pool
)

//初始化一个pool
func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     150,
		MaxActive:   180,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("AUTH", password); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

type Redis struct {
	sync.RWMutex
	Pool *redis.Pool
}

var RedisSource Redis

func init() {
	pool = newPool(sysconfig.SysConfig.Redis.Addr, sysconfig.SysConfig.Redis.Password)
	RedisSource.Pool = pool
	log.Print("连接池初始化成功")
	/*for{
		err:=RedisSource.RedisLRPop("lpop","order")
		if err!=nil{
			fmt.Println(err)
			break
		}
	}*/
	//b,_:=json.Marshal(models.Order{ShopID:1,Price:10,Count:1})
	//RedisSource.RedisLRPush("rpush","order", string(b))
	//RedisSource.RedisDelKey("order")
	go func() {
		//处理订单数据
		for {
			time.Sleep(5 * time.Second)
			length, err := RedisSource.RedisGetLen("order")
			if err != nil || length == 0 {
				fmt.Println("order中没值了")
				continue
			}
			data, err := RedisSource.RedisGetList("order", 0, length)
			if err != nil {
				continue
			}
			sql := "insert into orders (created_at,shop_id,user_id,count) values "
			var b bool
			m := make(map[uint]uint)
			for _, v := range data {
				var po models.PlaceOrder
				err := json.Unmarshal([]byte(cast.ToString(v)), &po)
				if err != nil {
					b = true
					break
				}
				if v, ok := m[po.ShopID]; ok {
					m[po.ShopID] = v + po.Count
				} else {
					m[po.ShopID] = po.Count
				}
				sql += "(now()," + cast.ToString(po.ShopID) + "," + cast.ToString(po.UserID) + "," + cast.ToString(po.Count) + "),"
			}
			if b {
				continue
			}
			sql = sql[:len(sql)-1]
			tx := GetDB().Begin()
			err = tx.Exec(sql).Error
			if err != nil {
				tx.Rollback()
				continue
			}
			for k, v := range m {
				err = tx.Exec("update shops set stock=stock-? where id=?", v, k).Error
				if err != nil {
					tx.Rollback()
					b = true
					break
				}
			}
			if b {
				continue
			}
			err = RedisSource.RedisLDel("ltrim", "order", length+1, -1)
			if err != nil {
				tx.Rollback()
				continue
			} else {
				tx.Commit()
			}
		}
	}()
}

//删除键
func (r *Redis) RedisDelKey(key interface{}) error {
	r.Lock()
	defer r.Unlock()
	conn := r.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("del", key)
	return err
}

//执行redis命令
func (r *Redis) RedisGetConn() redis.Conn {
	r.Lock()
	defer r.Unlock()
	conn := r.Pool.Get()
	return conn
}

//获取list列表长度
func (r *Redis) RedisGetLen(key interface{}) (int, error) {
	r.Lock()
	defer r.Unlock()
	conn := r.Pool.Get()
	defer conn.Close()
	v, err := redis.Int(conn.Do("llen", key))
	//fmt.Println(v)
	return v, err
}

//获取list列表数据
func (r *Redis) RedisGetList(key interface{}, start, end int) ([]interface{}, error) {
	r.Lock()
	defer r.Unlock()
	conn := r.Pool.Get()
	defer conn.Close()
	v, err := redis.Values(conn.Do("lrange", key, start, end))
	return v, err
}
func (r *Redis) RedisSetInterface(key, value interface{}) error {
	r.Lock()
	defer r.Unlock()
	conn := r.Pool.Get()
	defer conn.Close()
	_, err := conn.Do("set", key, value)
	return err
}
func (r *Redis) RedisGet(key interface{}) (interface{}, error) {
	r.Lock()
	defer r.Unlock()
	conn := r.Pool.Get()
	defer conn.Close()
	str, err := conn.Do("get", key)
	return str, err
}

//add最上面的一个值lpush 删除最下面的一个值rpush
func (r *Redis) RedisLRPush(method string, key, value interface{}) error {
	r.RLock()
	defer r.RUnlock()
	conn := r.Pool.Get()
	defer conn.Close()
	_, err := conn.Do(method, key, value)
	return err
}

//删除最上面的一个值lpop lpop mylist 删除最下面的一个值rpop  rpop mylist
func (r *Redis) RedisLRPop(method string, key interface{}) error {
	r.RLock()
	defer r.RUnlock()
	conn := r.Pool.Get()
	defer conn.Close()
	v, err := redis.String(conn.Do(method, key))
	fmt.Println(v, "=====")
	return err
}

//lrem: lrem mylist 0 "value"从mylist中删除全部等值value的元素   0为全部，负值为从尾
//ltrim ltrim mylist 1 -1 保留mylist中 1到末尾的值，即删除第一个值
func (r *Redis) RedisLDel(method string, key, start, end interface{}) error {
	r.RLock()
	defer r.RUnlock()
	conn := r.Pool.Get()
	defer conn.Close()
	_, err := conn.Do(method, key, start, end)
	return err
}
