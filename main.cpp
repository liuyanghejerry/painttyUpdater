#include "updater.h"
#include <QApplication>
#include <QTranslator>
#include <QStringList>
#include <QDebug>

void initTranslation()
{
    QString locale = QLocale(QLocale::system().uiLanguages().at(0)).name();

    QTranslator *qtTranslator = new QTranslator(qApp);
    qtTranslator->load(QString("qt_%1").arg(locale), ":/translation", "_", ".qm");
    QCoreApplication::installTranslator(qtTranslator);

    QTranslator *myappTranslator = new QTranslator(qApp);
    myappTranslator->load(QString("updater_%1").arg(locale), ":/translation", "_", ".qm");
    QCoreApplication::installTranslator(myappTranslator);
}

int main(int argc, char *argv[])
{
    QApplication a(argc, argv);
    initTranslation();
    Updater up;
    return a.exec();
}
