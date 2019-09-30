package models

import "github.com/jinzhu/gorm"

type Order struct {
	gorm.Model
	ShopID uint //商品id
	UserID uint //下单人
	Count  uint //购买数量
	Status uint //0 待付款 1已支付 2已失效
}
type Shop struct { //商品
	gorm.Model
	Total    int    //总量
	Stock    int    //库存
	StoreID  uint   //店铺id
	Price    int    //单价
	ShopName string //商品名称
}
type Store struct { //店铺
	gorm.Model
	StoreName string //店铺名称
}
type PlaceOrder struct {
	ShopID  uint //商品id
	StoreID uint //店铺id
	UserID  uint //下单人id
	Count   uint //数量
}
