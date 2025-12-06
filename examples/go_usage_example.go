package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// 示例: 如何在 Go 中使用 XrayHarmony 的功能

func Example_ParseShareURL() {
	// 解析 VMess 分享链接
	vmessURL := "vmess://eyJ2IjoiMiIsInBzIjoiVGVzdCIsImFkZCI6ImV4YW1wbGUuY29tIiwicG9ydCI6IjQ0MyIsImlkIjoiYWJjZDEyMzQtNTY3OC05YWJjLWRlZi0xMjM0NTY3ODkwYWIiLCJhaWQiOiIwIiwic2N5IjoiYXV0byIsIm5ldCI6InRjcCIsInRscyI6InRscyJ9"

	config, err := ParseVMessURL(vmessURL)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("协议: %s\n", config.Protocol)
	fmt.Printf("地址: %s:%d\n", config.Address, config.Port)
	fmt.Printf("UUID: %s\n", config.ID)
	fmt.Printf("备注: %s\n", config.Remark)
}

func Example_ConfigBuilder() {
	// 使用配置构建器创建完整配置
	builder := NewConfigBuilder()

	// 设置日志
	builder.SetLogLevel("warning")

	// 添加 SOCKS5 入站
	builder.AddSocksInbound(10808, "127.0.0.1", false, true)

	// 添加 VMess 出站
	builder.AddVMessOutbound(
		"example.com",
		443,
		"uuid-here",
		0,
		"auto",
	)

	// 添加直连出站
	builder.AddFreedomOutbound("direct")

	// 添加路由规则 - 中国 IP 直连
	builder.AddRoutingRule("field", "direct",
		[]string{"geosite:cn"},
		[]string{"geoip:cn", "geoip:private"},
	)

	// 设置 DNS
	builder.SetDNS(
		[]string{"8.8.8.8", "1.1.1.1"},
		map[string]string{},
	)

	// 启用统计
	builder.EnableStats()

	// 构建配置
	configJSON, err := builder.BuildJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("生成的配置:")
	fmt.Println(configJSON)
}

func Example_XrayInstance() {
	// 创建 Xray 实例
	instance := NewXrayInstance()

	// 使用配置构建器
	builder := NewConfigBuilder()
	builder.SetLogLevel("warning")
	builder.AddSocksInbound(10808, "127.0.0.1", false, true)
	builder.AddVMessOutbound("example.com", 443, "uuid", 0, "auto")

	configJSON, _ := builder.BuildJSON()

	// 加载配置
	err := instance.LoadConfig(configJSON)
	if err != nil {
		log.Fatal(err)
	}

	// 启动
	err = instance.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Xray 已启动:", instance.IsRunning())

	// 获取统计
	stats, _ := instance.GetStats()
	fmt.Println("统计信息:", stats)

	// 停止
	instance.Stop()
	fmt.Println("Xray 已停止")
}

