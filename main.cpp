#include "updater.h"
#include <QApplication>
#include <QTranslator>
#include <QStringList>
#include <QDebug>

void initTranslation()
{
    QString locale = QLocale(QLocale::system().uiLanguages().at(0)).name();
    qDebug()<<locale;

    QTranslator *qtTranslator = new QTranslator(qApp);
    qDebug()<<qtTranslator->load(QString("qt_%1").arg(locale), ":/translation", "_", ".qm");
    qDebug()<<QCoreApplication::installTranslator(qtTranslator);

    QTranslator *myappTranslator = new QTranslator(qApp);
    qDebug()<<myappTranslator->load(QString("updater_%1").arg(locale), ":/translation", "_", ".qm");
    qDebug()<<QCoreApplication::installTranslator(myappTranslator);
}

int main(int argc, char *argv[])
{
    QApplication a(argc, argv);
    initTranslation();
    Updater up;
    return a.exec();
}
