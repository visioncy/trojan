package controller

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"math/rand"
	strconv "strconv"
	"strings"
	"time"
	"trojan/core"
	"trojan/trojan"
)

// UserList 获取用户列表
func UserList(requestUser string) *ResponseBody {
	responseBody := ResponseBody{Msg: "success"}
	defer TimeCost(time.Now(), &responseBody)
	mysql := core.GetMysql()
	userList, err := mysql.GetData()
	if err != nil {
		responseBody.Msg = err.Error()
		return &responseBody
	}
	if requestUser != "admin" {
		findUser := false
		for _, user := range userList {
			if user.Username == requestUser {
				userList = []*core.User{user}
				findUser = true
				break
			}
		}
		if !findUser {
			userList = []*core.User{}
		}
	}

	//User加入连接数资料
	UserListFillData(userList)

	domain, port := trojan.GetDomainAndPort()
	responseBody.Data = map[string]interface{}{
		"domain":   domain,
		"port":     port,
		"userList": userList,
	}
	return &responseBody
}
func UserListFillData(userList []*core.User) []*core.User {
	var trojanStatusMap = trojan.TrojanStatusMap()
	for _, user := range userList {
		if trojanStatusMap[user.EncryptPass].Status != nil {
			user.IpCurrent = trojanStatusMap[user.EncryptPass].Status.IpCurrent
		}
	}

	return userList
}

// PageUserList 分页查询获取用户列表
func PageUserList(curPage int, pageSize int) *ResponseBody {
	responseBody := ResponseBody{Msg: "success"}
	defer TimeCost(time.Now(), &responseBody)
	mysql := core.GetMysql()
	pageData, err := mysql.PageList(curPage, pageSize)
	if err != nil {
		responseBody.Msg = err.Error()
		return &responseBody
	}
	domain, port := trojan.GetDomainAndPort()
	responseBody.Data = map[string]interface{}{
		"domain":   domain,
		"port":     port,
		"pageData": pageData,
	}
	return &responseBody
}

// CreateUser 创建用户
func CreateUser(username string, password string) *ResponseBody {
	responseBody := ResponseBody{Msg: "success"}
	defer TimeCost(time.Now(), &responseBody)
	if username == "admin" {
		responseBody.Msg = "不能创建用户名为admin的用户!"
		return &responseBody
	}
	mysql := core.GetMysql()
	if user := mysql.GetUserByName(username); user != nil {
		responseBody.Msg = "已存在用户名为: " + username + " 的用户!"
		return &responseBody
	}
	pass, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		responseBody.Msg = "Base64解码失败: " + err.Error()
		return &responseBody
	}
	if user := mysql.GetUserByPass(password); user != nil {
		responseBody.Msg = "已存在密码为: " + string(pass) + " 的用户!"
		return &responseBody
	}
	if err := mysql.CreateUser(username, password, string(pass)); err != nil {
		responseBody.Msg = err.Error()
	}
	return &responseBody
}

// UpdateUser 更新用户
func UpdateUser(id uint, username string, password string) *ResponseBody {
	responseBody := ResponseBody{Msg: "success"}
	defer TimeCost(time.Now(), &responseBody)
	if username == "admin" {
		responseBody.Msg = "不能更改用户名为admin的用户!"
		return &responseBody
	}
	mysql := core.GetMysql()
	userList, err := mysql.GetData(strconv.Itoa(int(id)))
	if err != nil {
		responseBody.Msg = err.Error()
		return &responseBody
	}
	if userList[0].Username != username {
		if user := mysql.GetUserByName(username); user != nil {
			responseBody.Msg = "已存在用户名为: " + username + " 的用户!"
			return &responseBody
		}
	}
	pass, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		responseBody.Msg = "Base64解码失败: " + err.Error()
		return &responseBody
	}
	if userList[0].Password != password {
		if user := mysql.GetUserByPass(password); user != nil {
			responseBody.Msg = "已存在密码为: " + string(pass) + " 的用户!"
			return &responseBody
		}
	}
	if err := mysql.UpdateUser(id, username, password, string(pass)); err != nil {
		responseBody.Msg = err.Error()
	}
	return &responseBody
}

// DelUser 删除用户
func DelUser(id uint) *ResponseBody {
	responseBody := ResponseBody{Msg: "success"}
	defer TimeCost(time.Now(), &responseBody)
	mysql := core.GetMysql()
	if err := mysql.DeleteUser(id); err != nil {
		responseBody.Msg = err.Error()
	} else {
		trojan.Restart()
	}
	return &responseBody
}

// SetExpire 设置用户过期
func SetExpire(id uint, useDays uint) *ResponseBody {
	responseBody := ResponseBody{Msg: "success"}
	defer TimeCost(time.Now(), &responseBody)
	mysql := core.GetMysql()
	if err := mysql.SetExpire(id, useDays); err != nil {
		responseBody.Msg = err.Error()
	}
	return &responseBody
}

