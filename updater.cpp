#include "updater.h"
#include <QHostAddress>
#include <QUrl>
#include <QStringList>
#include <QJsonDocument>
#include <QJsonObject>
#include <QJsonValue>
#include <QMessageBox>
#include "common.h"
#include "network/socket.h"


Updater::Updater(int & argc, char ** argv) :
    QApplication(argc, argv),
    socket(new Socket(this)),
    state_(State::READY)
{
    checkNewestVersion();
}

Updater::~Updater()
{
    socket->close();
}

void Updater::checkNewestVersion()
{
    connect(socket, &Socket::connected,
            [this](){
        state_ = State::CHK_VERSION;
        QJsonDocument doc;
        QJsonObject obj;
        obj.insert("request", QString("check"));
        // NOTICE: to ensure no ambigous,
        // all JSON should be in lower case
        QString locale = QLocale(QLocale::system().uiLanguages().at(0))
                .name().toLower();
        obj.insert("language", locale);
        obj.insert("platform", QString("windows x86"));
        doc.setObject(obj);
        socket->sendData(doc.toJson());
    });
    connect(socket, &Socket::error,
            [this](){
        switch(state_){
        case State::CHK_VERSION:
            state_ = State::CHK_ERROR;
            break;
        case State::DOWNLOAD_NEW:
            state_ = State::DOWNLOAD_ERROR;
            break;
        case State::OVERLAP:
            state_ = State::OVERLAP_ERROR;
            break;
        default:
            state_ = State::UNKNOWN_ERROR;
        }

        qDebug()<<socket->errorString();
        exit(1);
    });
    connect(socket, &Socket::newData,
            [this](const QByteArray& data){
        QJsonDocument doc = QJsonDocument::fromJson(data);
        QJsonObject obj = doc.object();
        if(obj.isEmpty()){
            state_ = State::CHK_ERROR;
            qDebug()<<"Check version error!";
            exit(1);
        }
        QJsonObject info = obj.value("info").toObject();
        if(info.isEmpty()){
            state_ = State::CHK_ERROR;
            qDebug()<<"Check version error!";
            exit(1);
        }

        // close connection right after we have the result.
        socket->close();

        QString version = info.value("version").toString().trimmed();
        QString changelog = info.value("changelog").toString();
        int level = info.value("level").toDouble();

        // try to use url fetched from remote server
        QUrl url = QUrl::fromUserInput(DOWNLOAD_URL);
        QString fetched_url = info.value("url").toString();

        // if we cannot get a valid url, use predefined url
        if(!fetched_url.isEmpty()){
            url = QUrl::fromUserInput(fetched_url);
        }

        QStringList commandList = qApp->arguments();
        // --version should be considered first
        int index = commandList.lastIndexOf("--version");
        // then we check if there is -v
        index = index>0?index:commandList.lastIndexOf("-v");
        if(index < 0 || index >= commandList.count()){
            qDebug()<<"parsing error!"<<"cannot find --version or -v";
            printUsage();
            exit(1);
        }
        QString old_version = commandList[index+1].trimmed();
        if(old_version.isEmpty()){
            qDebug()<<"parsing error!"<<"version number is empty";
            printUsage();
            exit(1);
        }
        if(version != old_version){
            QMessageBox msgBox;
            msgBox.setWindowTitle(tr("New version!"));
            if(level < 3) {
                msgBox.setIcon(QMessageBox::Information);
                msgBox.setText(tr("There's a new version of Mr.Paint.\n"
                                  "We suggest you download it here: %1")
                               .arg(url.toDisplayString()));
            }else{
                msgBox.setIcon(QMessageBox::Warning);
                msgBox.setText(tr("There's a critical update of Mr.Paint.\n"
                                  "You can connect to server ONLY if you've updated: %1")
                               .arg(url.toDisplayString()));
            }
            if(!changelog.isEmpty()){
                msgBox.setDetailedText(changelog);
            }

            msgBox.exec();
            exit(0);
        }

    });
    socket->connectToHost(QHostAddress(SERVER_ADDRESS), SERVER_PORT);

    return;
}

bool Updater::download()
{
    // TODO: auto download the new version
    return false;
}

bool Updater::overlap()
{
    // TODO: overlap the old version
    return false;
}

void Updater::printUsage()
{
    qDebug()<<"painttyUpdater "<<VERSION<<endl
           <<"Usage: "<<"painttyUpdater OPTIONS"<<endl<<endl
          <<"-v, --version VERSION: tell updater the current "
            "version of main program.";
}
