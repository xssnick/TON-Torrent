import React, {Component} from 'react';
import {EventsOff, EventsOn} from "../../wailsjs/runtime";
import {GetInfo} from "../../wailsjs/go/main/App";
import {textState} from "./Table";


export interface InfoProps {
    torrent: string
}

interface State {
    left: string
    path: string
    description: string
    downloadSpeed: string
    uploadSpeed: string
    downloaded: string
    size: string
    status: string
    added: string
    peers: string
    progress: string
}

export class InfoTorrentMenu extends Component<InfoProps,State> {
    constructor(props: InfoProps, state:State) {
        super(props, state);

        this.state = {
            left: "",
            path: "",
            description: "",
            downloadSpeed: "",
            uploadSpeed: "",
            downloaded: "",
            size: "",
            status: "",
            added: "",
            peers: "",
            progress: "",
        }
    }

    short(val: string) {
        return val.length > 40 ? val.slice(0, 40)+ "..." : val
    }

    copy(text: string) {
        return () => {navigator.clipboard.writeText(text).then()}
    }

    update() {
        GetInfo(this.props.torrent).then((tr: any)=>{
            this.setState({
                left: tr.TimeLeft,
                path: tr.Path,
                progress: tr.Progress,
                description: tr.Description,
                downloadSpeed: tr.Download,
                uploadSpeed: tr.Upload,
                downloaded: tr.Downloaded,
                size: tr.Size,
                status: tr.State,
                added: tr.AddedAt,
                peers: tr.Peers,
            });
        });
    }

    componentDidMount() {
        this.update()
        EventsOn("update_info", () => {
            this.update();
        })
    }
    componentWillUnmount() {
        EventsOff("update_info")
    }

    render() {
        return <div>
           <div className="info-progress-block">
               <span>Downloaded</span>
               <div className="info-progress-bar"><div className="filled" style={{width: this.state.progress+"%"}}/></div>
               <span className="percent">{this.state.progress}%</span>
           </div>
            <div className="info-data-block">
                { (this.state.status == "downloading" || this.state.status == "seeding") ? <div className="basic">
                    <div className="item" style={{width: "33%"}}><span className="field">Upload Speed: </span><span className="value">{this.state.uploadSpeed}</span></div>
                    <div className="item" style={{width: "40%"}}><span className="field">Download Speed: </span><span className="value">{this.state.downloadSpeed}</span></div>
                    <div className="item">{this.state.status == "downloading" ? <><span className="field">Remaining: </span><span className="value">{this.state.left}</span></> : ""}</div>
                </div> : <></> }
                <div className="basic">
                    <div className="item" style={{width: "33%"}}><span className="field">Size: </span><span className="value">{this.state.size}</span></div>
                    <div className="item" style={{width: "40%"}}><span className="field">Downloaded: </span><span className="value">{this.state.downloaded}</span></div>
                    <div className="item" ><span className="field">Peers: </span><span className="value">{this.state.peers}</span></div>
                </div>
                <div className="basic">
                    <div className="item" style={{width: "100%"}}><span className="field">Path: </span><span className="value">{this.short(this.state.path)}</span><button onClick={this.copy(this.state.path)}>Copy</button></div>
                </div>
                <div className="basic">
                    <div className="item" style={{width: "100%"}}><span className="field">Bag ID: </span><span className="value">{this.props.torrent.toUpperCase()}</span><button onClick={this.copy(this.props.torrent)}>Copy</button></div>
                </div>
                <div className="basic">
                    <div className="item" style={{width: "66%"}}><span className="field">Added at: </span><span className="value">{this.state.added}</span></div>
                </div>
                <div className="basic">
                    <div className="item"><span className="field">Status: </span><span className="value">{textState(this.state.status, Number(this.state.peers))}</span></div>
                </div>
                {this.state.description != "" ?
                    <div className="basic">
                        <div className="item" style={{width: "100%"}}><span className="field">Description: </span><span className="value">{this.state.description}</span></div>
                    </div> : ""}
            </div>
        </div>
    }
}