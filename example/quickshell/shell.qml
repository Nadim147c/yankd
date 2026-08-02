pragma ComponentBehavior: Bound
import qs.yankd

import QtQuick
import QtQuick.Layouts
import Quickshell
import Quickshell.Wayland
import Quickshell.Widgets

ShellRoot {
    LazyLoader {
        id: loader
        active: true
        PanelWindow {
            WlrLayershell.namespace: "quickshell:clipboard"
            WlrLayershell.keyboardFocus: WlrKeyboardFocus.OnDemand
            aboveWindows: true
            exclusiveZone: 0

            implicitWidth: body.width
            implicitHeight: body.height
            color: "transparent"
            Rectangle {
                id: body
                implicitWidth: content.width
                implicitHeight: content.height
                color: "#111111"
                radius: 20
                Body {
                    id: content
                }
            }
        }
    }
    component Body: ColumnLayout {
        id: col
        Component.onCompleted: Yankd.search("")
        Input {}
        Entries {}
        Preview {}
    }

    component Input: Item {
        implicitWidth: 500
        implicitHeight: 50
        Rectangle {
            implicitWidth: parent.width - 10
            implicitHeight: parent.height - 10
            anchors.centerIn: parent
            radius: 18
            bottomLeftRadius: 5
            bottomRightRadius: 5
            color: "#222222"
            TextInput {
                color: "#eeeeee"
                width: parent.width - 20
                anchors.centerIn: parent
                focus: true
                onTextChanged: Yankd.search(text)
                Keys.onPressed: event => {
                    if (event.key === Qt.Key_Escape) {
                        loader.active = false;
                        Quickshell.execDetached(["qs", "kill", "-p", Quickshell.shellDir]);
                        return;
                    }

                    // Ctrl+N or Down Arrow → move forward
                    if ((event.key === Qt.Key_N && event.modifiers === Qt.ControlModifier) || event.key === Qt.Key_Down) {
                        Yankd.next();
                        event.accepted = true;
                        return;
                    }

                    // Ctrl+P or Up Arrow → move backward
                    if ((event.key === Qt.Key_P && event.modifiers === Qt.ControlModifier) || event.key === Qt.Key_Up) {
                        Yankd.previous();
                        event.accepted = true;
                        return;
                    }

                    // Enter → trigger action
                    if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter) {
                        Yankd.set();
                        Quickshell.execDetached(["qs", "kill", "-p", Quickshell.shellDir]);
                        return;
                    }
                }
            }
        }
    }
    component Preview: Item {
        implicitWidth: 500
        implicitHeight: 200
        ClippingRectangle {
            implicitWidth: parent.width - 10
            implicitHeight: parent.height - 10
            anchors.centerIn: parent
            radius: 18
            topLeftRadius: 5
            topRightRadius: 5
            color: "#222222"
            Loader {
                anchors.fill: parent
                active: Yankd.previewMimeType.startsWith("image/")
                sourceComponent: Image {
                    anchors.fill: parent
                    source: `data:${Yankd.currentEntry.mimeType};base64,${Yankd.preview}`
                    fillMode: Image.PreserveAspectCrop
                }
            }
            Loader {
                anchors.fill: parent
                active: !Yankd.previewMimeType.startsWith("image/")
                sourceComponent: Text {
                    color: "#dddddd"
                    width: parent.width - 10
                    height: parent.height - 10
                    x: 5
                    y: 5
                    text: Yankd.preview
                }
            }
        }
    }

    component Entries: ClippingRectangle {
        color: "transparent"
        implicitHeight: 300
        implicitWidth: parent.width
        ListView {
            id: list
            implicitHeight: parent.height
            implicitWidth: parent.width - 20
            anchors.centerIn: parent
            model: Yankd.searchResult
            currentIndex: Yankd.searchIndex
            delegate: Item {
                id: entry
                width: list.width
                height: row.height + 5
                required property string mimeType
                required property string eventID
                required property string preview
                required property string score
                required property date time
                property bool isCurrent: ListView.isCurrentItem
                RowLayout {
                    id: row
                    width: entry.width - 10
                    anchors.centerIn: parent
                    spacing: 10
                    Text {
                        id: idTxt
                        text: `${entry.score}`.padStart(3, "0")
                        color: "#dddddd"
                    }
                    Item {
                        Layout.fillHeight: true
                        Layout.fillWidth: true
                        Text {
                            anchors.fill: parent
                            color: "#dddddd"
                            elide: Text.ElideRight
                            textFormat: Text.PlainText
                            text: entry.preview || `${entry.mimeType} from ${Qt.formatDateTime(entry.time, "dd MMM yyyy")}`
                        }
                    }
                }
                Rectangle {
                    anchors.fill: parent
                    visible: entry.isCurrent
                    color: "transparent"
                    radius: 5
                    border {
                        width: 2
                        color: "white"
                    }
                }
            }
        }
    }
}
