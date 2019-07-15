package config

import (
	"encoding/json"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"os"
)

type config struct {
	local       *gjson.Result
	nacos       *gjson.Result
	nacosClient config_client.IConfigClient
}

var conf config

func init() {
	var configFile = os.Getenv("ARES_CONFIG_FILE")
	if configFile == "" {
		configFile = "config.json"
	}
	local := readFile(configFile)
	conf = config{
		local: local,
	}
	if local.Get("nacos.serverConfigs").Exists() && local.Get("nacos.dataId").Exists() {
		initNacos(local)
	}
}

func initNacos(local *gjson.Result) {
	serverConfigList := arraylist.New()
	nacosServerConfigs := local.Get("nacos.serverConfigs").Array()
	for _, serverConfig := range nacosServerConfigs {
		serverConfigList.Add(constant.ServerConfig{
			IpAddr:      serverConfig.Get("ipAddr").String(),
			ContextPath: serverConfig.Get("contextPath").String(),
			Port:        serverConfig.Get("port").Uint(),
		})
	}
	var serverConfigs []constant.ServerConfig
	serverConfigsJson, _ := serverConfigList.ToJSON()
	json.Unmarshal(serverConfigsJson, &serverConfigs)
	nacosConfigClient, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig": constant.ClientConfig{
			TimeoutMs:            30 * 1000, //http请求超时时间，单位毫秒
			ListenInterval:       10 * 1000, //监听间隔时间，单位毫秒（仅在ConfigClient中有效）
			BeatInterval:         5 * 1000,  //心跳间隔时间，单位毫秒（仅在ServiceClient中有效）
			NamespaceId:          local.Get("nacos.namespaceId").String(),
			UpdateThreadNum:      20,   //更新服务的线程数
			NotLoadCacheAtStart:  true, //在启动时不读取本地缓存数据，true--不读取，false--读取
			UpdateCacheWhenEmpty: true, //当服务列表为空时是否更新本地缓存，true--更新,false--不更新
		},
	})
	if err != nil {
		os.Exit(1)
	}
	conf.nacosClient = nacosConfigClient
	nacosConfig, err := conf.nacosClient.GetConfig(vo.ConfigParam{
		DataId: local.Get("nacos.dataId").String(),
		Group:  local.Get("nacos.group").String(),
	})
	if err != nil {
		os.Exit(1)
	}
	conf.nacos = readString(nacosConfig)

	conf.nacosClient.ListenConfig(vo.ConfigParam{
		DataId: local.Get("nacos.dataId").String(),
		Group:  local.Get("nacos.group").String(),
		OnChange: func(namespace, group, dataId, data string) {
			conf.nacos = readString(data)
		},
	})
}

func readFile(path string) *gjson.Result {
	fh, err := os.Open(path)
	if err != nil {
		return nil
	}
	bytes, err := ioutil.ReadAll(fh)
	if err != nil {
		return nil
	}
	result := gjson.Parse(string(bytes[:]))
	return &result
}

func readString(data string) *gjson.Result {
	result := gjson.Parse(data)
	return &result
}

func GetLocalConfig() *gjson.Result {
	return conf.local
}

func GeNacosConfig() *gjson.Result {
	return conf.nacos
}

func GeNacosClient() config_client.IConfigClient {
	return conf.nacosClient
}

func GetArrayString(name string) []string {
	if conf.nacos != nil {
		if result := conf.nacos.Get(name); result.Exists() && result.IsArray() {
			return convertResult2ArrayString(result)
		}
	}
	if result := conf.local.Get(name); result.Exists() && result.IsArray() {
		return convertResult2ArrayString(result)
	}
	return []string{}
}

func GetArrayInt64(name string) []int64 {
	if conf.nacos != nil {
		if result := conf.nacos.Get(name); result.Exists() && result.IsArray() {
			return convertResult2ArrayInt64(result)
		}
	}
	if result := conf.local.Get(name); result.Exists() && result.IsArray() {
		return convertResult2ArrayInt64(result)
	}
	return []int64{}
}

func convertResult2ArrayString(result gjson.Result) []string {
	var arrayStrings []string
	for _, result := range result.Array() {
		arrayStrings = append(arrayStrings, result.String())
	}
	return arrayStrings
}

func convertResult2ArrayInt64(result gjson.Result) []int64 {
	var arrayStrings []int64
	for _, result := range result.Array() {
		arrayStrings = append(arrayStrings, result.Int())
	}
	return arrayStrings
}

func getStruct(name string, dst interface{}) error {
	if conf.nacos != nil {
		if result := conf.nacos.Get(name); result.Exists() && result.IsArray() {
			return convertResult2Struct(result, &dst)
		}
	}
	if result := conf.local.Get(name); result.Exists() && result.IsArray() {
		return convertResult2Struct(result, &dst)
	}
	return nil
}

func convertResult2Struct(result gjson.Result, dst interface{}) error {
	return json.Unmarshal([]byte(result.Raw), &dst)
}

func GetString(name string, defaultValues ...string) string {
	if conf.nacos != nil && conf.nacos.Get(name).Exists() {
		return conf.nacos.Get(name).String()
	}
	if conf.local.Get(name).Exists() {
		return conf.local.Get(name).String()
	}
	if len(defaultValues) > 0 {
		return defaultValues[0]
	}
	return ""
}

func GetBool(name string, defaultValues ...bool) bool {
	if conf.nacos != nil && conf.nacos.Get(name).Exists() {
		return conf.nacos.Get(name).Bool()
	}
	if conf.local.Get(name).Exists() {
		return conf.local.Get(name).Bool()
	}
	if len(defaultValues) > 0 {
		return defaultValues[0]
	}
	return false
}

func GetInt64(name string, defaultValues ...int64) int64 {
	if conf.nacos != nil && conf.nacos.Get(name).Exists() {
		return conf.nacos.Get(name).Int()
	}
	if conf.local.Get(name).Exists() {
		return conf.local.Get(name).Int()
	}
	if len(defaultValues) > 0 {
		return defaultValues[0]
	}
	return 0
}

func GetInt(name string, defaultValues ...int) int {
	var val int64
	if len(defaultValues) > 0 {
		val = GetInt64(name, int64(defaultValues[0]))
	} else {
		val = GetInt64(name)
	}
	return int(val)
}
