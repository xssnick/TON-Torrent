import React, {Component, MouseEvent} from 'react';
import {ExportMeta, GetTorrents, OpenFolder, RemoveTorrent, SetActive} from "../../wailsjs/go/main/App";
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime";
import Download from "../assets/images/icons/download.svg";
import Play from "../assets/images/icons/play.svg";
import Pause from "../assets/images/icons/pause.svg";
import Close from "../assets/images/icons/close.svg";

export interface SelectedTorrent {
    hash: string
    active: boolean
}

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
    contextShow: boolean
    contextItems: JSX.Element[]
    torrents: TorrentItem[]
}

export interface Filter {
    type: string
    search: string
}

export interface TableProps {
    filter: Filter
    onSelect: (items: SelectedTorrent[]) => void
}

export function Refresh() {
    EventsEmit("refresh");
}

export function textState(state: string) {
    switch (state) {
        case "seeding":
            return "Seeding";
        case "downloading":
            return "Downloading";
        case "fail":
            return "Failed";
        case "inactive":
            return "Inactive";
    }
    return "";
}

export class Table extends Component<TableProps,State> {
    constructor(props: TableProps, state:State) {
        super(props, state);

        this.state = {
            contextShow: false,
            contextItems: [],
            torrents: [],
        }
    }

    private tmpTorrents: TorrentItem[] = [/*{
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
    }*/];

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

            let list = newList.concat(this.tmpTorrents);
            this.setState({
                torrents: list
            });

            let selected = list.filter((tr)=>{return tr.selected});
            this.props.onSelect(selected.map<SelectedTorrent>((ti) => {
                return {
                    hash: ti.id,
                    active: ti.state == "downloading" || ti.state == "seeding",
                }
            }));
        });
    }

    componentDidMount() {
        EventsOn("update", () => {
            this.update();
        })
    }
    componentWillUnmount() {
        EventsOff("update")
    }

    clickRow(t: TorrentItem) {
        return (e: React.MouseEvent) => {
            // unselect old when no ctrl or command pressed
            if (!e.ctrlKey && !e.metaKey) {
                this.state.torrents.forEach((ti) => {
                    ti.selected = false
                })
            }

            if (!t.selected) {
                // select clicked
                t.selected = true
            }

            // copy to update state
            this.setState({
                torrents: this.state.torrents.slice(0, this.state.torrents.length)
            });

            // report selected to callback
            let selected = this.state.torrents.filter((tr)=>{return tr.selected});
            this.props.onSelect(selected.map<SelectedTorrent>((ti) => {
                return {
                    hash: ti.id,
                    active: ti.state == "downloading" || ti.state == "seeding",
                }
            }));
        }
    }

    renderTorrentsList() {
        let items = [];

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

            items.push(<tr className={t.selected ? "torrent-row torrent-selected" : "torrent-row"} key={t.id}
                           onClick={this.clickRow(t)} onDoubleClick={() => {OpenFolder(t.path).then()}}
                           onContextMenu={(e)=>{
                               e.preventDefault()

                               let menu = document.getElementById("menu")
                               let menuBack = document.getElementById("menu-back")
                               menu!.style.top =  e.pageY+"px";
                               menu!.style.left = e.pageX+"px";

                               let elems: JSX.Element[] = [];

                               elems.push(<div onClick={() => {
                                   OpenFolder(t.path).then()}}>
                                   <img src={Download} alt=""/><span>Open directory</span></div>)

                               if (t.state != "downloading" && t.state != "seeding" && t.state != "fail") {
                                   elems.push(<div onClick={() => {
                                       SetActive(t.id, true).then(Refresh)
                                   }}>
                                       <img src={Play} alt=""/><span>Start</span></div>)
                               }
                               if (t.state != "inactive" && t.state != "fail") {
                                   elems.push(<div onClick={() => {
                                       SetActive(t.id, false).then(Refresh)
                                   }}><img src={Pause} alt=""/><span>Pause</span></div>)
                               }
                               elems.push(<div onClick={() => {
                                   RemoveTorrent(t.id, false, false).then(Refresh)
                               }}><img src={Close} alt=""/><span>Remove</span></div>)
                               elems.push(<div onClick={() => {
                                   ExportMeta(t.id).then()
                               }}><img src={Close} alt=""/><span>Export .tonbag</span></div>)

                               this.setState((current) => ({ ...current, contextShow: true, contextItems: elems}))

                               document.body.addEventListener("click", () => {
                                  this.setState((current) => ({ ...current, contextShow: false, contextItems: []}))
                              }, { once: true });
                           }}>
                <td onMouseEnter={(e) =>{
                    let tip = document.getElementById("tip")
                    tip!.textContent = textState(t.state);
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
                }><div id={"state-"+t.id} className={"item-state "+t.state}></div></td>
                <td>{t.name}</td>
                <td className={"small"}>{t.size}</td>
                <td><div className="progress-block-small">
                    <span style={{textAlign:"left", width:"27px"}}>{t.progress}%</span>
                    <div className="progress-bar-small-form">
                        <div className="progress-bar-small" style={{width: t.progress+"%"}}></div>
                    </div></div></td>
                <td className={"small"}>{t.downloadSpeed}</td>
                <td className={"small"}>{t.uploadSpeed}</td>
            </tr>);
        }
        return items;
    }

    render() {
        return <table style={{fontSize:12}}>
            <span id="tip" className="tooltip"/>
            <div id="menu-back" className="context-backdrop" style={{visibility: this.state.contextShow ? "visible":"hidden"}}/>
            <div id="menu" className="context-menu" style={{visibility: this.state.contextShow ? "visible":"hidden"}}>
                {this.state.contextItems}
            </div>
            <thead>
            <tr>
                <th style={{width:"23px"}}></th>
                <th>Description</th>
                <th style={{width:"80px"}}>Size</th>
                <th style={{width:"150px"}}>Progress</th>
                <th style={{width:"90px"}}>Download</th>
                <th style={{width:"90px"}}>Upload</th>
            </tr>
            </thead>
            <tbody>
            {this.renderTorrentsList()}
            </tbody>
        </table>
    }
}