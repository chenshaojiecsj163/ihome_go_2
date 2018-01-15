package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	"github.com/astaxie/beego/orm"
	"ihome_go_2/models"
	"path"
	"strconv"
)

type HouseController struct {
	beego.Controller
}

func (this *HouseController) RetData(resp interface{}) {
	this.Data["json"] = resp
	this.ServeJSON()
}

func (this *HouseController) UploadHouseImage() {
	beego.Info("house UploadHouseImage is called")

	resp := Resp{Errno: models.RECODE_OK, Errmsg: models.RecodeText(models.RECODE_OK)}
	defer this.RetData(&resp)

	houseId := this.Ctx.Input.Param(":id")
	file, header, err := this.GetFile("house_image")
	if err != nil {
		resp.Errno = models.RECODE_SERVERERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}

	//创建buffer接受file
	buffer := make([]byte, header.Size)
	_, err = file.Read(buffer)
	if err != nil {
		resp.Errno = models.RECODE_IOERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}

	suffix := path.Base(header.Filename)
	_, fileId, err1 := models.FDFSUploadByBuffer(buffer, suffix)
	if err1 != nil {
		resp.Errno = models.RECODE_IOERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}
	//数据库相关操作
	o := orm.NewOrm()

	house := models.House{}
	house.Id, _ = strconv.Atoi(houseId)
	err = o.Read(&house)
	if err == orm.ErrNoRows || err == orm.ErrMissPK {
		resp.Errno = models.RECODE_DBERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}

	//判断房屋中的index_image_url是否为空
	if house.Index_image_url == "" {
		house.Index_image_url = fileId
		if _, err = o.Update(&house); err != nil {
			resp.Errno = models.RECODE_DBERR
			resp.Errmsg = models.RecodeText(resp.Errno)
			return
		}
	}

	imageinfo := models.HouseImage{Url: fileId, House: &house}

	if _, err = o.Insert(&imageinfo); err != nil {
		resp.Errno = models.RECODE_DBERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}
	//将fileid 拼接一个完整的url路径 + ip + port 返回给前端
	image_url := "http://192.168.137.100:9091/" + fileId

	url_map := make(map[string]interface{})
	url_map["avatar_url"] = image_url
	resp.Data = url_map

	return
}

func (this *HouseController) GetHouseInfo() {
	beego.Info("Houses GetHouseInfo is called")
	resp := Resp{Errno: models.RECODE_OK, Errmsg: models.RecodeText(models.RECODE_OK)}
	defer this.RetData(&resp)

	userId := this.GetSession("user_id")
	var houseInfo []models.House
	o := orm.NewOrm()

	qs := o.QueryTable("house").Filter("id", userId.(int))
	_, err := qs.All(&houseInfo)
	if err != nil {
		resp.Errno = models.RECODE_DBERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}
	//组织json
	//	data = make(map[string]interface{})
	//	data["houses"]=
	//	resp.Data = data
	return

}

func (this *HouseController) GetHouseInfoById() {
	beego.Info("Houses GetHouseInfoById is called")
	resp := Resp{Errno: models.RECODE_OK, Errmsg: models.RecodeText(models.RECODE_OK)}
	defer this.RetData(&resp)

	data := make(map[string]interface{})

	houseId := this.Ctx.Input.Param(":id")
	//	userId := this.GetSession("user_id")

	//从redis中读取数据
	//创建redis连接
	cache_conn, err := cache.NewCache("redis", `{"key":"ihome_go_2","conn":"127.0.0.1:6379","dbNum":"0"}`)
	if err != nil {
		resp.Errno = models.RECODE_DBERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}
	//尝试从redis中读取数据
	houseInfoKey := fmt.Sprintf("house_info_%s", houseId)
	houseInfoValue := cache_conn.Get(houseInfoKey)
	if houseInfoValue != nil {
		//????????????将查到信息返回

		resp.Data = data
		return
	}

	//redis中没有，从mysql数据库中读取
	o := orm.NewOrm()
	houseInfo := models.House{}
	houseInfo.Id, _ = strconv.Atoi(houseId)
	o.Read(&houseInfo)
	//????????????????????????

	//组织resp.data,返回
	resp.Data = data
	return

}
