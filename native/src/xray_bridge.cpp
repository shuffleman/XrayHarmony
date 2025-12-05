#include "xray_bridge.h"
#include <cstring>
#include <sstream>

namespace xray {

XrayBridge::XrayBridge() : instanceId_(-1) {
    instanceId_ = XrayNewInstance();
    if (instanceId_ < 0) {
        updateLastError();
    }
}

XrayBridge::~XrayBridge() {
    if (instanceId_ >= 0) {
        if (isRunning()) {
            stop();
        }
        XrayDeleteInstance(instanceId_);
    }
}

void XrayBridge::updateLastError() {
    char* error = XrayGetLastError();
    if (error != nullptr) {
        lastError_ = std::string(error);
        XrayFreeString(error);
    } else {
        lastError_.clear();
    }
}

bool XrayBridge::loadConfig(const std::string& configJSON) {
    if (instanceId_ < 0) {
        lastError_ = "Invalid instance";
        return false;
    }

    int result = XrayLoadConfig(instanceId_, configJSON.c_str());
    if (result != 0) {
        updateLastError();
        return false;
    }

    lastError_.clear();
    return true;
}

bool XrayBridge::loadConfigFromFile(const std::string& filePath) {
    if (instanceId_ < 0) {
        lastError_ = "Invalid instance";
        return false;
    }

    int result = XrayLoadConfigFromFile(instanceId_, filePath.c_str());
    if (result != 0) {
        updateLastError();
        return false;
    }

    lastError_.clear();
    return true;
}

bool XrayBridge::testConfig(const std::string& configJSON) {
    if (instanceId_ < 0) {
        lastError_ = "Invalid instance";
        return false;
    }

    int result = XrayTestConfig(instanceId_, configJSON.c_str());
    if (result != 0) {
        updateLastError();
        return false;
    }

    lastError_.clear();
    return true;
}

bool XrayBridge::start() {
    if (instanceId_ < 0) {
        lastError_ = "Invalid instance";
        return false;
    }

    int result = XrayStart(instanceId_);
    if (result != 0) {
        updateLastError();
        return false;
    }

    lastError_.clear();
    return true;
}

bool XrayBridge::stop() {
    if (instanceId_ < 0) {
        lastError_ = "Invalid instance";
        return false;
    }

    int result = XrayStop(instanceId_);
    if (result != 0) {
        updateLastError();
        return false;
    }

    lastError_.clear();
    return true;
}

bool XrayBridge::isRunning() {
    if (instanceId_ < 0) {
        return false;
    }

    int result = XrayIsRunning(instanceId_);
    return result == 1;
}

std::string XrayBridge::getStats() {
    if (instanceId_ < 0) {
        lastError_ = "Invalid instance";
        return "";
    }

    char* stats = XrayGetStats(instanceId_);
    if (stats == nullptr) {
        updateLastError();
        return "";
    }

    std::string result(stats);
    XrayFreeString(stats);
    lastError_.clear();
    return result;
}

std::string XrayBridge::getVersion() {
    char* version = XrayGetVersion();
    if (version == nullptr) {
        return "Unknown";
    }

    std::string result(version);
    XrayFreeString(version);
    return result;
}

std::string XrayBridge::getLastError() const {
    return lastError_;
}

long long XrayBridge::getInstanceId() const {
    return instanceId_;
}

// ================ VPNBridge 实现 ================

VPNBridge::VPNBridge(XrayBridge* xrayBridge)
    : managerId_(-1), xrayBridge_(xrayBridge) {
    if (xrayBridge_ == nullptr) {
        lastError_ = "XrayBridge is null";
        return;
    }

    long long xrayInstanceId = xrayBridge_->getInstanceId();
    if (xrayInstanceId < 0) {
        lastError_ = "Invalid Xray instance";
        return;
    }

    managerId_ = VPNNewManager(xrayInstanceId);
    if (managerId_ < 0) {
        updateLastError();
    }
}

VPNBridge::~VPNBridge() {
    if (managerId_ >= 0) {
        if (isRunning()) {
            stop();
        }
        VPNDeleteManager(managerId_);
    }
}

void VPNBridge::updateLastError() {
    char* error = XrayGetLastError();
    if (error != nullptr) {
        lastError_ = std::string(error);
        XrayFreeString(error);
    } else {
        lastError_.clear();
    }
}

bool VPNBridge::start(const VPNConfig& config) {
    std::string configJSON = configToJSON(config);
    return start(configJSON);
}

bool VPNBridge::start(const std::string& configJSON) {
    if (managerId_ < 0) {
        lastError_ = "Invalid VPN manager";
        return false;
    }

    int result = VPNStart(managerId_, configJSON.c_str());
    if (result != 0) {
        updateLastError();
        return false;
    }

    lastError_.clear();
    return true;
}

bool VPNBridge::stop() {
    if (managerId_ < 0) {
        lastError_ = "Invalid VPN manager";
        return false;
    }

    int result = VPNStop(managerId_);
    if (result != 0) {
        updateLastError();
        return false;
    }

    lastError_.clear();
    return true;
}

bool VPNBridge::isRunning() {
    if (managerId_ < 0) {
        return false;
    }

    int result = VPNIsRunning(managerId_);
    return result == 1;
}

VPNStats VPNBridge::getStats() {
    VPNStats stats = {false, "", 0};

    if (managerId_ < 0) {
        lastError_ = "Invalid VPN manager";
        return stats;
    }

    char* statsJSON = VPNGetStats(managerId_);
    if (statsJSON == nullptr) {
        updateLastError();
        return stats;
    }

    // 简单的 JSON 解析（实际应该使用 JSON 库）
    std::string jsonStr(statsJSON);
    XrayFreeString(statsJSON);

    // 解析 running
    size_t runningPos = jsonStr.find("\"running\":");
    if (runningPos != std::string::npos) {
        size_t truePos = jsonStr.find("true", runningPos);
        stats.running = (truePos != std::string::npos &&
                        truePos < runningPos + 20);
    }

    // 解析 socksAddr
    size_t socksPos = jsonStr.find("\"socksAddr\":\"");
    if (socksPos != std::string::npos) {
        socksPos += 13; // 跳过 "socksAddr":"
        size_t endPos = jsonStr.find("\"", socksPos);
        if (endPos != std::string::npos) {
            stats.socksAddr = jsonStr.substr(socksPos, endPos - socksPos);
        }
    }

    // 解析 mtu
    size_t mtuPos = jsonStr.find("\"mtu\":");
    if (mtuPos != std::string::npos) {
        mtuPos += 6; // 跳过 "mtu":
        size_t endPos = jsonStr.find_first_of(",}", mtuPos);
        if (endPos != std::string::npos) {
            std::string mtuStr = jsonStr.substr(mtuPos, endPos - mtuPos);
            stats.mtu = std::stoi(mtuStr);
        }
    }

    lastError_.clear();
    return stats;
}

std::string VPNBridge::getLastError() const {
    return lastError_;
}

std::string VPNBridge::configToJSON(const VPNConfig& config) {
    std::ostringstream oss;
    oss << "{";
    oss << "\"tunFd\":" << config.tunFd << ",";
    oss << "\"tunMTU\":" << config.tunMTU << ",";
    oss << "\"socksAddr\":\"" << config.socksAddr << "\",";

    // DNS 服务器数组
    oss << "\"dnsServers\":[";
    for (size_t i = 0; i < config.dnsServers.size(); ++i) {
        oss << "\"" << config.dnsServers[i] << "\"";
        if (i < config.dnsServers.size() - 1) {
            oss << ",";
        }
    }
    oss << "],";

    oss << "\"fakeDNS\":" << (config.fakeDNS ? "true" : "false") << ",";
    oss << "\"udp\":" << (config.udp ? "true" : "false") << ",";
    oss << "\"tcpConcurrent\":" << (config.tcpConcurrent ? "true" : "false");
    oss << "}";

    return oss.str();
}

} // namespace xray
