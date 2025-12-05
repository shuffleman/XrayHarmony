#ifndef XRAY_BRIDGE_H
#define XRAY_BRIDGE_H

#include <string>
#include <vector>

// C function declarations exported from Go
extern "C" {
    // Xray 核心函数
    char* XrayGetLastError();
    void XrayFreeString(char* str);
    long long XrayNewInstance();
    int XrayDeleteInstance(long long id);
    int XrayLoadConfig(long long id, const char* configJSON);
    int XrayLoadConfigFromFile(long long id, const char* filePath);
    int XrayStart(long long id);
    int XrayStop(long long id);
    int XrayIsRunning(long long id);
    char* XrayGetStats(long long id);
    int XrayTestConfig(long long id, const char* configJSON);
    char* XrayGetVersion();

    // VPN 相关函数
    long long VPNNewManager(long long xrayInstanceID);
    int VPNDeleteManager(long long id);
    int VPNStart(long long id, const char* configJSON);
    int VPNStop(long long id);
    int VPNIsRunning(long long id);
    char* VPNGetStats(long long id);
}

namespace xray {

class XrayBridge {
public:
    XrayBridge();
    ~XrayBridge();

    // 配置管理
    bool loadConfig(const std::string& configJSON);
    bool loadConfigFromFile(const std::string& filePath);
    bool testConfig(const std::string& configJSON);

    // 实例控制
    bool start();
    bool stop();
    bool isRunning();

    // 统计信息
    std::string getStats();
    std::string getVersion();

    // 错误处理
    std::string getLastError() const;

    // 获取实例 ID（用于创建 VPN 管理器）
    long long getInstanceId() const;

private:
    void updateLastError();

    long long instanceId_;
    std::string lastError_;
};

// VPN 配置结构
struct VPNConfig {
    int tunFd;
    int tunMTU;
    std::string socksAddr;
    std::vector<std::string> dnsServers;
    bool fakeDNS;
    bool udp;
    bool tcpConcurrent;
};

// VPN 统计信息结构
struct VPNStats {
    bool running;
    std::string socksAddr;
    int mtu;
};

class VPNBridge {
public:
    VPNBridge(XrayBridge* xrayBridge);
    ~VPNBridge();

    // VPN 控制
    bool start(const VPNConfig& config);
    bool start(const std::string& configJSON);
    bool stop();
    bool isRunning();

    // 统计信息
    VPNStats getStats();

    // 错误处理
    std::string getLastError() const;

private:
    void updateLastError();
    std::string configToJSON(const VPNConfig& config);

    long long managerId_;
    XrayBridge* xrayBridge_;
    std::string lastError_;
};

} // namespace xray

#endif // XRAY_BRIDGE_H
