package controllers

import (
	"baiwan/dataSource"
	"baiwan/models"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/kataras/iris"
	"github.com/spf13/cast"
	"log"
	"sync"
)

type OrderController struct {
	Ctx iris.Context
}

//服务器启动查询库存记录将库存放入redis
func init() {
	var shop []*models.Shop
	dataSource.GetDB().Find(&shop)
	//键 "shop-"+店铺id+"-"+商品id
	for _, v := range shop {
		//mstore.Store("shop-"+cast.ToString(v.StoreID)+"-"+cast.ToString(v.ID), v.Stock)
		dataSource.RedisSource.RedisSetInterface("shop-"+cast.ToString(v.StoreID)+"-"+cast.ToString(v.ID), v.Stock)
	}
}

var i = 0

func (c *OrderController) PostPlaceOrder() (r *models.Result) {
	//fmt.Println("================")
	var po models.PlaceOrder
	c.Ctx.ReadJSON(&po)
	r = &models.Result{}
	//获取该商品库存
	key := "shop-" + cast.ToString(po.StoreID) + "-" + cast.ToString(po.ShopID)
	fmt.Println(key)
	conn := dataSource.RedisSource.RedisGetConn()
	defer conn.Close()
	_, err := conn.Do("multi")
	if err != nil {
		log.Println("开启事务失败")
		models.CreateResult(r, "", "下单失败", errors.New("err"))
		return
	}
	store, err := redis.String(conn.Do("get", key))
	if err != nil {
		log.Println("修改redis库存失败")
		models.CreateResult(r, "", "下单失败", errors.New("err"))
		return
	}
	_, err = conn.Do("set", key, cast.ToUint(store)-po.Count)
	if err != nil {
		log.Println("修改redis库存失败")
		models.CreateResult(r, "", "下单失败", errors.New("err"))
		return
	}
	b, err := json.Marshal(po)
	if err != nil {
		models.CreateResult(r, "", "参数异常", errors.New("err"))
		return
	}
	//将数据缓存进redis
	_, err = conn.Do("rpush", "order", string(b))
	if err != nil {
		log.Println("缓存进redis失败")
		models.CreateResult(r, "", "下单失败", errors.New("err"))
		return
	}
	_, err = conn.Do("exec")
	if err != nil {
		models.CreateResult(r, "", "下单失败", errors.New("err"))
		return
	}
	var mu sync.Mutex
	mu.Lock()
	i++
	mu.Unlock()
	fmt.Println(i, "****************")
	models.CreateResult(r, "下单成功", "", nil)
	return
}
