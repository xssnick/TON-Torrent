import React, {Component} from 'react';
import {Modal} from "./Modal";
import {GetInfo, RemoveTorrent} from "../../wailsjs/go/main/App";
import {Refresh} from "./Table";
import FileLight from "../../public/light/file-popup.svg";
import FileDark from "../../public/dark/file-popup.svg";

interface State {
    names: string[]
}

interface RemoveConfirmModalProps {
    hashes: string[]
    onExit: () => void
    isDark: boolean
}

export class RemoveConfirmModal extends Component<RemoveConfirmModalProps, State> {
    constructor(props: any) {
        super(props);

        this.state = {
            names: [],
        };
    }

    async componentDidMount() {
        let names: string[] = [];
        for (const hash of this.props.hashes) {
            let info = await GetInfo(hash);
            names.push(info.Description);
        }
        this.setState((current)=>({...current, names}))
    }

    next = async (removeFiles: boolean) => {
        for (const hash of this.props.hashes) {
            await RemoveTorrent(hash, removeFiles, false)
        }
        this.props.onExit()
    }

    render() {
        return <Modal onHide={this.props.onExit} content={(
            <>
                <div className="add-torrent-block">
                    <span className="title">Are you sure you want to delete torrents below?</span>
                    <div className="files-block">{this.state.names.map((name)=>{
                        return <div className="file-item">
                            <img src={this.props.isDark ? FileDark : FileLight} alt={name}/>
                            <span className="name">{name}</span>
                        </div>
                    })}
                    </div>
                </div>
                <div className="modal-control">
                    <button className="second-button" style={{width:"115px"}} onClick={this.props.onExit}>
                        Cancel
                    </button>
                    <button className="second-button danger" style={{width:"115px"}}  onClick={()=>{this.next(true)}}>
                        Yes, delete all
                    </button>
                    <button className="main-button" style={{width:"115px"}} onClick={()=>{this.next(false).then(Refresh)}}>
                        Yes, keep files
                    </button>
                </div>
            </>
        )}/>;
    }
}