func Example_AssetManager() {
	// 创建资产管理器
	assetMgr := NewAssetManager("/tmp/xray_assets")

	// 获取 geoip 信息
	info, err := assetMgr.GetAssetInfo(AssetTypeGeoIP)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("GeoIP 存在: %v\n", info.Exists)

	if !info.Exists {
		// 下载 geoip
		fmt.Println("开始下载 geoip...")
		err = assetMgr.DownloadAsset(AssetTypeGeoIP, "", func(progress *DownloadProgress) {
			fmt.Printf("\r下载进度: %.2f%%", progress.Percentage)
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("\n下载完成!")
	}

	// 验证资产
	valid, err := assetMgr.VerifyAsset(AssetTypeGeoIP)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("GeoIP 有效: %v\n", valid)
}

func Example_Tun2Socks() {
	// 注意: 这个示例需要有效的 TUN 文件描述符
	// 在 HarmonyOS 中,TUN FD 由 VPN API 提供

	config := &Tun2SocksConfig{
		TunFd:     3, // 示例 FD
		SocksAddr: "127.0.0.1:10808",
		MTU:       1500,
		DNSAddr:   "8.8.8.8:53",
		FakeDNS:   false,
	}

	instance := NewTun2SocksInstance(config)

	// 启动 (需要有效的 TUN FD 和 SOCKS5 代理)
	// err := instance.Start()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println("Tun2Socks 运行中:", instance.IsRunning())

	// // 获取统计
	// bytesUp, bytesDown := instance.GetStats()
	// fmt.Printf("上行: %d, 下行: %d\n", bytesUp, bytesDown)

	fmt.Println("Tun2Socks 示例 (需要有效的 TUN FD)")
}

func Example_ProtocolConversion() {
	// 示例: 协议之间的转换

	// 1. 从分享链接创建配置
	vmessURL := "vmess://..."
	config, _ := ParseVMessURL(vmessURL)

	// 2. 转换为 JSON
	configJSON, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println("服务器配置 JSON:")
	fmt.Println(string(configJSON))

	// 3. 从配置生成分享链接
	newURL, _ := GenerateVMessURL(config)
	fmt.Println("\n生成的 VMess URL:")
	fmt.Println(newURL)

	// 4. 解析 VLESS 链接
	vlessURL := "vless://uuid@example.com:443?encryption=none&security=tls&type=tcp#MyServer"
	vlessConfig, _ := ParseVLESSURL(vlessURL)
	fmt.Printf("\nVLESS 服务器: %s:%d\n", vlessConfig.Address, vlessConfig.Port)
}

func Example_CompleteVPNSetup() {
	fmt.Println("=== 完整的 VPN 设置示例 ===\n")

	// 1. 下载必要的资产文件
	fmt.Println("步骤 1: 下载 geoip 和 geosite...")
	assetMgr := NewAssetManager("/data/xray_assets")

	// 下载 geoip (如果需要)
	if info, _ := assetMgr.GetAssetInfo(AssetTypeGeoIP); !info.Exists {
		assetMgr.DownloadAsset(AssetTypeGeoIP, "", nil)
	}

	// 下载 geosite (如果需要)
	if info, _ := assetMgr.GetAssetInfo(AssetTypeGeoSite); !info.Exists {
		assetMgr.DownloadAsset(AssetTypeGeoSite, "", nil)
	}

	// 2. 解析服务器配置
	fmt.Println("\n步骤 2: 解析服务器配置...")
	shareURL := "vmess://..."  // 你的分享链接
	serverConfig, _ := ParseShareURL(shareURL)
	fmt.Printf("服务器: %s:%d (%s)\n", serverConfig.Address, serverConfig.Port, serverConfig.Protocol)

	// 3. 创建 Xray 配置
	fmt.Println("\n步骤 3: 创建 Xray 配置...")
	builder := NewConfigBuilder()
	builder.SetLogLevel("warning")

	// SOCKS5 入站
	builder.AddSocksInbound(10808, "127.0.0.1", false, true)

	// 根据协议类型添加出站
	switch serverConfig.Protocol {
	case "vmess":
		builder.AddVMessOutbound(
			serverConfig.Address,
			serverConfig.Port,
			serverConfig.ID,
			serverConfig.AlterID,
			serverConfig.Security,
		)
	case "vless":
		builder.AddVLESSOutbound(
			serverConfig.Address,
			serverConfig.Port,
			serverConfig.ID,
			serverConfig.Flow,
			serverConfig.Encryption,
		)
	}

	// 添加直连和路由规则
	builder.AddFreedomOutbound("direct")
	builder.AddRoutingRule("field", "direct", []string{"geosite:cn"}, []string{"geoip:cn", "geoip:private"})
	builder.EnableStats()

	// 4. 启动 Xray
	fmt.Println("\n步骤 4: 启动 Xray...")
	xray := NewXrayInstanceWithAssets("/data/xray_assets")
	configJSON, _ := builder.BuildJSON()
	xray.LoadConfig(configJSON)
	xray.Start()
	fmt.Println("Xray 运行中:", xray.IsRunning())

	// 5. 启动 Tun2Socks (需要 TUN FD)
	fmt.Println("\n步骤 5: 启动 Tun2Socks...")
	// tunFd := getTunFdFromHarmonyOSVPNAPI()
	// tun2socks := NewTun2SocksInstance(&Tun2SocksConfig{
	// 	TunFd:     tunFd,
	// 	SocksAddr: "127.0.0.1:10808",
	// 	MTU:       1500,
	// 	DNSAddr:   "8.8.8.8:53",
	// })
	// tun2socks.Start()

	fmt.Println("\n=== VPN 已准备就绪 ===")
	fmt.Println("流量路径: 应用 → TUN → Tun2Socks → Xray SOCKS5 → 远程服务器")
}

func main() {
	fmt.Println("XrayHarmony Go 使用示例\n")
	fmt.Println("运行各个示例函数来查看用法:\n")

	// 取消注释以运行特定示例

	// Example_ParseShareURL()
	// Example_ConfigBuilder()
	// Example_XrayInstance()
	// Example_AssetManager()
	// Example_Tun2Socks()
	// Example_ProtocolConversion()
	// Example_CompleteVPNSetup()

	fmt.Println("所有示例已准备就绪!")
}
