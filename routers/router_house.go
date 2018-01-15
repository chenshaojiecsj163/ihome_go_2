package routers

import (
	"github.com/astaxie/beego"
	"ihome_go_2/controllers"
)

func init() {
	//请求当前用户已发布房源信息
	beego.Router("api/v1.0/user/houses", &controllers.HouseController{}, "get:GetHouseInfo")
	beego.Router("api/v1.0/houses/:id/images", &controllers.HouseController{}, "post:UploadHouseImage")
	beego.Router("api/v1.0/houses/:id", &controllers.HouseController{}, "get:GetHouseInfoById")

}
