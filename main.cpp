#include "updater.h"
#include <QApplication>
#include <QTranslator>
#include <QStringList>

void initTranslation()
{
    QTranslator *qtTranslator = new QTranslator(qApp);
//    QTranslator *myappTranslator = new QTranslator(qApp);

    QString locale = QLocale(QLocale::system().uiLanguages().at(0)).name();

    qtTranslator->load(QString("qt_%1").arg(locale), ":/translation", "_", ".qm");
//    myappTranslator->load(QString("paintty_%1").arg(locale), ":/translation", "_", ".qm");
    QCoreApplication::installTranslator(qtTranslator);
//    QCoreApplication::installTranslator(myappTranslator);
}

int main(int argc, char *argv[])
{
    Updater a(argc, argv);
    initTranslation();
    
    return a.exec();
}
