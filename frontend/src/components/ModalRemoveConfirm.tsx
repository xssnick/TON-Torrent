import React, {Component} from 'react';
import {baseModal} from "./Modal";
import {GetInfo, RemoveTorrent} from "../../wailsjs/go/main/App";
import {Refresh} from "./Table";

interface State {
    names: string[]
}

interface RemoveConfirmModalProps {
    hashes: string[]
    onExit: () => void;
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
        return baseModal(this.props.onExit, (
            <>
                <div className="add-torrent-block">
                    <span className="title big">Are you sure you want to delete torrents below?</span>
                    {this.state.names.map((name)=>{
                        return <span className="title name">"{name}"</span>
                    })}
                </div>
                <div className="modal-control" style={{width:"400px"}}>
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
        ));
    }
}