#ifndef COMMON_H
#define COMMON_H
#include <QtGlobal>
#include <QString>

namespace GlobalDef {
static const char* VERSION = "0.1";
const static char SETTINGS_NAME[] = "mrpaint.ini";

const static QString HOST_ADDR_IPV4("106.187.92.58");
const static QString HOST_ADDR_IPV6("2400:8900::f03c:91ff:fe70:bc64");
const static int HOST_UPDATER_PORT = 7071;

const static QString HOST_ADDR[] = {HOST_ADDR_IPV4,
                                    HOST_ADDR_IPV6};

#ifdef Q_OS_WIN32
static const char* DOWNLOAD_URL = "http://mrspaint.oss.aliyuncs.com/%E8%8C%B6%E7%BB%98%E5%90%9B_Alpha_x86.zip";
#elif Q_OS_UNIX
static const char* DOWNLOAD_URL = "";
#elif Q_OS_MAC
static const char* DOWNLOAD_URL = "http://mrspaint.oss.aliyuncs.com/%E8%8C%B6%E7%BB%98%E5%90%9B.app.zip";
#endif

}
#endif // COMMON_H
