#include "xray_bridge.h"
#include <cstring>

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

} // namespace xray
