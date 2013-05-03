#-------------------------------------------------
#
# Project created by QtCreator 2013-04-29T16:45:49
#
#-------------------------------------------------

QT       += core gui network

greaterThan(QT_MAJOR_VERSION, 4): QT += widgets

TARGET = updater
TEMPLATE = app

CONFIG += c++11


SOURCES += main.cpp\
    network/socket.cpp \
    updater.cpp

HEADERS  += \
    common.h \
    network/socket.h \
    updater.h

OTHER_FILES +=

RESOURCES += \
    resources.qrc

