import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {GetConfig, GetSpeedLimit, OpenDir, SaveConfig, SetSpeedLimit} from "../../wailsjs/go/main/App";

interface State {
    downloads: string
    addr: string
    addrValid: boolean
    uploadSpeed: string
    downloadSpeed: string

    err?: string
}

interface SettingsModalProps {
    onExit: () => void;
}

export class SettingsModal extends Component<SettingsModalProps, State> {
    constructor(props: any) {
        super(props);

        this.state = {
            downloads: "",
            addr: "",
            addrValid: true,
            uploadSpeed: "",
            downloadSpeed: "",
        };
    }

    componentDidMount() {
        GetConfig().then((cfg:any)=>{
            this.setState((current)=>({...current, downloads: cfg.DownloadsPath, addr: cfg.ListenAddr}))
        })
        GetSpeedLimit().then((lim: any) => {
            let d = lim.Download > 0 ? lim.Download.toString() : "";
            let u = lim.Upload > 0 ? lim.Upload.toString() : "";

            this.setState((current)=>({...current, uploadSpeed: u, downloadSpeed: d}))
        })
    }

    next = () => {
        let d = -1;
        let u = -1;
        if (this.state.downloadSpeed != "") {
            d = Number(this.state.downloadSpeed);
        }
        if (this.state.uploadSpeed != "") {
            u = Number(this.state.uploadSpeed);
        }

        SaveConfig(this.state.downloads, this.state.addr).then(()=>{
            SetSpeedLimit(d, u).then()
        });
        this.props.onExit()
    }

    render() {
        return baseModal(this.props.onExit, (
            <>
                <div className="add-torrent-block">
                    <span className="title">Downloads directory:</span>
                    <div className="create-input">
                        <button onClick={() => {
                            OpenDir().then((p: string) => {
                                if (p.length > 0) {
                                    this.setState((current) => ({...current, downloads: p}))
                                }
                            })
                        }}>Select</button>
                        <span>{
                            this.state.downloads.length > 30 ? "..."+this.state.downloads.slice(this.state.downloads.length-30,this.state.downloads.length) : this.state.downloads
                        }</span>
                    </div>
                    <div className="set-speed">
                        <div className="info">
                            <span className="title">Max upload KB/s</span>
                            <input type="text" pattern="[0-9]*" placeholder="∞" value={this.state.uploadSpeed} onChange={(e) =>{
                                if (e.target.validity.valid)
                                    this.setState((current) => ({...current, uploadSpeed: e.target.value}))
                            }}/>
                        </div>
                        <div className="info">
                            <span className="title">Max download KB/s</span>
                            <input type="text" pattern="[0-9]*" placeholder="∞" value={this.state.downloadSpeed} onChange={(e) =>{
                                if (e.target.validity.valid)
                                    this.setState((current) => ({...current, downloadSpeed: e.target.value}))
                            }}/>
                        </div>
                    </div>
                    <div className="set-speed">
                        <div className="info listen">
                            <span className="title">Storage listen address</span>
                            <input type="text" pattern="((\b25[0-5]|\b2[0-4][0-9]|\b[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})?(?::((6553[0-5])|(655[0-2][0-9])|(65[0-4][0-9]{2})|(6[0-4][0-9]{3})|([1-5][0-9]{4})|([0-5]{0,5})|([0-9]{1,4})))\b" style={{width: "200px"}} value={this.state.addr} onChange={(e) =>{
                                let valid = e.target.validity.valid
                                if (!valid) {
                                    if (!e.target.classList.contains("invalid"))
                                        e.target.classList.add("invalid")
                                } else {
                                    e.target.classList.remove("invalid")
                                }
                                this.setState((current) => ({...current, addr: e.target.value, addrValid: valid}))
                            }}/>
                        </div>
                    </div>
                    {this.state.err ? <span className="error">{this.state.err}</span> : ""}
                </div>
                <div className="modal-control">
                    <button className="second-button" onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="main-button" disabled={!this.state.addrValid} onClick={()=>{this.next()}}>
                        Save
                    </button>
                </div>
            </>
        ));
    }
}