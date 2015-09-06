/*
 *  Copyright (c) 2012 Kalle Kankare <kalle.kankare@iki.fi>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

#include <functional>

#include <QtDebug>
#include <QApplication>
#include <QCommandLineParser>
#include <QGraphicsPixmapItem>
#include <QGraphicsScene>
#include <QGraphicsSceneMouseEvent>
#include <QGraphicsSceneWheelEvent>
#include <QGraphicsView>
#include <QKeyEvent>
#include <QStack>
#include <QUndoCommand>
#include <QFileInfo>
#include <QDir>
#include <QJsonObject>
#include <QJsonArray>
#include <QJsonDocument>

#define CLIPPER_VERSION "1.0"

struct Options {
    QStringList infiles;
    QString outpattern;
    bool discardNoChange;
};

class ClipperScene;

class SceneChanger
{
public:
    virtual QString saveImage(const QImage &, int, bool) = 0;
    virtual bool deleteImage(const QString&) = 0;
    virtual bool popScene() = 0;

    enum Direction {
        NEXT,
        PREVIOUS,
        SUBIMAGE
    };

    virtual bool pushScene(const QImage&, Direction = NEXT, int = 0) = 0;

    virtual ClipperScene* getScene() = 0;
    virtual void pushCmd(QUndoCommand *) = 0;
    virtual void redo() = 0;
    virtual void undo() = 0;

    virtual void quit() = 0;
};

enum CustomItemTypes {
    ClipperRectType = QGraphicsItem::UserType + 1,
};

class ClipperRect : public QGraphicsRectItem
{
public:
    enum Style
    {
        LASSO_SELECTION,
        SELECTED,
        SAVED,
    };

    struct rect_t {
        QRectF rect;
        Style style;
        int position;
        QTransform transform;
        QTransform createTransform;
    };

    ClipperRect(QRectF rect = QRectF(), QGraphicsItem *parent = 0) :
        QGraphicsRectItem(rect, parent), m_style(LASSO_SELECTION), m_pos(0)
    {}

    void setStyle(Style style)
    {
        m_style = style;
        QPen pen;
        QBrush brush;

        switch(m_style)
        {
        case LASSO_SELECTION:
            pen.setColor(Qt::black);
            pen.setWidth(3);
            break;
        case SELECTED:
            pen.setColor(Qt::green);
            pen.setWidth(2);
            brush.setColor(Qt::darkGreen);
            brush.setStyle(Qt::BDiagPattern);
            setBrush(brush);
            break;
        case SAVED:
            pen.setColor(Qt::gray);
            pen.setWidth(2);
            brush.setColor(Qt::gray);
            brush.setStyle(Qt::BDiagPattern);
            setBrush(brush);
        default:
            break;
        }
        setPen(pen);
    }
    Style style() const { return m_style; }
    void setPosition(const int pos) {m_pos = pos;}
    int position() const { return m_pos; }
    virtual int type() const { return ClipperRectType; }
    QTransform createTransform() const { return m_createTransform; }
    void setCreateTransform(QTransform transform) { m_createTransform = transform; }

    QRectF createRect()
    {
        QTransform bak = transform();
        QRectF ret;
        setTransform(m_createTransform);
        ret = boundingRect();
        setTransform(bak);
        return ret;
    }

    rect_t save()
    {
        rect_t ret;
        ret.rect = rect();
        ret.style = m_style;
        ret.position = m_pos;
        ret.transform = transform();
        ret.createTransform = m_createTransform;
        return ret;
    }

    void load(const rect_t& r)
    {
        setRect(r.rect);
        m_style = r.style;
        m_pos = r.position;
        setTransform(r.transform);
        m_createTransform = r.createTransform;
    }
private:
    Style m_style;
    int m_pos;
    QTransform m_createTransform;
};

template<typename... Locals>
class ClipperCommand : public QUndoCommand
{
public:
    typedef std::tuple<Locals...> locals_t;

    template<typename C>
    ClipperCommand(C&& cmd): QUndoCommand(0)
    {
        m_cmd = std::bind(cmd, std::placeholders::_1, this);
    }

    template<typename... Locals2>
    void set(Locals2&&... locals)
    {
        m_local = std::forward_as_tuple(locals...);
    }

    template<std::size_t I>
    typename std::tuple_element<I,locals_t>::type get()
    {
        return std::get<I>(m_local);
    }

    virtual void undo() { m_cmd(false); }
    virtual void redo() { m_cmd(true); }
private:

    locals_t m_local;
    std::function<void(bool)> m_cmd;
};

class ClipperScene : public QGraphicsScene
{
public:
    ClipperScene(const QImage &img, SceneChanger* changer, QObject *parent = 0) :
        QGraphicsScene(parent), changer(changer), m_pic(0), m_rectpos(0),
        m_changed(false), m_size(img.size())
    {
        m_pic = addPixmap(QPixmap::fromImage(img));
        fitPicInView();
    }

protected:

    virtual void keyPressEvent(QKeyEvent *event)
    {
        switch(event->key())
        {
        case Qt::Key_Q:
            changer->quit();
            break;

        // toggle fullscreen
        case Qt::Key_F:
        {
            QGraphicsView *view = getView();
            if(view->isFullScreen())
                view->showNormal();
            else
                view->showFullScreen();
            break;
        }

        case Qt::Key_U:
            changer->undo();
            break;
        case Qt::Key_R:
            changer->redo();
            break;
        default:
            break;
        }
    }

    virtual void mouseReleaseEvent(QGraphicsSceneMouseEvent *event)
    {
        debugMouseEvent("release",event);

        QPointF p1 = event->buttonDownScenePos(event->button()),
            p2 = event->scenePos();

        SceneChanger *ch = changer;

        switch(event->button())
        {
        case Qt::LeftButton: {
            QRectF rect(p1,p2);
            rect = rect.normalized();

            // Add a new rectangle
            if(!rect.size().isNull())
            {
                qDebug() << "Added rect of size " << rect;
                typedef ClipperCommand<int> cmd_rect_t;
                changer->pushCmd(new cmd_rect_t(
                    [rect, ch] (bool redo, cmd_rect_t* ptr) {
                        if(redo) {
                            int pos = -1;
                            if (ptr->get<0>())
                                pos = ptr->get<0>();
                            ClipperRect *cr = ch->getScene()->addRect(rect, pos);
                            ptr->set(cr->position());
                        } else {
                            ch->getScene()->removeRect(ptr->get<0>());
                        }
                    }));
            }
            // A normal click
            else
            {
                ClipperRect *cr = getClipperRect(event->scenePos());

                // Save the whole current image
                if(!cr)
                {
                    rect = sceneRect();
                    bool changed = m_changed;
                    QTransform tr = m_pic->transform();
                    typedef ClipperCommand<QString,QImage,RectData> cmd_saveimg_t;
                    changer->pushCmd(new cmd_saveimg_t(
                            [rect, ch, changed, tr] (bool redo, cmd_saveimg_t* ptr) {
                                if(redo){
                                    QImage img = ch->getScene()->createImage(rect, tr);
                                    RectData saved = ch->getScene()->saveRects();
                                    ptr->set(ch->saveImage(img, 0, changed), img, saved);
                                    ch->popScene();
                                } else {
                                    ch->deleteImage(ptr->get<0>());
                                    ch->pushScene(ptr->get<1>(), SceneChanger::PREVIOUS);
                                    ch->getScene()->loadRects(ptr->get<2>());
                                }
                            }));
                }
                // save a subimage limited by the ClipperRect
                else
                {
                    rect = cr->boundingRect();
                    typedef ClipperCommand<QString> cmd_savesubimg_t;
                    changer->pushCmd(new cmd_savesubimg_t(
                            [rect, cr, ch] (bool redo,
                                cmd_savesubimg_t* ptr){
                                if(redo){
                                    QImage img = ch->getScene()->createImage(rect,
                                        cr->createTransform());
                                    ptr->set(ch->saveImage(img, cr->position(), true));
                                    cr->setStyle(ClipperRect::SAVED);
                                } else {
                                    ch->deleteImage(ptr->get<0>());
                                    cr->setStyle(ClipperRect::SELECTED);
                                }
                            }));
                }
            }
            break;
        }
        case Qt::RightButton:
        {
            ClipperRect *cr = getClipperRect(event->scenePos());
            QRectF rect;
            // Discard the whole image and switch to the next
            if (!cr)
            {
                rect = sceneRect();
                QTransform tr = m_pic->transform();
                typedef ClipperCommand<QImage,RectData> cmd_discardimg_t;
                changer->pushCmd(new cmd_discardimg_t (
                        [rect, ch, tr](bool redo, cmd_discardimg_t *ptr) {
                            if (redo) {
                                ptr->set(ch->getScene()->createImage(rect, tr), ch->getScene()->saveRects());
                                ch->popScene();
                            } else {
                                ch->pushScene(ptr->get<0>(), SceneChanger::PREVIOUS);
                                ch->getScene()->loadRects(ptr->get<1>());
                            }
                        }));
            }
            // Discard a rectangle unless it has already been saved
            else if(cr->style() != ClipperRect::SAVED)
            {
                rect = cr->boundingRect();
                int rectpos = cr->position();
                typedef ClipperCommand<QRectF, int> cmd_removerect_t;
                changer->pushCmd(new cmd_removerect_t(
                        [rect, rectpos, ch] (bool redo, cmd_removerect_t* ptr) {
                            if (redo) {
                                int pos = rectpos;
                                if (ptr->get<1>() != 0)
                                    pos = ptr->get<1>();
                                ch->getScene()->removeRect(pos);
                                ptr->set(rect, pos);
                            } else {
                                ClipperRect *re = ch->getScene()->addRect(
                                    ptr->get<0>(), ptr->get<1>());
                                ptr->set(ptr->get<0>(), re->position());
                            }
                        }));
             }
            break;
        }
        case Qt::MiddleButton:
        {
            ClipperRect *cr = getClipperRect(event->scenePos());
            // zoom to a given rectangle
            if (cr)
            {
                QRectF rect = cr->createRect();
                int imagepos = cr->position();
                typedef ClipperCommand<> cmd_zoomimg_t;
                QTransform tr = cr->createTransform();

                qDebug() << "Current transform" << m_pic->transform() << "createtransform" << tr;
                changer->pushCmd(new cmd_zoomimg_t(
                        [rect, ch, imagepos, tr](bool redo, cmd_zoomimg_t*){
                            if (redo) {
                                QImage img = ch->getScene()->createImage(rect, tr);
                                ch->pushScene(img, SceneChanger::SUBIMAGE, imagepos);
                            } else {
                                ch->popScene();
                            }}));
            }
            break;
        }
        default:
            break;
        }
    }

    virtual void wheelEvent(QGraphicsSceneWheelEvent *event)
    {
        qDebug() << "Wheel: " << " modifiers: " << event->modifiers() << " delta: " <<
            event->delta() << " orientation: " << event->orientation();

        QTransform tr;

        // rotation
        if(event->modifiers() & Qt::ShiftModifier)
        {
            qreal amount = 2;
            if(event->modifiers() & Qt::ControlModifier)
                amount = 90;
            qreal deg = amount * sgn(event->delta());
            tr = QTransform()
                .translate(m_size.width() / 2, m_size.height() / 2)
                .rotate(deg)
                .translate(- m_size.width() / 2, - m_size.height() / 2);
            m_changed = true;
        }
#if 0 // disable scaling for now
        // scaling
        else if (event->modifiers() & Qt::ControlModifier)
        {
            qreal factor = 1.15;
            if(event->delta() < 0 )
                factor = 1.0 / factor;
            tr = QTransform().scale(factor,factor);
            m_changed = true;
        }
#endif
        m_pic->setTransform(tr, true);
        for (int i = 0; i < rects.size(); i++)
            rects.at(i)->setTransform(tr, true);
    }

private:

    template <typename T> int sgn(T val)
    {
        return (T(0) < val) - (val < T(0));
    }

    ClipperRect *addRect(const QRectF &coords, int pos = -1)
    {
        ClipperRect *rect = new ClipperRect();

        if (pos < 0)
            pos = ++m_rectpos;

        rect->setRect(coords);
        rect->setStyle(ClipperRect::SELECTED);
        rect->setPosition(pos);
        rect->setCreateTransform(m_pic->transform());
        rects.push_back(rect);
        addItem(rect);
        rect->show();
        return rect;
    }

    ClipperRect *addRect(const ClipperRect::rect_t &r)
    {
        ClipperRect *ret = addRect(r.rect, r.position);
        ret->load(r);
        return ret;
    }

    void removeRect(int pos)
    {
        QList<ClipperRect*>::iterator it = rects.begin();
        for(;it != rects.end(); ++it)
            if((*it)->position() == pos)
            {
                removeItem(*it);
                rects.erase(it);
                break;
            }
    }

    void displayRects(bool visible)
    {
        for (int i = 0; i < rects.size(); i++)
            if(!visible)
                rects.at(i)->hide();
            else
                rects.at(i)->show();
    }

    QGraphicsItem* getItem(QPointF at)
    {
        return itemAt(at,QTransform());
    }

    bool pointsAtPic(QPointF at)
    {
        return getItem(at) == m_pic;
    }

    QRectF getSubItemRect(QPointF at)
    {
        QGraphicsItem *it = getItem(at);
        QRectF ret;
        if (!it || it->type() != ClipperRectType)
            return ret;

        return it->boundingRect();
    }
    ClipperRect *getClipperRect(QPointF at)
    {
        QGraphicsItem *it = getItem(at);
        if(!it || it->type() != ClipperRectType)
            return 0;
        return static_cast<ClipperRect*>(it);
    }

    QGraphicsView *getView()
    {
        return (views().count()) ? views().first() : 0;
    }

    void fitPicInView()
    {
        if(!m_pic || !views().count())
            return;

        QGraphicsView *view = getView();
        view->centerOn(m_pic);
        view->fitInView(m_pic,Qt::KeepAspectRatio);
    }

    QImage createImage(const QRectF &rect, QTransform tr = QTransform())
    {
        QImage img(rect.size().toSize(),QImage::Format_ARGB32);
        img.fill(Qt::white);

        QPainter pt(&img);
        displayRects(false);
        QTransform bak = m_pic->transform();
        m_pic->setTransform(tr);
        render(&pt, QRectF(), rect);
        m_pic->setTransform(bak);
        displayRects(true);
        return img;
    }

    void debugMouseEvent(char const* name, QGraphicsSceneMouseEvent *event)
    {
        qDebug() << name << ":" << event->button() << "pos: " <<
            event->buttonDownScenePos(event->button()) << " last: " <<
            event->scenePos();
    }

    typedef QList<ClipperRect::rect_t> RectData;

    RectData saveRects()
    {
        RectData ret;
        for (int i = 0; i < rects.size(); i++)
            ret.push_back(rects.at(i)->save());
        return ret;
    }
    void loadRects(const RectData &data)
    {
        for (int i = 0; i < data.size(); i++)
            addRect(data.at(i));
    }

    SceneChanger *changer;
    QGraphicsPixmapItem *m_pic;
    ClipperRect selectionRect;
    QList<ClipperRect*> rects;
    int m_rectpos;
    bool m_changed;
    QSize m_size;
};


class ClipperView : public QGraphicsView
{
public:
    explicit ClipperView(QWidget *parent = 0) : QGraphicsView(parent)
    {
        setRenderHints(QPainter::Antialiasing | QPainter::SmoothPixmapTransform);
        setDragMode(QGraphicsView::RubberBandDrag);
    }

    void setClipperScene(ClipperScene *scene)
    {
        setScene(scene);
        fitInView(scene->sceneRect(),Qt::KeepAspectRatio);
    }
    void resizeEvent(QResizeEvent *)
    {
        fitInView(scene()->sceneRect(),Qt::KeepAspectRatio);
    }
};

class ClipperJsonFile
{
    QJsonArray images;
public:

    void add(const QString& name, const QString& parent = QString())
    {
        QJsonObject item;
        item["name"] = name;
        if(!parent.isEmpty())
            item["parent"] = parent;

        images.push_back(item);
    }
    void remove(const QString& name)
    {
        QJsonArray::iterator it = images.begin();

        for (; it != images.end(); it++)
        {
            QJsonObject ob = (*it).toObject();
            if (ob["name"].toString() == name)
            {
                images.erase(it);
                break;
            }
        }
    }
    bool save(const QString& filename)
    {
        QJsonDocument doc(images);

        QString tmp = "";
        tmp.append(doc.toJson());
        qDebug() << "Generating the following json into " << filename << ":\n" << tmp;

        QFile fp(filename);
        if(!fp.open(QIODevice::WriteOnly))
        {
            qWarning() << "Could not open file " << filename << " for writing.";
            return false;
        }
        return fp.write(doc.toJson()) >= 0;
    }
};

class SceneHandler : public SceneChanger
{
public:
    SceneHandler(Options &opts) : sceneStack(), view(), opts(opts)
    {
        scenePosStack.push(0);
        popScene();
        cmdStack.setUndoLimit(10);
    }

    virtual QString saveImage(const QImage &img, int subImageNumber, bool changed)
    {
        if(img.isNull() || (!changed && opts.discardNoChange))
            return QString();

        QString name = opts.outpattern;
        QString parent;
        for (int i = 0; i < scenePosStack.size(); i++)
            name += QString("-") + QString::number(scenePosStack.at(i));

        if (subImageNumber > 0)
        {
            parent = name + ".jpg";
            name = name + "+" + QString::number(subImageNumber);
        }
        name += ".jpg";
        qDebug() << "Saving image to " << name;
        json.add(name, parent);
        img.save(name);

        return name;
    }

    virtual bool deleteImage(const QString& name)
    {
        qDebug() << "removing file " << name;
        if(name.isNull())
            return false;

        json.remove(name);
        return QFile::remove(name);
    }

    void quit()
    {
        json.save("clipper-images.json");
        QCoreApplication::exit();
    }

    // replaces current scene with a new one or quits if not available.
    bool popScene()
    {
        QGraphicsScene *top = 0;
        if(!sceneStack.empty())
            top = sceneStack.pop();

        if(sceneStack.empty())
        {
            QString fname;
            do {
                if(opts.infiles.empty())
                {
                    this->quit();
                    return false;
                }
                fname = opts.infiles.first();
                opts.infiles.removeFirst();
                ++scenePosStack.top();
                qDebug()<< "Filename is here" << fname << "and position " <<
                    scenePosStack.top();
            } while(newScene(fname).isNull());
        }
        else if(scenePosStack.size() > 1)
            scenePosStack.pop();
        view.setClipperScene(sceneStack.top());
        view.show();

        if(top)
            top->deleteLater();

        return true;
    }


    bool pushScene(const QImage& img, Direction dir = NEXT, int pos = 0)
    {
        if (dir == PREVIOUS)
            scenePosStack.top()--;
        else if(dir == SUBIMAGE)
            scenePosStack.push(pos);

        newScene(img);
        view.setClipperScene(sceneStack.top());

        return true;
    }

    ClipperScene *getScene()
    {
        return !sceneStack.empty() ? sceneStack.top() : 0;
    }

    void pushCmd(QUndoCommand *cmd) { cmdStack.push(cmd); }
    void undo() { cmdStack.undo(); }
    void redo() { cmdStack.redo(); }

private:

    QImage newScene(const QImage &img)
    {
        sceneStack.push(new ClipperScene(img, this));
        return img;
    }

    QImage newScene(const QString &name)
    {
        QImage img;
        if (!img.load(name))
            return QImage();
        return newScene(img);
    }

    QStack<ClipperScene* > sceneStack;
    QUndoStack cmdStack;
    QStack<int> scenePosStack;
    ClipperView view;
    Options opts;
    ClipperJsonFile json;
};

int main(int argc, char **argv)
{
    QApplication app(argc, argv);
    QCommandLineParser parser;
    Options opts;

    // Command line handling
    QCoreApplication::setApplicationName(argv[0]);
    QCoreApplication::setApplicationVersion(CLIPPER_VERSION);
    parser.addVersionOption();
    parser.addHelpOption();
    parser.setApplicationDescription("Image clipping application.");
    QCommandLineOption OutOpt(QStringList() << "o" << "out",
        "Pattern of output files.", "outpattern");
    parser.addOption(OutOpt);
    QCommandLineOption DiscardNoChange(QStringList() << "d" <<
        "discard-if-not-changed", "Discard the image when saving if not changed.");
    parser.addOption(DiscardNoChange);
    parser.process(app);
    opts.infiles = parser.positionalArguments();
    if (!parser.isSet(OutOpt))
    {
        qCritical() << "Error: --out is a required parameter.\n";
        parser.showHelp(1);
        Q_UNREACHABLE();
    }
    opts.outpattern = parser.value(OutOpt);

    QFileInfo info(opts.outpattern);
    if (!info.dir().exists())
    {
        qCritical() << "Error: --out directory" << info.dir().path() <<
            " does not exist.\n";
        parser.showHelp(1);
        Q_UNREACHABLE();
    }

    opts.discardNoChange = false;
    if (parser.isSet(DiscardNoChange))
        opts.discardNoChange = true;

    if (opts.infiles.empty())
        return 0;

    //DEBUG
    qDebug() << "infiles:";
    for (int i = 0; i < opts.infiles.size(); ++i)
        qDebug() << opts.infiles.at(i).toLocal8Bit().constData();
    qDebug() << "outpattern: " << opts.outpattern.toLocal8Bit().constData();

    // The processing
    SceneHandler sh(opts);
    return app.exec();
}
