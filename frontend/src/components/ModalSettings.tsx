import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {GetConfig, GetSpeedLimit, OpenDir, SaveConfig, SetSpeedLimit, OpenTunnelConfig} from "../../wailsjs/go/main/App";
import {BrowserOpenURL} from "../../wailsjs/runtime";

interface State {
    downloads: string
    tunnelConfig: string
    addr: string
    addrValid: boolean
    addrDaemonValid: boolean
    uploadSpeed: string
    downloadSpeed: string

    useTonutils: boolean
    seedFiles: boolean
    daemonMasterAddr: string
    daemonDB: string

    err?: string
}

interface SettingsModalProps {
    onExit: () => void;
    onTunnelConfig: (maxNoPayments: number, max: number) => void;
}

export class SettingsModal extends Component<SettingsModalProps, State> {
    constructor(props: any) {
        super(props);

        this.state = {
            downloads: "",
            addr: "",
            addrValid: true,
            addrDaemonValid: true,
            uploadSpeed: "",
            downloadSpeed: "",
            useTonutils: true,
            seedFiles: false,
            daemonDB: "",
            daemonMasterAddr: "",
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

        SaveConfig(this.state.downloads, this.state.useTonutils, this.state.seedFiles, this.state.addr, this.state.daemonMasterAddr, this.state.daemonDB, this.state.tunnelConfig).then(()=>{
            SetSpeedLimit(d, u).then()
        });
        this.props.onExit()
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

                                if (path != "") {
                                    this.props.onTunnelConfig(p.MaxFree, p.Max);
                                }
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
                        <label className="checkbox-file daemon">Use tonutils-storage implementation
                            <input type="checkbox" className="file-to-download" checked={this.state.useTonutils}
                                   onChange={(e) => {
                                       this.setState((current) => ({...current, useTonutils: !this.state.useTonutils}))
                                   }}/>
                            <span className="checkmark"></span>
                        </label>
                    </div>
                    <div className="set-speed" style={{display: this.state.useTonutils ? 'block' : 'none'}}>
                        <label className="checkbox-file daemon">Static seed mode
                            <input type="checkbox" className="file-to-download" checked={this.state.seedFiles}
                                   onChange={(e) => {
                                       this.setState((current) => ({...current, seedFiles: !this.state.seedFiles}))
                                   }}/>
                            <span className="checkmark"></span>
                        </label>
                    </div>
                    <div className="set-speed"
                         style={{display: (this.state.useTonutils && this.state.seedFiles) ? 'block' : 'none'}}>
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
                    <div className="daemon-config" style={{display: this.state.useTonutils ? 'none' : 'block'}}>
                        <span className="field-name">Daemon control address</span>
                        <input type="text"
                               pattern="((\b25[0-5]|\b2[0-4][0-9]|\b[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3})(?::((6553[0-5])|(655[0-2][0-9])|(65[0-4][0-9]{2})|(6[0-4][0-9]{3})|([1-5][0-9]{4})|([0-5]{0,5})|([0-9]{1,4})))\b"
                               value={this.state.daemonMasterAddr} onChange={(e) => {
                            let valid = e.target.validity.valid
                            if (!valid) {
                                if (!e.target.classList.contains("invalid"))
                                    e.target.classList.add("invalid")
                            } else {
                                e.target.classList.remove("invalid")
                            }
                            this.setState((current) => ({
                                ...current,
                                daemonMasterAddr: e.target.value,
                                addrDaemonValid: valid
                            }))
                        }}/>
                        <span className="field-name">Daemon DB path</span>
                        <div className="create-input">
                                <span>{
                                    this.state.daemonDB.length > 25 ? "..." + this.state.daemonDB.slice(this.state.daemonDB.length - 25, this.state.daemonDB.length) : this.state.daemonDB
                                }</span>
                            <button onClick={() => {
                                OpenDir().then((p: string) => {
                                    if (p.length > 0) {
                                        this.setState((current) => ({...current, daemonDB: p}))
                                    }
                                })
                            }}>Select
                            </button>
                        </div>
                    </div>
                    {this.state.err ? <span className="error">{this.state.err}</span> : ""}
                </div>
                <div className="modal-control">
                    <button className="second-button" onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="main-button" disabled={
                        (!this.state.addrValid && this.state.useTonutils && this.state.seedFiles)
                        || (this.state.addr.startsWith(":") && this.state.useTonutils && this.state.seedFiles)
                        || (!this.state.addrDaemonValid && !this.state.useTonutils)
                        || (this.state.daemonDB.length == 0 && !this.state.useTonutils)
                    } onClick={() => {
                        this.next()
                    }}>
                        Save
                    </button>
                </div>
                <div className="modal-version">
                    <span className="version">Version 1.6.0</span>
                    <span className="check" onClick={() => {
                        BrowserOpenURL("https://github.com/xssnick/TON-Torrent/releases")
                    }}>Check updates</span>
                </div>
            </>
        ));
    }
}