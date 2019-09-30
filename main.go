package main

import (
	"baiwan/controllers"
	"baiwan/sysconfig"
	"github.com/iris-contrib/middleware/cors"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/mvc"
)

func main() {
	app := iris.New()
	log := app.Logger()
	log.SetLevel("dubug")
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allows everything, use that to change the hosts.
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
	})
	app.Use(recover.New(), iris.Gzip, logger.New(), crs)
	app.AllowMethods(iris.MethodOptions)
	mvc.New(app.Party("/order")).Handle(new(controllers.OrderController))
	err := app.Run(iris.Addr(":" + sysconfig.SysConfig.Port))
	if err != nil {
		return
	}
}
