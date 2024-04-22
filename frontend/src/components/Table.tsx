import React, {Component} from 'react';
import {
    ExportMeta,
    GetTorrents,
    OpenFolder,
    SetActive,
    WantRemoveTorrent
} from "../../wailsjs/go/main/App";
import {EventsEmit, EventsOff, EventsOn} from "../../wailsjs/runtime";
import Play from "../../public/light/play.svg";
import Pause from "../../public/light/pause.svg";
import Close from "../../public/light/close.svg";
import OpenDir from "../assets/images/icons/open-folder.svg";
import Export from "../assets/images/icons/export.svg";
import Copy from "../assets/images/icons/copy.svg";

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
    peersNum: number
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

export function textState(state: string, peers: number) {
    switch (state) {
        case "seeding":
            return "Seeding";
        case "verifying":
            return "Verifying";
        case "downloading":
            return peers > 0 ? "Downloading" : "Searching for peers";
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
                    peersNum: t.PeersNum,
                    selected:  selected,
                })
            })

            this.setState({
                torrents: newList
            });

            let selected = newList.filter((tr)=>{return tr.selected && this.checkFilters(tr.name, tr.state)});
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
                // EventsEmit("select-torrent", ti.id);
                return {
                    hash: ti.id,
                    active: ti.state == "downloading" || ti.state == "seeding",
                }
            }));
        }
    }

    checkFilters = (name: string, state: string) => {
        if (this.props.filter.search.length > 0) {
            if (!name.toLowerCase().includes(this.props.filter.search.toLowerCase())) {
                return false;
            }
        }

        switch (this.props.filter.type) {
            case "Downloading":
                if(state != "downloading")
                    return false;
                break
            case "Seeding":
                if(state != "seeding")
                    return false;
                break
            case "Failed":
                if(state != "fail")
                    return false;
                break
            case "Active":
                if(state != "downloading" && state != "seeding")
                    return false;
                break
            case "Inactive":
                if(state != "fail" && state != "inactive")
                    return false;
                break
        }
        return true;
    }

    renderTorrentsList() {
        let items = [];

        for (let t of this.state.torrents) {
            if (!this.checkFilters(t.name, t.state)) {
                continue
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
                                   <img src={OpenDir} alt=""/><span>Open directory</span></div>)

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
                                   WantRemoveTorrent([t.id]).then(Refresh)
                               }}><img src={Close} alt=""/><span>Remove</span></div>)

                               elems.push(<div onClick={() => {
                                   ExportMeta(t.id).then()
                               }}><img src={Export} alt=""/><span>Export .tonbag</span></div>)

                               elems.push(<div onClick={() => {
                                   navigator.clipboard.writeText(t.id).then();
                               }}><img src={Copy} alt=""/><span>Copy bag ID</span></div>)

                               this.setState((current) => ({ ...current, contextShow: true, contextItems: elems}));

                               document.body.addEventListener("click", () => {
                                  this.setState((current) => ({ ...current, contextShow: false, contextItems: []}))
                              }, { once: true });
                           }}>
                <td style={{flexGrow:"1", width: "70px"}}><div className={"item-name"}><div className={"item-state-container"} onMouseEnter={(e) =>{
                    let tip = document.getElementById("tip");
                    tip!.textContent = textState(t.state, t.peersNum);
                    let rectItem = document.getElementById("state-"+t.id)!.getBoundingClientRect()
                    let rectTip = tip!.getBoundingClientRect();

                    tip!.style.top =  (rectItem.y - (rectTip.height + 12)).toString()+"px";
                    tip!.style.left = (rectItem.x - (rectTip.width/2 - rectItem.width/2)).toString()+"px";

                    tip!.style.opacity = "1";
                    tip!.style.visibility = "visible";
                }} onMouseLeave={
                    (e)=> {
                        let tip = document.getElementById("tip");
                        tip!.style.opacity = "0";
                        tip!.style.visibility = "hidden";
                    }
                }><div id={"state-"+t.id} className={"item-state "+(t.state == 'verifying' || (t.state == 'downloading' && t.peersNum == 0) ? 'searching' : t.state)}></div></div><span>{t.name}</span></div></td>
                <td style={{width:"130px"}}><div className="progress-block-small">
                    <span style={{textAlign:"left", width:"38px"}}>{t.progress}%</span>
                    <div className="progress-bar-small-form">
                        <div className="progress-bar-small" style={{width: t.progress+"%"}}></div>
                    </div>
                </div>
                </td>
                <td style={{width:"80px", justifyContent: "flex-end"}}>{t.size}</td>
                <td style={{width:"60px", justifyContent: "flex-end"}}>{t.peersNum}</td>
                <td style={{width:"95px", justifyContent: "flex-end"}}>{t.downloadSpeed}</td>
                <td style={{width:"95px", justifyContent: "flex-end", paddingRight:"12px"}}>{t.uploadSpeed}</td>
            </tr>);
        }
        return items;
    }

    render() {
        return <table style={{fontSize:12}}>
            <div id="menu-back" className="context-backdrop" style={{visibility: this.state.contextShow ? "visible":"hidden"}}/>
            <div id="menu" className="context-menu" style={{visibility: this.state.contextShow ? "visible":"hidden"}}>
                {this.state.contextItems}
            </div>
            <thead>
            <tr>
                <th style={{flexGrow:"1"}}>Name</th>
                <th style={{width:"130px"}}>Progress</th>
                <th style={{width:"90px", justifyContent: "flex-end"}}>Size</th>
                <th style={{width:"60px", justifyContent: "flex-end"}}>Peers</th>
                <th style={{width:"95px", justifyContent: "flex-end"}}>Download</th>
                <th style={{width:"95px", justifyContent: "flex-end", paddingRight:"12px"}}>Upload</th>
            </tr>
            </thead>
            <tbody style={{overflowX: "auto"}}>
            {this.renderTorrentsList()}
            </tbody>
        </table>
    }
}