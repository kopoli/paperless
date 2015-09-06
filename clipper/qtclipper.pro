#-------------------------------------------------
#
# Project created by QtCreator 2012-10-07T16:31:27
#
#-------------------------------------------------

QT       += core gui

greaterThan(QT_MAJOR_VERSION, 4): QT += widgets

CONFIG += c++11 warn_on release

release {
  DEFINES += QT_NO_DEBUG_OUTPUT
}
debug {
  QMAKE_CFLAGS_WARN_ON    = -Wall -Werror -Wundef -Wextra
  QMAKE_CXXFLAGS_WARN_ON  = $$QMAKE_CFLAGS_WARN_ON
}

TARGET = qtclipper
TEMPLATE = app

SOURCES += qtclipper.cpp
