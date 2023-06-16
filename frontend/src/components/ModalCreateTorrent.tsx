import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {CreateTorrent, ExportMeta, OpenDir, OpenFolderSelectFile} from "../../wailsjs/go/main/App";

interface State {
    createdStage: boolean
    canContinue: boolean
    path: string
    name: string

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
            path: "",
            name: "",
        }
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
        return baseModal(this.props.onExit, (
            <>
                <div style={this.state.createdStage ? {width: "455px"} : {display: "none"}} className="add-torrent-block">
                    {this.state.hash ? <div className="torrent-created">
                        <span className="header">Torrent successfully created!</span>
                        <input readOnly={true} onClick={(e)=>{ e.currentTarget.select()}} value={this.state.hash}/>
                        <button className="second-button black" style={{width:"170px"}} onClick={() => {
                            ExportMeta(this.state.hash!).then((file) => {OpenFolderSelectFile(file).then()})}
                        }>Export torrent file</button>
                    </div> : <div className="files-selector">
                        <div className="loader-block"><span className="loader"/><span className="loader-text">Creating torrent..</span></div>
                    </div>}
                </div>
                <div style={this.state.createdStage ? {display: "none"} : {width: "330px"}} className="add-torrent-block">
                    <span className="title">Torrent name</span>
                    <input className="torrent-name-input" maxLength={100} autoFocus={true} onInput={(e) => {
                        let val = e.currentTarget.value;
                        let can = val.length > 0 && this.state.path.length > 0;
                        this.setState((current) => ({...current, name: val, canContinue: can}));
                    }}/>
                    <span className="title">From directory</span>
                    <div className="create-input">
                        <button onClick={() => {
                            OpenDir().then((p: string) => {
                                if (p.length > 0) {
                                    let can = p.length > 0 && this.state.name.length > 0;
                                    this.setState((current) => ({...current, path: p, canContinue: can}))
                                }
                            })
                        }}>Select</button>
                        <span>{
                            this.state.path.length > 30 ? "..."+this.state.path.slice(this.state.path.length-30,this.state.path.length) : this.state.path
                        }</span>
                    </div>
                    {this.state.err ? <span className="error">{this.state.err}</span> : ""}
                </div>
                <div className="modal-control">
                    <button className="second-button" onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="main-button" disabled={!this.state.canContinue} onClick={()=>{this.next()}}>
                        Continue
                    </button>
                </div>
            </>
        ));
    }
}