// CancelExpire 取消设置用户过期
func CancelExpire(id uint) *ResponseBody {
	responseBody := ResponseBody{Msg: "success"}
	defer TimeCost(time.Now(), &responseBody)
	mysql := core.GetMysql()
	if err := mysql.CancelExpire(id); err != nil {
		responseBody.Msg = err.Error()
	}
	return &responseBody
}

// ClashSubInfo 获取clash订阅信息
func ClashSubInfo(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(200, "token is null")
		return
	}
	decodeByte, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		c.String(200, "token is error")
		return
	}
	if !gjson.GetBytes(decodeByte, "user").Exists() || !gjson.GetBytes(decodeByte, "pass").Exists() {
		c.String(200, "token is error")
		return
	}
	username := gjson.GetBytes(decodeByte, "user").String()
	password := gjson.GetBytes(decodeByte, "pass").String()

	mysql := core.GetMysql()
	user := mysql.GetUserByName(username)
	if user != nil {
		pass, _ := base64.StdEncoding.DecodeString(user.Password)
		if password == string(pass) {
			var wsData, wsHost string
			userInfo := fmt.Sprintf("upload=%d, download=%d", user.Upload, user.Download)
			if user.Quota != -1 {
				userInfo = fmt.Sprintf("%s, total=%d", userInfo, user.Quota)
			}
			if user.ExpiryDate != "" {
				utc, _ := time.LoadLocation("Asia/Shanghai")
				t, _ := time.ParseInLocation("2006-01-02", user.ExpiryDate, utc)
				userInfo = fmt.Sprintf("%s, expire=%d", userInfo, t.Unix())
			}
			c.Header("content-disposition", fmt.Sprintf("attachment; filename=%s", user.Username))
			c.Header("subscription-userinfo", userInfo)

			domain, port := trojan.GetDomainAndPort()
			name := fmt.Sprintf("%s:%d", domain, port)
			configData := string(core.Load(""))
			if gjson.Get(configData, "websocket").Exists() && gjson.Get(configData, "websocket.enabled").Bool() {
				if gjson.Get(configData, "websocket.host").Exists() {
					hostTemp := gjson.Get(configData, "websocket.host").String()
					if hostTemp != "" {
						wsHost = fmt.Sprintf(", headers: {Host: %s}", hostTemp)
					}
				}
				wsOpt := fmt.Sprintf("{path: %s%s}", gjson.Get(configData, "websocket.path").String(), wsHost)
				wsData = fmt.Sprintf(", network: ws, udp: true, ws-opts: %s", wsOpt)
			}
			proxyData := fmt.Sprintf("  - {name: %s, server: %s, port: %d, type: trojan, password: %s, sni: %s%s}",
				name, domain, port, password, domain, wsData)
			result := fmt.Sprintf(`proxies:
%s

proxy-groups:
  - name: PROXY
    type: select
    proxies:
      - %s

%s
`, proxyData, name, clashRules())
			c.String(200, result)
			return
		}
	}
	c.String(200, "token is error")
}
func ClashSubInfoMulti(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.String(200, "token is null")
		return
	}
	decodeByte, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		c.String(200, "token is error")
		return
	}
	if !gjson.GetBytes(decodeByte, "user").Exists() || !gjson.GetBytes(decodeByte, "pass").Exists() {
		c.String(200, "token is error")
		return
	}
	username := gjson.GetBytes(decodeByte, "user").String()
	password := gjson.GetBytes(decodeByte, "pass").String()

	mysql := core.GetMysql()
	user := mysql.GetUserByName(username)
	if user != nil {
		pass, _ := base64.StdEncoding.DecodeString(user.Password)
		if password == string(pass) {
			var wsData string
			var nwQuotaInfo string
			var nwExpireInfo string

			userInfo := fmt.Sprintf("upload=%d, download=%d", user.Upload, user.Download)
			nwInfo := fmt.Sprintf("已用=%.1fGB(下=%.1fGB,上=%.1fGB)", float64(user.Upload+user.Download)/1073741824,
				float64(user.Download)/1073741824, float64(user.Upload)/1073741824)

			if user.Quota != -1 {
				userInfo = fmt.Sprintf("%s, total=%d", userInfo, user.Quota)
				nwQuotaInfo = fmt.Sprintf("流量限制=%.1fGB", float64(user.Quota)/1073741824)
			} else {
				nwQuotaInfo = fmt.Sprintf("流量限制=无限")
			}
			if user.ExpiryDate != "" {
				utc, _ := time.LoadLocation("Asia/Shanghai")
				t, _ := time.ParseInLocation("2006-01-02", user.ExpiryDate, utc)
				userInfo = fmt.Sprintf("%s, expire=%d", userInfo, t.Unix())
				nwExpireInfo = fmt.Sprintf("用户=%s, 过期=%s", username, t.Format("2006-01-02"))
			} else {
				nwExpireInfo = fmt.Sprintf("用户=%s, 过期=无限期", username)
			}
			c.Header("content-disposition", fmt.Sprintf("attachment; filename=%s", user.Username))
			c.Header("subscription-userinfo", userInfo)

			domain, portMin, portMax, portNum, serverName, servers := trojan.GetDomain()
			//			name := fmt.Sprintf("%s:%d", domain, port)
			//			configData := string(core.Load(""))
			//			if gjson.Get(configData, "websocket").Exists() && gjson.Get(configData, "websocket.enabled").Bool() {
			//				if gjson.Get(configData, "websocket.host").Exists() {
			//					hostTemp := gjson.Get(configData, "websocket.host").String()
			//					if hostTemp != "" {
			//						wsHost = fmt.Sprintf(", headers: {Host: %s}", hostTemp)
			//					}
			//				}
			//				wsOpt := fmt.Sprintf("{path: %s%s}", gjson.Get(configData, "websocket.path").String(), wsHost)
			//				wsData = fmt.Sprintf(", network: ws, udp: true, ws-opts: %s", wsOpt)
			//			}
			//			proxyData := fmt.Sprintf("  - {name: %s, server: %s, port: %d, type: trojan, password: %s, sni: %s%s}",
			//				name, domain, port, password, domain, wsData)
			//			result := fmt.Sprintf(`proxies:
			//%s
			//
			//proxy-groups:
			// - name: PROXY
			//   type: select
			//   proxies:
			//     - %s
			//
			//%s
			//`, proxyData, name, clashRules())
			//
			//
			//	buffer.WriteString("a")
			//
			//}

			var result string
			//var portRnd int
			if portNum == 0 {
				portNum = 5
			}
			if portMin == 0 {
				portMin = 40000
			}
			if portMax == 0 {
				portMax = 50000
			}
			if serverName == "" {
				serverName = "USS-"
			}

			//如果域名没配置,使用默认域名填上
			if len(servers) == 0 {
				servers = []string{serverName + "," + domain}
			}

			nwInfoServers := []string{nwExpireInfo, nwQuotaInfo, nwInfo}

			//加入流量信息
			for srvi := 0; srvi < len(nwInfoServers); srvi++ {
				wsData = fmt.Sprintf(
					`trojan://%s@%s:%d#%s
`, password, "127.0.0.1", 555, nwInfoServers[srvi])
				result = result + wsData
			}

			//			for srvi := 0; srvi < len(servers); srvi++ {
			//				//var srvName, srvDomain string
			//				serverSplit := strings.Split(servers[srvi], ",")
			//				srvName := serverSplit[0]
			//				srvDomain := serverSplit[1]
			//				srvport := serverSplit[2]
			//				srvtype := serverSplit[3]
			//				srvhost := serverSplit[4]
			//				srvPath := serverSplit[5]
			//				//trojan://[PASSWORD]@[SERVERHOST]:[PORT]?type=[TYPE]&host=[HOST]&path=[PATH]#[NODE_NAME]
			//				//trojan://password@example.com:443?type=ws&host=yourdomain.com&path=/ray#MyTrojanNode
			//				for i := 1; i <= portNum; i++ {
			//					portRnd = rand.Intn(portMax-portMin) + portMin
			//					wsData = fmt.Sprintf(
			//						`trojan://%s@%s:%d#%s
			//`, password, srvDomain, portRnd, srvName+strconv.Itoa(i))
			//					result = result + wsData
			//				}
			//			}
			for srvi := 0; srvi < len(servers); srvi++ {
				serverSplit := strings.Split(servers[srvi], ",")
				srvName := serverSplit[0]
				srvDomain := serverSplit[1]

				// Initialize optional fields as empty strings
				srvport := ""
				srvtype := ""
				srvhost := ""
				srvPath := ""

				// Only assign if the index exists
				if len(serverSplit) > 2 {
					srvport = serverSplit[2]
				}
				if len(serverSplit) > 3 {
					srvtype = serverSplit[3]
				}
				if len(serverSplit) > 4 {
					srvhost = serverSplit[4]
				}
				if len(serverSplit) > 5 {
					srvPath = serverSplit[5]
				}

				// Build query parameters (only if they exist)
				var queryParams []string
				if srvtype != "" {
					queryParams = append(queryParams, "type="+srvtype)
				}
				if srvhost != "" {
					queryParams = append(queryParams, "host="+srvhost)
				}
				if srvPath != "" {
					queryParams = append(queryParams, "path="+srvPath)
				}

				queryString := ""
				if len(queryParams) > 0 {
					queryString = "?" + strings.Join(queryParams, "&")
				}

				// Use srvport if available, otherwise random port
				port := 0
				if srvport != "" {
					port, _ = strconv.Atoi(srvport)
				} else {
					port = rand.Intn(portMax-portMin) + portMin
				}

				//trojan://[PASSWORD]@[SERVERHOST]:[PORT]?type=[TYPE]&host=[HOST]&path=[PATH]#[NODE_NAME]
				//trojan://password@example.com:443?type=ws&host=yourdomain.com&path=/ray#MyTrojanNode
				for i := 1; i <= portNum; i++ {
					wsData = fmt.Sprintf(
						`trojan://%s@%s:%d%s#%s
`, password, srvDomain, port, queryString, srvName)
					result = result + wsData
				}
			}

			c.String(200, base64.StdEncoding.EncodeToString([]byte(result)))
			return
		}
	}
	c.String(200, "token is error")
}
