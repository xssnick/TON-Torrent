import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {
    GetConfig,
    GetSpeedLimit,
    OpenDir,
    SaveConfig,
    SetSpeedLimit,
    OpenTunnelConfig,
    ReinitApp
} from "../../wailsjs/go/main/App";
import {BrowserOpenURL} from "../../wailsjs/runtime";

interface State {
    downloads: string
    tunnelConfig: string
    addr: string
    addrValid: boolean
    uploadSpeed: string
    downloadSpeed: string

    seedFiles: boolean

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
            seedFiles: false,
            tunnelConfig: "",
        };
    }

    componentDidMount() {
        GetConfig().then((cfg:any)=>{
            this.setState((current)=>({...current,
                downloads: cfg.DownloadsPath,
                addr: cfg.ListenAddr,
                useTonutils: !cfg.UseDaemon,
                daemonDB: cfg.DaemonDBPath,
                daemonMasterAddr: cfg.DaemonControlAddr,
                seedFiles: cfg.SeedMode,
                tunnelConfig: cfg.TunnelConfig.NodesPoolConfigPath,
            }))
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

        SaveConfig(this.state.downloads, this.state.seedFiles, this.state.addr, this.state.tunnelConfig).then(()=>{
            SetSpeedLimit(d, u).then();
        });
        this.props.onExit();
    }

    render() {
        return baseModal(this.props.onExit, (
            <>
                <div style={{width: "287px"}} className="add-torrent-block">
                    <span className="title">Settings</span>
                    <span className="field-name">Downloads directory</span>
                    <div className="create-input">
                        <span>{
                            this.state.downloads.length > 25 ? "..." + this.state.downloads.slice(this.state.downloads.length - 25, this.state.downloads.length) : this.state.downloads
                        }</span>
                        <button onClick={() => {
                            OpenDir().then((p: string) => {
                                if (p.length > 0) {
                                    this.setState((current) => ({...current, downloads: p}))
                                }
                            })
                        }}>Select
                        </button>
                    </div>
                    <span style={{ marginTop: "7px" }} className="field-name">Tunnel config</span>
                    <div className="create-input">
                        <span>{this.state.tunnelConfig == "" ? "Not selected" : (this.state.tunnelConfig.length > 25 ? "..." + this.state.tunnelConfig.slice(this.state.tunnelConfig.length - 25, this.state.tunnelConfig.length) : this.state.tunnelConfig)}</span>
                        <button onClick={() => {
                            OpenTunnelConfig().then((p: any) => {
                                let path = "";
                                if (p) {
                                    path = p.Path;
                                }
                                this.setState((current) => ({...current, tunnelConfig: path}));
                            })
                        }}>Select
                        </button>
                    </div>
                    <div className="set-speed">
                        <div className="info">
                            <span className="field-name">Max upload KB/s</span>
                            <input type="text" pattern="[0-9]*" placeholder="No limit" value={this.state.uploadSpeed}
                                   onChange={(e) => {
                                       if (e.target.validity.valid)
                                           this.setState((current) => ({...current, uploadSpeed: e.target.value}))
                                   }}/>
                        </div>
                        <div className="info">
                            <span className="field-name">Max download KB/s</span>
                            <input type="text" pattern="[0-9]*" placeholder="No limit" value={this.state.downloadSpeed}
                                   onChange={(e) => {
                                       if (e.target.validity.valid)
                                           this.setState((current) => ({...current, downloadSpeed: e.target.value}))
                                   }}/>
                        </div>
                    </div>
                    <div className="set-speed">
                        <label className="checkbox-file daemon">Static seed mode
                            <input type="checkbox" className="file-to-download" checked={this.state.seedFiles}
                                   onChange={(e) => {
                                       this.setState((current) => ({...current, seedFiles: !this.state.seedFiles}))
                                   }}/>
                            <span className="checkmark"></span>
                        </label>
                    </div>
                    <div className="set-speed"
                         style={{display: this.state.seedFiles ? 'block' : 'none'}}>
                        <span className="field-name">External ip and port to listen on</span>
                        <input type="text"
                               pattern="((\b25[0-5]|\b2[0-4][0-9]|\b[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})(?::((6553[0-5])|(655[0-2][0-9])|(65[0-4][0-9]{2})|(6[0-4][0-9]{3})|([1-5][0-9]{4})|([0-5]{0,5})|([0-9]{1,4})))\b"
                               value={this.state.addr} onChange={(e) => {
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
                    {this.state.err ? <span className="error">{this.state.err}</span> : ""}
                </div>
                <div className="modal-control">
                    <button className="second-button" onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="main-button" disabled={
                        (!this.state.addrValid  && this.state.seedFiles)
                        || (this.state.addr.startsWith(":") && this.state.seedFiles)
                    } onClick={() => {
                        this.next()
                    }}>
                        Save
                    </button>
                </div>
                <div className="modal-version">
                    <span className="version">Version 1.7.1</span>
                    <span className="check" onClick={() => {
                        BrowserOpenURL("https://github.com/xssnick/TON-Torrent/releases")
                    }}>Check updates</span>
                </div>
            </>
        ));
    }
}