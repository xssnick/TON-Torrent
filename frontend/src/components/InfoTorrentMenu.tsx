import React, {Component} from 'react';
import {EventsOff, EventsOn} from "../../wailsjs/runtime";
import {GetInfo} from "../../wailsjs/go/main/App";
import {textState} from "./Table";
import {Simulate} from "react-dom/test-utils";
import click = Simulate.click;


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
        return val; //.length > 45 ? val.slice(0, 45)+ "..." : val
    }

    copy(text: string) {
        return (clicked: any) => {
            navigator.clipboard.writeText(text).then();
            clicked.target.classList.add('clicked');
            setTimeout(() => {
                clicked.target.classList.remove('clicked');
            },500);
        }
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
            <div className="info-data-block">
                <div className="basic">
                    <div className="item" style={{width: "20%"}}><span className="field">Progress</span></div>
                    <div className="item" style={{flexGrow: "1"}}>
                        <span className="value" style={{width: "43px"}}>{this.state.progress}%</span>
                        <div className="info-progress-bar">
                            <div className="filled" style={{width: this.state.progress+"%"}}/></div>
                        </div>
                </div>
                { (this.state.status == "downloading" || this.state.status == "seeding") ? <div className="basic">
                    <div className="item" style={{width: "20%"}}><span className="field">Upload Speed</span></div>
                    <div className="item" style={{width: "20%"}}><span className="value">{this.state.uploadSpeed}</span></div>
                    {this.state.status == "downloading" ? <>
                        <div className="item" style={{width: "20%"}}><span className="field">Download Speed</span></div>
                        <div className="item" style={{width: "15%"}}><span className="value">{this.state.downloadSpeed}</span></div>
                        <div className="item" style={{width: "12%"}}><span className="field">Remaining</span></div>
                        <div className="item" style={{flexGrow: "1", justifyContent: "flex-end"}}><span className="value">{this.state.left}</span></div>
                    </> : ""}
                </div> : <></> }
                <div className="basic">
                    <div className="item" style={{width: "20%"}}><span className="field">Downloaded</span></div>
                    <div className="item" style={{width: "20%"}}><span className="value">{this.state.downloaded}</span></div>
                    <div className="item" style={{width: "20%"}}><span className="field">Peers</span></div>
                    <div className="item" style={{width: "15%"}}><span className="value">{this.state.peers}</span></div>
                    <div className="item" style={{width: "12%"}}><span className="field">Size</span></div>
                    <div className="item" style={{flexGrow: "1", justifyContent: "flex-end"}}><span className="value">{this.state.size}</span></div>
                </div>
                <div className="basic">
                    <div className="item" style={{width: "20%"}}><span className="field">Path</span></div>
                    <div className="item" style={{flexGrow: "1", maxWidth: "75%"}}><span className="value" style={{maxWidth:"90%"}}>{this.short(this.state.path)}</span><button onClick={this.copy(this.state.path)}/></div>
                </div>
                <div className="basic">
                    <div className="item" style={{width: "20%"}}><span className="field">Bag ID</span></div>
                    <div className="item" style={{flexGrow: "1", maxWidth: "75%"}}><span className="value" style={{maxWidth:"90%"}}>{this.short(this.props.torrent.toUpperCase())}</span><button onClick={this.copy(this.props.torrent)}/></div>
                </div>
                <div className="basic">
                    <div className="item" style={{width: "20%"}}><span className="field">Added at</span></div>
                    <div className="item" style={{flexGrow: "1"}}><span className="value">{this.state.added}</span></div>
                </div>
                <div className="basic">
                    <div className="item" style={{width: "20%"}}><span className="field">Status</span></div>
                    <div className="item" style={{flexGrow: "1"}}><span className="value">{textState(this.state.status, Number(this.state.peers))}</span></div>
                </div>
                {this.state.description != "" ?
                    <div className="basic">
                        <div className="item" style={{width: "20%"}}><span className="field">Name</span></div>
                        <div className="item" style={{flexGrow: "1", maxWidth: "80%"}}><span className="value">{this.short(this.state.description)}</span></div>
                    </div> : ""}
            </div>
        </div>
    }
}