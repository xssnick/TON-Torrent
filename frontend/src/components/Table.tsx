import React, {Component} from 'react';
import {GetTorrents, OpenFolder} from "../../wailsjs/go/main/App";
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime"; // let's also import Component

export interface TorrentItem {
    id: string
    name: string
    state: string
    size: string
    uploadSpeed: string
    downloadSpeed: string
    path: string
    progress: number
    selected: boolean
}

interface State {
    torrents: TorrentItem[]
}

export interface Filter {
    type: string
    search: string
}

export interface TableProps {
    filter: Filter
    onSelect: (items: string[]) => void
}

export function refresh() {
    EventsEmit("refresh")
}

export class Table extends Component<TableProps,State> {
    private tmpTorrents: TorrentItem[] = [{
        id: "1",
        name: "Half Life 3",
        state: "seeding",
        size: "4.12 Gb",
        uploadSpeed: "1.12 Mb/s",
        downloadSpeed: "",
        progress: 100,
        selected: false,
        path: "",
    },{
        id: "2",
        name: "Tom Clancy: Ghost of Reconnect",
        state: "seeding",
        size: "12.89 Gb",
        uploadSpeed: "121.1 Kb/s",
        downloadSpeed: "",
        progress: 100,
        selected: false,
        path: "",
    },{
        id: "3",
        name: "Dota 3",
        state: "downloading",
        size: "20000.12 Gb",
        uploadSpeed: "1.92 Mb/s",
        downloadSpeed: "79.92 Mb/s",
        progress: 31.8,
        selected: false,
        path: "",
    },{
        id: "4",
        name: "Lost, Season 3",
        state: "fail",
        size: "189 Mb",
        uploadSpeed: "",
        downloadSpeed: "",
        progress: 0,
        selected: false,
        path: "",
    },{
        id: "5",
        name: "Dungeons & Dragons 5: Muhammed Ali",
        state: "inactive",
        size: "982 Mb",
        uploadSpeed: "",
        downloadSpeed: "",
        progress: 100,
        selected: false,
        path: "",
    }];

    update() {
        GetTorrents().then((tr)=>{
            let newList: TorrentItem[] = []
            tr.forEach((t)=> {
                let found = this.state.torrents.find((tf) => {
                    return tf.id == t.ID
                })

                let selected = found?.selected == true

                newList.push({
                    id: t.ID,
                    name: t.Name,
                    state: t.State,
                    size: t.Size,
                    uploadSpeed: t.Upload,
                    downloadSpeed: t.Download,
                    path: t.Path,
                    progress: t.Progress,
                    selected:  selected,
                })
            })

            this.setState({
                torrents: newList.concat(this.tmpTorrents)
            });
        });
    }

    // Before the component mounts, we initialise our state
    componentWillMount() {
        this.setState({
            torrents: [],
        });
    }

    componentWillUnmount() {
        EventsOff("update")
    }

        // After the component did mount, we set the state each second.
    componentDidMount() {
        EventsOn("update", () => {
            this.update();
        })
    }

    clickRow(t: TorrentItem) {
        return () => {
            if (!t.selected) {
                // we should have only 1 selected when regular click
                t.selected = true

                let unselect = (e: MouseEvent) => {
                    let target = e.target as HTMLElement;
                    if (target.classList.contains("top-button")) {
                        // not unselect when control button clicked
                        return
                    }

                    if (!e.ctrlKey && !e.metaKey) {
                        this.state.torrents.forEach((ti) => {
                            ti.selected = false
                        })

                        // copy to update state
                        this.setState({
                            torrents: this.state.torrents.slice(0, this.state.torrents.length)
                        });
                        this.props.onSelect([]);
                        document.body.removeEventListener("mouseup", unselect);
                    }
                }

                // unselect all when click on something else
                document.body.addEventListener("mouseup", unselect);
            } else {
                // just unselect
                t.selected = false
            }

            // copy to update state
            this.setState({
                torrents: this.state.torrents.slice(0, this.state.torrents.length)
            });
            let selected = this.state.torrents.filter((tr)=>{return tr.selected});
            this.props.onSelect(selected.map<string>((ti) => ti.id));
        }
    }

    renderTorrentsList() {
        let items = [];
        let i = 0;

        for (let t of this.state.torrents) {
            if (this.props.filter.search.length > 0) {
                if (!t.name.toLowerCase().includes(this.props.filter.search.toLowerCase())) {
                    continue
                }
            }

            switch (this.props.filter.type) {
                case "Downloading":
                    if(t.state != "downloading")
                        continue
                    break
                case "Seeding":
                    if(t.state != "seeding")
                        continue
                    break
                case "Failed":
                    if(t.state != "fail")
                        continue
                    break
                case "Active":
                    if(t.state != "downloading" && t.state != "seeding")
                        continue
                    break
                case "Inactive":
                    if(t.state != "fail" && t.state != "inactive")
                        continue
                    break
            }

            let addClass = "";
            if (i++ % 2 == 0) {
                addClass = "gray"
            }
            items.push(<tr className={t.selected ? "torrent-selected" : ""} key={t.id}
                           onClick={this.clickRow(t)} onDoubleClick={() => {OpenFolder(t.path).then(r => {})}}>
                <td className={addClass}><div id={"state-"+t.id} className={"item-state "+t.state} onMouseEnter={(e) =>{
                    let tip = document.getElementById("tip")
                    let rectItem = document.getElementById("state-"+t.id)!.getBoundingClientRect()
                    let rectTip = tip!.getBoundingClientRect()

                    tip!.style.top =  (rectItem.y - (rectTip.height + 5)).toString()+"px";
                    tip!.style.left = (rectItem.x - (rectTip.width/2 - rectItem.width/2)).toString()+"px";

                    tip!.style.opacity = "1";
                    tip!.style.visibility = "visible";
                }} onMouseLeave={
                    (e)=> {
                        let tip = document.getElementById("tip")
                        tip!.style.opacity = "0";
                        tip!.style.visibility = "hidden";
                    }
                }></div></td>
                <td className={addClass} style={{maxWidth:"150px"}}>{t.name}</td>
                <td className={"small "+addClass}>{t.size}</td>
                <td className={addClass}><div className="progress-block-small">
                    <span style={{textAlign:"left", width:"27px"}}>{t.progress}%</span>
                    <div className="progress-bar-small-form">
                        <div className="progress-bar-small" style={{width: t.progress+"%"}}></div>
                    </div></div></td>
                <td className={"small "+addClass}>{t.downloadSpeed}</td>
                <td className={"small "+addClass}>{t.uploadSpeed}</td>
            </tr>);
        }
        return items;
    }

    render() {
        return <table style={{fontSize:12}}>
            <thead>
            <tr>
                <th></th>
                <th>Description</th>
                <th>Size</th>
                <th>Progress</th>
                <th>Download</th>
                <th>Upload</th>
            </tr>
            </thead>
            <tbody>
                {this.renderTorrentsList()}
            </tbody>
        </table>
    }
}