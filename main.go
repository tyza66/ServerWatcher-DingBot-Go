package main

import (
	"fmt"
	"github.com/blinkbean/dingtalk"
	"github.com/jasonlvhit/gocron"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/v3/mem"
	"runtime"
	"time"
)

var bot = dingtalk.InitDingTalkWithSecret("86739bacb778f46fd978a70aa0c98a0de42ead542da9ce2552a322a233dc0c0f", "SEC039a95542d2cae2e174b4aca7654a6d2213fa36b0ca2b7c1fcf68f34817b152d")

func main() {
	start()
	sendMessage()
	//每天早上报告服务器基本信息
	gocron.Every(1).Day().At("09:30").Do(sendMessage)
	//每30秒检查服务器cpu和内存是否飙高
	gocron.Every(1).Second().Do(checkCpuAndMem)
	<-gocron.Start()
}

func start() {
	bot.SendMarkDownMessage("服务器信息", "服务器已启动，开始监控...")
}

type LSysInfo struct {
	MemAll         uint64
	MemFree        uint64
	MemUsed        uint64
	MemUsedPercent float64
	Days           int64
	Hours          int64
	Minutes        int64
	Seconds        int64

	CpuUsedPercent float64
	OS             string
	Arch           string
	CpuCores       int
}

func GetSysInfo() (info LSysInfo) {
	unit := uint64(1024 * 1024) // MB

	v, _ := mem.VirtualMemory()

	info.MemAll = v.Total
	info.MemFree = v.Free
	info.MemUsed = info.MemAll - info.MemFree
	// 注：使用SwapMemory或VirtualMemory，在不同系统中使用率不一样，因此直接计算一次
	info.MemUsedPercent = float64(info.MemUsed) / float64(info.MemAll) * 100.0 // v.UsedPercent
	info.MemAll /= unit
	info.MemUsed /= unit
	info.MemFree /= unit

	info.OS = runtime.GOOS
	info.Arch = runtime.GOARCH
	info.CpuCores = runtime.GOMAXPROCS(0)

	// 获取200ms内的CPU信息，太短不准确，也可以获几秒内的，但这样会有延时，因为要等待
	cc, _ := cpu.Percent(time.Millisecond*200, false)
	info.CpuUsedPercent = cc[0]

	// 获取开机时间
	boottime, _ := host.BootTime()
	ntime := time.Now().Unix()
	btime := time.Unix(int64(boottime), 0).Unix()
	deltatime := ntime - btime

	info.Seconds = int64(deltatime)
	info.Minutes = info.Seconds / 60
	info.Seconds -= info.Minutes * 60
	info.Hours = info.Minutes / 60
	info.Minutes -= info.Hours * 60
	info.Days = info.Hours / 24
	info.Hours -= info.Days * 24
	//fmt.Printf("info: %#v\n", info)
	return
}

func sendMessage() {
	message := "### 服务器基本信息"
	info := GetSysInfo()
	d, _ := disk.Usage("/")
	message += "  \n> 您的服务器已运行-" + fmt.Sprintf("%d", info.Days) + "天" + fmt.Sprintf("%d", info.Hours) + "小时" + fmt.Sprintf("%d", info.Minutes) + "分钟" + fmt.Sprintf("%d", info.Seconds) + "秒"
	message += "  \n- 系统运行内存使用率为：" + fmt.Sprintf("%.2f", info.MemUsedPercent) + "%"
	message += "  \n- 系统运行CPU使用率为：" + fmt.Sprintf("%.2f", info.CpuUsedPercent) + "%"
	message += "  \n- 系统物理磁盘磁盘使用率为：" + fmt.Sprintf("%.2f", d.UsedPercent) + "%"
	message += "  \n- 系统运行内存已用量为：" + fmt.Sprintf("%d", info.MemUsed) + "MB/" + fmt.Sprintf("%d", info.MemAll) + "MB"
	message += "  \n- 系统运行内存空闲量为：" + fmt.Sprintf("%d", info.MemFree) + "MB"
	message += "  \n- 系统运行CPU核心数为：" + fmt.Sprintf("%d", info.CpuCores) + "核"
	message += "  \n- 系统物理磁盘已用大小为：" + fmt.Sprintf("%d", d.Used/1024/1024/1024) + "GB/" + fmt.Sprintf("%d", d.Total/1024/1024/1024) + "GB"
	message += "  \n- 系统物理磁盘空闲大小为：" + fmt.Sprintf("%d", d.Free/1024/1024/1024) + "GB"
	bot.SendMarkDownMessage("服务器基本信息", message)
}

func checkCpuAndMem() {
	fmt.Println("检查服务器cpu和内存是否飙高")
	info := GetSysInfo()
	avg, _ := load.Avg()
	cpuNum := runtime.NumCPU()
	loadavg_max := float64(cpuNum) * 0.7
	if avg.Load1 > loadavg_max || avg.Load5 > loadavg_max || avg.Load15 > loadavg_max {
		bot.SendMarkDownMessage("服务器告警", "⚠️<font color=\"#d30c0c\">【警告】</font>您的云服务器当前负载过高，当前负载为<font color=\"#d30c0c\">"+fmt.Sprintf("%.2f", avg.Load1)+"</font>，请及时检查系统是否存在问题。")
	}
	if info.MemUsedPercent > 90.0 {
		bot.SendMarkDownMessage("服务器告警", "⚠️<font color=\"#d30c0c\">【警告】</font>您的云服务器当前内存使用率为<font color=\"#d30c0c\">"+fmt.Sprintf("%.2f", info.MemUsedPercent)+"%</font>，请及时检查系统是否存在问题。")
	}
	if info.CpuUsedPercent > 70.0 {
		bot.SendMarkDownMessage("服务器告警", "⚠️<font color=\"#d30c0c\">【警告】</font>您的云服务器当前CPU使用率为<font color=\"#d30c0c\">"+fmt.Sprintf("%.2f", info.CpuUsedPercent)+"%</font>，请及时检查系统是否存在问题。")
	}
	return
}
