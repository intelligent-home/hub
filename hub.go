package hub

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	// ReceiveTopic 设置设备主题, 需要提供设备编号
	ReceiveTopicFormat = "/receive/%s"
	// SendDevices 上报主题
	SendDevicesTopic = "/send/devices"
)

var (
	// MQTTSecretEnable 是否启用 mqtt 消息加密
	MQTTSecretEnable = false
	// MQTTMsgSecretKey  mqtt 消息加密 key, 如果使用3DES 必须保证 24 字节
	MQTTSecretKey = ""
	// MQTTServer mqtt 服务地址
	MQTTServer = "localhost:1883"
	// MQTTUser mqtt 服务连接用户名
	MQTTUser = ""
	// MQTTPassword mqtt 服务连接密码
	MQTTPassword = ""
	// DeviceNo 网关设备编号
	DeviceNo = ""
	// Pannel Address 控制面板地址
	DevicePannelAddress = ""
	// ReceiveTopic 设置设备主题
	ReceiveTopic = ""
	// store 存储网关设备状态信息
	store sync.Map
	// MQTTClient mqtt client
	MQTTClient   mqtt.Client
	MQTTClientID = fmt.Sprintf("controller-%d", time.Now().UnixMilli())
)

// Init will initialize device
func Init() error {
	// mqtt 消息加密
	if v := os.Getenv("MQTT_SECRET_ENABLE"); v != "" {
		secretEnable, err := strconv.ParseBool(v)
		if err == nil {
			MQTTSecretEnable = secretEnable
		}
	}
	// mqtt 消息加密 key
	if v := os.Getenv("MQTT_KEY"); v != "" {
		MQTTSecretKey = v
	}
	// mqtt server
	if v := os.Getenv("MQTT_SERVER"); v != "" {
		MQTTServer = v
	}
	// mqtt user
	if v := os.Getenv("MQTT_USER"); v != "" {
		MQTTUser = v
	}
	// mqtt password
	if v := os.Getenv("MQTT_PASSWORD"); v != "" {
		MQTTPassword = v
	}
	// mqtt client id
	if v := os.Getenv("MQTT_CLIENT_ID"); v != "" {
		MQTTClientID = v
	}
	// gateway device no.
	if v := os.Getenv("DEVICE_NO"); v != "" {
		DeviceNo = v
	}
	// receive topic
	ReceiveTopic = fmt.Sprintf(ReceiveTopicFormat, DeviceNo)
	// device control panel address
	if v := os.Getenv("DEVICE_PANEL_ADDRESS"); v != "" {
		DevicePannelAddress = v
	}
	slog.Info("[hub] bind gateway", "device_no", DeviceNo, "pannel_address", DevicePannelAddress)
	// mqtt client
	options := mqtt.NewClientOptions().
		AddBroker(MQTTServer).
		SetClientID(MQTTClientID). // 需要确保唯一性
		SetCleanSession(false).    // 保持会话, 与ClientID 相关,在ID不变情况下可以会保持会话及消息, 不保存 QoS 0 消息
		SetProtocolVersion(3)

	var kv []any
	kv = append(kv, "addr", MQTTServer, "secret", MQTTSecretEnable)
	if len(MQTTUser) > 0 {
		kv = append(kv, "user", MQTTUser)
		options.SetUsername(MQTTUser)
	}
	if len(MQTTPassword) > 0 {
		kv = append(kv, "password", "********")
		options.SetPassword(MQTTPassword)
	}
	slog.Info("[hub] MQTT server", kv...)

	MQTTClient = mqtt.NewClient(options)
	if token := MQTTClient.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("MQTT client connect", "error", token.Error())
		return token.Error()
	}
	slog.Info("[hub] MQTT client connected", "client_id", MQTTClientID)

	// 订阅指定网关设备消息
	MQTTClient.Subscribe(SendDevicesTopic, 1, func(client mqtt.Client, msg mqtt.Message) {
		var err error
		payload := msg.Payload()
		if MQTTSecretEnable {
			payload, err = Decrypt(payload)
			if err != nil {
				slog.Error("[hub] subscribe failed", "topic", msg.Topic(), "error", err)
				return
			}
		}
		// slog.Info("[ac-controller] receive subscribe msg", "payload", payload)
		var readStatusResult ReadStatusResult
		err = json.Unmarshal(payload, &readStatusResult)
		if err != nil {
			slog.Error("[hub] unmarshal failed", "error", err)
			return
		}
		// 存储网关设备信息
		store.Store(readStatusResult.DeviceNo, &readStatusResult)
		// msg.Ack()
		slog.Info("[hub] store device status", "device_no", readStatusResult.DeviceNo, "status", readStatusResult)
	})
	return nil
}

// Publish send mqtt command msg
func Publish(cmd Command) error {
	msg, err := cmd.Msg()
	if err != nil {
		slog.Error("[hub] create mqtt message failed", "command", cmd.Name(), "error", err)
		return err
	}
	if token := MQTTClient.Publish(ReceiveTopic, 1, false, msg); token.Wait() && token.Error() != nil {
		slog.Error("[hub] mqtt publish message failed", "command", cmd.Name(), "error", err)
		return err
	}
	slog.Info("[hub] mqtt publish message successful", "command", cmd.Name(), "msg", string(msg))
	return nil
}

// DeviceStatus get device previous status
func DeviceStatus(channel string) (Device, bool) {
	// 获取先前状态
	status, ok := store.Load(DeviceNo)
	if ok {
		prev, ok := status.(*ReadStatusResult)
		if ok {
			dev, ok := prev.Device(DevicePannelAddress, channel)
			if ok {
				return dev, true
			}
		}
	}
	return nil, false
}
