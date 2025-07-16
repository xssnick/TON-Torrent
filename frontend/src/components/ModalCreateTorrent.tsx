import React, {Component} from 'react';
import {Modal} from "./Modal";
import {
    CancelCreateTorrent,
    CreateTorrent,
    OpenDir,
    OpenFile,
} from "../../wailsjs/go/main/App";
import {EventsOff, EventsOn} from "../../wailsjs/runtime";

interface State {
    createdStage: boolean
    canContinue: boolean
    path: string
    name: string
    singleFile: boolean

    creationProgress: string

    hash?: string
    err?: string
}

interface CreateTorrentModalProps {
    onExit: () => void;
}

export class CreateTorrentModal extends Component<CreateTorrentModalProps, State> {
    constructor(props: CreateTorrentModalProps) {
        super(props);
        this.state = {
            err: "",
            canContinue: false,
            createdStage: false,
            singleFile: false,
            path: "",
            name: "",
            creationProgress: "0"
        }
    }

    componentDidMount() {
        EventsOn("update-create-progress", (progress: string) => {
            this.setState((current) => ({ ...current, creationProgress: progress }))
        })
    }
    componentWillUnmount() {
        CancelCreateTorrent().then()
        EventsOff("update-create-progress")
    }

    next = () => {
        if (!this.state.createdStage) {
            this.setState((current) => ({ ...current, canContinue: false, createdStage: true }))

            CreateTorrent(this.state.path, this.state.name).then((res: any) => {
                if (res.Hash) {
                    this.setState((current) => ({...current, canContinue: true, hash: res.Hash}))
                } else {
                    this.setState((current) => ({...current, canContinue: true, createdStage: false, err: res.Err}))
                }
            })
        } else {
            this.props.onExit()
        }
    }

    render() {
        return <Modal allowClose={!this.state.createdStage} onHide={this.props.onExit} content={(
            <>
                <div style={this.state.createdStage ? {width: "287px"} : {display: "none"}} className="add-torrent-block">
                    {this.state.hash ? <div className="torrent-created">
                        <div className="success"/>
                        <span className="title" style={{width: "70%"}}>Torrent successfully created!</span>
                    </div> : <><span className="title">Creating torrent...</span>
                        <div className="files-selector">
                            <div className="create-torrent-loader-block ">
                                <div className="create-progress-block">
                                    <span style={{width: "20%", textAlign: "center"}}>{this.state.creationProgress}%</span>
                                    <div className="create-progress-bar-form">
                                        <div className="create-progress-bar-small" style={{width: this.state.creationProgress+"%"}}></div>
                                    </div></div>
                            </div>
                        </div></>}
                </div>
                <div style={this.state.createdStage ? {display: "none"} : {width: "287px"}} className="add-torrent-block">
                    <span className="title">Create Torrent</span>
                    <input className="torrent-name-input" placeholder={"Name"} maxLength={100} autoFocus={true} onInput={(e) => {
                        let val = e.currentTarget.value;
                        let can = val.length > 0 && this.state.path.length > 0;
                        this.setState((current) => ({...current, name: val, canContinue: can}));
                    }}/>
                    <div className="create-input">
                        <span>{
                            this.state.path.length > 23 ? "..."+this.state.path.slice(this.state.path.length-23,this.state.path.length) : this.state.path
                        }</span>
                        <button onClick={() => {
                            let e = (p: string) => {
                                if (p.length > 0) {
                                    let can = p.length > 0 && this.state.name.length > 0;
                                    this.setState((current) => ({...current, path: p, canContinue: can}))
                                }
                            }

                            if (this.state.singleFile) {
                                OpenFile().then(e)
                            } else {
                                OpenDir().then(e)
                            }
                        }}>{this.state.singleFile ? "Select file" : "Select folder"}</button>
                    </div>
                    <div className="type-input">
                        <label className="checkbox-file single-file">Single File
                            <input type="checkbox" checked={this.state.singleFile}
                                   onChange={(e) => {
                                       this.setState((current) => ({...current, path: "", canContinue: false, singleFile: !this.state.singleFile}))
                                   }}/>
                            <span className="checkmark"></span>
                        </label>
                    </div>
                    {this.state.err ? <span className="error">{this.state.err}</span> : ""}
                </div>
                {(this.state.createdStage && this.state.hash) ? <div className="modal-control">
                        <button className="main-button" style={{width: "100%"}} onClick={()=>{this.next()}}>
                            Done
                        </button>
                    </div>:
                    <div className="modal-control">
                    <button className="second-button" onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="main-button" disabled={!this.state.canContinue} onClick={()=>{this.next()}}>
                        Continue
                    </button>
                </div>}
            </>
        )}/>;
    }
}