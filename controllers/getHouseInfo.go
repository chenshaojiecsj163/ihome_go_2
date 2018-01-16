package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	"github.com/astaxie/beego/orm"
	"ihome_go_2/models"
	"path"
	"strconv"
	"time"
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

	//查询头像url
	var user = models.User{}
	if err := o.QueryTable("user").Filter("id", userId).One(&user); err != nil {
		resp.Errno = models.RECODE_DBERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}
	//组织头像url
	user_avatar_url := fmt.Sprintf("%s/%s", "http://192.168.137.100:9091", user.Avatar_url)

	qs := o.QueryTable("house").Filter("id", userId.(int))
	_, err := qs.All(&houseInfo)
	if err != nil {
		resp.Errno = models.RECODE_DBERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}

	//组织json
	houseList := []interface{}{}
	for _, house := range houseInfo {

		o.LoadRelated(&house, "Area")
		house_image_url := fmt.Sprintf("%s/%s", "http://192.168.137.100:9091", house.Index_image_url)
		m_house := map[string]interface{}{
			"address":     house.Address,
			"area_name":   house.Area.Name,
			"ctime":       house.Ctime,
			"house_id":    house.Id,
			"img_url":     house_image_url, //house.Index_image_url,
			"order_count": house.Order_count,
			"price":       house.Price,
			"room_count":  house.Room_count,
			"title":       house.Title,
			"user_avatar": user_avatar_url,
		}
		houseList = append(houseList, m_house)
	}
	data := make(map[string]interface{})
	data["houses"] = houseList
	resp.Data = data
	return

}

func mhouseinfo(m_house models.House, houseInfo map[string]interface{}) {
	beego.Info("mhouseinfo is called")

	houseInfo["acreage"] = m_house.Acreage
	houseInfo["address"] = m_house.Address
	houseInfo["beds"] = m_house.Beds
	houseInfo["capacity"] = m_house.Capacity

	houseInfo["deposit"] = m_house.Deposit
	houseInfo["hid"] = m_house.Id
	houseInfo["max_days"] = m_house.Max_days
	houseInfo["min_days"] = m_house.Min_days
	houseInfo["price"] = m_house.Price
	houseInfo["room_count"] = m_house.Acreage
	houseInfo["title"] = m_house.Title

	houseInfo["unit"] = m_house.Unit
	houseInfo["user_id"] = m_house.User.Id
	houseInfo["user_name"] = m_house.User.Name

	ava_url := fmt.Sprintf("%s/%s", "http://192.168.137.100:9091", m_house.User.Avatar_url)
	houseInfo["user_avatar"] = ava_url

	facility := []int{}
	for _, facility_ := range m_house.Facilities {
		facility = append(facility, facility_.Id)
	}
	houseInfo["facilities"] = facility

	imageUrl := []string{}
	for _, imgurl := range m_house.Images {
		url := fmt.Sprintf("%s/%s", "http://192.168.137.100:9091", imgurl.Url)
		imageUrl = append(imageUrl, url)
	}
	houseInfo["img_urls"] = imageUrl

	//评论。。。。。。。。又不会
	houseInfo["comments"] = ""
}
func (this *HouseController) GetHouseInfoById() {
	beego.Info("Houses GetHouseInfoById is called")
	resp := Resp{Errno: models.RECODE_OK, Errmsg: models.RecodeText(models.RECODE_OK)}
	defer this.RetData(&resp)

	data := make(map[string]interface{})

	houseId := this.Ctx.Input.Param(":id")
	userId := this.GetSession("user_id")

	//从redis中读取数据
	//创建redis连接
	cache_conn, err := cache.NewCache("redis", `{"key":"ihome_go_2","conn":"127.0.0.1:6379","dbNum":"0"}`)
	if err != nil {
		resp.Errno = models.RECODE_DBERR
		resp.Errmsg = models.RecodeText(resp.Errno)
		return
	}
	//尝试从redis中读取数据,不会写。。
	//????????????
	houseInfoKey := fmt.Sprintf("house_info_%s", houseId)
	houseInfoValue := cache_conn.Get(houseInfoKey)
	if houseInfoValue != nil {
		house_info := map[string]interface{}{}
		json.Unmarshal(houseInfoValue.([]byte), &house_info)
		data["house"] = house_info
		data["user_id"] = userId
		resp.Data = data
		return
	}
	//redis中没有，从mysql数据库中读取
	o := orm.NewOrm()
	houseInfo := models.House{}
	houseInfo.Id, _ = strconv.Atoi(houseId)
	o.Read(&houseInfo)
	//这儿也不会。抄的
	o.LoadRelated(&houseInfo, "Area")
	o.LoadRelated(&houseInfo, "User")
	o.LoadRelated(&houseInfo, "Images")
	o.LoadRelated(&houseInfo, "Facilities")

	//房屋信息存入redis

	beego.Info("---------------,", houseInfo, "----------------------")
	//...........不想写了，啥都不会
	a_houseInfo := make(map[string]interface{})
	mhouseinfo(houseInfo, a_houseInfo)
	houseInfoValue, _ = json.Marshal(a_houseInfo)
	cache_conn.Put(houseInfoKey, houseInfoValue, 3600*time.Second)
	data["user_id"] = userId
	data["house"] = a_houseInfo
	//组织resp.data,返回
	resp.Data = data
	return

}
