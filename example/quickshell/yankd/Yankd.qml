pragma Singleton

import QtQuick
import Quickshell
import Quickshell.Io

Singleton {
    id: root

    property string binName: "yankd"
    function search(query: string) {
        searchProc.exec([binName, "search", "--format=json", query]);
    }

    function set(query: string) {
        Quickshell.execDetached([binName, "set", `${currentEntry.eventID}`]);
    }

    function pauseDaemon(state: string) {
        if (!state)
            state = "toggle";
        Quickshell.execDetached([binName, "daemon", "pause", state]);
    }

    property int searchIndex: 0
    onSearchIndexChanged: updatePreview()
    function setIndex(i: int) {
        searchIndex = Math.min(Math.max(i, 0), searchResult.length - 1);
    }
    function next() {
        searchIndex = (searchIndex + 1) % searchResult.length;
    }
    function previous() {
        if (searchIndex === 0)
            searchIndex = searchResult.length - 1;
        searchIndex = searchIndex - 1;
    }

    property string previewMimeType: ""
    property string preview: ""
    function updatePreview() {
        previewProc.exec([binName, "get", "--primary", "--quiet", "--base64", `${searchResult[searchIndex].eventID}`]);
    }
    property YankdSearchEntry currentEntry: searchResult[searchIndex]
    property list<YankdSearchEntry> searchResult
    onSearchResultChanged: searchIndex = 0

    Component {
        id: yankdSearchEntry
        YankdSearchEntry {}
    }

    Process {
        id: previewProc
        stdout: StdioCollector {
            onStreamFinished: {
                root.preview = text;
                root.previewMimeType = root.currentEntry.mimeType;
            }
        }
    }
    Process {
        id: searchProc
        stdout: StdioCollector {
            onStreamFinished: {
                if (text.length === 0) {
                    root.searchResult = [];
                    return;
                }
                const data = JSON.parse(text);
                const res = [];
                for (const entry of data) {
                    res.push(yankdSearchEntry.createObject(root, {
                        eventID: entry.id,
                        mimeType: entry.mime_type,
                        time: new Date(entry.time),
                        preview: entry.preview,
                        score: Math.round(entry.score)
                    }));
                }
                root.searchResult = res;
                root.updatePreview();
            }
        }
    }
}
