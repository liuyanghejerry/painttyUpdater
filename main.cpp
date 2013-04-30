#include "updater.h"
#include <QApplication>

int main(int argc, char *argv[])
{
    Updater a(argc, argv);
    
    return a.exec();
}
