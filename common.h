#ifndef COMMON_H
#define COMMON_H
#include <QtGlobal>

//static const char* SERVER_ADDRESS = "106.187.92.58";
static const char* SERVER_ADDRESS = "127.0.0.1";
static const int SERVER_PORT = 7071;
static const char* VERSION = "0.1";

#ifdef Q_OS_WIN32
static const char* DOWNLOAD_URL = "http://mrspaint.oss.aliyuncs.com/%E8%8C%B6%E7%BB%98%E5%90%9B_Alpha_x86.zip";
#elif Q_OS_UNIX
static const char* DOWNLOAD_URL = "";
#elif Q_OS_MAC
static const char* DOWNLOAD_URL = "http://mrspaint.oss.aliyuncs.com/%E8%8C%B6%E7%BB%98%E5%90%9B.app.zip";
#endif

#endif // COMMON_H
