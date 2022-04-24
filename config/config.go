package config

import (
	"bytes"
	"fmt"
	//"log"
	"os"

	"github.com/goinggo/mapstructure"
	"github.com/spf13/viper"
)

// FileAddr 远程/本地
type FileAddrType string

// FileType 支持的配置文件类型
type FileType string

const (
	LocalFile       FileAddrType = "local"
	RemoteFile      FileAddrType = "remote"
	JsonFile        FileType     = "json"
	YamlFile        FileType     = "yaml"
	TomlFile        FileType     = "toml"
	HCLFile         FileType     = "hcl"
	INIFile         FileType     = "ini"
	EnvFile         FileType     = "env"
	PropertiesFiles FileType     = "properties"
	LincxLockConf   string       = "lincxlock"
)
// Config 配置
type Config struct {
	V *viper.Viper
}

// type ConfigFile struct{
// 	 FAType FileAddrType
// 	 FType FileType
// 	 FName string
// 	 FPath string
// }
func init(){
	LockConf=&LockConfig{}
	TypeList = []LockType{RedisLock,EtcdLock,ZKLock}
}

// SetLockConf 设置LockConf的属性值
func SetLockConf(locktype string,timeout int,hosts []string)error{
	if len(hosts)==0{
		return fmt.Errorf("hosts is nil")
	}
	if !checkType(LockType(locktype)){
		return fmt.Errorf("this lock type not supported")
	}
	if timeout<0{
		return fmt.Errorf("timeout must >= 0")
	}
	LockConf.Locktype = LockType(locktype)
	LockConf.Timeout = timeout
	LockConf.Hosts = hosts
	return nil
}


func checkType(locktype LockType)bool{
	n:=len(TypeList)
	for i:=0;i<n;i++{
		if locktype==TypeList[i]{
			return true
		}
	}
	return false
}
// LoadLocalFile 加载文件
func LoadLocalFile(fileAddr FileAddrType, fileType FileType, filePath, fileName string) error {
	c := &Config{}
	c.V = viper.New()
	
	err := c.LoadConfigFile(fileAddr, fileType,  filePath,fileName)
	
	if err != nil {
		return err
	}
	lockMap:=c.V.GetStringMap(LincxLockConf)
	if err := mapstructure.Decode(lockMap, LockConf); err != nil {
		return err
	}
	return nil
}

// LoadConfigFile 加载配置文件
func (c *Config) LoadConfigFile(fileAddr FileAddrType, fileType FileType, filePath, fileName string) error {

	if fileAddr == RemoteFile {
		
		c.V.AddRemoteProvider(string(fileType),
			filePath,
			fileName,
		)
		c.V.SetConfigType(string(fileType))
		err := c.V.ReadRemoteConfig()
		if err != nil {
			return err
		}
		
	} else {
		c.V.AddConfigPath(filePath)
		c.V.SetConfigName(fileName)
		c.V.SetConfigType(string(fileType))
		err := c.V.ReadInConfig()		
		if err != nil {
			return err
		}
		
	}
	return nil

}


// LoadConfigFromYaml 从yaml文件读取配置
func LoadConfigFromYaml(c *Config) error {
	c.V = viper.New()
	// 设置配置文件的名字
	c.V.SetConfigName("config")

	// 添加配置文件所在的路径注意在Linux环境下%GOPATH要替换为$GOPATH
	c.V.AddConfigPath("%GOPATH/src/etcdlincx")
	c.V.AddConfigPath("./")

	// 添加配置文件类型
	c.V.SetConfigType("yaml")
	if err := c.V.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// LoadConfigFromIo 从IO读取配置
func LoadConfigFromIo(c *Config) error {
	c.V = viper.New()
	f, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	confLength, _ := f.Seek(0, 2)
	configData := make([]byte, confLength)

	f.Seek(0, 0)

	f.Read(configData)

	c.V.SetConfigType("yaml")

	err = c.V.ReadConfig(bytes.NewBuffer(configData))

	if err != nil {
		return err
	}

	return nil
}

// EnvConfigPrefix 从本地环境变量读取配置
func EnvConfigPrefix(c *Config) error {
	c.V = viper.New()

	//BindEnv($1,$2)
	// 如果只传入一个参数，则会提取指定的环境变量$1，如果设置了前缀，则会自动补全 前缀_$1
	//如果传入两个参数则不会补全前缀，直接获取第二参数中传入的环境变量$2

	os.Setenv("LOG_LEVEL", "INFO")

	return nil
}
