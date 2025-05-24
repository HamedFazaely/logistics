package conf

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
)

type Configuration struct {
	DBName               string
	DBUser               string
	DBPass               string
	DBHost               string
	DBPort               string
	MQHost               string
	MQPort               string
	MQUser               string
	MQPassword           string
	SMSQueueName         string
	SMSReplyQueueName    string
	StatusQueueName      string
	StatusReplyQueueName string
	OrderExpirationHours int
	ServerListenPort     string
	Development          bool
}

func (c *Configuration) GetDBDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}

func (c *Configuration) GetMQConnString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%s/", c.MQUser, c.MQPassword, c.MQHost, c.MQPort)
}

var conf *Configuration

var once sync.Once

func GetConffiguration() *Configuration {
	once.Do(func() {
		conf = &Configuration{}
		conf.DBName = getStringEnv("DB_NAME", "logistics")
		conf.DBUser = getStringEnv("DB_USER", "hamed")
		conf.DBPass = getStringEnv("DB_PASSWORD", "")
		conf.DBHost = getStringEnv("DB_HOST", "localhost")
		conf.DBPort = getStringEnv("DB_PORT", "3306")
		conf.MQHost = getStringEnv("MQ_HOST", "localhost")
		conf.MQPort = getStringEnv("MQ_PORT", "5672")
		conf.MQUser = getStringEnv("MQ_USER", "hamed")
		conf.MQPassword = getStringEnv("MQ_PASSWORD", "")
		conf.SMSQueueName = getStringEnv("SMS_Q_NAME", "send_sms")
		conf.SMSReplyQueueName = getStringEnv("SMS_REPLY_Q_NAME", "sms_reply")
		conf.StatusQueueName = getStringEnv("STATUS_Q_NAME", "stats_q")
		conf.StatusReplyQueueName = getStringEnv("STATUS_REPLY_Q_NAME", "status_reply")
		conf.OrderExpirationHours = getIntEnv("ORDER_EXPIRATION_HOURS", 10)
		conf.ServerListenPort = getStringEnv("SERVER_LISTEN_PORT", "8080")
		conf.Development = getBoolEnv("DEVELOPMENT", true)

	})
	if conf.DBPass == "" {
		panic(errors.New("database password must be set"))
	}
	if conf.MQPassword == "" {
		panic(errors.New("mq password must be set"))
	}
	return conf
}

func getStringEnv(envName string, defVal string) string {
	val := os.Getenv(envName)
	if val == "" {
		return defVal
	}
	return val
}

func getIntEnv(envName string, defval int) int {
	val := os.Getenv(envName)
	if val == "" {
		return defval
	}
	ret, err := strconv.Atoi(val)
	if err != nil {
		panic(err)
	}
	return ret
}

func getBoolEnv(envName string, defVal bool) bool {
	v := os.Getenv(envName)
	if v == "" {
		return defVal
	}
	return v == "true"
}
