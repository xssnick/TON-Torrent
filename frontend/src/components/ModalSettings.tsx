import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {GetConfig, OpenDir, SaveConfig} from "../../wailsjs/go/main/App";

interface State {
    downloads: string

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
        };
    }

    componentDidMount() {
        GetConfig().then((cfg:any)=>{
            this.setState({
                downloads: cfg.DownloadsPath,
            })
        })
    }

    next = () => {
        SaveConfig(this.state.downloads).then();
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
                            <span className="title">Max upload speed</span>
                            <input type="text" placeholder="∞"/>
                        </div>
                        <div className="info">
                            <span className="title">Max download speed</span>
                            <input type="text" placeholder="∞"/>
                        </div>
                    </div>
                    <div className="set-speed">
                        <div className="info">
                            <span className="title">Storage listen address</span>
                            <input type="text"/>
                        </div>
                    </div>
                    {this.state.err ? <span className="error">{this.state.err}</span> : ""}
                </div>
                <div className="modal-control">
                    <button className="second-button" onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="main-button" onClick={()=>{this.next()}}>
                        Save
                    </button>
                </div>
            </>
        ));
    }
}