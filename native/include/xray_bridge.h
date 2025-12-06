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

    // 获取实例 ID
    long long getInstanceId() const;

private:
    void updateLastError();

    long long instanceId_;
    std::string lastError_;
};

} // namespace xray

#endif // XRAY_BRIDGE_H
