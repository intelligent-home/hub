package hub

import (
	"encoding/json"
	"strconv"
	"time"
)

// Command mqtt command
type Command interface {
	// Name command name
	Name() string
	// Msg command mqtt message
	Msg() ([]byte, error)
}

// SetPanelsCommand 设置面板指令
type SetPanelsCommand struct {
	// Command 指令名称, 如 set_panels
	Command string `json:"command" jsonschema:"description=指令名称"`
	// Serial 序列号
	Serial string `json:"serial" jsonschema:"description=序列号"`
	// DeviceNo 设备编号, 如 c49878f1e235
	DeviceNo string `json:"device_no" jsonschema:"description=设备编号"`
	// Panels 控制面板列表
	Panels []Panel `json:"panels" jsonschema:"description=控制面板列表"`
	// TimeStamp 时间戳
	Timestamp string `json:"time_stamp" jsonschema:"description=时间戳"`
}

// NewSetPanelsCommand 创建设置面板指令
func NewSetPanelsCommand() *SetPanelsCommand {
	serial := strconv.Itoa(int(time.Now().UnixMilli()))
	return &SetPanelsCommand{
		Command:   "set_panels",
		Serial:    serial,
		DeviceNo:  DeviceNo,
		Timestamp: serial,
	}
}

// NewSetPanelsCommandWithDevice create a set panels command with device
func NewSetPanelsCommandWithDevice(device Device) *SetPanelsCommand {
	command := NewSetPanelsCommand()
	pannel := Panel{
		Address: DevicePannelAddress,
		ChildDev: []Device{
			device,
		},
	}
	command.Panels = []Panel{pannel}
	return command
}

// Name the command name
func (c *SetPanelsCommand) Name() string {
	return "set_panels"
}

// Msg 设置面板 mqtt 消息
func (c *SetPanelsCommand) Msg() ([]byte, error) {
	// 构建 mqtt 消息
	mqttMsg, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	if MQTTSecretEnable {
		return Encrypt(mqttMsg)
	}
	return mqttMsg, nil
}

type Panel struct {
	// Address 控制面板地址
	Address string `json:"address" jsonschema:"description=控制面板地址"`
	// ChildDev 子设备列表
	ChildDev []Device `json:"child_dev" josnschema:"description=子设备列表"`
}

// Device 设备
type Device map[string]string

// CommandResult 指令执行结果
type CommandResult struct {
	// Command 指令名称
	Command string `json:"command"`
	// Serial 序列号, 用于反馈于指定请求判断
	Serial string `json:"serial"`
	// DeviceNo 设备编号
	DeviceNo string `json:"device_no"`
	// Result 执行结果, 1:成功, 0:失败
	Result string `json:"result"`
}

// ReadStatusResult 读取网关设备状态结果
type ReadStatusResult struct {
	Timestamp string `json:"time_stamp"`
	// Command 指令名称
	Command string `json:"command"`
	// Serial 序列号, 用于反馈于指定请求判断
	Serial string `json:"serial"`
	// DeviceNo 设备编号
	DeviceNo string `json:"device_no"`
	// Panels 控制面板列表
	Panels []Panel `json:"panels"`
}

// NewReadStatusResult  创建设备状态结果
func NewReadStatusResult(deviceNo string) *ReadStatusResult {
	return &ReadStatusResult{
		Command:  "read_status",
		DeviceNo: deviceNo,
	}
}

// Device 查找指定设备, addr 为 panel 地址, channel 为管道号
func (r *ReadStatusResult) Device(addr string, channel string) (Device, bool) {
	for _, p := range r.Panels {
		if p.Address == addr {
			for _, dev := range p.ChildDev {
				if dev["channel"] == channel {
					return dev, true
				}
			}
		}
	}
	return nil, false
}

// ReadStatusCommand 读取设备状态指令
type ReadStatusCommand struct {
	Timestamp string `json:"time_stamp"`
	Command   string `json:"command"`
	Serial    string `json:"serial"`
}

// NewReadStatusCommand 创建读取设备指令
func NewReadStatusCommand() *ReadStatusCommand {
	return &ReadStatusCommand{Command: "read_status"}
}

// Name command name
func (c *ReadStatusCommand) Name() string {
	return "read_status"
}

// Msg 指令 mqtt 消息
func (c *ReadStatusCommand) Msg() ([]byte, error) {
	serial := strconv.Itoa(int(time.Now().UnixMilli()))
	c.Timestamp = serial
	c.Serial = serial
	// 构建 mqtt 消息
	mqttMsg, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	if MQTTSecretEnable {
		return Encrypt(mqttMsg)
	}
	return mqttMsg, nil
}
